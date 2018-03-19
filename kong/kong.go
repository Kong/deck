package kong

import (
	"fmt"
	"net/http"
	"net/url"
)

type service struct {
	client *Kong
}

type Kong struct {
	client  *http.Client
	BaseURL *url.URL
	common  service
	Sample  *SampleService
}

func New(client *http.Client) *Kong {
	if client == nil {
		client = http.DefaultClient
	}
	kong := new(Kong)
	kong.client = client
	kong.common.client = kong
	kong.Sample = (*SampleService)(&kong.common)

	return kong
}

func (k *Kong) Do() {
	fmt.Println("Kong.Do()")
}
