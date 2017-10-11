package seedco

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/orijtech/otils"
)

const baseURL = "https://api.seed.co/v1"

type Client struct {
	rt http.RoundTripper
	mu sync.RWMutex

	_authToken string
}

func (c *Client) doAuthAndReq(req *http.Request) ([]byte, http.Header, error) {
	bearerToken := fmt.Sprintf("Bearer %s", c.authToken())
	req.Header.Set("Authorization", bearerToken)
	return c.doReq(req)
}

func (c *Client) doReq(req *http.Request) ([]byte, http.Header, error) {
	res, err := c.httpClient().Do(req)
	if err != nil {
		return nil, nil, err
	}
	if res.Body != nil {
		defer res.Body.Close()
	}
	if !otils.StatusOK(res.StatusCode) {
		return nil, res.Header, fmt.Errorf(res.Status)
	}
	blob, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, res.Header, err
	}
	return blob, res.Header, nil
}

func (c *Client) authToken() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c._authToken
}

const EnvBearerTokenKey = "SEEDCO_BEARER_TOKEN"

var errMissingBearerTokenFromEnv = fmt.Errorf("missing %q value from environment", EnvBearerTokenKey)

func NewClientFromEnv() (*Client, error) {
	token := strings.TrimSpace(os.Getenv(EnvBearerTokenKey))
	if token == "" {
		return nil, errMissingBearerTokenFromEnv
	}
	return NewClientWithToken(token)
}

func NewClientWithToken(token string) (*Client, error) {
	return &Client{_authToken: token}, nil
}

func (c *Client) httpClient() *http.Client {
	c.mu.RLock()
	rt := c.rt
	c.mu.RUnlock()
	return &http.Client{Transport: rt}
}

func (c *Client) SetHTTPRoundTripper(rt http.RoundTripper) {
	c.mu.Lock()
	c.rt = rt
	c.mu.Unlock()
}

func (c *Client) SetAuthToken(token string) {
	c.mu.Lock()
	c._authToken = token
	c.mu.Unlock()
}

type Error struct {
	Message string `json:"message"`
}

var _ error = (*Error)(nil)

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}

func NewClient() (*Client, error) {
	return new(Client), nil
}
