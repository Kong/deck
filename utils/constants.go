package utils

import (
	"github.com/kong/go-kong/kong"
)

const (
	defaultTimeout     = 60000
	defaultSlots       = 10000
	defaultWeight      = 100
	defaultConcurrency = 10
)

var (
	serviceDefaults = kong.Service{
		Protocol:       kong.String("http"),
		ConnectTimeout: kong.Int(defaultTimeout),
		WriteTimeout:   kong.Int(defaultTimeout),
		ReadTimeout:    kong.Int(defaultTimeout),
	}
	routeDefaults = kong.Route{
		PreserveHost:  kong.Bool(false),
		RegexPriority: kong.Int(0),
		StripPath:     kong.Bool(true),
		Protocols:     kong.StringSlice("http", "https"),
	}
	targetDefaults = kong.Target{
		Weight: kong.Int(defaultWeight),
	}
	upstreamDefaults = kong.Upstream{
		Slots: kong.Int(defaultSlots),
		Healthchecks: &kong.Healthcheck{
			Active: &kong.ActiveHealthcheck{
				Concurrency: kong.Int(defaultConcurrency),
				Healthy: &kong.Healthy{
					HTTPStatuses: []int{200, 302},
					Interval:     kong.Int(0),
					Successes:    kong.Int(0),
				},
				HTTPPath: kong.String("/"),
				Type:     kong.String("http"),
				Timeout:  kong.Int(1),
				Unhealthy: &kong.Unhealthy{
					HTTPFailures: kong.Int(0),
					TCPFailures:  kong.Int(0),
					Timeouts:     kong.Int(0),
					Interval:     kong.Int(0),
					HTTPStatuses: []int{429, 404, 500, 501, 502, 503, 504, 505},
				},
			},
			Passive: &kong.PassiveHealthcheck{
				Healthy: &kong.Healthy{
					HTTPStatuses: []int{
						200, 201, 202, 203, 204, 205,
						206, 207, 208, 226, 300, 301, 302, 303, 304, 305,
						306, 307, 308,
					},
					Successes: kong.Int(0),
				},
				Unhealthy: &kong.Unhealthy{
					HTTPFailures: kong.Int(0),
					TCPFailures:  kong.Int(0),
					Timeouts:     kong.Int(0),
					HTTPStatuses: []int{429, 500, 503},
				},
			},
		},
		HashOn:           kong.String("none"),
		HashFallback:     kong.String("none"),
		HashOnCookiePath: kong.String("/"),
	}
	consumerGroupPluginDefault = kong.ConsumerGroupPlugin{
		Config: kong.Configuration{
			"window_type": "sliding",
		},
	}
	defaultsRestrictedFields = map[string][]string{
		"Service":  {"ID", "Name"},
		"Route":    {"ID", "Name"},
		"Target":   {"ID", "Target"},
		"Upstream": {"ID", "Name"},
	}
)

const (
	// ImplementationTypeKongGateway indicates an implementation backed by Kong Gateway.
	ImplementationTypeKongGateway = "kong-gateway"
)
