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
	"regexp"
	"time"

	"github.com/google/uuid"
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

type AuthPayloadCreator interface {
	AuthPayload(clientId string) ([]byte, error)
	Profile() string
}

type Client interface {
	Authenticate(ctx context.Context, creds AuthPayloadCreator) (*types.Session, error)
}

type client struct {
	// The client id is provided by Wealthsimple
	// and its issued to all clients
	ClientId string

	// The WSSID is the id that seems to be used
	// to identify the current device
	// WSSID string
	WSSID string

	// ShouldRemember2FA configures the remeber me
	// token to be remebered by WS and possibly reused in
	// the future
	ShouldRemember2FA bool

	// OTP remember me token is used to re-authenticate
	// with Wealthsimple, skipping 2FA
	RefreshOtpToken string

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

func NewClientFromSession(s *types.Session) Client {
	c := &client{
		ShouldRemember2FA: true,
		WSSID:             s.WSSID,
		ClientId:          s.ClientId,
		RefreshOtpToken:   s.RefreshOtpToken,
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

// fetchWssidIfNotSet does a simple call to WS to just get
// the WSSID and stores it in cookies
func (c *client) fetchWssidIfNotSet(ctx context.Context) error {
	if c.WSSID != "" {
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
	wssidCookie, wssFound := lo.Find(c.client.Jar.Cookies(wsUrl), func(cookie *http.Cookie) bool {
		return cookie.Name == "wssdi"
	})

	if !wssFound {
		return fmt.Errorf("wssdi not found in first request: %v", c.client.Jar.Cookies(wsUrl))
	}

	c.WSSID = wssidCookie.Value
	log.Infow("Id for client is now", "id", c.WSSID)
	return nil
}

func (c *client) fetchClientIdIfNotSet(ctx context.Context) error {
	if c.ClientId != "" {
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
	defer resp.Body.Close()

	bits, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	responseStr := string(bits)

	re := regexp.MustCompile(`(?i)<script.*src="(.+/app-[a-f0-9]+\.js)`)
	matches := re.FindStringSubmatch(responseStr)
	if len(matches) <= 1 {
		return fmt.Errorf("couldn't find app JS URL in login page response body")
	}

	appJSURL := matches[1]

	// go to the app url to find the clientId
	req, err = http.NewRequestWithContext(ctx, http.MethodGet, appJSURL, nil)
	if err != nil {
		return err
	}
	resp, err = c.client.Do(req)
	if err != nil {
		return err
	}

	bits, err = io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	responseStr = string(bits)

	// Look for clientId in the app JS file
	re = regexp.MustCompile(`(?i)production:.*clientId:"([a-f0-9]+)"`)
	matches = re.FindStringSubmatch(responseStr)
	if len(matches) <= 1 {
		return fmt.Errorf("couldn't find clientId in app JS file response body")
	}
	c.ClientId = matches[1]
	log.Infow("Using clientId", "clientId", c.ClientId)
	return nil
}

// Authenticate is the main entry point where we call the auth API
// to fetch credentials
func (c *client) Authenticate(ctx context.Context, creds AuthPayloadCreator) (*types.Session, error) {
	log.Infow("Starting authentication", "creds", creds)

	err := c.fetchWssidIfNotSet(ctx)
	if err != nil {
		return nil, err
	}

	err = c.fetchClientIdIfNotSet(ctx)
	if err != nil {
		return nil, err
	}

	tokenReqBody, err := creds.AuthPayload(c.ClientId)
	if err != nil {
		return nil, err
	}

	log.Infow("Sending request credential payload", "payload", string(tokenReqBody))

	// Try to get the bearer token by providing the credentials
	req, err := http.NewRequestWithContext(ctx, endpoints.AuthToken.Method, endpoints.AuthToken.String(), bytes.NewBuffer(tokenReqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Add("x-ws-profile", creds.Profile())
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
			dreq, _ := httputil.DumpRequest(req, false)
			dr, _ := httputil.DumpResponse(resp, true)
			return nil, fmt.Errorf("unable to authenticate, req: %s, res: %s", string(dreq), string(dr))
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

	s, err := ParseSessionFromBody(body)
	if err != nil {
		return nil, err
	}
	s.ClientId = c.ClientId
	s.WSSID = c.WSSID
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
		c.RefreshOtpToken = resp.Header.Get("x-wealthsimple-otp-claim")
	}

	return resp, nil
}

func (c *client) RoundTrip(req *http.Request) (*http.Response, error) {
	ExtendHeaders(&req.Header, c.WSSID, c.RefreshOtpToken)
	return c.transport.RoundTrip(req)
}

func ParseSessionFromBody(body []byte) (*types.Session, error) {
	type serializedInput struct {
		AccessToken      string `json:"access_token"`
		RefreshToken     string `json:"refresh_token"`
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
		AccessToken:  st.AccessToken,
		RefreshToken: st.RefreshToken,
		Expiry:       &exp,
		SessionId:    uuid.Must(uuid.NewRandom()).String(),
	}, nil

}
