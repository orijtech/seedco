package seedco_test

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/orijtech/seedco"
)

type backend struct {
	route string
}

var _ http.RoundTripper = (*backend)(nil)

func TestAuthToken(t *testing.T) {
	client, err := seedco.NewClientWithToken(testToken1)
	if err != nil {
		t.Fatal(err)
	}
	client.SetHTTPRoundTripper(&backend{route: authRoute})

	tests := [...]struct {
		username  string
		password  string
		wantErr   string
		wantToken *seedco.Token
	}{
		0: {"", "", "blank", nil},
		1: {"foo-username", "", "passwords must be non-blank", nil},
		2: {"foo-username", "foo", "invalid credentials", nil},
		3: {"foo-username", "foo-password", "", tokenFromFile("./testdata/token1.json")},
	}

	for i, tt := range tests {
		token, err := client.AuthToken(tt.username, tt.password)
		if tt.wantErr != "" {
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("#%d:\ngot=(%v)\nwant match=(%v)", i, err, tt.wantErr)
			}
			continue
		}
		if err != nil {
			t.Errorf("#%d: unexpected err: %v", i, err)
			continue
		}
		if tt.wantToken == nil {
			t.Errorf("#%d: unexpectedly wantToken is nil", i)
			continue
		}
		if g, w := token, tt.wantToken; !reflect.DeepEqual(g, w) {
			t.Errorf("#%d:\ngot: %+v\nwant:%+v\n", i, g, w)
		}
	}
}

func TestRefreshToken(t *testing.T) {
	client, err := seedco.NewClientWithToken(testToken1)
	if err != nil {
		t.Fatal(err)
	}
	client.SetHTTPRoundTripper(&backend{route: refreshTokenRoute})

	tests := [...]struct {
		refreshToken string
		wantErr      string
		wantToken    *seedco.Token
	}{
		0: {"", "blank", nil},
		1: {"foo-fake-refresh-token", "invalid credentials", nil},
		2: {"foo-refresh-token", "", tokenFromFile("./testdata/token2.json")},
	}

	for i, tt := range tests {
		token, err := client.RefreshToken(tt.refreshToken)
		if tt.wantErr != "" {
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("#%d:\ngot=(%v)\nwant match=(%v)", i, err, tt.wantErr)
			}
			continue
		}
		if err != nil {
			t.Errorf("#%d: unexpected err: %v", i, err)
			continue
		}
		if tt.wantToken == nil {
			t.Errorf("#%d: unexpectedly wantToken is nil", i)
			continue
		}
		if g, w := token, tt.wantToken; !reflect.DeepEqual(g, w) {
			t.Errorf("#%d:\ngot: %+v\nwant:%+v\n", i, g, w)
		}
	}
}

const (
	testToken1 = "test-token-1"

	authRoute = "auth"

	apiVersionRoute = "api-version"

	refreshTokenRoute = "refresh-token"
)

func (b *backend) RoundTrip(req *http.Request) (*http.Response, error) {
	if req == nil {
		return makeResp("non-nil request wanted", http.StatusBadRequest, nil)
	}

	switch b.route {
	case authRoute:
		return authRoundTrip(req)
	case refreshTokenRoute:
		return refreshTokenRoundTrip(req)
	case apiVersionRoute:
		return apiVersionRoundTrip(req)
	default:
		return makeResp("unimplemented", http.StatusBadRequest, nil)
	}
}

var basicAuthMap = map[string]string{
	"foo-username": "foo-password",
}

func authRoundTrip(req *http.Request) (*http.Response, error) {
	username, recvPassword, ok := req.BasicAuth()
	if !ok {
		return makeResp(`expecting "username" and "password" to be set`, http.StatusBadRequest, nil)
	}
	savedPass := basicAuthMap[username]
	if savedPass != recvPassword {
		return makeResp(`invalid credentials`, http.StatusForbidden, nil)
	}
	return respFromFile("./testdata/token-resp.json")
}

var allowedRefreshTokens = map[string]string{
	"foo-refresh-token": "token-resp-1",
}

func refreshTokenRoundTrip(req *http.Request) (*http.Response, error) {
	// Ensure that the Content-Type is "application/json"
	if g, w := req.Header.Get("Content-Type"), "application/json"; g != w {
		return makeResp(fmt.Sprintf("contentType: got=%q want=%q", g, w), http.StatusBadRequest, nil)
	}
	blob, err := ioutil.ReadAll(req.Body)
	_ = req.Body.Close()
	if err != nil {
		return makeResp(err.Error(), http.StatusBadRequest, nil)
	}
	recv := make(map[string]string)
	if err := json.Unmarshal(blob, &recv); err != nil {
		return makeResp(err.Error(), http.StatusBadRequest, nil)
	}
	tokenBasename := allowedRefreshTokens[recv["refresh_token"]]
	if tokenBasename == "" {
		return makeResp(`invalid credentials`, http.StatusUnauthorized, nil)
	}
	srcFilepath := fmt.Sprintf("./testdata/%s.json", tokenBasename)
	return respFromFile(srcFilepath)
}

func makeResp(status string, code int, body io.ReadCloser) (*http.Response, error) {
	return &http.Response{
		Body:       body,
		Header:     make(http.Header),
		Status:     status,
		StatusCode: code,
	}, nil
}

func respFromFile(path string) (*http.Response, error) {
	f, err := os.Open(path)
	if err != nil {
		return makeResp(err.Error(), http.StatusInternalServerError, nil)
	}
	return makeResp("200 OK", http.StatusOK, f)
}

func tokenFromFile(path string) *seedco.Token {
	blob, err := ioutil.ReadFile(path)
	if err != nil {
		return nil
	}
	token := new(seedco.Token)
	if err := json.Unmarshal(blob, token); err != nil {
		return nil
	}
	return token
}
