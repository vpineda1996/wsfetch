package client

import (
	"context"
	"strings"
	"time"

	"github.com/samber/lo"
	"github.com/vpineda1996/wsfetch/pkg/client/generated"
)

// Transactions implements Client.
func (c *client) Transactions(ctx context.Context, accountIds []AccountId, from time.Time, until *time.Time) (map[AccountId][]Transaction, error) {
	// Convert AccountId slice to string slice
	accountIdStrs := make([]string, len(accountIds))
	for i, id := range accountIds {
		accountIdStrs[i] = string(id)
	}

	res, err := generated.FetchActivityFeedItems(ctx, c.tradeClient, lo.ToPtr(25), nil, &generated.ActivityCondition{
		AccountIds: accountIdStrs,
		EndDate:    until,
	}, []generated.ActivitiesOrderBy{generated.ActivitiesOrderByOccurredAtDesc})
	if err != nil {
		return nil, err
	}

	transactions, err := c.activityListToTransaction(ctx, res)
	if err != nil {
		return nil, err
	}

	// Convert slice to map
	result := make(map[AccountId][]Transaction)
	for _, t := range transactions {
		id := AccountId(t.Account)
		if _, ok := result[id]; !ok {
			result[id] = []Transaction{}
		}
		result[id] = append(result[id], t)
	}

	return result, nil
}

func (c *client) activityListToTransaction(ctx context.Context, res *generated.FetchActivityFeedItemsResponse) ([]Transaction, error) {
	var trns []Transaction
	for _, e := range res.ActivityFeedItems.GetEdges() {
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

		description, err := GetActivityDescription(ctx, c, &nd.Activity)
		if err != nil {
			return nil, err
		}

		tr := Transaction{
			Date:        *nd.OccurredAt,
			Merchant:    *merch,
			Category:    string(nd.Type),
			Account:     nd.AccountId,
			Amount:      amnt,
			Description: description,
		}
		trns = append(trns, tr)
	}
	return trns, nil
}
