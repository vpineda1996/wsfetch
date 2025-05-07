package client

import (
	"context"
	"time"

	"github.com/samber/lo"
	"github.com/vpnda/wsfetch/pkg/client/generated"
)

// GetActivities implements Client.
func (c *client) GetActivities(ctx context.Context, accountIds []AccountId, from *time.Time, until *time.Time) (map[AccountId][]generated.Activity, error) {
	// Convert AccountId slice to string slice
	accountIdStrs := make([]string, len(accountIds))
	for i, id := range accountIds {
		accountIdStrs[i] = string(id)
	}

	res, err := generated.FetchActivityFeedItems(ctx, c.tradeClient, lo.ToPtr(25), nil, &generated.ActivityCondition{
		AccountIds: accountIdStrs,
		StartDate:  from,
		EndDate:    until,
	}, []generated.ActivitiesOrderBy{generated.ActivitiesOrderByOccurredAtDesc})
	if err != nil {
		return nil, err
	}

	// Convert slice to map
	result := make(map[AccountId][]generated.Activity)
	for _, e := range res.GetActivityFeedItems().GetEdges() {
		nd := e.GetNode()
		id := AccountId(nd.GetAccountId())
		if _, ok := result[id]; !ok {
			result[id] = []generated.Activity{}
		}
		result[id] = append(result[id], nd.Activity)
	}

	return result, nil
}
