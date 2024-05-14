package cash

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Khan/genqlient/graphql"
	"github.com/samber/lo"
	"github.com/vpineda1996/wsfetch/pkg/base"
	"github.com/vpineda1996/wsfetch/pkg/endpoints"
	"github.com/vpineda1996/wsfetch/pkg/services/cash/generated"
)

type Transaction struct {
	Date              time.Time
	Merchant          string
	Category          string
	Account           string
	OriginalStatement string
	Amount            string
}

// Client is able to make requests to Wealthsimple using graphql queries
type Client interface {
	Transactions(ctx context.Context, from time.Time, to time.Time) ([]Transaction, error)
}

type client struct {
	// Authenticated client from WS
	investClient graphql.Client

	// Trade Client
	tradeClient graphql.Client

	// Profile or User id, normally in the form
	// user-psde1sas14
	Identities *base.ClientIdentifiers
}

func NewClient(ctx context.Context, c *base.Wealthsimple) (Client, error) {
	cids, err := c.ClientIdentifiers(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch profile id: %w", err)
	}
	tradeInnerClient := *c
	tradeInnerClient.Profile = base.Trade
	return &client{
		investClient: graphql.NewClient(endpoints.MyWealthsimpleGetGraphQl.String(), c),
		tradeClient:  graphql.NewClient(endpoints.MyWealthsimpleGetGraphQl.String(), &tradeInnerClient),
		Identities:   cids,
	}, nil
}

// Transactions implements Client.
func (c *client) Transactions(ctx context.Context, from time.Time, to time.Time) ([]Transaction, error) {
	accounts, err := generated.FetchAllAccounts(
		ctx,
		c.investClient,
		c.Identities.IdentityId,
		25,
		nil,
	)

	if err != nil {
		return nil, err
	}

	var account *generated.Account
	for _, e := range accounts.GetIdentity().Accounts.Edges {
		if e.Node.GetUnifiedAccountType() == "CASH" {
			account = &e.Node.Account
			break
		}
	}
	if account == nil {
		return nil, fmt.Errorf("unable to find cash account")
	}

	oneMonthAgo := time.Now().Add(-20 * time.Hour * 24)
	res, err := generated.FetchActivityList(ctx, c.tradeClient, 25, nil,
		[]string{account.Id}, nil, oneMonthAgo)
	if err != nil {
		return nil, err
	}

	return activityListToTransaction(res)
}

func activityListToTransaction(res *generated.FetchActivityListResponse) ([]Transaction, error) {
	var trns []Transaction
	for _, e := range res.GetActivities().GetEdges() {
		nd := e.GetNode()
		amnt := nd.Amount
		if !strings.EqualFold(nd.AmountSign, "positive") {
			amnt = "-" + amnt
		}
		merch, _ := lo.Find([]*string{
			nd.BillPayPayeeNickname,
			nd.ETransferName,
			nd.P2pHandle,
			nd.InstitutionName,
			lo.ToPtr("unknown"),
		}, func(s *string) bool { return s != nil })

		tr := Transaction{
			Date:     nd.OccurredAt,
			Merchant: *merch,
			Category: nd.Type,
			Account:  nd.AccountId,
			Amount:   amnt,
		}
		trns = append(trns, tr)
	}
	return trns, nil
}
