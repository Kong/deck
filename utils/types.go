package utils

import (
	"github.com/hbagdi/go-kong/kong"
	"github.com/hbagdi/go-kong/kong/custom"
)

// KongRawState contains all of Kong Data
type KongRawState struct {
	Services []*kong.Service
	Routes   []*kong.Route

	Plugins []*kong.Plugin
	// TODO add plugin schema

	Upstreams []*kong.Upstream
	Targets   []*kong.Target

	Certificates []*kong.Certificate
	SNIs         []*kong.SNI

	Consumers      []*kong.Consumer
	CustomEntities []*custom.Entity
}
