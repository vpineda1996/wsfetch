package creds

import (
	"context"
	"fmt"

	"github.com/samber/lo"
	"github.com/vpineda1996/wsfetch/pkg/auth/types"
)

type StaticTokenFetcher types.Session

var (
	_ SessionFetcher = &StaticTokenFetcher{}
)

// GetSession gets credentials from a persistent client so it is easier to
func (t StaticTokenFetcher) GetSession(ctx context.Context) (*types.Session, error) {
	if isSessionActive(types.Session(t)) {
		return nil, fmt.Errorf("session has expired")
	}
	return lo.ToPtr(types.Session(t)), nil
}
