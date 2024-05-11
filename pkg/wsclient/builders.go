package wsclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	ghttputil "net/http/httputil"
	"net/url"

	"github.com/samber/lo"
	"github.com/vpineda1996/wsfetch/internal/httputil"
	"github.com/vpineda1996/wsfetch/pkg/auth/creds"
	"github.com/vpineda1996/wsfetch/pkg/auth/types"
	"github.com/vpineda1996/wsfetch/pkg/endpoints"
	"go.uber.org/zap"
)

func DefaultAuthClient(pc types.PasswordCredentials) *WsClient {
	return AuthClientFromFetcher(creds.NewDefaultFetcher(pc))
}

func StatictAuthClient(session types.Session) *WsClient {
	return AuthClientFromFetcher(creds.StaticTokenFetcher(session))
}

type WsClient struct {
	// cred fetcher
	Fetcher creds.SessionFetcher

	// internal http client
	delegate *http.Client

	// cooke jar
	Jar http.CookieJar
}

var (
	log = lo.Must(zap.NewProduction()).Sugar()
)

func (c *WsClient) Do(r *http.Request) (*http.Response, error) {
	auth, err := c.Fetcher.GetSession(r.Context())
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve credentials: %w", err)
	}

	err = c.hydrateSessionJar(r.Context(), r.URL, auth)
	if err != nil {
		return nil, fmt.Errorf("error hydrating session jar: %w", err)
	}

	httputil.ExtendHeaders(&r.Header, auth.DeviceId)
	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", auth.BearerToken))

	return c.delegate.Do(r)
}

// TODO refactor this into a wrapper of th do method
func (c *WsClient) hydrateSessionJar(ctx context.Context, originalUrl *url.URL, sess *types.Session) error {
	if originalUrl.Host != endpoints.MyWeathSimple {
		log.Infow("Skipping hydration", "targetHost", originalUrl.Host)
		return nil
	}

	data := map[string]map[string]string{
		"session": {
			"access_token": sess.BearerToken,
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

	httputil.ExtendHeaders(&req.Header, sess.DeviceId)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", sess.BearerToken))
	c.Jar.SetCookies(lo.Must(url.Parse("my.wealthsimple.com")), []*http.Cookie{
		{
			Name:  "wssdi",
			Value: sess.DeviceId,
		},
	})

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

func AuthClientFromFetcher(fetcher creds.SessionFetcher) *WsClient {
	jar := lo.Must(cookiejar.New(nil))
	return &WsClient{
		Fetcher: fetcher,
		Jar:     jar,
		delegate: &http.Client{
			Jar: jar,
			Transport: httputil.RoundTripperFunc(func(r *http.Request) (*http.Response, error) {
				d, _ := ghttputil.DumpRequest(r, false)
				fmt.Println(string(d))
				return http.DefaultTransport.RoundTrip(r)
			}),
		},
	}
}
