package dump

import (
	"github.com/hbagdi/go-kong/kong"
	"github.com/hbagdi/go-kong/kong/custom"
)

type KongRawState struct {
	Services       []*kong.Service
	Routes         []*kong.Route
	Plugins        []*kong.Plugin
	Upstreams      []*kong.Upstream
	Targets        []*kong.Target
	Certificates   []*kong.Certificate
	SNIs           []*kong.SNI
	Consumers      []*kong.Consumer
	CustomEntities []*custom.Entity
}
