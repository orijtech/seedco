package seedco_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/orijtech/seedco/v1"
)

func TestListBalances(t *testing.T) {
	client, err := seedco.NewClientWithToken(testToken1)
	if err != nil {
		t.Fatal(err)
	}
	client.SetHTTPRoundTripper(&backend{route: listBalancesRoute})

	tests := [...]struct {
		token   string
		wantErr string
	}{
		0: {token: token1},
		1: {token: "invalid-token", wantErr: "unauthorized"},
		2: {token: errToken, wantErr: "bad balance"},
	}

	for i, tt := range tests {
		client.SetAuthToken(tt.token)
		balances, err := client.ListBalances()
		if tt.wantErr != "" {
			if err == nil {
				t.Errorf("#%d: want non-nil error", i)
			}
			continue
		}

		if err != nil {
			t.Errorf("#%d: unexpected error: %v", i, err)
			continue
		}
		if len(balances) == 0 {
			t.Errorf("#%d expected balances", i)
		}
	}
}

const (
	token1   = "token-1"
	token2   = "token-2"
	errToken = "err-token"
)

var tokenToBankAccountsMap = map[string]string{
	token1:   "bank-acct1",
	token2:   "bank-acct2",
	errToken: "bank-acct-with-err",
}

func listBalancesRoundTrip(req *http.Request) (*http.Response, error) {
	token, badRes, err := ensureBearerTokenAuthd(req)
	if badRes != nil || err != nil {
		return badRes, err
	}
	basename := tokenToBankAccountsMap[token]
	if basename == "" {
		return makeResp("not authorized", http.StatusUnauthorized, nil)
	}
	fullPath := fmt.Sprintf("./testdata/%s.json", basename)
	return respFromFile(fullPath)
}
