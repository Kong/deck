package kong

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/pkg/errors"

	"github.com/kong/go-kong/kong/custom"
)

const defaultBaseURL = "http://localhost:8001"

var pageSize = 1000

type service struct {
	client *Client
}

var (
	defaultCtx = context.Background()
)

// Client talks to the Admin API or control plane of a
// Kong cluster
type Client struct {
	client         *http.Client
	baseURL        string
	common         service
	Consumers      *ConsumerService
	Services       *Svcservice
	Routes         *RouteService
	CACertificates *CACertificateService
	Certificates   *CertificateService
	Plugins        *PluginService
	SNIs           *SNIService
	Upstreams      *UpstreamService
	Targets        *TargetService
	Workspaces     *WorkspaceService
	Admins         *AdminService
	RBACUsers      *RBACUserService

	credentials *credentialService
	KeyAuths    *KeyAuthService
	BasicAuths  *BasicAuthService
	HMACAuths   *HMACAuthService
	JWTAuths    *JWTAuthService
	MTLSAuths   *MTLSAuthService
	ACLs        *ACLService

	Oauth2Credentials *Oauth2Service

	logger         io.Writer
	debug          bool
	CustomEntities *CustomEntityService

	custom.Registry
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
func NewClient(baseURL *string, client *http.Client) (*Client, error) {
	if client == nil {
		client = http.DefaultClient
	}
	kong := new(Client)
	kong.client = client
	var rootURL string
	if baseURL != nil {
		rootURL = *baseURL
	} else {
		rootURL = defaultBaseURL
	}
	url, err := url.ParseRequestURI(rootURL)
	if err != nil {
		return nil, errors.Wrap(err, "parsing URL")
	}
	kong.baseURL = url.String()

	kong.common.client = kong
	kong.Consumers = (*ConsumerService)(&kong.common)
	kong.Services = (*Svcservice)(&kong.common)
	kong.Routes = (*RouteService)(&kong.common)
	kong.Plugins = (*PluginService)(&kong.common)
	kong.Certificates = (*CertificateService)(&kong.common)
	kong.CACertificates = (*CACertificateService)(&kong.common)
	kong.SNIs = (*SNIService)(&kong.common)
	kong.Upstreams = (*UpstreamService)(&kong.common)
	kong.Targets = (*TargetService)(&kong.common)
	kong.Workspaces = (*WorkspaceService)(&kong.common)
	kong.Admins = (*AdminService)(&kong.common)
	kong.RBACUsers = (*RBACUserService)(&kong.common)

	kong.credentials = (*credentialService)(&kong.common)
	kong.KeyAuths = (*KeyAuthService)(&kong.common)
	kong.BasicAuths = (*BasicAuthService)(&kong.common)
	kong.HMACAuths = (*HMACAuthService)(&kong.common)
	kong.JWTAuths = (*JWTAuthService)(&kong.common)
	kong.MTLSAuths = (*MTLSAuthService)(&kong.common)
	kong.ACLs = (*ACLService)(&kong.common)

	kong.Oauth2Credentials = (*Oauth2Service)(&kong.common)

	kong.CustomEntities = (*CustomEntityService)(&kong.common)
	kong.Registry = custom.NewDefaultRegistry()

	for i := 0; i < len(defaultCustomEntities); i++ {
		err := kong.Register(defaultCustomEntities[i].Type(),
			&defaultCustomEntities[i])
		if err != nil {
			return nil, err
		}
	}
	kong.logger = os.Stderr
	return kong, nil
}

// Do executes a HTTP request and returns a response
func (c *Client) Do(ctx context.Context, req *http.Request,
	v interface{}) (*Response, error) {
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

	response := newResponse(resp)

	///check for API errors
	if err = hasError(resp); err != nil {
		return response, err
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
	return response, err
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

// Status returns the status of a Kong node
func (c *Client) Status(ctx context.Context) (*Status, error) {

	req, err := c.NewRequest("GET", "/status", nil, nil)
	if err != nil {
		return nil, err
	}

	var s Status
	_, err = c.Do(ctx, req, &s)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// Root returns the response of GET request on root of
// Admin API (GET /).
func (c *Client) Root(ctx context.Context) (map[string]interface{}, error) {
	req, err := c.NewRequest("GET", "/", nil, nil)
	if err != nil {
		return nil, err
	}

	var root map[string]interface{}
	_, err = c.Do(ctx, req, &root)
	if err != nil {
		return nil, err
	}
	return root, nil
}
