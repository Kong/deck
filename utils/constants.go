package utils

import "github.com/hbagdi/go-kong/kong"

var (
	serviceDefaults = kong.Service{
		Port:           kong.Int(80),
		Retries:        kong.Int(5),
		Protocol:       kong.String("http"),
		Path:           kong.String("/"),
		ConnectTimeout: kong.Int(60000),
		WriteTimeout:   kong.Int(60000),
		ReadTimeout:    kong.Int(60000),
	}
	routeDefaults = kong.Route{
		PreserveHost:  kong.Bool(false),
		RegexPriority: kong.Int(0),
		StripPath:     kong.Bool(true),
		Protocols:     kong.StringSlice("http", "https"),
	}
	targetDefaults = kong.Target{
		Weight: kong.Int(100),
	}
	upstreamDefaults = kong.Upstream{
		Slots: kong.Int(10000),
		Healthchecks: &kong.Healthcheck{
			Active: &kong.ActiveHealthcheck{
				Concurrency: kong.Int(10),
				Healthy: &kong.Healthy{
					HTTPStatuses: []int{200, 302},
					Interval:     kong.Int(0),
					Successes:    kong.Int(0),
				},
				HTTPPath: kong.String("/"),
				Timeout:  kong.Int(1),
				Unhealthy: &kong.Unhealthy{
					HTTPFailures: kong.Int(0),
					TCPFailures:  kong.Int(0),
					Timeouts:     kong.Int(0),
					HTTPStatuses: []int{429, 404, 500, 501, 502, 503, 504, 505},
				},
			},
			Passive: &kong.PassiveHealthcheck{
				Healthy: &kong.Healthy{
					HTTPStatuses: []int{200, 201, 202, 203, 204, 205,
						206, 207, 208, 226, 300, 301, 302, 303, 304, 305,
						306, 307, 308},
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
)
