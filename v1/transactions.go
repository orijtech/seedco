package seedco

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/orijtech/otils"
)

type Status string

const (
	Pending Status = "pending"
	Settled Status = "settled"
)

type SearchParams struct {
	Status    Status    `json:"status,omitempty"`
	StartDate time.Time `json:"start_date,omitempty"`
	EndDate   time.Time `json:"end_date,omitempty"`
	Offset    int       `json:"offset,omitempty"`
	Limit     int       `json:"limit,omitempty"`

	MaxPageNumber int64 `json:"max_page_number,omitempty"`
}

var errAlreadyCanceled = errors.New("already canceled")

func makeCanceler() (<-chan bool, func() error) {
	var once sync.Once
	cancelChan := make(chan bool, 1)
	cancelFn := func() error {
		var err error = errAlreadyCanceled
		once.Do(func() {
			err = nil
			close(cancelChan)
		})
		return err
	}
	return cancelChan, cancelFn
}

type SearchResults struct {
	PagesChan <-chan *TransactionPage
	Cancel    func() error
}

type TransactionPage struct {
	Transactions []*Transaction `json:"transactions,omitempty"`
	PageNumber   int64          `json:"p,omitempty"`
	Err          error          `json:"err,omitempty"`
}

type recvTransactions struct {
	Transactions []*Transaction `json:"results,omitempty"`
}

const defaultLimit = int(1000)

func (c *Client) ListTransactions(sp *SearchParams) (*SearchResults, error) {
	if sp == nil {
		sp = new(SearchParams)
	}

	maxPageNumber := sp.MaxPageNumber
	exceedsMaxPage := func(pn int64) bool {
		return maxPageNumber > 0 && pn >= maxPageNumber
	}

	cancelChan, cancelFn := makeCanceler()
	pagesChan := make(chan *TransactionPage)
	go func() {
		defer close(pagesChan)

		throttle := time.NewTicker(150 * time.Millisecond)
		spc := new(SearchParams)
		*spc = *sp
		if spc.Limit <= 0 {
			spc.Limit = defaultLimit
		}
		if spc.Offset <= 0 {
			spc.Offset = 0
		}

		pageNumber := int64(0)

		for {
			qv, err := otils.ToURLValues(spc)
			tPage := &TransactionPage{
				PageNumber: pageNumber,
			}
			if err != nil {
				tPage.Err = err
				pagesChan <- tPage
				return
			}

			fullURL := fmt.Sprintf("%s/public/transactions", baseURL)
			if len(qv) > 0 {
				fullURL = fmt.Sprintf("%s?%s", fullURL, qv.Encode())
			}
			req, err := http.NewRequest("GET", fullURL, nil)
			if err != nil {
				tPage.Err = err
				pagesChan <- tPage
				return
			}

			blob, _, err := c.doAuthAndReq(req)
			recvT := new(recvTransactions)
			if err := json.Unmarshal(blob, recvT); err != nil {
				tPage.Err = err
				pagesChan <- tPage
			} else if len(recvT.Transactions) > 0 {
				tPage.Transactions = recvT.Transactions
				pagesChan <- tPage
			}

			pageNumber += 1
			if len(tPage.Transactions) == 0 || exceedsMaxPage(pageNumber) {
				return
			}

			select {
			case <-throttle.C:
			case <-cancelChan:
				return
			}

			// Next increment the offset
			spc.Offset += spc.Limit
		}
	}()

	sr := &SearchResults{
		PagesChan: pagesChan,
		Cancel:    cancelFn,
	}
	return sr, nil
}

type Transaction struct {
	ID string `json:"id,omitempty"`

	CheckingAccountID string `json:"checking_account_id,omitempty"`

	AmountCents float64 `json:"amount,omitempty"`

	// Status possible values are Settled, Pending.
	Status Status `json:"status,omitempty"`

	Attachments []*Attachment `json:"attachments,omitempty"`

	// Date is when the transaction occured.
	Date *time.Time `json:"date,omitempty"`

	Category    string `json:"category,omitempty"`
	Description string `json:"description,omitempty"`

	// Memo is the user entered memorandum for the transaction.
	Memo string `json:"memo,omitempty"`
}

type Attachment struct {
	// TODO: Fill this in from seedco MGMT
}
