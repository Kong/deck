package konnect

import (
	"context"
	"encoding/json"
	"github.com/pkg/errors"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

var (
	defaultCtx = context.Background()
)

type service struct {
	client *Client
}

// Client talks to the Konnect API.
type Client struct {
	client  *http.Client
	baseURL string
	common  service
	Auth    *AuthService
	logger  io.Writer
	debug   bool
}

// NewClient returns a Client which talks to Konnect's API.
func NewClient(httpClient *http.Client) (*Client, error) {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	client := new(Client)
	client.client = httpClient
	url, err := url.ParseRequestURI(baseURL)
	if err != nil {
		return nil, errors.Wrap(err, "parsing URL")
	}
	client.baseURL = url.String()

	client.common.client = client
	client.Auth = (*AuthService)(&client.common)
	client.logger = os.Stderr
	return client, nil
}

// Do executes a HTTP request and returns a response
func (c *Client) Do(ctx context.Context, req *http.Request,
	v interface{}) (*http.Response, error) {
	var err error
	if req == nil {
		return nil, errors.New("request cannot be nil")
	}
	if ctx == nil {
		ctx = defaultCtx
	}
	req = req.WithContext(ctx)

	// log the request
	err = c.logRequest(req)
	if err != nil {
		return nil, err
	}

	//Make the request
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "making HTTP request")
	}

	// log the response
	err = c.logResponse(resp)
	if err != nil {
		return nil, err
	}

	// check for API errors
	if err = hasError(resp); err != nil {
		return resp, err
	}

	// Call Close on exit
	defer func() {
		e := resp.Body.Close()
		if e != nil {
			err = e
		}
	}()

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
	return resp, err
}

// SetDebugMode enables or disables logging of
// the request to the logger set by SetLogger().
// By default, debug logging is disabled.
func (c *Client) SetDebugMode(enableDebug bool) {
	c.debug = enableDebug
}

func (c *Client) logRequest(r *http.Request) error {
	if !c.debug {
		return nil
	}
	dump, err := httputil.DumpRequestOut(r, true)
	if err != nil {
		return err
	}
	_, err = c.logger.Write(append(dump, '\n'))
	return err
}

func (c *Client) logResponse(r *http.Response) error {
	if !c.debug {
		return nil
	}
	dump, err := httputil.DumpResponse(r, true)
	if err != nil {
		return err
	}
	_, err = c.logger.Write(append(dump, '\n'))
	return err
}

// SetLogger sets the debug logger, defaults to os.StdErr
func (c *Client) SetLogger(w io.Writer) {
	if w == nil {
		return
	}
	c.logger = w
}
