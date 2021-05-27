package file

import (
	"encoding/json"
	"reflect"
	"testing"

	ghodss "github.com/ghodss/yaml"
	"github.com/kong/go-kong/kong"
	"github.com/stretchr/testify/assert"
)

var (
	jsonString = `{                                                                           
  "name": "rate-limiting",                                                  
  "config": {                                                               
    "minute": 10
  },                                                                        
  "service": "foo",                                                         
  "route": "bar",                                                         
  "consumer": "baz",                                                        
  "enabled": true,                                                          
  "run_on": "first",                                                        
  "protocols": [                                                            
    "http"
  ]                                                                         
}`
	yamlString = `
name: rate-limiting
config:
  minute: 10
service: foo
consumer: baz
route: bar
enabled: true
run_on: first
protocols:
- http
`
)

func Test_sortKey(t *testing.T) {
	tests := []struct {
		name        string
		sortable    sortable
		expectedKey string
	}{
		{
			sortable: &FService{
				Service: kong.Service{
					Name: kong.String("my-service"),
					ID:   kong.String("my-id"),
				},
			},
			expectedKey: "my-service",
		},
		{
			sortable: &FService{
				Service: kong.Service{
					ID: kong.String("my-id"),
				},
			},
			expectedKey: "my-id",
		},
		{
			sortable:    FService{},
			expectedKey: "",
		},
		{
			sortable: &FRoute{
				Route: kong.Route{
					Name: kong.String("my-route"),
					ID:   kong.String("my-id"),
				},
			},
			expectedKey: "my-route",
		},
		{
			sortable: FRoute{
				Route: kong.Route{
					ID: kong.String("my-id"),
				},
			},
			expectedKey: "my-id",
		},
		{
			sortable:    FRoute{},
			expectedKey: "",
		},
		{
			sortable: FUpstream{
				Upstream: kong.Upstream{
					Name: kong.String("my-upstream"),
					ID:   kong.String("my-id"),
				},
			},
			expectedKey: "my-upstream",
		},
		{
			sortable: FUpstream{
				Upstream: kong.Upstream{
					ID: kong.String("my-id"),
				},
			},
			expectedKey: "my-id",
		},
		{
			sortable:    FUpstream{},
			expectedKey: "",
		},
		{
			sortable: FTarget{
				Target: kong.Target{
					Target: kong.String("my-target"),
					ID:     kong.String("my-id"),
				},
			},
			expectedKey: "my-target",
		},
		{
			sortable: FTarget{
				Target: kong.Target{
					ID: kong.String("my-id"),
				},
			},
			expectedKey: "my-id",
		},
		{
			sortable:    FTarget{},
			expectedKey: "",
		},
		{
			sortable: FCertificate{
				Cert: kong.String("my-certificate"),
				ID:   kong.String("my-id"),
			},
			expectedKey: "my-certificate",
		},
		{
			sortable: FCertificate{
				ID: kong.String("my-id"),
			},
			expectedKey: "my-id",
		},
		{
			sortable:    FCertificate{},
			expectedKey: "",
		},
		{
			sortable: FCACertificate{
				CACertificate: kong.CACertificate{
					Cert: kong.String("my-ca-certificate"),
					ID:   kong.String("my-id"),
				},
			},
			expectedKey: "my-ca-certificate",
		},
		{
			sortable: FCACertificate{
				CACertificate: kong.CACertificate{
					ID: kong.String("my-id"),
				},
			},
			expectedKey: "my-id",
		},
		{
			sortable:    FCACertificate{},
			expectedKey: "",
		},
		{
			sortable: FPlugin{
				Plugin: kong.Plugin{
					Name: kong.String("my-plugin"),
					ID:   kong.String("my-id"),
				},
			},
			expectedKey: "my-plugin",
		},
		{
			sortable: FPlugin{
				Plugin: kong.Plugin{
					Name: kong.String("my-plugin"),
					ID:   kong.String("my-id"),
					Consumer: &kong.Consumer{
						ID: kong.String("my-consumer-id"),
					},
				},
			},
			expectedKey: "my-pluginmy-consumer-id",
		},
		{
			sortable: FPlugin{
				Plugin: kong.Plugin{
					Name: kong.String("my-plugin"),
					ID:   kong.String("my-id"),
					Route: &kong.Route{
						ID: kong.String("my-route-id"),
					},
				},
			},
			expectedKey: "my-pluginmy-route-id",
		},
		{
			sortable: FPlugin{
				Plugin: kong.Plugin{
					Name: kong.String("my-plugin"),
					ID:   kong.String("my-id"),
					Service: &kong.Service{
						ID: kong.String("my-service-id"),
					},
				},
			},
			expectedKey: "my-pluginmy-service-id",
		},

		{
			sortable: FPlugin{
				Plugin: kong.Plugin{
					ID: kong.String("my-id"),
				},
			},
			expectedKey: "my-id",
		},
		{
			sortable:    FPlugin{},
			expectedKey: "",
		},
		{
			sortable: &FConsumer{
				Consumer: kong.Consumer{
					Username: kong.String("my-consumer"),
					ID:       kong.String("my-id"),
				},
			},
			expectedKey: "my-consumer",
		},
		{
			sortable: &FConsumer{
				Consumer: kong.Consumer{
					ID: kong.String("my-id"),
				},
			},
			expectedKey: "my-id",
		},
		{
			sortable:    FConsumer{},
			expectedKey: "",
		},
		{
			sortable: &FServicePackage{
				Name: kong.String("my-service-package"),
				ID:   kong.String("my-id"),
			},
			expectedKey: "my-service-package",
		},
		{
			sortable: &FServicePackage{
				ID: kong.String("my-id"),
			},
			expectedKey: "my-id",
		},
		{
			sortable:    FServicePackage{},
			expectedKey: "",
		},
		{
			sortable: &FServiceVersion{
				Version: kong.String("my-service-version"),
				ID:      kong.String("my-id"),
			},
			expectedKey: "my-service-version",
		},
		{
			sortable: &FServiceVersion{
				ID: kong.String("my-id"),
			},
			expectedKey: "my-id",
		},
		{
			sortable:    FServiceVersion{},
			expectedKey: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := tt.sortable.sortKey()
			if key != tt.expectedKey {
				t.Errorf("Expected %v, but is %v", tt.expectedKey, key)
			}
		})
	}
}

func TestPluginUnmarshalYAML(t *testing.T) {
	var p FPlugin
	assert := assert.New(t)
	assert.Nil(ghodss.Unmarshal([]byte(yamlString), &p))
	assert.Equal(kong.Plugin{
		Name:      p.Name,
		Config:    p.Config,
		Enabled:   p.Enabled,
		RunOn:     p.RunOn,
		Protocols: p.Protocols,
		Service: &kong.Service{
			ID: kong.String("foo"),
		},
		Consumer: &kong.Consumer{
			ID: kong.String("baz"),
		},
		Route: &kong.Route{
			ID: kong.String("bar"),
		},
	}, p.Plugin)
}

func TestPluginUnmarshalJSON(t *testing.T) {
	var p FPlugin
	assert := assert.New(t)
	assert.Nil(json.Unmarshal([]byte(jsonString), &p))
	assert.Equal(kong.Plugin{
		Name:      p.Name,
		Config:    p.Config,
		Enabled:   p.Enabled,
		RunOn:     p.RunOn,
		Protocols: p.Protocols,
		Service: &kong.Service{
			ID: kong.String("foo"),
		},
		Consumer: &kong.Consumer{
			ID: kong.String("baz"),
		},
		Route: &kong.Route{
			ID: kong.String("bar"),
		},
	}, p.Plugin)
}

func Test_unwrapURL(t *testing.T) {
	type args struct {
		urlString string
		fService  *FService
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			args: args{
				urlString: "https://foo.com:8008/bar",
				fService: &FService{
					Service: kong.Service{
						Host:     kong.String("foo.com"),
						Port:     kong.Int(8008),
						Protocol: kong.String("https"),
						Path:     kong.String("/bar"),
					},
				},
			},
			wantErr: false,
		},
		{
			args: args{
				urlString: "https://foo.com/bar",
				fService: &FService{
					Service: kong.Service{
						Host:     kong.String("foo.com"),
						Protocol: kong.String("https"),
						Path:     kong.String("/bar"),
						Port:     kong.Int(443),
					},
				},
			},
			wantErr: false,
		},
		{
			args: args{
				urlString: "https://foo.com:4224/",
				fService: &FService{
					Service: kong.Service{
						Host:     kong.String("foo.com"),
						Protocol: kong.String("https"),
						Path:     kong.String("/"),
						Port:     kong.Int(4224),
					},
				},
			},
			wantErr: false,
		},
		{
			args: args{
				urlString: "https://foo.com/",
				fService: &FService{
					Service: kong.Service{
						Host:     kong.String("foo.com"),
						Protocol: kong.String("https"),
						Path:     kong.String("/"),
						Port:     kong.Int(443),
					},
				},
			},
			wantErr: false,
		},
		{
			args: args{
				urlString: "http://foo.com:4242",
				fService: &FService{
					Service: kong.Service{
						Host:     kong.String("foo.com"),
						Protocol: kong.String("http"),
						Port:     kong.Int(4242),
					},
				},
			},
			wantErr: false,
		},
		{
			args: args{
				urlString: "http://foo.com",
				fService: &FService{
					Service: kong.Service{
						Host:     kong.String("foo.com"),
						Protocol: kong.String("http"),
						Port:     kong.Int(80),
					},
				},
			},
			wantErr: false,
		},
		{
			args: args{
				urlString: "grpc://foocom",
				fService: &FService{
					Service: kong.Service{
						Host:     kong.String("foocom"),
						Protocol: kong.String("grpc"),
						Port:     kong.Int(80),
					},
				},
			},
			wantErr: false,
		},
		{
			args: args{
				urlString: "foo.com/sdf",
				fService: &FService{
					Service: kong.Service{},
				},
			},
			wantErr: true,
		},
		{
			args: args{
				urlString: "foo.com",
				fService: &FService{
					Service: kong.Service{},
				},
			},
			wantErr: true,
		},
		{
			args: args{
				urlString: "42:",
				fService: &FService{
					Service: kong.Service{},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := FService{}
			if err := unwrapURL(tt.args.urlString, &in); (err != nil) != tt.wantErr {
				t.Errorf("unwrapURL() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(tt.args.fService, &in) {
				t.Errorf("unwrapURL() got = %v, want = %v", &in, tt.args.fService)
			}
		})
	}
}
