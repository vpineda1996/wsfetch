package client

import (
	"context"
	"fmt"

	"github.com/vpineda1996/wsfetch/pkg/client/generated"
)

func (c *client) SecurityIDToSymbol(ctx context.Context, securityID string) (SecuritySymbol, error) {
	symbol := fmt.Sprintf("[%s]", securityID)

	if c.SecurityMarketDataCacheGetter != nil {
		marketData, err := c.GetSecurityMarketData(ctx, securityID)
		if err != nil {
			return "", err
		}
		if marketData != nil && marketData.Stock != nil {
			symbol = marketData.Stock.Symbol

			if marketData.Stock.PrimaryExchange != nil {
				symbol = fmt.Sprintf("%s:%s", *marketData.Stock.PrimaryExchange, symbol)
			}
		}
	}

	return SecuritySymbol(symbol), nil
}

func (c *client) GetSecurityMarketData(ctx context.Context, securityID string) (*generated.SecurityMarketData, error) {
	if c.SecurityMarketDataCacheGetter != nil {
		cachedValue, ok := c.SecurityMarketDataCacheGetter(securityID)
		if ok && cachedValue != nil {
			return cachedValue, nil
		}
	}

	marketData, err := generated.FetchSecurityMarketData(ctx, c.tradeClient, securityID)
	if err != nil {
		return nil, err
	}

	if c.SecurityMarketDataCacheSetter != nil {
		c.SecurityMarketDataCacheSetter(securityID, &marketData.Security.SecurityMarketData)
	}

	return &marketData.Security.SecurityMarketData, nil
}
