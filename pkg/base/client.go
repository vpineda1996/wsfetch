package base

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	ghttputil "net/http/httputil"
	"net/url"

	"github.com/samber/lo"
	"github.com/vpnda/wsfetch/internal/httputil"
	"github.com/vpnda/wsfetch/pkg/auth/creds"
	"github.com/vpnda/wsfetch/pkg/auth/types"
	"github.com/vpnda/wsfetch/pkg/endpoints"
	"go.uber.org/zap"
)

func DefaultAuthClient(pc types.PasswordCredentials) *Wealthsimple {
	return AuthClientFromFetcher(creds.NewDefaultFetcher(pc))
}

func AuthClientFromSession(session *types.Session) *Wealthsimple {
	return AuthClientFromFetcher(creds.NewFetcherFromExistingSession(session))
}

func StatictAuthClient(session types.Session) *Wealthsimple {
	return AuthClientFromFetcher(creds.StaticTokenFetcher(session))
}

type WsProfile string

const (
	Invest WsProfile = "invest"
	Trade            = "trade"
	Tax              = "tax"
)

type Wealthsimple struct {
	// cred fetcher
	Fetcher creds.SessionFetcher

	// internal http client
	delegate *http.Client

	// cooke jar
	Jar http.CookieJar

	// Profile to authenticate to Wealthsimple, defaults
	// to invest
	Profile WsProfile
}

var (
	log               = lo.Must(zap.NewProduction()).Sugar()
	myWealthsimpleUrl = lo.Must(url.Parse(endpoints.MyWeathSimple))
)

type TokenInformation struct {
	UserId            string
	IdentityId        string
	ProfileToClientId map[WsProfile]string
}

func (c *Wealthsimple) GetTokenInformation(ctx context.Context) (*TokenInformation, error) {
	sess, err := c.Fetcher.GetSession(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve credentials: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, endpoints.AuthTokenInfo.Method, endpoints.AuthTokenInfo.String(), nil)
	if err != nil {
		return nil, err
	}

	c.populateStandardAndAuthHeaders(req, sess)
	res, err := c.delegate.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cannot fetch client identifiers: %d", res.StatusCode)
	}

	bodyBits, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var jsonResponse struct {
		IdentityId string `json:"identity_canonical_id"`
		UserId     string `json:"user_canonical_id"`
		Profiles   map[WsProfile]struct {
			Default string `json:"default"`
		} `json:"profiles"`
	}
	err = json.Unmarshal(bodyBits, &jsonResponse)
	if err != nil {
		return nil, fmt.Errorf("invalid response body: %s, error: %w", string(bodyBits), err)
	}

	prfToClientId := map[WsProfile]string{}
	for k, v := range jsonResponse.Profiles {
		prfToClientId[k] = v.Default
	}

	return &TokenInformation{
		UserId:            jsonResponse.UserId,
		IdentityId:        jsonResponse.IdentityId,
		ProfileToClientId: prfToClientId,
	}, nil
}

func (c *Wealthsimple) Do(r *http.Request) (*http.Response, error) {
	sess, err := c.Fetcher.GetSession(r.Context())
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve credentials: %w", err)
	}

	err = c.hydrateSessionJar(r.Context(), r.URL, sess)
	if err != nil {
		return nil, fmt.Errorf("error hydrating session jar: %w", err)
	}

	c.populateStandardAndAuthHeaders(r, sess)
	return c.delegate.Do(r)
}

func (c *Wealthsimple) populateStandardAndAuthHeaders(req *http.Request, sess *types.Session) {
	httputil.ExtendHeaders(&req.Header, sess.ClientId)
	req.Header.Set("x-ws-profile", string(c.Profile))
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", sess.AccessToken))
}

// TODO refactor this into a wrapper of th do method
func (c *Wealthsimple) hydrateSessionJar(ctx context.Context, originalUrl *url.URL, sess *types.Session) error {
	if originalUrl.Host != endpoints.MyWeathSimple {
		log.Infow("Skipping hydration", "targetHost", originalUrl.Host)
		return nil
	}

	// If session ID is already in the cookie jar exit early
	if myWsCookes := c.Jar.Cookies(myWealthsimpleUrl); len(myWsCookes) != 0 {
		_, sessionIdFound := lo.Find(myWsCookes, func(ck *http.Cookie) bool {
			return ck.Name == "_session_id"
		})
		if sessionIdFound {
			return nil
		}
	}

	data := map[string]map[string]string{
		"session": {
			"access_token": sess.AccessToken,
		},
	}
	serializedData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		endpoints.MyWealthsimpleSession.String(), bytes.NewBuffer(serializedData))
	if err != nil {
		return err
	}

	c.populateStandardAndAuthHeaders(req, sess)
	res, err := c.delegate.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		d, _ := ghttputil.DumpResponse(res, true)
		return fmt.Errorf("response has invalid code: %s", string(d))
	}

	return nil
}

func AuthClientFromFetcher(fetcher creds.SessionFetcher) *Wealthsimple {
	jar := lo.Must(cookiejar.New(nil))
	return &Wealthsimple{
		Fetcher: fetcher,
		Jar:     jar,
		delegate: &http.Client{
			Jar: jar,
			// Transport: httputil.RoundTripperFunc(func(r *http.Request) (*http.Response, error) {
			// 	d, _ := ghttputil.DumpRequest(r, true)
			// 	fmt.Println(string(d))
			// 	res, err := http.DefaultTransport.RoundTrip(r)
			// 	if err == nil {
			// 		d, _ := ghttputil.DumpResponse(res, true)
			// 		fmt.Println(string(d))
			// 	}
			// 	return res, err
			// }),
		},
		Profile: Invest,
	}
}
