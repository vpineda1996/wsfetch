package httputil

import (
	"net/http"
	"time"
)

func ExtendHeaders(h *http.Header, deviceId string) {
	basicHeaders := map[string]string{
		"Accept":                "application/json",
		"Content-Type":          "application/json",
		"Date":                  time.Now().UTC().Format(time.RFC1123),
		"X-Wealthsimple-Client": "@wealthsimple/wealthsimple",
	}

	for k, v := range basicHeaders {
		h.Set(k, v)
	}

	if deviceId != "" {
		h.Set("x-ws-device-id", deviceId)
		h.Set("x-ws-session-id", "user_"+deviceId)
	}
}
