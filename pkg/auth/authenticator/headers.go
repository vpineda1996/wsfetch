package authenticator

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/vpineda1996/wsfetch/pkg/auth/types"
)

func ExtendHeaders(h *http.Header, deviceId string, otpClaim string) {
	basicHeaders := map[string]string{
		"Accept":                "application/json",
		"Content-Type":          "application/json",
		"Date":                  time.Now().UTC().Format(time.RFC1123),
		"X-Wealthsimple-Client": "@wealthsimple/wealthsimple",
	}

	for k, v := range basicHeaders {
		h.Add(k, v)
	}

	if deviceId != "" {
		h.Add("x-ws-device-id", deviceId)
		h.Add("x-ws-session-id", "user_"+deviceId)
	}

	if otpClaim != "" {
		h.Add("x-wealthsimple-otp-claim", otpClaim)
	}
}

func InjectAuthClaimWithCode(headers *http.Header, twoFa types.TwoFactorAuthRequest, code string, remember bool) {
	twoFa.InjectToHeaders(headers)
	headers.Add("x-wealthsimple-otp", fmt.Sprintf("%s;remember=%t", code, remember))
}

func Parse2FAHeaders(headers http.Header) (types.TwoFactorAuthRequest, error) {
	ret := types.TwoFactorAuthRequest{}
	if headers.Get("x-wealthsimple-otp-required") != "" {
		v, err := strconv.ParseBool(headers.Get("x-wealthsimple-otp-required"))
		if err != nil {
			return ret, err
		}
		ret.Required = v
		if !ret.Required {
			return ret, nil
		}
	}

	if headers.Get("x-wealthsimple-otp-authenticated-claim") == "" ||
		headers.Get("x-wealthsimple-otp") == "" {
		return ret, fmt.Errorf("unable to parse auth claim: %v", headers)
	}

	ret.AuthenticatedClaim = headers.Get("x-wealthsimple-otp-authenticated-claim")
	for _, s := range strings.Split(headers.Get("x-wealthsimple-otp"), ";") {
		sTrim := strings.TrimSpace(s)
		if strings.HasPrefix(sTrim, "method") {
			ret.Method = strings.Split(sTrim, "=")[1]
		}
	}

	return ret, nil
}
