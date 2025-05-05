package client

import (
	"context"

	"github.com/samber/lo"
	"github.com/vpnda/wsfetch/pkg/client/generated"
)

func (c *client) GetAccounts(ctx context.Context) ([]generated.Account, error) {
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

	var accountList []generated.Account
	for _, e := range accounts.GetIdentity().Accounts.Edges {
		accountList = append(accountList, e.Node.Account)
	}
	return accountList, nil
}
