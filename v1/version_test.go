package seedco_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/orijtech/seedco/v1"
)

func TestAPIVersion(t *testing.T) {
	client, err := seedco.NewClientWithToken(testToken1)
	if err != nil {
		t.Fatal(err)
	}
	client.SetHTTPRoundTripper(&backend{route: apiVersionRoute})

	tests := [...]struct {
		token   string
		wantErr string

		wantAPIVersion *seedco.APIVersion
	}{
		0: {"token2", "", apiVersionFromFile("./testdata/api-version-2.json")},
		1: {"foo", "invalid credentials", nil},
		2: {"blanko", "invalid credentials", nil},
		3: {"", "invalid credentials", nil},
		4: {"token3", "Unauthorized", nil},
	}

	for i, tt := range tests {
		client.SetAuthToken(tt.token)
		apiVersion, err := client.APIVersion()
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
		if g, w := apiVersion, tt.wantAPIVersion; !reflect.DeepEqual(g, w) {
			t.Errorf("#%d:\ngot: %+v\nwant:%+v\n", i, g, w)
		}
	}
}

var tokenToVersionMap = map[string]string{
	"token2": "api-version-2-resp",
	"token3": "api-version-with-errs",
}

func ensureBearerTokenAuthd(req *http.Request) (string, *http.Response, error) {
	bearerToken := req.Header.Get("Authorization")
	bearerStr := "Bearer "
	idx := strings.Index(bearerToken, bearerStr)
	if idx < 0 {
		res, err := makeResp(`expecting Bearer token to have been set`, http.StatusUnauthorized, nil)
		return "", res, err
	}
	bearerToken = bearerToken[idx+len(bearerStr):]
	return bearerToken, nil, nil
}

func apiVersionRoundTrip(req *http.Request) (*http.Response, error) {
	bearerToken, badRes, err := ensureBearerTokenAuthd(req)
	if badRes != nil || err != nil {
		return badRes, err
	}
	basename := tokenToVersionMap[bearerToken]
	if basename == "" {
		return makeResp(`invalid credentials`, http.StatusUnauthorized, nil)
	}
	fullPath := fmt.Sprintf("./testdata/%s.json", basename)
	return respFromFile(fullPath)
}

func apiVersionFromFile(path string) *seedco.APIVersion {
	blob, err := ioutil.ReadFile(path)
	if err != nil {
		return nil
	}
	vers := new(seedco.APIVersion)
	if err := json.Unmarshal(blob, vers); err != nil {
		return nil
	}
	return vers
}
