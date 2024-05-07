package creds

import (
	"context"
	"time"

	"github.com/vpineda1996/wsfetch/pkg/auth/authenticator"
	"github.com/vpineda1996/wsfetch/pkg/auth/types"
)

// SessionFetcher fetches an active session
type SessionFetcher interface {
	GetSession(context.Context) (*types.Session, error)
}

// DefaultFetcher must not be copied, it should be passed
// by reference
type defaultFetcher struct {
	creds        types.PasswordCredentials
	sessionCache types.Session
}

var (
	_ SessionFetcher = &defaultFetcher{}
)

func NewDefaultFetcher(pc types.PasswordCredentials) SessionFetcher {
	return &defaultFetcher{
		creds: pc,
	}
}

func (df *defaultFetcher) GetSession(ctx context.Context) (*types.Session, error) {
	if isSessionActive(df.sessionCache) {
		return &df.sessionCache, nil
	}

	aClient, err := authenticator.NewPersistentClient()
	if err != nil {
		return nil, err
	}

	sess, err := aClient.Authenticate(ctx, df.creds)
	if err != nil {
		return nil, err
	}

	df.sessionCache = *sess
	return sess, nil
}

func isSessionActive(session types.Session) bool {
	return session.Expiry != nil && time.Until(*session.Expiry) > 5*time.Second
}
