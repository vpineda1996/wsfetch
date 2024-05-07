package types

import "encoding/json"

// PasswordCredentials represents the credentials that customers use
type PasswordCredentials struct {
	Username string
	Password string
}

func (pc PasswordCredentials) AuthPayload() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"grant_type":     "password",
		"username":       pc.Username,
		"password":       pc.Password,
		"skip_provision": true,
		"otp_claim":      nil,
		// TODO: remove unecessary permissions here
		"scope": "invest.read trade.read tax.read",
		// magic string from WS
		"client_id": "4da53ac2b03225bed1550eba8e4611e086c7b905a3855e6ed12ea08c246758fa",
	})
}
