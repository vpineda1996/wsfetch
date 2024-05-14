package types

import "net/http"

// feches the 2FA code given a claim. The claim is
// in JWT Format
type TwoFactorCodeFetcher interface {
	Fetch(twoFaHeader TwoFactorAuthRequest) (string, error)
}

type TwoFactorCodeFetcherFunc func(twoFaHeader TwoFactorAuthRequest) (string, error)

func (c TwoFactorCodeFetcherFunc) Fetch(req TwoFactorAuthRequest) (string, error) {
	return c(req)
}

type TwoFactorAuthRequest struct {
	// Required indicates if 2FA is needed
	Required bool
	// Method is the type of request (eg sms)
	Method string
	// JWT token to authenticate
	AuthenticatedClaim string
}

func (t TwoFactorAuthRequest) InjectToHeaders(headers *http.Header) {
	headers.Add("x-wealthsimple-otp-authenticated-claim", t.AuthenticatedClaim)
}
