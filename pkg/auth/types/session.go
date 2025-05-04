package types

import (
	"encoding/json"
	"time"
)

type Session struct {
	// Access token
	AccessToken string

	// Refresh token
	RefreshToken string

	// Refresh OTP token
	RefreshOtpToken string

	// WSSID
	WSSID string

	// Session ID
	SessionId string

	// Client ID
	// Example: e123edb34b8bb66baeefbeef07275cc5
	ClientId string

	// Expiry
	Expiry *time.Time
}

func (s *Session) AuthPayload(clientId string) ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"grant_type":    "refresh_token",
		"refresh_token": s.RefreshToken,
		"client_id":     clientId,
	})
}

func (s *Session) Profile() string {
	return "invest"
}
