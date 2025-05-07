package client

import (
	"context"

	"github.com/samber/lo"
	"github.com/vpnda/wsfetch/pkg/client/generated"
)

func (c *client) GetAccounts(ctx context.Context) ([]generated.AccountWithFinancials, error) {
	accounts, err := generated.FetchAllAccountFinancials(
		ctx,
		c.tradeClient,
		c.Identities.IdentityId,
		nil,
		lo.ToPtr(25),
		nil,
	)

	if err != nil {
		return nil, err
	}

	var accountList []generated.AccountWithFinancials
	for _, e := range accounts.GetIdentity().Accounts.Edges {
		accountList = append(accountList, e.Node.AccountWithFinancials)
	}
	return accountList, nil
}

// GetAccount implements Client.
func (c *client) GetAccount(ctx context.Context, accountId string) (*generated.AccountWithFinancials, error) {
	accounts, err := c.GetAccounts(ctx)
	if err != nil {
		return nil, err
	}
	for _, account := range accounts {
		if account.Id == accountId {
			return &account, nil
		}
	}
	return nil, ErrNoAccountFound
}
