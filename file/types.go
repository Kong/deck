package file

import "github.com/hbagdi/go-kong/kong"

type service struct {
	kong.Service `yaml:",inline"`
	Routes       []*route
	Plugins      []*plugin
}

type route struct {
	kong.Route `yaml:",inline"`
	Plugins    []*plugin
}

type upstream struct {
	kong.Upstream `yaml:",inline"`
	Targets       []*target
}

type target struct {
	kong.Target `yaml:",inline"`
}

type certificate struct {
	kong.Certificate `yaml:",inline"`
}

type plugin struct {
	kong.Plugin `yaml:",inline"`
}

type consumer struct {
	kong.Consumer `yaml:",inline"`
	Plugins       []*plugin
}

type fileStructure struct {
	Services     []service
	Upstreams    []upstream
	Certificates []certificate
	Plugins      []plugin
	Consumers    []consumer
}
