package client

import (
	"context"
	"time"

	"github.com/vpnda/wsfetch/pkg/client/generated"
)

func NewCachingClient(c Client) *cachingClient {
	marketDataCache := make(map[string]*generated.SecurityMarketData)
	accountCache := make(map[string]*generated.AccountWithFinancials)

	return &cachingClient{
		delegate: c,
		securityMarketDataCacheGetter: func(securityID string) (*generated.SecurityMarketData, bool) {
			data, ok := marketDataCache[securityID]
			return data, ok
		},
		securityMarketDataCacheSetter: func(securityID string, data *generated.SecurityMarketData) {
			marketDataCache[securityID] = data
		},
		accountCacheGetter: func(accountID string) (*generated.AccountWithFinancials, bool) {
			data, ok := accountCache[accountID]
			return data, ok
		},
		accountCacheSetter: func(accountID string, data *generated.AccountWithFinancials) {
			accountCache[accountID] = data
		},
	}
}

// GetAccounts implements Client.
func (c *cachingClient) GetAccounts(ctx context.Context) ([]generated.AccountWithFinancials, error) {
	accounts, err := c.delegate.GetAccounts(ctx)
	if err != nil {
		return nil, err
	}

	for _, account := range accounts {
		c.accountCacheSetter(account.Id, &account)
	}

	return accounts, nil
}

// GetAccount implements Client.
func (c *cachingClient) GetAccount(ctx context.Context, accountId string) (*generated.AccountWithFinancials, error) {
	if account, ok := c.accountCacheGetter(accountId); ok {
		return account, nil
	}
	account, err := c.delegate.GetAccount(ctx, accountId)
	if err != nil {
		return nil, err
	}
	c.accountCacheSetter(accountId, account)
	return account, nil
}

// GetActivities implements Client.
func (c *cachingClient) GetActivities(ctx context.Context, accountIds []AccountId, from *time.Time, until *time.Time) (map[AccountId][]generated.Activity, error) {
	return c.delegate.GetActivities(ctx, accountIds, from, until)
}

// SecurityIDToSymbol implements Client.
func (c *cachingClient) GetSecurityMarketData(ctx context.Context, securityID string) (*generated.SecurityMarketData, error) {
	if marketData, ok := c.securityMarketDataCacheGetter(securityID); ok {
		return marketData, nil
	}
	marketData, err := c.delegate.GetSecurityMarketData(ctx, securityID)
	if err != nil {
		return nil, err
	}
	c.securityMarketDataCacheSetter(securityID, marketData)
	return marketData, nil
}
