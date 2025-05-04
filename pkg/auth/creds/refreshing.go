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
	creds        authenticator.AuthPayloadCreator
	sessionCache *types.Session
}

var (
	_ SessionFetcher = &defaultFetcher{}
)

func NewDefaultFetcher(pc types.PasswordCredentials) SessionFetcher {
	return &defaultFetcher{
		creds: pc,
	}
}

func NewFetcherFromExistingSession(s *types.Session) SessionFetcher {
	return &defaultFetcher{
		sessionCache: s,
		creds:        s,
	}
}

// GetSession gets credentials from a persistent client so it is easier to
func (df *defaultFetcher) GetSession(ctx context.Context) (*types.Session, error) {
	if isSessionActive(df.sessionCache) {
		return df.sessionCache, nil
	}

	var client authenticator.Client
	if df.sessionCache != nil {
		client = authenticator.NewClientFromSession(df.sessionCache)
	} else {
		client = authenticator.NewClient()
	}

	sess, err := client.Authenticate(ctx, df.creds)
	if err != nil {
		return nil, err
	}

	df.sessionCache = sess
	return sess, nil
}

func isSessionActive(session *types.Session) bool {
	return session != nil && session.Expiry != nil && time.Until(*session.Expiry) > 5*time.Second
}
