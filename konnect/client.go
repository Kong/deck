package konnect

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/pkg/errors"
)

var defaultCtx = context.Background()

type service struct {
	client         *Client
	controlPlaneID string
}

// Client talks to the Konnect API.
type Client struct {
	client                *http.Client
	baseURL               string
	common                service
	Auth                  *AuthService
	ServicePackages       *ServicePackageService
	ServiceVersions       *ServiceVersionService
	Documents             *DocumentService
	ControlPlanes         *ControlPlaneService
	ControlPlaneRelations *ControlPlaneRelationsService
	logger                io.Writer
	debug                 bool
}

// ClientOpts contains configuration options for a new Client.
type ClientOpts struct {
	BaseURL string
}

// NewClient returns a Client which talks to Konnect's API.
func NewClient(httpClient *http.Client, opts ClientOpts) (*Client, error) {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	client := new(Client)
	client.client = httpClient
	url, err := url.ParseRequestURI(opts.BaseURL)
	if err != nil {
		return nil, errors.Wrap(err, "parsing URL")
	}
	client.baseURL = url.String()

	client.common.client = client
	client.Auth = (*AuthService)(&client.common)
	client.ServicePackages = (*ServicePackageService)(&client.common)
	client.ServiceVersions = (*ServiceVersionService)(&client.common)
	client.Documents = (*DocumentService)(&client.common)
	client.ControlPlanes = (*ControlPlaneService)(&client.common)
	client.ControlPlaneRelations = (*ControlPlaneRelationsService)(&client.common)
	client.logger = os.Stderr
	return client, nil
}

// SetControlPlaneID sets the kong control-plane ID in the client.
// This is used to inject the control-plane ID in requests as needed.
func (c *Client) SetControlPlaneID(cpID string) {
	c.common.controlPlaneID = cpID
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

	// Make the request
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
