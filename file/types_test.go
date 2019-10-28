package file

import (
	"encoding/json"
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
	var p Plugin
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
	var p Plugin
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
