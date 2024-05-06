package auth

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type HeaderOpts struct {
	DeviceId  string
	SessionId string
}

func BuildHeaders(opts HeaderOpts) http.Header {
	basicHeaders := map[string]string{
		"Accept":                "application/json",
		"Content-Type":          "application/json",
		"Date":                  time.Now().UTC().Format(time.RFC1123),
		"X-Wealthsimple-Client": "@wealthsimple/wealthsimple",
	}

	res := make(http.Header)
	for k, v := range basicHeaders {
		res.Add(k, v)
	}

	if opts.DeviceId != "" {
		res.Add("x-ws-device-id", opts.DeviceId)
	}

	if opts.SessionId != "" {
		res.Add("x-ws-session-id", opts.SessionId)
	}

	return res
}

type TwoFactorAuthHeaders struct {
	Required           bool
	Method             string
	AuthenticatedClaim string
}

func Parse2FAHeaders(headers http.Header) (TwoFactorAuthHeaders, error) {
	ret := TwoFactorAuthHeaders{}
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
