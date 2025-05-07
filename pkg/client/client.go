package client

import (
	"context"
	"fmt"
	"time"

	"github.com/Khan/genqlient/graphql"
	"github.com/vpnda/wsfetch/pkg/base"
	"github.com/vpnda/wsfetch/pkg/client/generated"
	"github.com/vpnda/wsfetch/pkg/endpoints"
)

type AccountId string
type SecuritySymbol string

// Client is able to make requests to Wealthsimple using graphql queries
type Client interface {
	GetAccount(ctx context.Context, accountId string) (*generated.AccountWithFinancials, error)
	GetAccounts(ctx context.Context) ([]generated.AccountWithFinancials, error)
	GetActivities(ctx context.Context, accountIds []AccountId, from *time.Time, until *time.Time) (map[AccountId][]generated.Activity, error)

	GetSecurityMarketData(ctx context.Context, securityID string) (*generated.SecurityMarketData, error)
}

var (
	_ Client = &client{}
	_ Client = &cachingClient{}
)

type client struct {
	// Trade Client
	tradeClient graphql.Client

	// Profile or User id, normally in the form
	// user-psde1sas14
	Identities *base.TokenInformation
}

type cachingClient struct {
	delegate Client
	// Cache functions for security market data
	securityMarketDataCacheGetter func(securityID string) (*generated.SecurityMarketData, bool)
	securityMarketDataCacheSetter func(securityID string, data *generated.SecurityMarketData)

	// Cache functions for account data
	accountCacheGetter func(accountID string) (*generated.AccountWithFinancials, bool)
	accountCacheSetter func(accountID string, data *generated.AccountWithFinancials)
}

func NewClient(ctx context.Context, c *base.Wealthsimple) (Client, error) {
	cids, err := c.GetTokenInformation(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch profile id: %w", err)
	}
	tradeInnerClient := *c
	tradeInnerClient.Profile = base.Trade
	return &client{
		tradeClient: graphql.NewClient(endpoints.MyWealthsimpleGetGraphQl.String(), &tradeInnerClient),
		Identities:  cids,
	}, nil
}
