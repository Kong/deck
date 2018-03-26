package kong

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
)

type service struct {
	client *Client
}

// Client talks to the Admin API or control plane of a
// Kong cluster
type Client struct {
	client  *http.Client
	BaseURL *url.URL
	common  service
	Sample  *SampleService
}

// Response is a Kong Admin API response. It wraps http.Response.
type Response struct {
	*http.Response
	//other Kong specific fields
}

// Status respresents current status of a Kong node.
type Status struct {
	Database struct {
		Reachable bool `json:"reachable"`
	} `json:"database"`
	Server struct {
		ConnectionsAccepted int `json:"connections_accepted"`
		ConnectionsActive   int `json:"connections_active"`
		ConnectionsHandled  int `json:"connections_handled"`
		ConnectionsReading  int `json:"connections_reading"`
		ConnectionsWaiting  int `json:"connections_waiting"`
		ConnectionsWriting  int `json:"connections_writing"`
		TotalRequests       int `json:"total_requests"`
	} `json:"server"`
}

// NewClient returns a Client which talks to Admin API of Kong
func NewClient(client *http.Client) *Client {
	if client == nil {
		client = http.DefaultClient
	}
	kong := new(Client)
	kong.client = client
	kong.common.client = kong
	kong.Sample = (*SampleService)(&kong.common)

	return kong
}

func newResponse(res *http.Response) *Response {
	return &Response{Response: res}
}

// Do executes a HTTP request and returns a response
func (c *Client) Do(ctx context.Context, req *http.Request, v interface{}) (*Response, error) {
	var err error
	if req == nil {
		return nil, errors.New("Request object cannot be nil")
	}
	req = req.WithContext(ctx)
	//Make the request
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	// Call Close on exit
	defer func() {
		var e error
		e = resp.Body.Close()
		if e != nil {
			err = e
		}
	}()
	response := newResponse(resp)

	//TODO check for response

	// response
	if v != nil {
		if writer, ok := v.(io.Writer); ok {
			_, err = io.Copy(writer, resp.Body)
			if err != nil {
				return nil, err
			}
		} else {
			err = json.NewDecoder(resp.Body).Decode(v)
			if err != nil {
				return nil, err
			}
		}
	}
	return response, err
}

// Status returns the status of a Kong node
func (c *Client) Status() (*Status, error) {

	req, err := http.NewRequest("GET", "http://localhost:8001/status", nil)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	ctx := context.Background()

	var s Status
	_, err = c.Do(ctx, req, &s)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return &s, nil
}
