package file

import "github.com/hbagdi/go-kong/kong"

type service struct {
	kong.Service `yaml:",inline,omitempty"`
	Routes       []*route  `yaml:",omitempty"`
	Plugins      []*plugin `yaml:",omitempty"`
}

type route struct {
	kong.Route `yaml:",inline,omitempty"`
	Plugins    []*plugin `yaml:",omitempty"`
}

type upstream struct {
	kong.Upstream `yaml:",inline,omitempty"`
	Targets       []*target `yaml:",omitempty"`
}

type target struct {
	kong.Target `yaml:",inline,omitempty"`
}

type certificate struct {
	kong.Certificate `yaml:",inline,omitempty"`
}

type plugin struct {
	kong.Plugin `yaml:",inline,omitempty"`
}

type consumer struct {
	kong.Consumer `yaml:",inline,omitempty"`
	Plugins       []*plugin `yaml:",omitempty"`
}

type fileStructure struct {
	Services     []service     `yaml:",omitempty"`
	Upstreams    []upstream    `yaml:",omitempty"`
	Certificates []certificate `yaml:",omitempty"`
	Plugins      []plugin      `yaml:",omitempty"`
	Consumers    []consumer    `yaml:",omitempty"`
}
