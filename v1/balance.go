package seedco

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Balance struct {
	// Accessible is the value in cents indicating
	// how much balance that you can use.
	// Accessible balance can easily change when
	// PendingDebits settle. That is why
	// TotalAvailable is the balance that is
	// safe to spend.
	Accessible float64 `json:"accessible,omitempty"`

	// PendingDebits indicates the total value of
	// debit transactions that haven't yet settled.
	PendingDebits float64 `json:"pending_debits,omitempty"`

	// PendingCredits indicates the total value of
	// credit transactions that haven't yet settled.
	PendingCredits float64 `json:"pending_credits,omitempty"`

	// ScheduledDebits indicates the total
	// balanced of scheduled debits.
	ScheduledDebits float64 `json:"scheduled_debits,omitempty"`

	// Settled is the total balance of settled transactions.
	Settled float64 `json:"settled,omitempty"`

	// Lockbox is the total balance  in the lockbox.
	Lockbox float64 `json:"lockbox,omitempty"`

	// TotalAvailable is the total balance that is safe to spend:
	// TotalAvailable = Accessible - PendingDebits - ScheduledDebits.
	TotalAvailable float64 `json:"total_available,omitempty"`

	CheckingAccountID string `json:"checking_account_id,omitempty"`
}

type listBalancesResponse struct {
	Balances []*Balance `json:"results"`
	Errors   []*Error   `json:"errors"`
}

func (c *Client) ListBalances() ([]*Balance, error) {
	fullURL := fmt.Sprintf("%s/public/balance", baseURL)
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, err
	}
	blob, _, err := c.doAuthAndReq(req)
	if err != nil {
		return nil, err
	}
	lbr := new(listBalancesResponse)
	if err := json.Unmarshal(blob, lbr); err != nil {
		return nil, err
	}
	if err := flattenErrs(lbr.Errors); err != nil {
		return nil, err
	}
	return lbr.Balances, nil
}
