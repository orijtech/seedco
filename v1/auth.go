package seedco

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type Token struct {
	AccessToken  string  `json:"access_token"`
	ExpiresIn    float64 `json:"expires_in"`
	Permissions  string  `json:"permissions"`
	RefreshToken string  `json:"refresh_token"`
	TokenType    string  `json:"token_type"`
}

type tokenListing struct {
	Tokens []*Token `json:"results"`
	Errors []*Error `json:"errors"`
}

var (
	errNoToken       = errors.New("no token could be parsed")
	errRefreshToken  = errors.New("refreshToken cannot be blank")
	errBlankUsername = errors.New("usernames must be non-blank")
	errBlankPassword = errors.New("passwords must be non-blank")
)

func (c *Client) AuthToken(username, password string) (*Token, error) {
	if username == "" {
		return nil, errBlankUsername
	}
	if password == "" {
		return nil, errBlankPassword
	}
	fullURL := fmt.Sprintf("%s/public/auth/token", baseURL)
	req, err := http.NewRequest("POST", fullURL, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(username, password)
	return c.doReqAndParseOutTokenFromTokenListing(req)
}

func (c *Client) doReqAndParseOutTokenFromTokenListing(req *http.Request) (*Token, error) {
	blob, _, err := c.doReq(req)
	if err != nil {
		return nil, err
	}
	tkl := new(tokenListing)
	if err := json.Unmarshal(blob, tkl); err != nil {
		return nil, err
	}
	if err := flattenErrs(tkl.Errors); err != nil {
		return nil, err
	}
	var token *Token
	if len(tkl.Tokens) > 0 {
		token = tkl.Tokens[0]
	}
	if token == nil {
		return nil, errNoToken
	}
	return token, nil
}

func (c *Client) RefreshToken(refreshToken string) (*Token, error) {
	if refreshToken == "" {
		return nil, errRefreshToken
	}
	blob, err := json.Marshal(map[string]string{"refresh_token": refreshToken})
	if err != nil {
		return nil, err
	}
	fullURL := fmt.Sprintf("%s/public/auth/token/refresh", refreshToken)
	req, err := http.NewRequest("POST", fullURL, bytes.NewReader(blob))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return c.doReqAndParseOutTokenFromTokenListing(req)
}
