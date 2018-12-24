package kong

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/go-querystring/query"
)

func (c *Client) newRequest(method, endpoint string, qs interface{},
	body interface{}) (*http.Request, error) {
	// TODO introduce method as a type, method as string
	// in http package is doomsday
	if endpoint == "" {
		return nil, errors.New("endpoint can't be nil")
	}
	//TODO verify endpoint is preceded with /

	validator, ok := body.(Validator)
	if ok && !validator.Valid() {
		return nil, errors.New("validation on entity in the request failed")
	}

	//body to be sent in JSON
	var buf []byte
	if body != nil {
		var err error
		buf, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
	}

	//Create a new request
	req, err := http.NewRequest(method, c.baseURL+endpoint,
		bytes.NewBuffer(buf))
	if body != nil {
		req.Header.Add("Content-Type", "application/json")
	}

	if err != nil {
		return nil, err
	}
	//Add query string if any
	if qs != nil {
		values, err := query.Values(qs)
		if err != nil {
			return nil, err
		}
		req.URL.RawQuery = values.Encode()
	}
	return req, nil
}
