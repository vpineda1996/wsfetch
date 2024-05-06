package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"

	"github.com/samber/lo"
	"github.com/vpineda1996/wsfetch/pkg/endpoints"
	"go.uber.org/zap"
)

type Session struct {
	// Example: e123edb34b8bb66baeefbeef07275cc5
	DeviceId string
	// TODO do we need session id?
	SessionId string
	// Bearer token
	BearerToken string
}

type PasswordCredentials struct {
	Username string
	Password string
}

func (pc PasswordCredentials) Payload() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"grant_type":     "password",
		"username":       pc.Username,
		"password":       pc.Password,
		"skip_provision": true,
		"otp_claim":      nil,
		// only grant read permissions for now
		"scope": "invest.read trade.read",
		// magic string from WS
		"client_id": "4da53ac2b03225bed1550eba8e4611e086c7b905a3855e6ed12ea08c246758fa",
	})
}

type Client struct {
	// client with authenticated information
	client *http.Client

	// client id should be the device ID on
	// sessions
	id string
}

var (
	log   = lo.Must(zap.NewProduction()).Sugar()
	wsUrl = lo.Must(url.Parse("https://wealthsimple.com"))
)

func NewClient() *Client {
	return &Client{
		id: "test",
		client: &http.Client{
			Jar: lo.Must(cookiejar.New(&cookiejar.Options{})),
		},
	}
}

func (c *Client) Authenticate(ctx context.Context, creds PasswordCredentials) error {
	log.Infow("Starting authentication", "creds", creds)

	// call "my.wealthsimple.com/app/login" to get cookie
	req, err := http.NewRequestWithContext(ctx, endpoints.LoginSplash.Method, endpoints.LoginSplash.String(), nil)
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

	c.id = wssCookie.Value
	log.Infow("Id for client is now", "id", c.id)

	tokenReqBody, err := creds.Payload()
	if err != nil {
		return err
	}

	// call "https://api.production.wealthsimple.com/v1/oauth/v2/token"
	req, err = http.NewRequestWithContext(ctx, endpoints.AuthToken.Method, endpoints.AuthToken.String(), bytes.NewBuffer(tokenReqBody))
	if err != nil {
		return err
	}

	resp, err = c.client.Do(req)
	if err != nil {
		return err
	}
	log.Infow("First try at token", "respStatus", resp.Status)
	log.Infow("Header for second call", "header", resp.Header)

	// parse header token
	twoFaHeader, err := Parse2FAHeaders(resp.Header)
	if err != nil {
		return err
	}
	log.Infow("2FA parsed header result", "2FA", twoFaHeader)

	// if 401 try otp
	//
	return nil

}
