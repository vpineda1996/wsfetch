package authenticator

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strconv"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"github.com/vpnda/wsfetch/internal/httputil"
	"github.com/vpnda/wsfetch/pkg/auth/types"
	"github.com/vpnda/wsfetch/pkg/endpoints"
)

func Test_Client_2FA(t *testing.T) {
	var (
		testAuthClaim           = "someJwtToken"
		testSomeAuthPayload     = "someAuthPayload"
		testSomeOtpClaim        = "someOtpClaim"
		testSomeRememberMeToken = "someRememberedOtpClaim"
		ctx                     = context.Background()
	)

	testCases := []struct {
		name           string
		rememberMe     bool
		remeberMeToken string
		codeFetchErr   error
		expectedErr    error
	}{
		{
			name: "base case",
		},
		{
			name:       "remember me enabled",
			rememberMe: true,
		},
		{
			name:         "handle code fetch error",
			rememberMe:   true,
			codeFetchErr: errors.New("some error"),
			expectedErr:  errors.New("some error"),
		},
		{
			name:           "remember me enabled; have token",
			rememberMe:     true,
			remeberMeToken: testSomeRememberMeToken,
		},
		{
			name:           "remember me enabled; have token",
			rememberMe:     false,
			remeberMeToken: testSomeRememberMeToken,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)
			c := &client{
				ShouldRemember2FA: tc.rememberMe,
				RefreshOtpToken:   tc.remeberMeToken,
			}
			newFromExisting(c)
			callCount := 0
			c.codeFetcher = types.TwoFactorCodeFetcherFunc(func(twoFaHeader types.TwoFactorAuthRequest) (string, error) {
				g.Expect(twoFaHeader.AuthenticatedClaim).To(Equal(testAuthClaim))
				if tc.codeFetchErr != nil {
					return "", tc.codeFetchErr
				}
				return "123456", nil
			})

			// handle the first request call
			c.transport = httputil.RoundTripperFunc(func(r *http.Request) (*http.Response, error) {
				g.Expect(r.URL.Host).To(ContainSubstring(endpoints.RootHostname))
				g.Expect(r.URL.Path).To(Equal(endpoints.AuthToken.Path))

				buf, err := io.ReadAll(r.Body)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(string(buf)).To(Equal(testSomeAuthPayload))

				g.Expect(r.Header.Get("x-wealthsimple-otp-authenticated-claim")).To(Equal(testAuthClaim))
				g.Expect(r.Header.Get("x-wealthsimple-otp")).To(ContainSubstring("123456"))
				g.Expect(r.Header.Get("x-wealthsimple-otp")).To(ContainSubstring("remember=" + strconv.FormatBool(tc.rememberMe)))
				callCount += 1
				headers := http.Header{}
				headers.Add("x-wealthsimple-otp-claim", testSomeOtpClaim)
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     headers,
				}, nil
			})

			// trigger
			resp, err := c.resolve2FA(ctx, types.TwoFactorAuthRequest{
				AuthenticatedClaim: testAuthClaim,
			}, []byte(testSomeAuthPayload))

			if tc.expectedErr != nil {
				g.Expect(err).To(Equal(tc.expectedErr))
				return
			}
			g.Expect(err).NotTo(HaveOccurred())

			// validate
			g.Expect(callCount).To(Equal(1))
			g.Expect(resp.StatusCode).To(Equal(http.StatusOK))
			if tc.rememberMe {
				g.Expect(c.RefreshOtpToken).To(Equal(testSomeOtpClaim))
			} else {
				g.Expect(c.RefreshOtpToken).To(Equal(tc.remeberMeToken))
			}
		})
	}
}

func Test_ParseSessionFromBody(t *testing.T) {
	g := NewWithT(t)
	// setup
	token := "{\"access_token\":\"eyJhbGciOiJSUzI1NiJ9.eyJzd\"," +
		"\"token_type\":\"Bearer\",\"expires_in\":3600," +
		"\"refresh_token\":\"8lahDQ_5ock11Tjvf6tR_pfEEDbddPeaMtTHAo9S4Ds\"," +
		"\"scope\":\"invest.read trade.read tax.read\",\"created_at\":1715053576}"

	sess, err := ParseSessionFromBody([]byte(token))
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(sess.RefreshToken).ToNot(BeEmpty())
	g.Expect(sess.AccessToken).ToNot(BeEmpty())
	g.Expect(time.Until(*sess.Expiry)).Should(BeNumerically("~", 3500*time.Second, 3700*time.Second))
}
