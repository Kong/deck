package file

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/hbagdi/go-kong/kong"
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

func Test_copyToCert(t *testing.T) {
	type args struct {
		certificate FCertificate
	}
	tests := []struct {
		name string
		args args
		want cert
	}{
		{
			name: "basic sanity",
			args: args{
				certificate: FCertificate{
					Certificate: kong.Certificate{
						Key:  kong.String("key"),
						Cert: kong.String("cert"),
						ID:   kong.String("cert-id"),
						SNIs: kong.StringSlice("0.example.com", "1.example.com"),
						Tags: kong.StringSlice("tag1", "tag2"),
					},
				},
			},
			want: cert{
				Key:  kong.String("key"),
				Cert: kong.String("cert"),
				ID:   kong.String("cert-id"),
				SNIs: []*sni{
					{Name: kong.String("0.example.com")},
					{Name: kong.String("1.example.com")},
				},
				Tags: kong.StringSlice("tag1", "tag2"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := copyToCert(tt.args.certificate); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("copyToCert() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_copyFromCert(t *testing.T) {
	type args struct {
		cert        cert
		certificate *FCertificate
	}
	tests := []struct {
		name string
		args args
		want *FCertificate
	}{
		{
			name: "basic sanity",
			args: args{
				cert: cert{
					Key:  kong.String("key"),
					Cert: kong.String("cert"),
					ID:   kong.String("cert-id"),
					SNIs: []*sni{
						{Name: kong.String("0.example.com")},
						{Name: kong.String("1.example.com")},
					},
					Tags: kong.StringSlice("tag1", "tag2"),
				},
				certificate: &FCertificate{},
			},
			want: &FCertificate{
				Certificate: kong.Certificate{
					Key:  kong.String("key"),
					Cert: kong.String("cert"),
					ID:   kong.String("cert-id"),
					SNIs: kong.StringSlice("0.example.com", "1.example.com"),
					Tags: kong.StringSlice("tag1", "tag2"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			copyFromCert(tt.args.cert, tt.args.certificate)

			if !reflect.DeepEqual(tt.args.certificate, tt.want) {
				t.Errorf("copyFromCert() = %v, want %v", tt.args.certificate, tt.want)
			}
		})
	}
}
