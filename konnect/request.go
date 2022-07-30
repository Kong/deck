package konnect

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/go-querystring/query"
)

// NewRequest creates a request based on the inputs.
// endpoint should be relative to the baseURL specified during
// client creation.
// body is always marshaled into JSON.
func (c *Client) NewRequest(method, endpoint string, qs interface{},
	body interface{},
) (*http.Request, error) {
	if endpoint == "" {
		return nil, fmt.Errorf("endpoint can't be nil")
	}
	// body to be sent in JSON
	var buf []byte
	if body != nil {
		var err error
		buf, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
	}

	// Create a new request
	req, err := http.NewRequest(method, c.baseURL+endpoint,
		bytes.NewBuffer(buf))
	if err != nil {
		return nil, err
	}

	// add body if needed
	if body != nil {
		req.Header.Add("Content-Type", "application/json")
	}

	// add bearer token
	if c.token != "" {
		req.Header.Add("Authorization", "Bearer "+c.token)
	}

	// add query string if any
	if qs != nil {
		values, err := query.Values(qs)
		if err != nil {
			return nil, err
		}
		req.URL.RawQuery = values.Encode()
	}
	return req, nil
}
