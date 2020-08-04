package file

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/kong/go-kong/kong"
	"github.com/stretchr/testify/assert"
	yaml "gopkg.in/yaml.v2"
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

func TestPluginUnmarshalYAML(t *testing.T) {
	var p FPlugin
	assert := assert.New(t)
	assert.Nil(yaml.Unmarshal([]byte(yamlString), &p))
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
			fmt.Printf("\n\n%+v", in)
			if !reflect.DeepEqual(tt.args.fService, &in) {
				t.Errorf("unwrapURL() got = %v, want = %v", &in, tt.args.fService)
			}
		})
	}
}
