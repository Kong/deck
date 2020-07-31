package utils

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/hbagdi/go-kong/kong"
	"github.com/hbagdi/go-kong/kong/custom"
	"github.com/pkg/errors"
)

// KongRawState contains all of Kong Data
type KongRawState struct {
	Services []*kong.Service
	Routes   []*kong.Route

	Plugins []*kong.Plugin

	Upstreams []*kong.Upstream
	Targets   []*kong.Target

	Certificates   []*kong.Certificate
	SNIs           []*kong.SNI
	CACertificates []*kong.CACertificate

	Consumers      []*kong.Consumer
	CustomEntities []*custom.Entity

	KeyAuths    []*kong.KeyAuth
	HMACAuths   []*kong.HMACAuth
	JWTAuths    []*kong.JWTAuth
	BasicAuths  []*kong.BasicAuth
	ACLGroups   []*kong.ACLGroup
	Oauth2Creds []*kong.Oauth2Credential
	MTLSAuths   []*kong.MTLSAuth
}

// ErrArray holds an array of errors.
type ErrArray struct {
	Errors []error
}

// Error returns a pretty string of errors present.
func (e ErrArray) Error() string {
	if len(e.Errors) == 0 {
		return "nil"
	}
	var res string

	res = strconv.Itoa(len(e.Errors)) + " errors occurred:\n"
	for _, err := range e.Errors {
		res += fmt.Sprintf("\t%v\n", err)
	}
	return res
}

// KongClientConfig holds config details to use to talk to a Kong server.
type KongClientConfig struct {
	Address   string
	Workspace string

	TLSServerName string

	TLSCACert string

	TLSSkipVerify bool
	Debug         bool

	Headers []string
}

// HeaderRoundTripper injects Headers into requests
// made via RT.
type HeaderRoundTripper struct {
	headers []string
	rt      http.RoundTripper
}

// RoundTrip satisfies the RoundTripper interface.
func (t *HeaderRoundTripper) RoundTrip(req *http.Request) (*http.Response,
	error) {
	newRequest := new(http.Request)
	*newRequest = *req
	newRequest.Header = make(http.Header, len(req.Header))
	for k, s := range req.Header {
		newRequest.Header[k] = append([]string(nil), s...)
	}
	for _, s := range t.headers {
		split := strings.SplitN(s, ":", 2)
		if len(split) >= 2 {
			newRequest.Header[split[0]] = append([]string(nil), split[1])
		}
	}
	return t.rt.RoundTrip(newRequest)
}

// GetKongClient returns a Kong client
func GetKongClient(opt KongClientConfig) (*kong.Client, error) {

	var tlsConfig tls.Config
	if opt.TLSSkipVerify {
		tlsConfig.InsecureSkipVerify = true
	}
	if opt.TLSServerName != "" {
		tlsConfig.ServerName = opt.TLSServerName
	}

	if opt.TLSCACert != "" {
		certPool := x509.NewCertPool()
		ok := certPool.AppendCertsFromPEM([]byte(opt.TLSCACert))
		if !ok {
			return nil, errors.New("failed to load TLSCACert")
		}
		tlsConfig.RootCAs = certPool
	}

	c := &http.Client{}
	defaultTransport := http.DefaultTransport.(*http.Transport)
	defaultTransport.TLSClientConfig = &tlsConfig
	c.Transport = defaultTransport
	if len(opt.Headers) > 0 {
		c.Transport = &HeaderRoundTripper{
			headers: opt.Headers,
			rt:      defaultTransport,
		}
	}
	address := CleanAddress(opt.Address)

	url, err := url.ParseRequestURI(address)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse kong address")
	}
	if opt.Workspace != "" {
		url.Path = path.Join(url.Path, opt.Workspace)
	}

	kongClient, err := kong.NewClient(kong.String(url.String()), c)
	if err != nil {
		return nil, errors.Wrap(err, "creating client for Kong's Admin API")
	}
	if opt.Debug {
		kongClient.SetDebugMode(true)
		kongClient.SetLogger(os.Stderr)
	}
	return kongClient, nil
}

// CleanAddress removes trailling / from a URL.
func CleanAddress(address string) string {
	re := regexp.MustCompile("[/]+$")
	return re.ReplaceAllString(address, "")
}
