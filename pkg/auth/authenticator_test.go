package auth

import (
	"context"
	"testing"
)

func Test_Client_Authenticate(t *testing.T) {

	ctx := context.Background()
	c := NewClient()

	c.Authenticate(ctx, PasswordCredentials{})
}
