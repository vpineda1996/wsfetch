package client

import (
	"context"
	"fmt"
	"time"

	"github.com/Khan/genqlient/graphql"
	"github.com/vpineda1996/wsfetch/pkg/base"
	"github.com/vpineda1996/wsfetch/pkg/client/generated"
	"github.com/vpineda1996/wsfetch/pkg/endpoints"
)

type Transaction struct {
	Date              time.Time
	Merchant          string
	Category          string
	Account           string
	OriginalStatement string
	Amount            string
	Description       string
}

type AccountId string
type SecuritySymbol string

// Client is able to make requests to Wealthsimple using graphql queries
type Client interface {
	GetAccounts(ctx context.Context) ([]generated.Account, error)
	Transactions(ctx context.Context, accountIds []AccountId, from time.Time, until *time.Time) (map[AccountId][]Transaction, error)

	SecurityIDToSymbol(ctx context.Context, s string) (SecuritySymbol, error)
}

type client struct {
	// Trade Client
	tradeClient graphql.Client

	// Profile or User id, normally in the form
	// user-psde1sas14
	Identities *base.TokenInformation

	// Cache functions for security market data
	SecurityMarketDataCacheGetter func(securityID string) (*generated.SecurityMarketData, bool)
	SecurityMarketDataCacheSetter func(securityID string, data *generated.SecurityMarketData)
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
