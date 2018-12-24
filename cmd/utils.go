package cmd

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"os"
	"strings"

	"github.com/hbagdi/go-kong/kong"
	"github.com/pkg/errors"
)

type kongClientConfig struct {
	Address string

	Headers []string

	TLSSkipVerify bool
	TLSServerName string

	TLSCACert string

	Debug bool
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
func GetKongClient(opt kongClientConfig) (*kong.Client, error) {

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
	kongClient, err := kong.NewClient(kong.String(opt.Address), c)
	if err != nil {
		return nil, errors.Wrap(err, "creating client for Kong's Admin API")
	}
	if opt.Debug {
		kongClient.SetDebugMode(true)
		kongClient.SetLogger(os.Stderr)
	}
	return kongClient, nil
}
