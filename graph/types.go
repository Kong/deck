package graph

import "github.com/hbagdi/go-kong/kong"

// KongGraphState can store a graph of Kong Entities
type KongGraphState struct {
	service []Service
	routes  []Route
	plugins []Plugin
	//upstreams    []Upstream
	//targets      []Target
	//certificates []Certificate
	//snis         []SNI
	//consumers    []Consumer
}

// Meta contains additional information for an entity
type Meta struct {
	Name   *string
	Global *bool
	Type   *string
}

type Service struct {
	kong.Service
	Meta          Meta
	Routes        []*Route
	PluginDetails []*Plugin
	Plugins       []*string
}

type Route struct {
	kong.Route
	Service       Service // back reference
	Meta          Meta
	PluginDetails []*Plugin
	Plugins       []*string
}

type Plugin struct {
	kong.Plugin
	Meta    Meta
	Service Service
	Route   Route
}
