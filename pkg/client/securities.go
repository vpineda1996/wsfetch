package client

import (
	"context"
	"fmt"

	"github.com/vpnda/wsfetch/pkg/client/generated"
)

func SecuritySymbolFromMarketData(marketData *generated.SecurityMarketData) (SecuritySymbol, error) {
	symbol := marketData.Stock.Symbol
	if marketData.Stock.PrimaryExchange != nil {
		symbol = fmt.Sprintf("%s:%s", *marketData.Stock.PrimaryExchange, symbol)
	}
	return SecuritySymbol(symbol), nil
}

func (c *client) GetSecurityMarketData(ctx context.Context, securityID string) (*generated.SecurityMarketData, error) {
	marketData, err := generated.FetchSecurityMarketData(ctx, c.tradeClient, securityID)
	if err != nil {
		return nil, err
	}

	return &marketData.Security.SecurityMarketData, nil
}
