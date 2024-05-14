package authenticator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/samber/lo"
	"github.com/vpineda1996/wsfetch/pkg/auth/cfetch"
	"github.com/vpineda1996/wsfetch/pkg/auth/types"
	"github.com/vpineda1996/wsfetch/pkg/endpoints"
	"go.uber.org/zap"
)

var (
	log   = lo.Must(zap.NewProduction()).Sugar()
	wsUrl = lo.Must(url.Parse("https://wealthsimple.com"))
)

type Client interface {
	Authenticate(ctx context.Context, creds types.PasswordCredentials) (*types.Session, error)
}

type client struct {
	// The client id is provided by Wealthsimple
	// and its issued to all clients
	Id string `json:"id"`

	// ShouldRemember2FA configures the remeber me
	// token to be remebered by WS and possibly reused in
	// the future
	ShouldRemember2FA bool `json:"shouldRemember2FA"`

	// OTP remember me token is used to re-authenticate
	// with Wealthsimple, skipping 2FA
	RemeberMeToken string `json:"rememberMeToken"`

	// 2FA fetcher
	codeFetcher types.TwoFactorCodeFetcher

	// client with authenticated information
	client *http.Client

	// Transport used for requests
	transport http.RoundTripper
}

func NewClient() Client {
	c := &client{
		ShouldRemember2FA: true,
	}
	newFromExisting(c)
	return c
}

func newFromExisting(c *client) {
	c.client = &http.Client{
		Jar:       lo.Must(cookiejar.New(&cookiejar.Options{})),
		Transport: c,
	}
	c.codeFetcher = cfetch.NewCli()
	c.transport = http.DefaultTransport
}

var (
	_ http.RoundTripper = &client{}
	_ Client            = &client{}
)

// fetchClientIdIfNotSet does a simple call to WS to just get
// the client ID and stores it in cookies
func (c *client) fetchClientIdIfNotSet(ctx context.Context) error {
	if c.Id != "" {
		return nil
	}

	// call "my.wealthsimple.com/app/login" to get cookie
	req, err := http.NewRequestWithContext(ctx, endpoints.MyWealthsimpleLoginSplash.Method, endpoints.MyWealthsimpleLoginSplash.String(), nil)
	if err != nil {
		return err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	log.Infow("Queried for wssdi", "respStatus", resp.Status)

	// find wssdi cookie
	wssCookie, wssFound := lo.Find(c.client.Jar.Cookies(wsUrl), func(cookie *http.Cookie) bool {
		return cookie.Name == "wssdi"
	})

	if !wssFound {
		return fmt.Errorf("wssdi not found in first request: %v", c.client.Jar.Cookies(wsUrl))
	}

	c.Id = wssCookie.Value
	log.Infow("Id for client is now", "id", c.Id)
	return nil
}

// Authenticate is the main entry point where we call the auth API
// to fetch credentials
func (c *client) Authenticate(ctx context.Context, creds types.PasswordCredentials) (*types.Session, error) {
	log.Infow("Starting authentication", "creds", creds)

	err := c.fetchClientIdIfNotSet(ctx)
	if err != nil {
		return nil, err
	}

	tokenReqBody, err := creds.AuthPayload()
	if err != nil {
		return nil, err
	}

	log.Infow("Sending request credential payload", "payload", string(tokenReqBody))

	// Try to get the bearer token by providing the credentials
	req, err := http.NewRequestWithContext(ctx, endpoints.AuthToken.Method, endpoints.AuthToken.String(), bytes.NewBuffer(tokenReqBody))
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	log.Infow("First attempt at getting token", "respStatus", resp.Status)
	if resp.StatusCode == http.StatusUnauthorized {
		// parse header to see if its 2FA the issue
		twoFaHeader, err := Parse2FAHeaders(resp.Header)
		if err != nil {
			return nil, fmt.Errorf("error parsing 2FA: %w", err)
		}

		// fail the problem is not 2FA
		if !twoFaHeader.Required {
			dr, _ := httputil.DumpResponse(resp, true)
			return nil, fmt.Errorf("unable to authenticate, dump: %s", string(dr))
		}

		resp, err = c.resolve2FA(ctx, twoFaHeader, tokenReqBody)
		if err != nil {
			return nil, err
		}
	}

	if resp.StatusCode != http.StatusOK {
		dr, _ := httputil.DumpResponse(resp, true)
		return nil, fmt.Errorf("unable to authorize request: %s", dr)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	s, err := ParseSessionFromBody(c.Id, body)
	if err != nil {
		return nil, err
	}
	log.Infow("Resolved session", "sessionCreds", s)
	return s, nil

}

func (c *client) resolve2FA(ctx context.Context, twoFaHeader types.TwoFactorAuthRequest, authPayload []byte) (*http.Response, error) {
	log.Infow("Handling 2FA", "2FAHeader", twoFaHeader)

	// acutally authorize the token with this request
	req, err := http.NewRequestWithContext(ctx, endpoints.AuthToken.Method, endpoints.AuthToken.String(), bytes.NewBuffer(authPayload))
	if err != nil {
		return nil, err
	}

	// fetch the 2FA token from user, we don't care how,
	code, err := c.codeFetcher.Fetch(twoFaHeader)
	if err != nil {
		return nil, err
	}
	InjectAuthClaimWithCode(&req.Header, twoFaHeader, code, c.ShouldRemember2FA)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusOK && c.ShouldRemember2FA {
		c.RemeberMeToken = resp.Header.Get("x-wealthsimple-otp-claim")
	}

	return resp, nil
}

func (c *client) RoundTrip(req *http.Request) (*http.Response, error) {
	ExtendHeaders(&req.Header, c.Id, c.RemeberMeToken)
	return c.transport.RoundTrip(req)
}

func ParseSessionFromBody(deviceId string, body []byte) (*types.Session, error) {
	type serializedInput struct {
		AccessToken      string `json:"access_token"`
		ExpiresInSeconds int    `json:"expires_in"`
	}
	var st serializedInput
	err := json.Unmarshal(body, &st)
	if err != nil {
		return nil, err
	}

	if st.AccessToken == "" || st.ExpiresInSeconds == 0 {
		return nil, fmt.Errorf("unable to parse session (%+v), %s", st, string(body))
	}

	exp := time.Now().Add(time.Duration(st.ExpiresInSeconds) * time.Second)
	return &types.Session{
		DeviceId:    deviceId,
		BearerToken: st.AccessToken,
		Expiry:      &exp,
	}, nil

}
