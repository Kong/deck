package kong

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
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

func (k *Kong) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	var err error
	if req == nil {
		return nil, errors.New("Request object cannot be nil")
	}
	//Make the request
	rawResp, err := k.client.Do(req)
	return rawResp, err
}

func (k *Kong) Status() {

	req, err := http.NewRequest("GET", "http://localhost:8001/status", nil)

	if err != nil {
		log.Println(err)
		return
	}
	ctx := context.Background()

	res, err := k.Do(ctx, req)
	if err != nil {
		log.Println(err)
		return
	}
	defer res.Body.Close()
	buf, err := ioutil.ReadAll(res.Body)

	fmt.Println(string(buf))
}
