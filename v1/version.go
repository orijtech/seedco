package seedco

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

var (
	errNilAPIVersion = errors.New("no APIVersion received")
)

func (c *Client) APIVersion() (*APIVersion, error) {
	fullURL := fmt.Sprintf("%s/public/api/client-version", baseURL)
	req, err := http.NewRequest("POST", fullURL, nil)
	if err != nil {
		return nil, err
	}
	blob, _, err := c.doAuthAndReq(req)
	if err != nil {
		return nil, err
	}
	avr := new(apiVersionResponse)
	if err := json.Unmarshal(blob, avr); err != nil {
		return nil, err
	}
	if err := flattenErrs(avr.Errors); err != nil {
		return nil, err
	}
	if len(avr.Results) == 0 {
		return nil, errNilAPIVersion
	}
	return avr.Results[0], nil
}

func flattenErrs(errsList []*Error) error {
	if len(errsList) == 0 {
		return nil
	}
	var strErrs []string
	for _, err := range errsList {
		if s := err.Error(); s != "" {
			strErrs = append(strErrs, s)
		}
	}
	if len(strErrs) == 0 {
		return nil
	}
	return errors.New(strings.Join(strErrs, "\n"))
}

type apiVersionResponse struct {
	Errors  []*Error      `json:"errors,omitempty"`
	Results []*APIVersion `json:"results,omitempty"`
}

type APIVersion struct {
	Version   string     `json:"api_version,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`

	ID string `json:"id,omitempty"`

	IndividualID string `json:"individual_id,omitempty"`
}
