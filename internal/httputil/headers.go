package httputil

import (
	"net/http"
	"time"
)

func ExtendHeaders(h *http.Header, wssid string) {
	basicHeaders := map[string]string{
		"Accept":                "application/json",
		"Content-Type":          "application/json",
		"Date":                  time.Now().UTC().Format(time.RFC1123),
		"X-Wealthsimple-Client": "@wealthsimple/wealthsimple",
	}

	for k, v := range basicHeaders {
		h.Set(k, v)
	}

	if wssid != "" {
		h.Set("x-ws-device-id", wssid)
		h.Set("x-ws-session-id", "user_"+wssid)
	}
}
