package cash

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/vpineda1996/wsfetch/pkg/wsclient"
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
	client *wsclient.WsClient
}

type graphQlRequest struct {
	OperationName string `json:"OperationName"`
	Variables     map[string]string
	Query         string `json:"query"`
}

func NewClient(c *wsclient.WsClient) Client {
	return &client{
		client: c,
	}
}

func (c *client) graphql(ctx context.Context, query string) ([]Transaction, error) {
	body, err := json.Marshal(graphQlRequest{
		OperationName: "clientSessionsQuery",
		Variables:     map[string]string{},
		Query:         "query clientSessionsQuery($clientId: ID!) {\n  client(id: $clientId) {\n    id\n    profile {\n      preferred_first_name\n      full_legal_name {\n        first_name\n        last_name\n        __typename\n      }\n      date_of_birth\n      identity_id\n      __typename\n    }\n    tier\n    jurisdiction\n    sri_interested\n    created_at\n    onboarding {\n      flow\n      states {\n        complete\n        risk_score\n        kyc_application\n        email_confirmation\n        signatures\n        e_signatures\n        debit_account\n        __typename\n      }\n      selected_product\n      funded\n      __typename\n    }\n    two_factor {\n      device_type\n      __typename\n    }\n    session {\n      id\n      email\n      unconfirmed_email\n      email_confirmed\n      email_updates\n      pending_co_owner_onboarding\n      phone\n      roles\n      feature_flags\n      block_user_due_to_risk_survey\n      skips_risk_survey\n      recovery_code\n      search {\n        api_key\n        indices\n        application_id\n        roles\n        feature_flags\n        __typename\n      }\n      impersonated\n      churned\n      is_advisor\n      is_employer\n      is_advised\n      is_halal\n      requires_pep_review\n      has_account\n      has_draft_transfers\n      call_required\n      digital_suitability\n      locale\n      theme\n      earn_rewards_hidden\n      global_notifications {\n        type\n        details\n        dismissed\n        dismissable\n        priority\n        __typename\n      }\n      trading_attributes {\n        tax_loss_harvest\n        __typename\n      }\n      reassessment_required\n      force_reassessment\n      reassessment_in_progress\n      __typename\n    }\n    __typename\n  }\n}",
	})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://my.wealthsimple.com/graphql", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	d, _ := httputil.DumpResponse(res, true)
	fmt.Println(string(d))

	return []Transaction{}, nil
}

// Transactions implements Client.
func (c *client) Transactions(ctx context.Context, from time.Time, to time.Time) ([]Transaction, error) {

	return c.graphql(ctx, "todo")
}
