package types

import "time"

type Session struct {
	// Example: e123edb34b8bb66baeefbeef07275cc5
	DeviceId string
	// Bearer token
	BearerToken string
	// Expiry
	Expiry *time.Time
}
