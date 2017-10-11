package seedco_test

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"testing"

	"github.com/orijtech/seedco/v1"
)

func TestListTransactions(t *testing.T) {
	client, err := seedco.NewClientWithToken(testToken1)
	if err != nil {
		t.Fatal(err)
	}
	client.SetHTTPRoundTripper(&backend{route: listTransactionsRoute})

	tests := [...]struct {
		params        *seedco.SearchParams
		wantCount     int
		wantPageCount int
	}{
		0: {&seedco.SearchParams{Limit: 2, Offset: 0, MaxPageNumber: 1}, 2, 1},
		1: {&seedco.SearchParams{Limit: 2, Offset: 0}, 4, 2},
	}

	for i, tt := range tests {
		pres, err := client.ListTransactions(tt.params)
		if err != nil {
			t.Errorf("#%d: unexpected error: %v", i, err)
			continue
		}
		pageCount := 0
		n := 0
		for page := range pres.PagesChan {
			pageCount += 1
			if err := page.Err; err != nil {
				t.Errorf("pageNumber: %d err: %v", page.PageNumber, err)
				continue
			}
			n += len(page.Transactions)
		}
		if g, w := n, tt.wantCount; g != w {
			t.Errorf("#%d: itemCount: got=%d want=%d", i, g, w)
		}
		if g, w := pageCount, tt.wantPageCount; g != w {
			t.Errorf("#%d: pageCount: got=%d want=%d", i, g, w)
		}
	}
}

func listTransactionsRoundTrip(req *http.Request) (*http.Response, error) {
	_, badRes, err := ensureBearerTokenAuthd(req)
	if badRes != nil || err != nil {
		return badRes, err
	}
	query := req.URL.Query()
	limit, _ := strconv.Atoi(query.Get("limit"))
	offset, _ := strconv.Atoi(query.Get("offset"))
	fullPath := fmt.Sprintf("./testdata/transactions-%d,%d.json", offset, limit)
	if fi, err := os.Stat(fullPath); err != nil || fi == nil {
		prc, pwc := io.Pipe()
		go func() {
			fmt.Fprintf(pwc, `{}`)
			_ = pwc.Close()
		}()
		return makeResp(`200 OK`, http.StatusOK, prc)
	}
	return respFromFile(fullPath)
}
