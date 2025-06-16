//go:build integration

package integration

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/fatih/color"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	deckDump "github.com/kong/go-database-reconciler/pkg/dump"
	"github.com/kong/go-database-reconciler/pkg/utils"
	"github.com/kong/go-kong/kong"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	// missing Enable
	svc1 = []*kong.Service{
		{
			ID:             kong.String("58076db2-28b6-423b-ba39-a797193017f7"),
			Name:           kong.String("svc1"),
			ConnectTimeout: kong.Int(60000),
			Host:           kong.String("mockbin.org"),
			Port:           kong.Int(80),
			Protocol:       kong.String("http"),
			ReadTimeout:    kong.Int(60000),
			Retries:        kong.Int(5),
			WriteTimeout:   kong.Int(60000),
			Tags:           nil,
		},
	}

	// latest
	svc1_207 = []*kong.Service{
		{
			ID:             kong.String("58076db2-28b6-423b-ba39-a797193017f7"),
			Name:           kong.String("svc1"),
			ConnectTimeout: kong.Int(60000),
			Host:           kong.String("mockbin.org"),
			Port:           kong.Int(80),
			Protocol:       kong.String("http"),
			ReadTimeout:    kong.Int(60000),
			Retries:        kong.Int(5),
			WriteTimeout:   kong.Int(60000),
			Enabled:        kong.Bool(true),
			Tags:           nil,
		},
	}

	defaultCPService = []*kong.Service{
		{
			ID:             kong.String("58076db2-28b6-423b-ba39-a797193017f7"),
			Name:           kong.String("default"),
			ConnectTimeout: kong.Int(60000),
			Host:           kong.String("mockbin-default.org"),
			Port:           kong.Int(80),
			Protocol:       kong.String("http"),
			ReadTimeout:    kong.Int(60000),
			Retries:        kong.Int(5),
			WriteTimeout:   kong.Int(60000),
			Enabled:        kong.Bool(true),
			Tags:           nil,
		},
	}

	testCPService = []*kong.Service{
		{
			ID:             kong.String("58076db2-28b6-423b-ba39-a797193017f7"),
			Name:           kong.String("test"),
			ConnectTimeout: kong.Int(60000),
			Host:           kong.String("mockbin-test.org"),
			Port:           kong.Int(80),
			Protocol:       kong.String("http"),
			ReadTimeout:    kong.Int(60000),
			Retries:        kong.Int(5),
			WriteTimeout:   kong.Int(60000),
			Enabled:        kong.Bool(true),
			Tags:           nil,
		},
	}

	// missing RequestBuffering, ResponseBuffering, Service, PathHandling
	route1_143 = []*kong.Route{
		{
			ID:                      kong.String("87b6a97e-f3f7-4c47-857a-7464cb9e202b"),
			Name:                    kong.String("r1"),
			Paths:                   []*string{kong.String("/r1")},
			PreserveHost:            kong.Bool(false),
			Protocols:               []*string{kong.String("http"), kong.String("https")},
			RegexPriority:           kong.Int(0),
			StripPath:               kong.Bool(true),
			HTTPSRedirectStatusCode: kong.Int(301),
		},
	}

	// missing RequestBuffering, ResponseBuffering
	// PathHandling set to v1
	route1_151 = []*kong.Route{
		{
			ID:                      kong.String("87b6a97e-f3f7-4c47-857a-7464cb9e202b"),
			Name:                    kong.String("r1"),
			Paths:                   []*string{kong.String("/r1")},
			PathHandling:            kong.String("v1"),
			PreserveHost:            kong.Bool(false),
			Protocols:               []*string{kong.String("http"), kong.String("https")},
			RegexPriority:           kong.Int(0),
			StripPath:               kong.Bool(true),
			HTTPSRedirectStatusCode: kong.Int(301),
			Service: &kong.Service{
				ID: kong.String("58076db2-28b6-423b-ba39-a797193017f7"),
			},
		},
	}

	// missing RequestBuffering, ResponseBuffering
	route1_205_214 = []*kong.Route{
		{
			ID:                      kong.String("87b6a97e-f3f7-4c47-857a-7464cb9e202b"),
			Name:                    kong.String("r1"),
			Paths:                   []*string{kong.String("/r1")},
			PathHandling:            kong.String("v0"),
			PreserveHost:            kong.Bool(false),
			Protocols:               []*string{kong.String("http"), kong.String("https")},
			RegexPriority:           kong.Int(0),
			StripPath:               kong.Bool(true),
			HTTPSRedirectStatusCode: kong.Int(301),
			Service: &kong.Service{
				ID: kong.String("58076db2-28b6-423b-ba39-a797193017f7"),
			},
		},
	}

	// latest
	route1_20x = []*kong.Route{
		{
			ID:                      kong.String("87b6a97e-f3f7-4c47-857a-7464cb9e202b"),
			Name:                    kong.String("r1"),
			Paths:                   []*string{kong.String("/r1")},
			PathHandling:            kong.String("v0"),
			PreserveHost:            kong.Bool(false),
			Protocols:               []*string{kong.String("http"), kong.String("https")},
			RegexPriority:           kong.Int(0),
			StripPath:               kong.Bool(true),
			HTTPSRedirectStatusCode: kong.Int(301),
			RequestBuffering:        kong.Bool(true),
			ResponseBuffering:       kong.Bool(true),
			Service: &kong.Service{
				ID: kong.String("58076db2-28b6-423b-ba39-a797193017f7"),
			},
		},
	}

	// has run-on set to 'first'
	plugin_143_151 = []*kong.Plugin{ //nolint:revive,stylecheck
		{
			Name: kong.String("basic-auth"),
			Protocols: []*string{
				kong.String("grpc"),
				kong.String("grpcs"),
				kong.String("http"),
				kong.String("https"),
			},
			Enabled: kong.Bool(true),
			Config: kong.Configuration{
				"anonymous":        "58076db2-28b6-423b-ba39-a797193017f7",
				"hide_credentials": false,
			},
			RunOn: kong.String("first"),
		},
	}

	// latest
	plugin = []*kong.Plugin{
		{
			Name: kong.String("basic-auth"),
			Protocols: []*string{
				kong.String("grpc"),
				kong.String("grpcs"),
				kong.String("http"),
				kong.String("https"),
			},
			Enabled: kong.Bool(true),
			Config: kong.Configuration{
				"anonymous":        "58076db2-28b6-423b-ba39-a797193017f7",
				"hide_credentials": false,
			},
		},
	}

	plugin36 = []*kong.Plugin{
		{
			Name: kong.String("basic-auth"),
			Protocols: []*string{
				kong.String("grpc"),
				kong.String("grpcs"),
				kong.String("http"),
				kong.String("https"),
			},
			Enabled: kong.Bool(true),
			Config: kong.Configuration{
				"anonymous":        "58076db2-28b6-423b-ba39-a797193017f7",
				"hide_credentials": false,
				"realm":            string("service"),
			},
		},
	}

	plugin_on_entities = []*kong.Plugin{ //nolint:revive,stylecheck
		{
			Name: kong.String("prometheus"),
			Protocols: []*string{
				kong.String("grpc"),
				kong.String("grpcs"),
				kong.String("http"),
				kong.String("https"),
			},
			Enabled: kong.Bool(true),
			Config: kong.Configuration{
				"per_consumer": false,
			},
			Service: &kong.Service{
				ID: kong.String("58076db2-28b6-423b-ba39-a797193017f7"),
			},
		},
		{
			Name: kong.String("prometheus"),
			Protocols: []*string{
				kong.String("grpc"),
				kong.String("grpcs"),
				kong.String("http"),
				kong.String("https"),
			},
			Enabled: kong.Bool(true),
			Config: kong.Configuration{
				"per_consumer": false,
			},
			Route: &kong.Route{
				ID: kong.String("87b6a97e-f3f7-4c47-857a-7464cb9e202b"),
			},
		},
		{
			Name: kong.String("prometheus"),
			Protocols: []*string{
				kong.String("grpc"),
				kong.String("grpcs"),
				kong.String("http"),
				kong.String("https"),
			},
			Enabled: kong.Bool(true),
			Config: kong.Configuration{
				"per_consumer": false,
			},
			Consumer: &kong.Consumer{
				ID: kong.String("d2965b9b-0608-4458-a9f8-0b93d88d03b8"),
			},
		},
	}

	plugin_on_entities3x = []*kong.Plugin{ //nolint:revive,stylecheck
		{
			Name: kong.String("prometheus"),
			Protocols: []*string{
				kong.String("grpc"),
				kong.String("grpcs"),
				kong.String("http"),
				kong.String("https"),
			},
			Enabled: kong.Bool(true),
			Config: kong.Configuration{
				"bandwidth_metrics":       false,
				"latency_metrics":         false,
				"per_consumer":            false,
				"status_code_metrics":     false,
				"upstream_health_metrics": false,
			},
			Service: &kong.Service{
				ID: kong.String("58076db2-28b6-423b-ba39-a797193017f7"),
			},
		},
		{
			Name: kong.String("prometheus"),
			Protocols: []*string{
				kong.String("grpc"),
				kong.String("grpcs"),
				kong.String("http"),
				kong.String("https"),
			},
			Enabled: kong.Bool(true),
			Config: kong.Configuration{
				"bandwidth_metrics":       false,
				"latency_metrics":         false,
				"per_consumer":            false,
				"status_code_metrics":     false,
				"upstream_health_metrics": false,
			},
			Route: &kong.Route{
				ID: kong.String("87b6a97e-f3f7-4c47-857a-7464cb9e202b"),
			},
		},
		{
			Name: kong.String("prometheus"),
			Protocols: []*string{
				kong.String("grpc"),
				kong.String("grpcs"),
				kong.String("http"),
				kong.String("https"),
			},
			Enabled: kong.Bool(true),
			Config: kong.Configuration{
				"bandwidth_metrics":       false,
				"latency_metrics":         false,
				"per_consumer":            false,
				"status_code_metrics":     false,
				"upstream_health_metrics": false,
			},
			Consumer: &kong.Consumer{
				ID: kong.String("d2965b9b-0608-4458-a9f8-0b93d88d03b8"),
			},
		},
	}

	plugin_on_entities381x = []*kong.Plugin{ //nolint:revive,stylecheck
		{
			Name: kong.String("prometheus"),
			Protocols: []*string{
				kong.String("grpc"),
				kong.String("grpcs"),
				kong.String("http"),
				kong.String("https"),
			},
			Enabled: kong.Bool(true),
			Config: kong.Configuration{
				"ai_metrics":              false,
				"bandwidth_metrics":       false,
				"latency_metrics":         false,
				"per_consumer":            false,
				"status_code_metrics":     false,
				"upstream_health_metrics": false,
			},
			Service: &kong.Service{
				ID: kong.String("58076db2-28b6-423b-ba39-a797193017f7"),
			},
		},
		{
			Name: kong.String("prometheus"),
			Protocols: []*string{
				kong.String("grpc"),
				kong.String("grpcs"),
				kong.String("http"),
				kong.String("https"),
			},
			Enabled: kong.Bool(true),
			Config: kong.Configuration{
				"ai_metrics":              false,
				"bandwidth_metrics":       false,
				"latency_metrics":         false,
				"per_consumer":            false,
				"status_code_metrics":     false,
				"upstream_health_metrics": false,
			},
			Route: &kong.Route{
				ID: kong.String("87b6a97e-f3f7-4c47-857a-7464cb9e202b"),
			},
		},
		{
			Name: kong.String("prometheus"),
			Protocols: []*string{
				kong.String("grpc"),
				kong.String("grpcs"),
				kong.String("http"),
				kong.String("https"),
			},
			Enabled: kong.Bool(true),
			Config: kong.Configuration{
				"ai_metrics":              false,
				"bandwidth_metrics":       false,
				"latency_metrics":         false,
				"per_consumer":            false,
				"status_code_metrics":     false,
				"upstream_health_metrics": false,
			},
			Consumer: &kong.Consumer{
				ID: kong.String("d2965b9b-0608-4458-a9f8-0b93d88d03b8"),
			},
		},
	}

	plugin_on_entities310x = []*kong.Plugin{ //nolint:revive,stylecheck
		{
			Name: kong.String("prometheus"),
			Protocols: []*string{
				kong.String("grpc"),
				kong.String("grpcs"),
				kong.String("http"),
				kong.String("https"),
			},
			Enabled: kong.Bool(true),
			Config: kong.Configuration{
				"ai_metrics":              false,
				"bandwidth_metrics":       false,
				"latency_metrics":         false,
				"per_consumer":            false,
				"status_code_metrics":     false,
				"upstream_health_metrics": false,
				"wasm_metrics":            false,
			},
			Service: &kong.Service{
				ID: kong.String("58076db2-28b6-423b-ba39-a797193017f7"),
			},
		},
		{
			Name: kong.String("prometheus"),
			Protocols: []*string{
				kong.String("grpc"),
				kong.String("grpcs"),
				kong.String("http"),
				kong.String("https"),
			},
			Enabled: kong.Bool(true),
			Config: kong.Configuration{
				"ai_metrics":              false,
				"bandwidth_metrics":       false,
				"latency_metrics":         false,
				"per_consumer":            false,
				"status_code_metrics":     false,
				"upstream_health_metrics": false,
				"wasm_metrics":            false,
			},
			Route: &kong.Route{
				ID: kong.String("87b6a97e-f3f7-4c47-857a-7464cb9e202b"),
			},
		},
		{
			Name: kong.String("prometheus"),
			Protocols: []*string{
				kong.String("grpc"),
				kong.String("grpcs"),
				kong.String("http"),
				kong.String("https"),
			},
			Enabled: kong.Bool(true),
			Config: kong.Configuration{
				"ai_metrics":              false,
				"bandwidth_metrics":       false,
				"latency_metrics":         false,
				"per_consumer":            false,
				"status_code_metrics":     false,
				"upstream_health_metrics": false,
				"wasm_metrics":            false,
			},
			Consumer: &kong.Consumer{
				ID: kong.String("d2965b9b-0608-4458-a9f8-0b93d88d03b8"),
			},
		},
	}

	plugin_on_entitiesKonnect = []*kong.Plugin{ //nolint:revive,stylecheck
		{
			Name: kong.String("prometheus"),
			Protocols: []*string{
				kong.String("grpc"),
				kong.String("grpcs"),
				kong.String("http"),
				kong.String("https"),
			},
			Enabled: kong.Bool(true),
			Config: kong.Configuration{
				"ai_metrics":              false,
				"bandwidth_metrics":       false,
				"latency_metrics":         false,
				"per_consumer":            false,
				"status_code_metrics":     false,
				"upstream_health_metrics": false,
				"wasm_metrics":            false,
			},
			Service: &kong.Service{
				ID: kong.String("58076db2-28b6-423b-ba39-a797193017f7"),
			},
		},
		{
			Name: kong.String("prometheus"),
			Protocols: []*string{
				kong.String("grpc"),
				kong.String("grpcs"),
				kong.String("http"),
				kong.String("https"),
			},
			Enabled: kong.Bool(true),
			Config: kong.Configuration{
				"ai_metrics":              false,
				"bandwidth_metrics":       false,
				"latency_metrics":         false,
				"per_consumer":            false,
				"status_code_metrics":     false,
				"upstream_health_metrics": false,
				"wasm_metrics":            false,
			},
			Route: &kong.Route{
				ID: kong.String("87b6a97e-f3f7-4c47-857a-7464cb9e202b"),
			},
		},
		{
			Name: kong.String("prometheus"),
			Protocols: []*string{
				kong.String("grpc"),
				kong.String("grpcs"),
				kong.String("http"),
				kong.String("https"),
			},
			Enabled: kong.Bool(true),
			Config: kong.Configuration{
				"ai_metrics":              false,
				"bandwidth_metrics":       false,
				"latency_metrics":         false,
				"per_consumer":            false,
				"status_code_metrics":     false,
				"upstream_health_metrics": false,
				"wasm_metrics":            false,
			},
			Consumer: &kong.Consumer{
				ID: kong.String("d2965b9b-0608-4458-a9f8-0b93d88d03b8"),
			},
		},
	}

	upstream_pre31 = []*kong.Upstream{ //nolint:revive,stylecheck
		{
			Name:      kong.String("upstream1"),
			Algorithm: kong.String("round-robin"),
			Slots:     kong.Int(10000),
			Healthchecks: &kong.Healthcheck{
				Threshold: kong.Float64(0),
				Active: &kong.ActiveHealthcheck{
					Concurrency: kong.Int(10),
					Healthy: &kong.Healthy{
						HTTPStatuses: []int{200, 302},
						Interval:     kong.Int(0),
						Successes:    kong.Int(0),
					},
					HTTPPath:               kong.String("/"),
					Type:                   kong.String("http"),
					Timeout:                kong.Int(1),
					HTTPSVerifyCertificate: kong.Bool(true),
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
					Type: kong.String("http"),
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
		},
	}

	// latest
	upstream = []*kong.Upstream{
		{
			Name:      kong.String("upstream1"),
			Algorithm: kong.String("round-robin"),
			Slots:     kong.Int(10000),
			Healthchecks: &kong.Healthcheck{
				Threshold: kong.Float64(0),
				Active: &kong.ActiveHealthcheck{
					Concurrency: kong.Int(10),
					Healthy: &kong.Healthy{
						HTTPStatuses: []int{200, 302},
						Interval:     kong.Int(0),
						Successes:    kong.Int(0),
					},
					HTTPPath:               kong.String("/"),
					Type:                   kong.String("http"),
					Timeout:                kong.Int(1),
					HTTPSVerifyCertificate: kong.Bool(true),
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
					Type: kong.String("http"),
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
			UseSrvName:       kong.Bool(false),
		},
	}

	target = []*kong.Target{
		{
			Target: kong.String("198.51.100.11:80"),
			Upstream: &kong.Upstream{
				ID: kong.String("a6f89ffc-1e53-4b01-9d3d-7a142bcd"),
			},
			Weight: kong.Int(100),
		},
	}

	targetZeroWeight = []*kong.Target{
		{
			Target: kong.String("198.51.100.11:80"),
			Upstream: &kong.Upstream{
				ID: kong.String("a6f89ffc-1e53-4b01-9d3d-7a142bcd"),
			},
			Weight: kong.Int(0),
		},
	}

	rateLimitingPlugin = []*kong.Plugin{
		{
			Name: kong.String("rate-limiting"),
			Config: kong.Configuration{
				"day":                 nil,
				"fault_tolerant":      true,
				"header_name":         nil,
				"hide_client_headers": false,
				"hour":                nil,
				"limit_by":            "consumer",
				"minute":              float64(123),
				"month":               nil,
				"path":                nil,
				"policy":              "cluster",
				"redis_database":      float64(0),
				"redis_host":          nil,
				"redis_password":      nil,
				"redis_port":          float64(6379),
				"redis_server_name":   nil,
				"redis_ssl":           false,
				"redis_ssl_verify":    false,
				"redis_timeout":       float64(2000),
				"second":              nil,
				"year":                nil,
			},
			Enabled: kong.Bool(true),
			RunOn:   nil,
			Protocols: []*string{
				kong.String("grpc"),
				kong.String("grpcs"),
				kong.String("http"),
				kong.String("https"),
			},
			Tags: nil,
		},
	}

	consumer = []*kong.Consumer{
		{
			Username: kong.String("yolo"),
			ID:       kong.String("d2965b9b-0608-4458-a9f8-0b93d88d03b8"),
		},
	}

	consumerGroupsConsumers = []*kong.Consumer{
		{
			Username: kong.String("foo"),
		},
		{
			Username: kong.String("bar"),
		},
		{
			Username: kong.String("baz"),
		},
	}

	consumerGroups = []*kong.ConsumerGroupObject{
		{
			ConsumerGroup: &kong.ConsumerGroup{
				Name: kong.String("silver"),
			},
			Consumers: []*kong.Consumer{
				{
					Username: kong.String("bar"),
				},
				{
					Username: kong.String("baz"),
				},
			},
		},
		{
			ConsumerGroup: &kong.ConsumerGroup{
				Name: kong.String("gold"),
			},
			Consumers: []*kong.Consumer{
				{
					Username: kong.String("foo"),
				},
			},
		},
	}

	consumerGroupsWithTags = []*kong.ConsumerGroupObject{
		{
			ConsumerGroup: &kong.ConsumerGroup{
				Name: kong.String("silver"),
				Tags: kong.StringSlice("tag1", "tag3"),
			},
			Consumers: []*kong.Consumer{
				{
					Username: kong.String("bar"),
				},
				{
					Username: kong.String("baz"),
				},
			},
		},
		{
			ConsumerGroup: &kong.ConsumerGroup{
				Name: kong.String("gold"),
				Tags: kong.StringSlice("tag1", "tag2"),
			},
			Consumers: []*kong.Consumer{
				{
					Username: kong.String("foo"),
				},
			},
		},
	}

	consumerGroupsWithRLA = []*kong.ConsumerGroupObject{
		{
			ConsumerGroup: &kong.ConsumerGroup{
				Name: kong.String("silver"),
			},
			Consumers: []*kong.Consumer{
				{
					Username: kong.String("bar"),
				},
			},
			Plugins: []*kong.ConsumerGroupPlugin{
				{
					Name: kong.String("rate-limiting-advanced"),
					Config: kong.Configuration{
						"limit":                  []any{float64(7)},
						"retry_after_jitter_max": float64(1),
						"window_size":            []any{float64(60)},
						"window_type":            "sliding",
					},
					ConsumerGroup: &kong.ConsumerGroup{
						ID: kong.String("521a90ad-36cb-4e31-a5db-1d979aee40d1"),
					},
				},
			},
		},
		{
			ConsumerGroup: &kong.ConsumerGroup{
				Name: kong.String("gold"),
			},
			Consumers: []*kong.Consumer{
				{
					Username: kong.String("foo"),
				},
			},
			Plugins: []*kong.ConsumerGroupPlugin{
				{
					Name: kong.String("rate-limiting-advanced"),
					Config: kong.Configuration{
						"limit":                  []any{float64(10)},
						"retry_after_jitter_max": float64(1),
						"window_size":            []any{float64(60)},
						"window_type":            "sliding",
					},
					ConsumerGroup: &kong.ConsumerGroup{
						ID: kong.String("92177268-b134-42f9-909a-36f9d2d3d5e7"),
					},
				},
			},
		},
	}

	consumerGroupsWithTagsAndRLA = []*kong.ConsumerGroupObject{
		{
			ConsumerGroup: &kong.ConsumerGroup{
				Name: kong.String("silver"),
				Tags: kong.StringSlice("tag1", "tag3"),
			},
			Consumers: []*kong.Consumer{
				{
					Username: kong.String("bar"),
				},
			},
			Plugins: []*kong.ConsumerGroupPlugin{
				{
					Name: kong.String("rate-limiting-advanced"),
					Config: kong.Configuration{
						"limit":                  []any{float64(7)},
						"retry_after_jitter_max": float64(1),
						"window_size":            []any{float64(60)},
						"window_type":            "sliding",
					},
					ConsumerGroup: &kong.ConsumerGroup{
						ID: kong.String("521a90ad-36cb-4e31-a5db-1d979aee40d1"),
					},
				},
			},
		},
		{
			ConsumerGroup: &kong.ConsumerGroup{
				Name: kong.String("gold"),
				Tags: kong.StringSlice("tag1", "tag2"),
			},
			Consumers: []*kong.Consumer{
				{
					Username: kong.String("foo"),
				},
			},
			Plugins: []*kong.ConsumerGroupPlugin{
				{
					Name: kong.String("rate-limiting-advanced"),
					Config: kong.Configuration{
						"limit":                  []any{float64(10)},
						"retry_after_jitter_max": float64(1),
						"window_size":            []any{float64(60)},
						"window_type":            "sliding",
					},
					ConsumerGroup: &kong.ConsumerGroup{
						ID: kong.String("92177268-b134-42f9-909a-36f9d2d3d5e7"),
					},
				},
			},
		},
	}

	consumerGroupsWithRLAApp = []*kong.ConsumerGroupObject{
		{
			ConsumerGroup: &kong.ConsumerGroup{
				Name: kong.String("silver"),
			},
			Consumers: []*kong.Consumer{
				{
					Username: kong.String("bar"),
				},
			},
			Plugins: []*kong.ConsumerGroupPlugin{
				{
					Name: kong.String("rate-limiting-advanced"),
					Config: kong.Configuration{
						"limit":                  []any{float64(7)},
						"retry_after_jitter_max": float64(1),
						"window_size":            []any{float64(60)},
						"window_type":            string("sliding"),
					},
					ConsumerGroup: &kong.ConsumerGroup{
						ID: kong.String("f79972fe-e9a0-40b5-8dc6-f1bf3758b86b"),
					},
				},
			},
		},
		{
			ConsumerGroup: &kong.ConsumerGroup{
				Name: kong.String("gold"),
			},
			Consumers: []*kong.Consumer{
				{
					Username: kong.String("foo"),
				},
			},
			Plugins: []*kong.ConsumerGroupPlugin{
				{
					Name: kong.String("rate-limiting-advanced"),
					Config: kong.Configuration{
						"limit":                  []any{float64(10)},
						"retry_after_jitter_max": float64(1),
						"window_size":            []any{float64(60)},
						"window_type":            string("sliding"),
					},
					ConsumerGroup: &kong.ConsumerGroup{
						ID: kong.String("8eea863e-460c-4019-895a-1e80cb08699d"),
					},
				},
			},
		},
	}

	consumerGroupAppPlugins = []*kong.Plugin{
		{
			Name: kong.String("rate-limiting-advanced"),
			Config: kong.Configuration{
				"consumer_groups":         []any{string("silver"), string("gold")},
				"dictionary_name":         string("kong_rate_limiting_counters"),
				"enforce_consumer_groups": bool(true),
				"header_name":             nil,
				"hide_client_headers":     bool(false),
				"identifier":              string("consumer"),
				"limit":                   []any{float64(5)},
				"namespace":               string("dNRC6xKsRL8Koc1uVYA4Nki6DLW7XIdx"),
				"path":                    nil,
				"redis": map[string]any{
					"cluster_addresses":   nil,
					"connect_timeout":     nil,
					"database":            float64(0),
					"host":                nil,
					"keepalive_backlog":   nil,
					"keepalive_pool_size": float64(30),
					"password":            nil,
					"port":                nil,
					"read_timeout":        nil,
					"send_timeout":        nil,
					"sentinel_addresses":  nil,
					"sentinel_master":     nil,
					"sentinel_password":   nil,
					"sentinel_role":       nil,
					"sentinel_username":   nil,
					"server_name":         nil,
					"ssl":                 false,
					"ssl_verify":          false,
					"timeout":             float64(2000),
					"username":            nil,
				},
				"retry_after_jitter_max": float64(0),
				"strategy":               string("local"),
				"sync_rate":              float64(-1),
				"window_size":            []any{float64(60)},
				"window_type":            string("sliding"),
			},
			Enabled:   kong.Bool(true),
			Protocols: []*string{kong.String("grpc"), kong.String("grpcs"), kong.String("http"), kong.String("https")},
		},
		{
			Name: kong.String("key-auth"),
			Config: kong.Configuration{
				"anonymous":        nil,
				"hide_credentials": false,
				"key_in_body":      false,
				"key_in_header":    true,
				"key_in_query":     true,
				"key_names":        []interface{}{"apikey"},
				"run_on_preflight": true,
			},
			Enabled:   kong.Bool(true),
			Protocols: []*string{kong.String("http"), kong.String("https")},
		},
	}

	consumerGroupScopedPlugins = []*kong.Plugin{
		{
			Name: kong.String("rate-limiting-advanced"),
			ConsumerGroup: &kong.ConsumerGroup{
				ID: kong.String("77e6691d-67c0-446a-9401-27be2b141aae"),
			},
			Config: kong.Configuration{
				"consumer_groups":         nil,
				"dictionary_name":         string("kong_rate_limiting_counters"),
				"disable_penalty":         bool(false),
				"enforce_consumer_groups": bool(false),
				"error_code":              float64(429),
				"error_message":           string("API rate limit exceeded"),
				"header_name":             nil,
				"hide_client_headers":     bool(false),
				"identifier":              string("consumer"),
				"limit":                   []any{float64(10)},
				"namespace":               string("gold"),
				"path":                    nil,
				"redis": map[string]any{
					"cluster_addresses":   nil,
					"connect_timeout":     nil,
					"database":            float64(0),
					"host":                nil,
					"keepalive_backlog":   nil,
					"keepalive_pool_size": float64(30),
					"password":            nil,
					"port":                nil,
					"read_timeout":        nil,
					"send_timeout":        nil,
					"sentinel_addresses":  nil,
					"sentinel_master":     nil,
					"sentinel_password":   nil,
					"sentinel_role":       nil,
					"sentinel_username":   nil,
					"server_name":         nil,
					"ssl":                 false,
					"ssl_verify":          false,
					"timeout":             float64(2000),
					"username":            nil,
				},
				"retry_after_jitter_max": float64(1),
				"strategy":               string("local"),
				"sync_rate":              float64(-1),
				"window_size":            []any{float64(60)},
				"window_type":            string("sliding"),
			},
			Enabled:   kong.Bool(true),
			Protocols: []*string{kong.String("grpc"), kong.String("grpcs"), kong.String("http"), kong.String("https")},
		},
		{
			Name: kong.String("rate-limiting-advanced"),
			ConsumerGroup: &kong.ConsumerGroup{
				ID: kong.String("5bcbd3a7-030b-4310-bd1d-2721ff85d236"),
			},
			Config: kong.Configuration{
				"consumer_groups":         nil,
				"dictionary_name":         string("kong_rate_limiting_counters"),
				"disable_penalty":         bool(false),
				"enforce_consumer_groups": bool(false),
				"error_code":              float64(429),
				"error_message":           string("API rate limit exceeded"),
				"header_name":             nil,
				"hide_client_headers":     bool(false),
				"identifier":              string("consumer"),
				"limit":                   []any{float64(7)},
				"namespace":               string("silver"),
				"path":                    nil,
				"redis": map[string]any{
					"cluster_addresses":   nil,
					"connect_timeout":     nil,
					"database":            float64(0),
					"host":                nil,
					"keepalive_backlog":   nil,
					"keepalive_pool_size": float64(30),
					"password":            nil,
					"port":                nil,
					"read_timeout":        nil,
					"send_timeout":        nil,
					"sentinel_addresses":  nil,
					"sentinel_master":     nil,
					"sentinel_password":   nil,
					"sentinel_role":       nil,
					"sentinel_username":   nil,
					"server_name":         nil,
					"ssl":                 false,
					"ssl_verify":          false,
					"timeout":             float64(2000),
					"username":            nil,
				},
				"retry_after_jitter_max": float64(1),
				"strategy":               string("local"),
				"sync_rate":              float64(-1),
				"window_size":            []any{float64(60)},
				"window_type":            string("sliding"),
			},
			Enabled:   kong.Bool(true),
			Protocols: []*string{kong.String("grpc"), kong.String("grpcs"), kong.String("http"), kong.String("https")},
		},
		{
			Name: kong.String("rate-limiting-advanced"),
			Config: kong.Configuration{
				"consumer_groups":         nil,
				"dictionary_name":         string("kong_rate_limiting_counters"),
				"disable_penalty":         bool(false),
				"enforce_consumer_groups": bool(false),
				"error_code":              float64(429),
				"error_message":           string("API rate limit exceeded"),
				"header_name":             nil,
				"hide_client_headers":     bool(false),
				"identifier":              string("consumer"),
				"limit":                   []any{float64(5)},
				"namespace":               string("silver"),
				"path":                    nil,
				"redis": map[string]any{
					"cluster_addresses":   nil,
					"connect_timeout":     nil,
					"database":            float64(0),
					"host":                nil,
					"keepalive_backlog":   nil,
					"keepalive_pool_size": float64(30),
					"password":            nil,
					"port":                nil,
					"read_timeout":        nil,
					"send_timeout":        nil,
					"sentinel_addresses":  nil,
					"sentinel_master":     nil,
					"sentinel_password":   nil,
					"sentinel_role":       nil,
					"sentinel_username":   nil,
					"server_name":         nil,
					"ssl":                 false,
					"ssl_verify":          false,
					"timeout":             float64(2000),
					"username":            nil,
				},
				"retry_after_jitter_max": float64(1),
				"strategy":               string("local"),
				"sync_rate":              float64(-1),
				"window_size":            []any{float64(60)},
				"window_type":            string("sliding"),
			},
			Enabled:   kong.Bool(true),
			Protocols: []*string{kong.String("grpc"), kong.String("grpcs"), kong.String("http"), kong.String("https")},
		},
		{
			Name: kong.String("key-auth"),
			Config: kong.Configuration{
				"anonymous":        nil,
				"hide_credentials": false,
				"key_in_body":      false,
				"key_in_header":    true,
				"key_in_query":     true,
				"key_names":        []interface{}{"apikey"},
				"run_on_preflight": true,
			},
			Enabled:   kong.Bool(true),
			Protocols: []*string{kong.String("http"), kong.String("https")},
		},
	}

	consumerGroupScopedPlugins36x = []*kong.Plugin{
		{
			Name: kong.String("rate-limiting-advanced"),
			ConsumerGroup: &kong.ConsumerGroup{
				ID: kong.String("77e6691d-67c0-446a-9401-27be2b141aae"),
			},
			Config: kong.Configuration{
				"consumer_groups":         nil,
				"dictionary_name":         string("kong_rate_limiting_counters"),
				"disable_penalty":         bool(false),
				"enforce_consumer_groups": bool(false),
				"error_code":              float64(429),
				"error_message":           string("API rate limit exceeded"),
				"header_name":             nil,
				"hide_client_headers":     bool(false),
				"identifier":              string("consumer"),
				"limit":                   []any{float64(10)},
				"namespace":               string("gold"),
				"path":                    nil,
				"redis": map[string]any{
					"cluster_addresses":   nil,
					"connect_timeout":     nil,
					"database":            float64(0),
					"host":                nil,
					"keepalive_backlog":   nil,
					"keepalive_pool_size": float64(256),
					"password":            nil,
					"port":                nil,
					"read_timeout":        nil,
					"send_timeout":        nil,
					"sentinel_addresses":  nil,
					"sentinel_master":     nil,
					"sentinel_password":   nil,
					"sentinel_role":       nil,
					"sentinel_username":   nil,
					"server_name":         nil,
					"ssl":                 false,
					"ssl_verify":          false,
					"timeout":             float64(2000),
					"username":            nil,
				},
				"retry_after_jitter_max": float64(1),
				"strategy":               string("local"),
				"sync_rate":              float64(-1),
				"window_size":            []any{float64(60)},
				"window_type":            string("sliding"),
			},
			Enabled:   kong.Bool(true),
			Protocols: []*string{kong.String("grpc"), kong.String("grpcs"), kong.String("http"), kong.String("https")},
		},
		{
			Name: kong.String("rate-limiting-advanced"),
			ConsumerGroup: &kong.ConsumerGroup{
				ID: kong.String("5bcbd3a7-030b-4310-bd1d-2721ff85d236"),
			},
			Config: kong.Configuration{
				"consumer_groups":         nil,
				"dictionary_name":         string("kong_rate_limiting_counters"),
				"disable_penalty":         bool(false),
				"enforce_consumer_groups": bool(false),
				"error_code":              float64(429),
				"error_message":           string("API rate limit exceeded"),
				"header_name":             nil,
				"hide_client_headers":     bool(false),
				"identifier":              string("consumer"),
				"limit":                   []any{float64(7)},
				"namespace":               string("silver"),
				"path":                    nil,
				"redis": map[string]any{
					"cluster_addresses":   nil,
					"connect_timeout":     nil,
					"database":            float64(0),
					"host":                nil,
					"keepalive_backlog":   nil,
					"keepalive_pool_size": float64(256),
					"password":            nil,
					"port":                nil,
					"read_timeout":        nil,
					"send_timeout":        nil,
					"sentinel_addresses":  nil,
					"sentinel_master":     nil,
					"sentinel_password":   nil,
					"sentinel_role":       nil,
					"sentinel_username":   nil,
					"server_name":         nil,
					"ssl":                 false,
					"ssl_verify":          false,
					"timeout":             float64(2000),
					"username":            nil,
				},
				"retry_after_jitter_max": float64(1),
				"strategy":               string("local"),
				"sync_rate":              float64(-1),
				"window_size":            []any{float64(60)},
				"window_type":            string("sliding"),
			},
			Enabled:   kong.Bool(true),
			Protocols: []*string{kong.String("grpc"), kong.String("grpcs"), kong.String("http"), kong.String("https")},
		},
		{
			Name: kong.String("rate-limiting-advanced"),
			Config: kong.Configuration{
				"consumer_groups":         nil,
				"dictionary_name":         string("kong_rate_limiting_counters"),
				"disable_penalty":         bool(false),
				"enforce_consumer_groups": bool(false),
				"error_code":              float64(429),
				"error_message":           string("API rate limit exceeded"),
				"header_name":             nil,
				"hide_client_headers":     bool(false),
				"identifier":              string("consumer"),
				"limit":                   []any{float64(5)},
				"namespace":               string("silver"),
				"path":                    nil,
				"redis": map[string]any{
					"cluster_addresses":   nil,
					"connect_timeout":     nil,
					"database":            float64(0),
					"host":                nil,
					"keepalive_backlog":   nil,
					"keepalive_pool_size": float64(256),
					"password":            nil,
					"port":                nil,
					"read_timeout":        nil,
					"send_timeout":        nil,
					"sentinel_addresses":  nil,
					"sentinel_master":     nil,
					"sentinel_password":   nil,
					"sentinel_role":       nil,
					"sentinel_username":   nil,
					"server_name":         nil,
					"ssl":                 false,
					"ssl_verify":          false,
					"timeout":             float64(2000),
					"username":            nil,
				},
				"retry_after_jitter_max": float64(1),
				"strategy":               string("local"),
				"sync_rate":              float64(-1),
				"window_size":            []any{float64(60)},
				"window_type":            string("sliding"),
			},
			Enabled:   kong.Bool(true),
			Protocols: []*string{kong.String("grpc"), kong.String("grpcs"), kong.String("http"), kong.String("https")},
		},
		{
			Name: kong.String("key-auth"),
			Config: kong.Configuration{
				"anonymous":        nil,
				"hide_credentials": false,
				"key_in_body":      false,
				"key_in_header":    true,
				"key_in_query":     true,
				"key_names":        []interface{}{"apikey"},
				"run_on_preflight": true,
			},
			Enabled:   kong.Bool(true),
			Protocols: []*string{kong.String("http"), kong.String("https")},
		},
	}

	consumerGroupScopedPlugins37x = []*kong.Plugin{
		{
			Name: kong.String("rate-limiting-advanced"),
			ConsumerGroup: &kong.ConsumerGroup{
				ID: kong.String("77e6691d-67c0-446a-9401-27be2b141aae"),
			},
			Config: kong.Configuration{
				"consumer_groups":         nil,
				"dictionary_name":         string("kong_rate_limiting_counters"),
				"disable_penalty":         bool(false),
				"enforce_consumer_groups": bool(false),
				"error_code":              float64(429),
				"error_message":           string("API rate limit exceeded"),
				"header_name":             nil,
				"hide_client_headers":     bool(false),
				"identifier":              string("consumer"),
				"limit":                   []any{float64(10)},
				"namespace":               string("gold"),
				"path":                    nil,
				"redis": map[string]any{
					"cluster_addresses":   nil,
					"connect_timeout":     nil,
					"database":            float64(0),
					"host":                nil,
					"keepalive_backlog":   nil,
					"keepalive_pool_size": float64(256),
					"password":            nil,
					"port":                nil,
					"read_timeout":        nil,
					"send_timeout":        nil,
					"sentinel_addresses":  nil,
					"sentinel_master":     nil,
					"sentinel_password":   nil,
					"sentinel_role":       nil,
					"sentinel_username":   nil,
					"server_name":         nil,
					"ssl":                 false,
					"ssl_verify":          false,
					"timeout":             float64(2000),
					"username":            nil,
				},
				"retry_after_jitter_max": float64(1),
				"strategy":               string("local"),
				"sync_rate":              float64(-1),
				"window_size":            []any{float64(60)},
				"window_type":            string("sliding"),
			},
			Enabled:   kong.Bool(true),
			Protocols: []*string{kong.String("grpc"), kong.String("grpcs"), kong.String("http"), kong.String("https")},
		},
		{
			Name: kong.String("rate-limiting-advanced"),
			ConsumerGroup: &kong.ConsumerGroup{
				ID: kong.String("5bcbd3a7-030b-4310-bd1d-2721ff85d236"),
			},
			Config: kong.Configuration{
				"consumer_groups":         nil,
				"dictionary_name":         string("kong_rate_limiting_counters"),
				"disable_penalty":         bool(false),
				"enforce_consumer_groups": bool(false),
				"error_code":              float64(429),
				"error_message":           string("API rate limit exceeded"),
				"header_name":             nil,
				"hide_client_headers":     bool(false),
				"identifier":              string("consumer"),
				"limit":                   []any{float64(7)},
				"namespace":               string("silver"),
				"path":                    nil,
				"redis": map[string]any{
					"cluster_addresses":   nil,
					"connect_timeout":     nil,
					"database":            float64(0),
					"host":                nil,
					"keepalive_backlog":   nil,
					"keepalive_pool_size": float64(256),
					"password":            nil,
					"port":                nil,
					"read_timeout":        nil,
					"send_timeout":        nil,
					"sentinel_addresses":  nil,
					"sentinel_master":     nil,
					"sentinel_password":   nil,
					"sentinel_role":       nil,
					"sentinel_username":   nil,
					"server_name":         nil,
					"ssl":                 false,
					"ssl_verify":          false,
					"timeout":             float64(2000),
					"username":            nil,
				},
				"retry_after_jitter_max": float64(1),
				"strategy":               string("local"),
				"sync_rate":              float64(-1),
				"window_size":            []any{float64(60)},
				"window_type":            string("sliding"),
			},
			Enabled:   kong.Bool(true),
			Protocols: []*string{kong.String("grpc"), kong.String("grpcs"), kong.String("http"), kong.String("https")},
		},
		{
			Name: kong.String("rate-limiting-advanced"),
			Config: kong.Configuration{
				"consumer_groups":         nil,
				"dictionary_name":         string("kong_rate_limiting_counters"),
				"disable_penalty":         bool(false),
				"enforce_consumer_groups": bool(false),
				"error_code":              float64(429),
				"error_message":           string("API rate limit exceeded"),
				"header_name":             nil,
				"hide_client_headers":     bool(false),
				"identifier":              string("consumer"),
				"limit":                   []any{float64(5)},
				"namespace":               string("silver"),
				"path":                    nil,
				"redis": map[string]any{
					"cluster_addresses":   nil,
					"connect_timeout":     nil,
					"database":            float64(0),
					"host":                nil,
					"keepalive_backlog":   nil,
					"keepalive_pool_size": float64(256),
					"password":            nil,
					"port":                nil,
					"read_timeout":        nil,
					"send_timeout":        nil,
					"sentinel_addresses":  nil,
					"sentinel_master":     nil,
					"sentinel_password":   nil,
					"sentinel_role":       nil,
					"sentinel_username":   nil,
					"server_name":         nil,
					"ssl":                 false,
					"ssl_verify":          false,
					"timeout":             float64(2000),
					"username":            nil,
				},
				"retry_after_jitter_max": float64(1),
				"strategy":               string("local"),
				"sync_rate":              float64(-1),
				"window_size":            []any{float64(60)},
				"window_type":            string("sliding"),
			},
			Enabled:   kong.Bool(true),
			Protocols: []*string{kong.String("grpc"), kong.String("grpcs"), kong.String("http"), kong.String("https")},
		},
		{
			Name: kong.String("key-auth"),
			Config: kong.Configuration{
				"anonymous":        nil,
				"hide_credentials": false,
				"key_in_body":      false,
				"key_in_header":    true,
				"key_in_query":     true,
				"key_names":        []interface{}{"apikey"},
				"realm":            nil, // This is present on 3.7.x+
				"run_on_preflight": true,
			},
			Enabled:   kong.Bool(true),
			Protocols: []*string{kong.String("http"), kong.String("https")},
		},
	}

	consumerGroupScopedPluginsKonnect = []*kong.Plugin{
		{
			Name: kong.String("rate-limiting-advanced"),
			ConsumerGroup: &kong.ConsumerGroup{
				ID: kong.String("77e6691d-67c0-446a-9401-27be2b141aae"),
			},
			Config: kong.Configuration{
				"compound_identifier":     nil,
				"consumer_groups":         nil,
				"dictionary_name":         string("kong_rate_limiting_counters"),
				"disable_penalty":         bool(false),
				"enforce_consumer_groups": bool(false),
				"error_code":              float64(429),
				"error_message":           string("API rate limit exceeded"),
				"header_name":             nil,
				"hide_client_headers":     bool(false),
				"identifier":              string("consumer"),
				"limit":                   []any{float64(10)},
				"lock_dictionary_name":    string("kong_locks"),
				"namespace":               string("gold"),
				"path":                    nil,
				"redis": map[string]any{
					"cluster_addresses":   nil,
					"connect_timeout":     nil,
					"database":            float64(0),
					"host":                nil,
					"keepalive_backlog":   nil,
					"keepalive_pool_size": float64(256),
					"password":            nil,
					"port":                nil,
					"read_timeout":        nil,
					"send_timeout":        nil,
					"sentinel_addresses":  nil,
					"sentinel_master":     nil,
					"sentinel_password":   nil,
					"sentinel_role":       nil,
					"sentinel_username":   nil,
					"server_name":         nil,
					"ssl":                 false,
					"ssl_verify":          false,
					"timeout":             float64(2000),
					"username":            nil,
				},
				"retry_after_jitter_max": float64(1),
				"strategy":               string("local"),
				"sync_rate":              float64(-1),
				"window_size":            []any{float64(60)},
				"window_type":            string("sliding"),
			},
			Enabled:   kong.Bool(true),
			Protocols: []*string{kong.String("grpc"), kong.String("grpcs"), kong.String("http"), kong.String("https")},
		},
		{
			Name: kong.String("rate-limiting-advanced"),
			ConsumerGroup: &kong.ConsumerGroup{
				ID: kong.String("5bcbd3a7-030b-4310-bd1d-2721ff85d236"),
			},
			Config: kong.Configuration{
				"compound_identifier":     nil,
				"consumer_groups":         nil,
				"dictionary_name":         string("kong_rate_limiting_counters"),
				"disable_penalty":         bool(false),
				"enforce_consumer_groups": bool(false),
				"error_code":              float64(429),
				"error_message":           string("API rate limit exceeded"),
				"header_name":             nil,
				"hide_client_headers":     bool(false),
				"identifier":              string("consumer"),
				"limit":                   []any{float64(7)},
				"lock_dictionary_name":    string("kong_locks"),
				"namespace":               string("silver"),
				"path":                    nil,
				"redis": map[string]any{
					"cluster_addresses":   nil,
					"connect_timeout":     nil,
					"database":            float64(0),
					"host":                nil,
					"keepalive_backlog":   nil,
					"keepalive_pool_size": float64(256),
					"password":            nil,
					"port":                nil,
					"read_timeout":        nil,
					"send_timeout":        nil,
					"sentinel_addresses":  nil,
					"sentinel_master":     nil,
					"sentinel_password":   nil,
					"sentinel_role":       nil,
					"sentinel_username":   nil,
					"server_name":         nil,
					"ssl":                 false,
					"ssl_verify":          false,
					"timeout":             float64(2000),
					"username":            nil,
				},
				"retry_after_jitter_max": float64(1),
				"strategy":               string("local"),
				"sync_rate":              float64(-1),
				"window_size":            []any{float64(60)},
				"window_type":            string("sliding"),
			},
			Enabled:   kong.Bool(true),
			Protocols: []*string{kong.String("grpc"), kong.String("grpcs"), kong.String("http"), kong.String("https")},
		},
		{
			Name: kong.String("rate-limiting-advanced"),
			Config: kong.Configuration{
				"compound_identifier":     nil,
				"consumer_groups":         nil,
				"dictionary_name":         string("kong_rate_limiting_counters"),
				"disable_penalty":         bool(false),
				"enforce_consumer_groups": bool(false),
				"error_code":              float64(429),
				"error_message":           string("API rate limit exceeded"),
				"header_name":             nil,
				"hide_client_headers":     bool(false),
				"identifier":              string("consumer"),
				"limit":                   []any{float64(5)},
				"lock_dictionary_name":    string("kong_locks"),
				"namespace":               string("silver"),
				"path":                    nil,
				"redis": map[string]any{
					"cluster_addresses":   nil,
					"connect_timeout":     nil,
					"database":            float64(0),
					"host":                nil,
					"keepalive_backlog":   nil,
					"keepalive_pool_size": float64(256),
					"password":            nil,
					"port":                nil,
					"read_timeout":        nil,
					"send_timeout":        nil,
					"sentinel_addresses":  nil,
					"sentinel_master":     nil,
					"sentinel_password":   nil,
					"sentinel_role":       nil,
					"sentinel_username":   nil,
					"server_name":         nil,
					"ssl":                 false,
					"ssl_verify":          false,
					"timeout":             float64(2000),
					"username":            nil,
				},
				"retry_after_jitter_max": float64(1),
				"strategy":               string("local"),
				"sync_rate":              float64(-1),
				"window_size":            []any{float64(60)},
				"window_type":            string("sliding"),
			},
			Enabled:   kong.Bool(true),
			Protocols: []*string{kong.String("grpc"), kong.String("grpcs"), kong.String("http"), kong.String("https")},
		},
		{
			Name: kong.String("key-auth"),
			Config: kong.Configuration{
				"anonymous":        nil,
				"hide_credentials": false,
				"identity_realms":  []any{map[string]any{"id": nil, "region": nil, "scope": string("cp")}},
				"key_in_body":      false,
				"key_in_header":    true,
				"key_in_query":     true,
				"key_names":        []interface{}{"apikey"},
				"realm":            nil, // This is present on 3.7.x+
				"run_on_preflight": true,
			},
			Enabled:   kong.Bool(true),
			Protocols: []*string{kong.String("http"), kong.String("https")},
		},
	}

	consumerGroupScopedPlugins381x = []*kong.Plugin{
		{
			Name: kong.String("rate-limiting-advanced"),
			ConsumerGroup: &kong.ConsumerGroup{
				ID: kong.String("77e6691d-67c0-446a-9401-27be2b141aae"),
			},
			Config: kong.Configuration{
				"consumer_groups":         nil,
				"dictionary_name":         string("kong_rate_limiting_counters"),
				"disable_penalty":         bool(false),
				"enforce_consumer_groups": bool(false),
				"error_code":              float64(429),
				"error_message":           string("API rate limit exceeded"),
				"header_name":             nil,
				"hide_client_headers":     bool(false),
				"identifier":              string("consumer"),
				"limit":                   []any{float64(10)},
				"namespace":               string("gold"),
				"path":                    nil,
				"redis": map[string]any{
					"cluster_addresses":        nil,
					"cluster_max_redirections": float64(5),
					"cluster_nodes":            nil,
					"connect_timeout":          float64(2000),
					"connection_is_proxied":    bool(false),
					"database":                 float64(0),
					"host":                     string("127.0.0.1"),
					"keepalive_backlog":        nil,
					"keepalive_pool_size":      float64(256),
					"password":                 nil,
					"port":                     float64(6379),
					"read_timeout":             float64(2000),
					"send_timeout":             float64(2000),
					"sentinel_addresses":       nil,
					"sentinel_master":          nil,
					"sentinel_nodes":           nil,
					"sentinel_password":        nil,
					"sentinel_role":            nil,
					"sentinel_username":        nil,
					"server_name":              nil,
					"ssl":                      false,
					"ssl_verify":               false,
					"timeout":                  float64(2000),
					"username":                 nil,
				},
				"retry_after_jitter_max": float64(1),
				"strategy":               string("local"),
				"sync_rate":              float64(-1),
				"window_size":            []any{float64(60)},
				"window_type":            string("sliding"),
			},
			Enabled:   kong.Bool(true),
			Protocols: []*string{kong.String("grpc"), kong.String("grpcs"), kong.String("http"), kong.String("https")},
		},
		{
			Name: kong.String("rate-limiting-advanced"),
			ConsumerGroup: &kong.ConsumerGroup{
				ID: kong.String("5bcbd3a7-030b-4310-bd1d-2721ff85d236"),
			},
			Config: kong.Configuration{
				"consumer_groups":         nil,
				"dictionary_name":         string("kong_rate_limiting_counters"),
				"disable_penalty":         bool(false),
				"enforce_consumer_groups": bool(false),
				"error_code":              float64(429),
				"error_message":           string("API rate limit exceeded"),
				"header_name":             nil,
				"hide_client_headers":     bool(false),
				"identifier":              string("consumer"),
				"limit":                   []any{float64(7)},
				"namespace":               string("silver"),
				"path":                    nil,
				"redis": map[string]any{
					"cluster_addresses":        nil,
					"cluster_max_redirections": float64(5),
					"cluster_nodes":            nil,
					"connect_timeout":          float64(2000),
					"connection_is_proxied":    bool(false),
					"database":                 float64(0),
					"host":                     string("127.0.0.1"),
					"keepalive_backlog":        nil,
					"keepalive_pool_size":      float64(256),
					"password":                 nil,
					"port":                     float64(6379),
					"read_timeout":             float64(2000),
					"send_timeout":             float64(2000),
					"sentinel_addresses":       nil,
					"sentinel_master":          nil,
					"sentinel_nodes":           nil,
					"sentinel_password":        nil,
					"sentinel_role":            nil,
					"sentinel_username":        nil,
					"server_name":              nil,
					"ssl":                      false,
					"ssl_verify":               false,
					"timeout":                  float64(2000),
					"username":                 nil,
				},
				"retry_after_jitter_max": float64(1),
				"strategy":               string("local"),
				"sync_rate":              float64(-1),
				"window_size":            []any{float64(60)},
				"window_type":            string("sliding"),
			},
			Enabled:   kong.Bool(true),
			Protocols: []*string{kong.String("grpc"), kong.String("grpcs"), kong.String("http"), kong.String("https")},
		},
		{
			Name: kong.String("rate-limiting-advanced"),
			Config: kong.Configuration{
				"consumer_groups":         nil,
				"dictionary_name":         string("kong_rate_limiting_counters"),
				"disable_penalty":         bool(false),
				"enforce_consumer_groups": bool(false),
				"error_code":              float64(429),
				"error_message":           string("API rate limit exceeded"),
				"header_name":             nil,
				"hide_client_headers":     bool(false),
				"identifier":              string("consumer"),
				"limit":                   []any{float64(5)},
				"namespace":               string("silver"),
				"path":                    nil,
				"redis": map[string]any{
					"cluster_addresses":        nil,
					"cluster_max_redirections": float64(5),
					"cluster_nodes":            nil,
					"connect_timeout":          float64(2000),
					"connection_is_proxied":    bool(false),
					"database":                 float64(0),
					"host":                     string("127.0.0.1"),
					"keepalive_backlog":        nil,
					"keepalive_pool_size":      float64(256),
					"password":                 nil,
					"port":                     float64(6379),
					"read_timeout":             float64(2000),
					"send_timeout":             float64(2000),
					"sentinel_addresses":       nil,
					"sentinel_master":          nil,
					"sentinel_nodes":           nil,
					"sentinel_password":        nil,
					"sentinel_role":            nil,
					"sentinel_username":        nil,
					"server_name":              nil,
					"ssl":                      false,
					"ssl_verify":               false,
					"timeout":                  float64(2000),
					"username":                 nil,
				},
				"retry_after_jitter_max": float64(1),
				"strategy":               string("local"),
				"sync_rate":              float64(-1),
				"window_size":            []any{float64(60)},
				"window_type":            string("sliding"),
			},
			Enabled:   kong.Bool(true),
			Protocols: []*string{kong.String("grpc"), kong.String("grpcs"), kong.String("http"), kong.String("https")},
		},
		{
			Name: kong.String("key-auth"),
			Config: kong.Configuration{
				"anonymous":        nil,
				"hide_credentials": false,
				"key_in_body":      false,
				"key_in_header":    true,
				"key_in_query":     true,
				"key_names":        []interface{}{"apikey"},
				"realm":            nil, // This is present on 3.7.x+
				"run_on_preflight": true,
			},
			Enabled:   kong.Bool(true),
			Protocols: []*string{kong.String("http"), kong.String("https")},
		},
	}

	consumerGroupScopedPlugins390x = []*kong.Plugin{
		{
			Name: kong.String("rate-limiting-advanced"),
			ConsumerGroup: &kong.ConsumerGroup{
				ID: kong.String("77e6691d-67c0-446a-9401-27be2b141aae"),
			},
			Config: kong.Configuration{
				"compound_identifier":     nil,
				"consumer_groups":         nil,
				"dictionary_name":         string("kong_rate_limiting_counters"),
				"disable_penalty":         bool(false),
				"enforce_consumer_groups": bool(false),
				"error_code":              float64(429),
				"error_message":           string("API rate limit exceeded"),
				"header_name":             nil,
				"hide_client_headers":     bool(false),
				"identifier":              string("consumer"),
				"limit":                   []any{float64(10)},
				"lock_dictionary_name":    string("kong_locks"),
				"namespace":               string("gold"),
				"path":                    nil,
				"redis": map[string]any{
					"cluster_addresses":        nil,
					"cluster_max_redirections": float64(5),
					"cluster_nodes":            nil,
					"connect_timeout":          float64(2000),
					"connection_is_proxied":    bool(false),
					"database":                 float64(0),
					"host":                     string("127.0.0.1"),
					"keepalive_backlog":        nil,
					"keepalive_pool_size":      float64(256),
					"password":                 nil,
					"port":                     float64(6379),
					"read_timeout":             float64(2000),
					"redis_proxy_type":         nil,
					"send_timeout":             float64(2000),
					"sentinel_addresses":       nil,
					"sentinel_master":          nil,
					"sentinel_nodes":           nil,
					"sentinel_password":        nil,
					"sentinel_role":            nil,
					"sentinel_username":        nil,
					"server_name":              nil,
					"ssl":                      false,
					"ssl_verify":               false,
					"timeout":                  float64(2000),
					"username":                 nil,
				},
				"retry_after_jitter_max": float64(1),
				"strategy":               string("local"),
				"sync_rate":              float64(-1),
				"window_size":            []any{float64(60)},
				"window_type":            string("sliding"),
			},
			Enabled:   kong.Bool(true),
			Protocols: []*string{kong.String("grpc"), kong.String("grpcs"), kong.String("http"), kong.String("https")},
		},
		{
			Name: kong.String("rate-limiting-advanced"),
			ConsumerGroup: &kong.ConsumerGroup{
				ID: kong.String("5bcbd3a7-030b-4310-bd1d-2721ff85d236"),
			},
			Config: kong.Configuration{
				"compound_identifier":     nil,
				"consumer_groups":         nil,
				"dictionary_name":         string("kong_rate_limiting_counters"),
				"disable_penalty":         bool(false),
				"enforce_consumer_groups": bool(false),
				"error_code":              float64(429),
				"error_message":           string("API rate limit exceeded"),
				"header_name":             nil,
				"hide_client_headers":     bool(false),
				"identifier":              string("consumer"),
				"limit":                   []any{float64(7)},
				"lock_dictionary_name":    string("kong_locks"),
				"namespace":               string("silver"),
				"path":                    nil,
				"redis": map[string]any{
					"cluster_addresses":        nil,
					"cluster_max_redirections": float64(5),
					"cluster_nodes":            nil,
					"connect_timeout":          float64(2000),
					"connection_is_proxied":    bool(false),
					"database":                 float64(0),
					"host":                     string("127.0.0.1"),
					"keepalive_backlog":        nil,
					"keepalive_pool_size":      float64(256),
					"password":                 nil,
					"port":                     float64(6379),
					"read_timeout":             float64(2000),
					"redis_proxy_type":         nil,
					"send_timeout":             float64(2000),
					"sentinel_addresses":       nil,
					"sentinel_master":          nil,
					"sentinel_nodes":           nil,
					"sentinel_password":        nil,
					"sentinel_role":            nil,
					"sentinel_username":        nil,
					"server_name":              nil,
					"ssl":                      false,
					"ssl_verify":               false,
					"timeout":                  float64(2000),
					"username":                 nil,
				},
				"retry_after_jitter_max": float64(1),
				"strategy":               string("local"),
				"sync_rate":              float64(-1),
				"window_size":            []any{float64(60)},
				"window_type":            string("sliding"),
			},
			Enabled:   kong.Bool(true),
			Protocols: []*string{kong.String("grpc"), kong.String("grpcs"), kong.String("http"), kong.String("https")},
		},
		{
			Name: kong.String("rate-limiting-advanced"),
			Config: kong.Configuration{
				"compound_identifier":     nil,
				"consumer_groups":         nil,
				"dictionary_name":         string("kong_rate_limiting_counters"),
				"disable_penalty":         bool(false),
				"enforce_consumer_groups": bool(false),
				"error_code":              float64(429),
				"error_message":           string("API rate limit exceeded"),
				"header_name":             nil,
				"hide_client_headers":     bool(false),
				"identifier":              string("consumer"),
				"limit":                   []any{float64(5)},
				"lock_dictionary_name":    string("kong_locks"),
				"namespace":               string("silver"),
				"path":                    nil,
				"redis": map[string]any{
					"cluster_addresses":        nil,
					"cluster_max_redirections": float64(5),
					"cluster_nodes":            nil,
					"connect_timeout":          float64(2000),
					"connection_is_proxied":    bool(false),
					"database":                 float64(0),
					"host":                     string("127.0.0.1"),
					"keepalive_backlog":        nil,
					"keepalive_pool_size":      float64(256),
					"password":                 nil,
					"port":                     float64(6379),
					"read_timeout":             float64(2000),
					"redis_proxy_type":         nil,
					"send_timeout":             float64(2000),
					"sentinel_addresses":       nil,
					"sentinel_master":          nil,
					"sentinel_nodes":           nil,
					"sentinel_password":        nil,
					"sentinel_role":            nil,
					"sentinel_username":        nil,
					"server_name":              nil,
					"ssl":                      false,
					"ssl_verify":               false,
					"timeout":                  float64(2000),
					"username":                 nil,
				},
				"retry_after_jitter_max": float64(1),
				"strategy":               string("local"),
				"sync_rate":              float64(-1),
				"window_size":            []any{float64(60)},
				"window_type":            string("sliding"),
			},
			Enabled:   kong.Bool(true),
			Protocols: []*string{kong.String("grpc"), kong.String("grpcs"), kong.String("http"), kong.String("https")},
		},
		{
			Name: kong.String("key-auth"),
			Config: kong.Configuration{
				"anonymous":        nil,
				"hide_credentials": false,
				"key_in_body":      false,
				"key_in_header":    true,
				"key_in_query":     true,
				"key_names":        []interface{}{"apikey"},
				"realm":            nil, // This is present on 3.7.x+
				"run_on_preflight": true,
			},
			Enabled:   kong.Bool(true),
			Protocols: []*string{kong.String("http"), kong.String("https")},
		},
	}

	consumerGroupScopedPlugins310x = []*kong.Plugin{
		{
			Name: kong.String("rate-limiting-advanced"),
			ConsumerGroup: &kong.ConsumerGroup{
				ID: kong.String("77e6691d-67c0-446a-9401-27be2b141aae"),
			},
			Config: kong.Configuration{
				"compound_identifier":     nil,
				"consumer_groups":         nil,
				"dictionary_name":         string("kong_rate_limiting_counters"),
				"disable_penalty":         bool(false),
				"enforce_consumer_groups": bool(false),
				"error_code":              float64(429),
				"error_message":           string("API rate limit exceeded"),
				"header_name":             nil,
				"hide_client_headers":     bool(false),
				"identifier":              string("consumer"),
				"limit":                   []any{float64(10)},
				"lock_dictionary_name":    string("kong_locks"),
				"namespace":               string("gold"),
				"path":                    nil,
				"redis": map[string]any{
					"cluster_addresses":        nil,
					"cluster_max_redirections": float64(5),
					"cluster_nodes":            nil,
					"connect_timeout":          float64(2000),
					"connection_is_proxied":    bool(false),
					"database":                 float64(0),
					"host":                     string("127.0.0.1"),
					"keepalive_backlog":        nil,
					"keepalive_pool_size":      float64(256),
					"password":                 nil,
					"port":                     float64(6379),
					"read_timeout":             float64(2000),
					"redis_proxy_type":         nil,
					"send_timeout":             float64(2000),
					"sentinel_addresses":       nil,
					"sentinel_master":          nil,
					"sentinel_nodes":           nil,
					"sentinel_password":        nil,
					"sentinel_role":            nil,
					"sentinel_username":        nil,
					"server_name":              nil,
					"ssl":                      false,
					"ssl_verify":               false,
					"timeout":                  float64(2000),
					"username":                 nil,
				},
				"retry_after_jitter_max": float64(1),
				"strategy":               string("local"),
				"sync_rate":              float64(-1),
				"window_size":            []any{float64(60)},
				"window_type":            string("sliding"),
			},
			Enabled:   kong.Bool(true),
			Protocols: []*string{kong.String("grpc"), kong.String("grpcs"), kong.String("http"), kong.String("https")},
		},
		{
			Name: kong.String("rate-limiting-advanced"),
			ConsumerGroup: &kong.ConsumerGroup{
				ID: kong.String("5bcbd3a7-030b-4310-bd1d-2721ff85d236"),
			},
			Config: kong.Configuration{
				"compound_identifier":     nil,
				"consumer_groups":         nil,
				"dictionary_name":         string("kong_rate_limiting_counters"),
				"disable_penalty":         bool(false),
				"enforce_consumer_groups": bool(false),
				"error_code":              float64(429),
				"error_message":           string("API rate limit exceeded"),
				"header_name":             nil,
				"hide_client_headers":     bool(false),
				"identifier":              string("consumer"),
				"limit":                   []any{float64(7)},
				"lock_dictionary_name":    string("kong_locks"),
				"namespace":               string("silver"),
				"path":                    nil,
				"redis": map[string]any{
					"cluster_addresses":        nil,
					"cluster_max_redirections": float64(5),
					"cluster_nodes":            nil,
					"connect_timeout":          float64(2000),
					"connection_is_proxied":    bool(false),
					"database":                 float64(0),
					"host":                     string("127.0.0.1"),
					"keepalive_backlog":        nil,
					"keepalive_pool_size":      float64(256),
					"password":                 nil,
					"port":                     float64(6379),
					"read_timeout":             float64(2000),
					"redis_proxy_type":         nil,
					"send_timeout":             float64(2000),
					"sentinel_addresses":       nil,
					"sentinel_master":          nil,
					"sentinel_nodes":           nil,
					"sentinel_password":        nil,
					"sentinel_role":            nil,
					"sentinel_username":        nil,
					"server_name":              nil,
					"ssl":                      false,
					"ssl_verify":               false,
					"timeout":                  float64(2000),
					"username":                 nil,
				},
				"retry_after_jitter_max": float64(1),
				"strategy":               string("local"),
				"sync_rate":              float64(-1),
				"window_size":            []any{float64(60)},
				"window_type":            string("sliding"),
			},
			Enabled:   kong.Bool(true),
			Protocols: []*string{kong.String("grpc"), kong.String("grpcs"), kong.String("http"), kong.String("https")},
		},
		{
			Name: kong.String("rate-limiting-advanced"),
			Config: kong.Configuration{
				"compound_identifier":     nil,
				"consumer_groups":         nil,
				"dictionary_name":         string("kong_rate_limiting_counters"),
				"disable_penalty":         bool(false),
				"enforce_consumer_groups": bool(false),
				"error_code":              float64(429),
				"error_message":           string("API rate limit exceeded"),
				"header_name":             nil,
				"hide_client_headers":     bool(false),
				"identifier":              string("consumer"),
				"limit":                   []any{float64(5)},
				"lock_dictionary_name":    string("kong_locks"),
				"namespace":               string("silver"),
				"path":                    nil,
				"redis": map[string]any{
					"cluster_addresses":        nil,
					"cluster_max_redirections": float64(5),
					"cluster_nodes":            nil,
					"connect_timeout":          float64(2000),
					"connection_is_proxied":    bool(false),
					"database":                 float64(0),
					"host":                     string("127.0.0.1"),
					"keepalive_backlog":        nil,
					"keepalive_pool_size":      float64(256),
					"password":                 nil,
					"port":                     float64(6379),
					"read_timeout":             float64(2000),
					"redis_proxy_type":         nil,
					"send_timeout":             float64(2000),
					"sentinel_addresses":       nil,
					"sentinel_master":          nil,
					"sentinel_nodes":           nil,
					"sentinel_password":        nil,
					"sentinel_role":            nil,
					"sentinel_username":        nil,
					"server_name":              nil,
					"ssl":                      false,
					"ssl_verify":               false,
					"timeout":                  float64(2000),
					"username":                 nil,
				},
				"retry_after_jitter_max": float64(1),
				"strategy":               string("local"),
				"sync_rate":              float64(-1),
				"window_size":            []any{float64(60)},
				"window_type":            string("sliding"),
			},
			Enabled:   kong.Bool(true),
			Protocols: []*string{kong.String("grpc"), kong.String("grpcs"), kong.String("http"), kong.String("https")},
		},
		{
			Name: kong.String("key-auth"),
			Config: kong.Configuration{
				"anonymous":        nil,
				"hide_credentials": false,
				"identity_realms":  []any{map[string]any{"id": nil, "region": nil, "scope": string("cp")}},
				"key_in_body":      false,
				"key_in_header":    true,
				"key_in_query":     true,
				"key_names":        []interface{}{"apikey"},
				"realm":            nil, // This is present on 3.7.x+
				"run_on_preflight": true,
			},
			Enabled:   kong.Bool(true),
			Protocols: []*string{kong.String("http"), kong.String("https")},
		},
	}
)

const complexQueryForDegraphqlRoute = `query SearchPosts($filters: PostsFilters) {
  posts(filter: $filters) {
    id
    title
    author
  }
}
`

const errorConsumerGroupPolicies = "a rate-limiting-advanced plugin with config.consumer_groups\n" +
	"and/or config.enforce_consumer_groups was found. Please use Consumer Groups scoped\n" +
	"Plugins when running against Kong Enterprise 3.4.0 and above.\n\n" +
	"Check https://docs.konghq.com/gateway/latest/kong-enterprise/consumer-groups/ for more information"

// test scope:
//   - 1.4.3
func Test_Sync_ServicesRoutes_Till_1_4_3(t *testing.T) {
	// setup stage
	client, err := getTestClient()
	require.NoError(t, err)

	// ignore entities fields based on Kong version
	ignoreFields := []cmp.Option{
		cmpopts.IgnoreFields(kong.Route{}, "Service"),
	}

	tests := []struct {
		name          string
		kongFile      string
		expectedState utils.KongRawState
	}{
		{
			name:     "creates a service",
			kongFile: "testdata/sync/001-create-a-service/kong.yaml",
			expectedState: utils.KongRawState{
				Services: svc1,
			},
		},
		{
			name:     "create services and routes",
			kongFile: "testdata/sync/002-create-services-and-routes/kong.yaml",
			expectedState: utils.KongRawState{
				Services: svc1,
				Routes:   route1_143,
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runWhen(t, "kong", "<=1.4.3")
			setup(t)

			require.NoError(t, sync(context.Background(), tc.kongFile))
			testKongState(t, client, false, false, tc.expectedState, ignoreFields)
		})
	}
}

// test scope:
//   - 1.5.1
//   - 1.5.0.11+enterprise
func Test_Sync_ServicesRoutes_Till_1_5_1(t *testing.T) {
	// setup stage
	client, err := getTestClient()
	require.NoError(t, err)

	tests := []struct {
		name          string
		kongFile      string
		expectedState utils.KongRawState
	}{
		{
			name:     "creates a service",
			kongFile: "testdata/sync/001-create-a-service/kong.yaml",
			expectedState: utils.KongRawState{
				Services: svc1,
			},
		},
		{
			name:     "create services and routes",
			kongFile: "testdata/sync/002-create-services-and-routes/kong.yaml",
			expectedState: utils.KongRawState{
				Services: svc1,
				Routes:   route1_151,
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runWhen(t, "kong", ">1.4.3 <=1.5.1")
			setup(t)

			require.NoError(t, sync(context.Background(), tc.kongFile))
			testKongState(t, client, false, false, tc.expectedState, nil)
		})
	}
}

// test scope:
//   - 2.0.5
//   - 2.1.4
func Test_Sync_ServicesRoutes_From_2_0_5_To_2_1_4(t *testing.T) {
	// setup stage
	client, err := getTestClient()
	require.NoError(t, err)

	tests := []struct {
		name          string
		kongFile      string
		expectedState utils.KongRawState
	}{
		{
			name:     "creates a service",
			kongFile: "testdata/sync/001-create-a-service/kong.yaml",
			expectedState: utils.KongRawState{
				Services: svc1,
			},
		},
		{
			name:     "create services and routes",
			kongFile: "testdata/sync/002-create-services-and-routes/kong.yaml",
			expectedState: utils.KongRawState{
				Services: svc1,
				Routes:   route1_205_214,
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runWhen(t, "kong", ">=2.0.5 <=2.1.4")
			setup(t)

			require.NoError(t, sync(context.Background(), tc.kongFile))
			testKongState(t, client, false, false, tc.expectedState, nil)
		})
	}
}

// test scope:
//   - 2.2.2
//   - 2.3.3
//   - 2.4.1
//   - 2.5.1
//   - 2.6.0
//   - 2.2.1.3+enterprise
//   - 2.3.3.4+enterprise
//   - 2.4.1.3+enterprise
//   - 2.5.1.2+enterprise
func Test_Sync_ServicesRoutes_From_2_2_1_to_2_6_0(t *testing.T) {
	// setup stage
	client, err := getTestClient()
	require.NoError(t, err)

	tests := []struct {
		name          string
		kongFile      string
		expectedState utils.KongRawState
	}{
		{
			name:     "creates a service",
			kongFile: "testdata/sync/001-create-a-service/kong.yaml",
			expectedState: utils.KongRawState{
				Services: svc1,
			},
		},
		{
			name:     "create services and routes",
			kongFile: "testdata/sync/002-create-services-and-routes/kong.yaml",
			expectedState: utils.KongRawState{
				Services: svc1,
				Routes:   route1_20x,
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runWhen(t, "kong", ">2.2.1 <=2.6.0")
			setup(t)

			require.NoError(t, sync(context.Background(), tc.kongFile))
			testKongState(t, client, false, false, tc.expectedState, nil)
		})
	}
}

// test scope:
//   - 2.7.0
//   - 2.6.0.2+enterprise
//   - 2.7.0.0+enterprise
//   - 2.8.0.0+enterprise
func Test_Sync_ServicesRoutes_From_2_6_9_Till_2_8_0(t *testing.T) {
	// setup stage
	client, err := getTestClient()
	require.NoError(t, err)

	tests := []struct {
		name          string
		kongFile      string
		expectedState utils.KongRawState
	}{
		{
			name:     "creates a service",
			kongFile: "testdata/sync/001-create-a-service/kong.yaml",
			expectedState: utils.KongRawState{
				Services: svc1_207,
			},
		},
		{
			name:     "create services and routes",
			kongFile: "testdata/sync/002-create-services-and-routes/kong.yaml",
			expectedState: utils.KongRawState{
				Services: svc1_207,
				Routes:   route1_20x,
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runWhen(t, "kong", ">2.6.9 <3.0.0")
			setup(t)

			require.NoError(t, sync(context.Background(), tc.kongFile))
			testKongState(t, client, false, false, tc.expectedState, nil)
		})
	}
}

// test scope:
//   - 3.x
func Test_Sync_ServicesRoutes_From_3x(t *testing.T) {
	// setup stage
	client, err := getTestClient()
	require.NoError(t, err)

	tests := []struct {
		name          string
		kongFile      string
		expectedState utils.KongRawState
	}{
		{
			name:     "creates a service",
			kongFile: "testdata/sync/001-create-a-service/kong3x.yaml",
			expectedState: utils.KongRawState{
				Services: svc1_207,
			},
		},
		{
			name:     "create services and routes",
			kongFile: "testdata/sync/002-create-services-and-routes/kong3x.yaml",
			expectedState: utils.KongRawState{
				Services: svc1_207,
				Routes:   route1_20x,
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runWhen(t, "kong", ">=3.0.0")
			setup(t)

			require.NoError(t, sync(context.Background(), tc.kongFile))
			testKongState(t, client, false, false, tc.expectedState, nil)
		})
	}
}

// test scope:
//   - konnect
func Test_Sync_ServicesRoutes_Konnect(t *testing.T) {
	// setup stage
	client, err := getTestClient()
	require.NoError(t, err)

	tests := []struct {
		name          string
		kongFile      string
		expectedState utils.KongRawState
	}{
		{
			name:     "creates a service",
			kongFile: "testdata/sync/001-create-a-service/kong3x.yaml",
			expectedState: utils.KongRawState{
				Services: svc1_207,
			},
		},
		{
			name:     "create services and routes",
			kongFile: "testdata/sync/002-create-services-and-routes/kong3x.yaml",
			expectedState: utils.KongRawState{
				Services: svc1_207,
				Routes:   route1_20x,
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runWhen(t, "konnect", "")
			setup(t)

			require.NoError(t, sync(context.Background(), tc.kongFile))
			testKongState(t, client, true, false, tc.expectedState, nil)
		})
	}
}

// test scope:
//   - 1.4.3
func Test_Sync_BasicAuth_Plugin_1_4_3(t *testing.T) {
	// setup stage
	client, err := getTestClient()
	require.NoError(t, err)

	tests := []struct {
		name            string
		kongFile        string
		initialKongFile string
		expectedState   utils.KongRawState
	}{
		{
			name:     "create a plugin",
			kongFile: "testdata/sync/003-create-a-plugin/kong.yaml",
			expectedState: utils.KongRawState{
				Plugins: plugin_143_151,
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runWhen(t, "kong", "==1.4.3")
			setup(t)

			require.NoError(t, sync(context.Background(), tc.kongFile))
			testKongState(t, client, false, false, tc.expectedState, nil)
		})
	}
}

// test scope:
//   - 1.5.0.11+enterprise
func Test_Sync_BasicAuth_Plugin_Earlier_Than_1_5_1(t *testing.T) {
	// setup stage
	client, err := getTestClient()
	require.NoError(t, err)

	tests := []struct {
		name            string
		kongFile        string
		initialKongFile string
		expectedState   utils.KongRawState
	}{
		{
			name:     "create a plugin",
			kongFile: "testdata/sync/003-create-a-plugin/kong.yaml",
			expectedState: utils.KongRawState{
				Plugins: plugin,
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runWhen(t, "kong", "<1.5.1 !1.4.3")
			setup(t)

			require.NoError(t, sync(context.Background(), tc.kongFile))
			testKongState(t, client, false, false, tc.expectedState, nil)
		})
	}
}

// test scope:
//   - 1.5.1
func Test_Sync_BasicAuth_Plugin_1_5_1(t *testing.T) {
	// setup stage
	client, err := getTestClient()
	require.NoError(t, err)

	tests := []struct {
		name            string
		kongFile        string
		initialKongFile string
		expectedState   utils.KongRawState
	}{
		{
			name:     "create a plugin",
			kongFile: "testdata/sync/003-create-a-plugin/kong.yaml",
			expectedState: utils.KongRawState{
				Plugins: plugin_143_151,
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runWhen(t, "kong", "==1.5.1")
			setup(t)

			require.NoError(t, sync(context.Background(), tc.kongFile))
			testKongState(t, client, false, false, tc.expectedState, nil)
		})
	}
}

// test scope:
//   - 2.0.5
//   - 2.1.4
//   - 2.2.2
//   - 2.3.3
//   - 2.4.1
//   - 2.5.1
//   - 2.6.0
//   - 2.7.0
//   - 2.1.4.6+enterprise
//   - 2.2.1.3+enterprise
//   - 2.3.3.4+enterprise
//   - 2.4.1.3+enterprise
//   - 2.5.1.2+enterprise
//   - 2.6.0.2+enterprise
//   - 2.7.0.0+enterprise
//   - 2.8.0.0+enterprise
func Test_Sync_BasicAuth_Plugin_From_2_0_5_Till_2_8_0(t *testing.T) {
	// setup stage
	client, err := getTestClient()
	require.NoError(t, err)

	tests := []struct {
		name            string
		kongFile        string
		initialKongFile string
		expectedState   utils.KongRawState
	}{
		{
			name:     "create a plugin",
			kongFile: "testdata/sync/003-create-a-plugin/kong.yaml",
			expectedState: utils.KongRawState{
				Plugins: plugin,
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runWhen(t, "kong", ">=2.0.5 <3.0.0")
			setup(t)

			require.NoError(t, sync(context.Background(), tc.kongFile))
			testKongState(t, client, false, false, tc.expectedState, nil)
		})
	}
}

// test scope:
// - >=3.0 <3.6.0
func Test_Sync_BasicAuth_Plugin_From_3x(t *testing.T) {
	// setup stage
	client, err := getTestClient()
	require.NoError(t, err)

	tests := []struct {
		name            string
		kongFile        string
		initialKongFile string
		expectedState   utils.KongRawState
	}{
		{
			name:     "create a plugin",
			kongFile: "testdata/sync/003-create-a-plugin/kong3x.yaml",
			expectedState: utils.KongRawState{
				Plugins: plugin,
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runWhen(t, "kong", ">=3.0.0 <3.6.0")
			setup(t)

			require.NoError(t, sync(context.Background(), tc.kongFile))
			testKongState(t, client, false, false, tc.expectedState, nil)
		})
	}
}

// test scope:
//   - 3.6+
func Test_Sync_BasicAuth_Plugin_From_36(t *testing.T) {
	// setup stage
	client, err := getTestClient()
	require.NoError(t, err)

	tests := []struct {
		name            string
		kongFile        string
		initialKongFile string
		expectedState   utils.KongRawState
	}{
		{
			name:     "create a plugin",
			kongFile: "testdata/sync/003-create-a-plugin/kong3x.yaml",
			expectedState: utils.KongRawState{
				Plugins: plugin36,
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runWhen(t, "kong", ">=3.6.0")
			setup(t)

			require.NoError(t, sync(context.Background(), tc.kongFile))
			testKongState(t, client, false, false, tc.expectedState, nil)
		})
	}
}

// test scope:
//   - konnect
func Test_Sync_BasicAuth_Plugin_Konnect(t *testing.T) {
	// setup stage
	client, err := getTestClient()
	require.NoError(t, err)

	tests := []struct {
		name            string
		kongFile        string
		initialKongFile string
		expectedState   utils.KongRawState
	}{
		{
			name:     "create a plugin",
			kongFile: "testdata/sync/003-create-a-plugin/kong3x.yaml",
			expectedState: utils.KongRawState{
				Plugins: plugin36,
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runWhen(t, "konnect", "")
			setup(t)

			require.NoError(t, sync(context.Background(), tc.kongFile))
			testKongState(t, client, true, false, tc.expectedState, nil)
		})
	}
}

// test scope:
//   - 1.4.3
//   - 1.5.1
//   - 1.5.0.11+enterprise
func Test_Sync_Upstream_Target_Till_1_5_2(t *testing.T) {
	// setup stage
	client, err := getTestClient()
	require.NoError(t, err)

	// ignore entities fields based on Kong version
	ignoreFields := []cmp.Option{
		cmpopts.IgnoreFields(kong.Healthcheck{}, "Threshold"),
	}

	tests := []struct {
		name          string
		kongFile      string
		expectedState utils.KongRawState
	}{
		{
			name:     "creates an upstream and target",
			kongFile: "testdata/sync/004-create-upstream-and-target/kong.yaml",
			expectedState: utils.KongRawState{
				Upstreams: upstream_pre31,
				Targets:   target,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runWhen(t, "kong", "<=1.5.2")
			setup(t)

			require.NoError(t, sync(context.Background(), tc.kongFile))
			testKongState(t, client, false, false, tc.expectedState, ignoreFields)
		})
	}
}

// test scope:
//   - 2.0.5
//   - 2.1.4
//   - 2.2.2
//   - 2.3.3
//   - 2.4.1
//   - 2.5.1
//   - 2.6.0
//   - 2.7.0
//   - 2.1.4.6+enterprise
//   - 2.2.1.3+enterprise
//   - 2.3.3.4+enterprise
//   - 2.4.1.3+enterprise
//   - 2.5.1.2+enterprise
//   - 2.6.0.2+enterprise
//   - 2.7.0.0+enterprise
//   - 2.8.0.0+enterprise
func Test_Sync_Upstream_Target_From_2x(t *testing.T) {
	// setup stage
	client, err := getTestClient()
	require.NoError(t, err)

	tests := []struct {
		name          string
		kongFile      string
		expectedState utils.KongRawState
	}{
		{
			name:     "creates an upstream and target",
			kongFile: "testdata/sync/004-create-upstream-and-target/kong.yaml",
			expectedState: utils.KongRawState{
				Upstreams: upstream_pre31,
				Targets:   target,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runWhen(t, "kong", ">=2.1.0 <3.0.0")
			setup(t)

			require.NoError(t, sync(context.Background(), tc.kongFile))
			testKongState(t, client, false, false, tc.expectedState, nil)
		})
	}
}

// test scope:
//   - 3.0
func Test_Sync_Upstream_Target_From_30(t *testing.T) {
	// setup stage
	client, err := getTestClient()
	require.NoError(t, err)

	tests := []struct {
		name          string
		kongFile      string
		expectedState utils.KongRawState
	}{
		{
			name:     "creates an upstream and target",
			kongFile: "testdata/sync/004-create-upstream-and-target/kong3x.yaml",
			expectedState: utils.KongRawState{
				Upstreams: upstream_pre31,
				Targets:   target,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runWhen(t, "kong", ">=3.0.0 <3.1.0")
			setup(t)

			require.NoError(t, sync(context.Background(), tc.kongFile))
			testKongState(t, client, false, false, tc.expectedState, nil)
		})
	}
}

// test scope:
//   - 3.x
func Test_Sync_Upstream_Target_From_3x(t *testing.T) {
	// setup stage
	client, err := getTestClient()
	require.NoError(t, err)

	tests := []struct {
		name          string
		kongFile      string
		expectedState utils.KongRawState
	}{
		{
			name:     "creates an upstream and target",
			kongFile: "testdata/sync/004-create-upstream-and-target/kong3x.yaml",
			expectedState: utils.KongRawState{
				Upstreams: upstream,
				Targets:   target,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runWhen(t, "kong", ">=3.1.0")
			setup(t)

			require.NoError(t, sync(context.Background(), tc.kongFile))
			testKongState(t, client, false, false, tc.expectedState, nil)
		})
	}
}

// test scope:
//   - konnect
func Test_Sync_Upstream_Target_Konnect(t *testing.T) {
	// setup stage
	client, err := getTestClient()
	require.NoError(t, err)

	tests := []struct {
		name          string
		kongFile      string
		expectedState utils.KongRawState
	}{
		{
			name:     "creates an upstream and target",
			kongFile: "testdata/sync/004-create-upstream-and-target/kong3x.yaml",
			expectedState: utils.KongRawState{
				Upstreams: upstream,
				Targets:   target,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runWhen(t, "konnect", "")
			setup(t)

			require.NoError(t, sync(context.Background(), tc.kongFile))
			testKongState(t, client, true, false, tc.expectedState, nil)
		})
	}
}

// test scope:
//   - 2.4.1
//   - 2.5.1
//   - 2.6.0
//   - 2.7.0
//   - 2.4.1.3+enterprise
//   - 2.5.1.2+enterprise
//   - 2.6.0.2+enterprise
//   - 2.7.0.0+enterprise
//   - 2.8.0.0+enterprise
func Test_Sync_Upstreams_Target_ZeroWeight_2x(t *testing.T) {
	// setup stage
	client, err := getTestClient()
	require.NoError(t, err)

	tests := []struct {
		name          string
		kongFile      string
		expectedState utils.KongRawState
	}{
		{
			name:     "creates an upstream and target with weight equals to zero",
			kongFile: "testdata/sync/005-create-upstream-and-target-weight/kong.yaml",
			expectedState: utils.KongRawState{
				Upstreams: upstream_pre31,
				Targets:   targetZeroWeight,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runWhen(t, "kong", ">=2.4.1 <3.0.0")
			setup(t)

			require.NoError(t, sync(context.Background(), tc.kongFile))
			testKongState(t, client, false, false, tc.expectedState, nil)
		})
	}
}

// test scope:
//   - 3.0
func Test_Sync_Upstreams_Target_ZeroWeight_30(t *testing.T) {
	// setup stage
	client, err := getTestClient()
	require.NoError(t, err)

	tests := []struct {
		name          string
		kongFile      string
		expectedState utils.KongRawState
	}{
		{
			name:     "creates an upstream and target with weight equals to zero",
			kongFile: "testdata/sync/005-create-upstream-and-target-weight/kong3x.yaml",
			expectedState: utils.KongRawState{
				Upstreams: upstream_pre31,
				Targets:   targetZeroWeight,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runWhen(t, "kong", ">=3.0.0 <3.1.0")
			setup(t)

			require.NoError(t, sync(context.Background(), tc.kongFile))
			testKongState(t, client, false, false, tc.expectedState, nil)
		})
	}
}

// test scope:
//   - 3.x
func Test_Sync_Upstreams_Target_ZeroWeight_3x(t *testing.T) {
	// setup stage
	client, err := getTestClient()
	require.NoError(t, err)

	tests := []struct {
		name          string
		kongFile      string
		expectedState utils.KongRawState
	}{
		{
			name:     "creates an upstream and target with weight equals to zero",
			kongFile: "testdata/sync/005-create-upstream-and-target-weight/kong3x.yaml",
			expectedState: utils.KongRawState{
				Upstreams: upstream,
				Targets:   targetZeroWeight,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runWhen(t, "kong", ">=3.1.0")
			setup(t)

			require.NoError(t, sync(context.Background(), tc.kongFile))
			testKongState(t, client, false, false, tc.expectedState, nil)
		})
	}
}

// test scope:
//   - konnect
func Test_Sync_Upstreams_Target_ZeroWeight_Konnect(t *testing.T) {
	// setup stage
	client, err := getTestClient()
	require.NoError(t, err)

	tests := []struct {
		name          string
		kongFile      string
		expectedState utils.KongRawState
	}{
		{
			name:     "creates an upstream and target with weight equals to zero",
			kongFile: "testdata/sync/005-create-upstream-and-target-weight/kong3x.yaml",
			expectedState: utils.KongRawState{
				Upstreams: upstream,
				Targets:   targetZeroWeight,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runWhen(t, "konnect", "")
			setup(t)

			require.NoError(t, sync(context.Background(), tc.kongFile))
			testKongState(t, client, true, false, tc.expectedState, nil)
		})
	}
}

func Test_Sync_RateLimitingPlugin(t *testing.T) {
	// setup stage
	client, err := getTestClient()
	require.NoError(t, err)

	tests := []struct {
		name          string
		kongFile      string
		expectedState utils.KongRawState
	}{
		{
			name:     "fill defaults",
			kongFile: "testdata/sync/006-fill-defaults-rate-limiting/kong.yaml",
			expectedState: utils.KongRawState{
				Plugins: rateLimitingPlugin,
			},
		},
		{
			name:     "fill defaults with dedup",
			kongFile: "testdata/sync/007-fill-defaults-rate-limiting-dedup/kong.yaml",
			expectedState: utils.KongRawState{
				Plugins: rateLimitingPlugin,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runWhen(t, "kong", "==2.7.0")
			setup(t)

			require.NoError(t, sync(context.Background(), tc.kongFile))
			testKongState(t, client, false, false, tc.expectedState, nil)
		})
	}
}

// test scope:
//   - 1.5.0.11+enterprise
func Test_Sync_FillDefaults_Earlier_Than_1_5_1(t *testing.T) {
	// setup stage
	client, err := getTestClient()
	require.NoError(t, err)

	// ignore entities fields based on Kong version
	ignoreFields := []cmp.Option{
		cmpopts.IgnoreFields(kong.Route{}, "Service"),
		cmpopts.IgnoreFields(kong.Healthcheck{}, "Threshold"),
	}

	tests := []struct {
		name          string
		kongFile      string
		expectedState utils.KongRawState
	}{
		{
			name:     "creates a service",
			kongFile: "testdata/sync/008-create-simple-entities/kong.yaml",
			expectedState: utils.KongRawState{
				Services:  svc1,
				Routes:    route1_151,
				Plugins:   plugin,
				Targets:   target,
				Upstreams: upstream_pre31,
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runWhen(t, "kong", "<1.5.1 !1.4.3")
			setup(t)

			require.NoError(t, sync(context.Background(), tc.kongFile))
			testKongState(t, client, false, false, tc.expectedState, ignoreFields)
		})
	}
}

// test scope:
//   - 2.0.5
//   - 2.1.4
func Test_Sync_FillDefaults_From_2_0_5_To_2_1_4(t *testing.T) {
	// setup stage
	client, err := getTestClient()
	require.NoError(t, err)

	tests := []struct {
		name          string
		kongFile      string
		expectedState utils.KongRawState
	}{
		{
			name:     "create services and routes",
			kongFile: "testdata/sync/008-create-simple-entities/kong.yaml",
			expectedState: utils.KongRawState{
				Services:  svc1,
				Routes:    route1_205_214,
				Upstreams: upstream_pre31,
				Targets:   target,
				Plugins:   plugin,
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runWhen(t, "kong", ">=2.0.5 <=2.1.4")
			setup(t)

			require.NoError(t, sync(context.Background(), tc.kongFile))
			testKongState(t, client, false, false, tc.expectedState, nil)
		})
	}
}

// test scope:
//   - 2.2.2
//   - 2.3.3
//   - 2.4.1
//   - 2.5.1
//   - 2.6.0
//   - 2.2.1.3+enterprise
//   - 2.3.3.4+enterprise
//   - 2.4.1.3+enterprise
//   - 2.5.1.2+enterprise
func Test_Sync_FillDefaults_From_2_2_1_to_2_6_0(t *testing.T) {
	// setup stage
	client, err := getTestClient()
	require.NoError(t, err)

	tests := []struct {
		name          string
		kongFile      string
		expectedState utils.KongRawState
	}{
		{
			name:     "create services and routes",
			kongFile: "testdata/sync/008-create-simple-entities/kong.yaml",
			expectedState: utils.KongRawState{
				Services:  svc1,
				Routes:    route1_20x,
				Upstreams: upstream_pre31,
				Targets:   target,
				Plugins:   plugin,
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runWhen(t, "kong", ">2.2.1 <=2.6.0")
			setup(t)

			require.NoError(t, sync(context.Background(), tc.kongFile))
			testKongState(t, client, false, false, tc.expectedState, nil)
		})
	}
}

// test scope:
//   - 2.7.0
//   - 2.6.0.2+enterprise
//   - 2.7.0.0+enterprise
//   - 2.8.0.0+enterprise
func Test_Sync_FillDefaults_From_2_6_9(t *testing.T) {
	// setup stage
	client, err := getTestClient()
	require.NoError(t, err)

	tests := []struct {
		name          string
		kongFile      string
		expectedState utils.KongRawState
	}{
		{
			name:     "creates entities with minimum configuration",
			kongFile: "testdata/sync/008-create-simple-entities/kong.yaml",
			expectedState: utils.KongRawState{
				Services:  svc1_207,
				Routes:    route1_20x,
				Plugins:   plugin,
				Targets:   target,
				Upstreams: upstream_pre31,
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runWhen(t, "kong", ">2.6.9 <3.0.0")
			setup(t)

			require.NoError(t, sync(context.Background(), tc.kongFile))
			testKongState(t, client, false, false, tc.expectedState, nil)
		})
	}
}

func Test_Sync_SkipCACert_2x(t *testing.T) {
	// setup stage
	client, err := getTestClient()
	require.NoError(t, err)

	tests := []struct {
		name          string
		kongFile      string
		expectedState utils.KongRawState
	}{
		{
			name:     "syncing with --skip-ca-certificates should ignore CA certs",
			kongFile: "testdata/sync/009-skip-ca-cert/kong.yaml",
			expectedState: utils.KongRawState{
				Services:       svc1_207,
				CACertificates: []*kong.CACertificate{},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// ca_certificates first appeared in 1.3, but we limit to 2.7+
			// here because the schema changed and the entities aren't the same
			// across all versions, even though the skip functionality works the same.
			runWhen(t, "kong", ">=2.7.0 <3.0.0")
			setup(t)

			require.NoError(t, sync(context.Background(), tc.kongFile, "--skip-ca-certificates"))
			testKongState(t, client, false, false, tc.expectedState, nil)
		})
	}
}

// test scope:
// - >=3.0.0
func Test_Sync_SkipCACert_3x(t *testing.T) {
	// setup stage
	client, err := getTestClient()
	require.NoError(t, err)

	tests := []struct {
		name          string
		kongFile      string
		expectedState utils.KongRawState
	}{
		{
			name:     "syncing with --skip-ca-certificates should ignore CA certs",
			kongFile: "testdata/sync/009-skip-ca-cert/kong3x.yaml",
			expectedState: utils.KongRawState{
				Services:       svc1_207,
				CACertificates: []*kong.CACertificate{},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// ca_certificates first appeared in 1.3, but we limit to 2.7+
			// here because the schema changed and the entities aren't the same
			// across all versions, even though the skip functionality works the same.
			runWhen(t, "kong", ">=3.0.0")
			setup(t)

			require.NoError(t, sync(context.Background(), tc.kongFile, "--skip-ca-certificates"))
			testKongState(t, client, false, false, tc.expectedState, nil)
		})
	}
}

// test scope:
// - konnect
func Test_Sync_SkipCACert_Konnect(t *testing.T) {
	// setup stage
	client, err := getTestClient()
	require.NoError(t, err)

	tests := []struct {
		name          string
		kongFile      string
		expectedState utils.KongRawState
	}{
		{
			name:     "syncing with --skip-ca-certificates should ignore CA certs",
			kongFile: "testdata/sync/009-skip-ca-cert/kong3x.yaml",
			expectedState: utils.KongRawState{
				Services:       svc1_207,
				CACertificates: []*kong.CACertificate{},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// ca_certificates first appeared in 1.3, but we limit to 2.7+
			// here because the schema changed and the entities aren't the same
			// across all versions, even though the skip functionality works the same.
			runWhen(t, "konnect", "")
			setup(t)

			require.NoError(t, sync(context.Background(), tc.kongFile, "--skip-ca-certificates"))
			testKongState(t, client, true, false, tc.expectedState, nil)
		})
	}
}

func Test_Sync_RBAC_2x(t *testing.T) {
	// setup stage
	client, err := getTestClient()
	require.NoError(t, err)

	tests := []struct {
		name          string
		kongFile      string
		expectedState utils.KongRawState
	}{
		{
			name:     "rbac",
			kongFile: "testdata/sync/xxx-rbac-endpoint-permissions/kong.yaml",
			expectedState: utils.KongRawState{
				RBACRoles: []*kong.RBACRole{
					{
						Name:    kong.String("workspace-portal-admin"),
						Comment: kong.String("Full access to Dev Portal related endpoints in the workspace"),
					},
				},
				RBACEndpointPermissions: []*kong.RBACEndpointPermission{
					{
						Workspace: kong.String("default"),
						Endpoint:  kong.String("/developers"),
						Actions:   []*string{kong.String("read"), kong.String("delete"), kong.String("create"), kong.String("update")},
						Negative:  kong.Bool(false),
					},
					{
						Workspace: kong.String("default"),
						Endpoint:  kong.String("/developers/*"),
						Actions:   []*string{kong.String("read"), kong.String("delete"), kong.String("create"), kong.String("update")},
						Negative:  kong.Bool(false),
					},
					{
						Workspace: kong.String("default"),
						Endpoint:  kong.String("/files"),
						Actions:   []*string{kong.String("read"), kong.String("delete"), kong.String("create"), kong.String("update")},
						Negative:  kong.Bool(false),
					},
					{
						Workspace: kong.String("default"),
						Endpoint:  kong.String("/files/*"),
						Actions:   []*string{kong.String("read"), kong.String("delete"), kong.String("create"), kong.String("update")},
						Negative:  kong.Bool(false),
					},
					{
						Workspace: kong.String("default"),
						Endpoint:  kong.String("/kong"),
						Actions:   []*string{kong.String("read"), kong.String("delete"), kong.String("create"), kong.String("update")},
						Negative:  kong.Bool(false),
					},
					{
						Workspace: kong.String("default"),
						Endpoint:  kong.String("/rbac/*"),
						Actions:   []*string{kong.String("read"), kong.String("delete"), kong.String("create"), kong.String("update")},
						Negative:  kong.Bool(true),
					},
					{
						Workspace: kong.String("default"),
						Endpoint:  kong.String("/rbac/*/*"),
						Actions:   []*string{kong.String("read"), kong.String("delete"), kong.String("create"), kong.String("update")},
						Negative:  kong.Bool(true),
					},
					{
						Workspace: kong.String("default"),
						Endpoint:  kong.String("/rbac/*/*/*"),
						Actions:   []*string{kong.String("read"), kong.String("delete"), kong.String("create"), kong.String("update")},
						Negative:  kong.Bool(true),
					},
					{
						Workspace: kong.String("default"),
						Endpoint:  kong.String("/rbac/*/*/*/*"),
						Actions:   []*string{kong.String("read"), kong.String("delete"), kong.String("create"), kong.String("update")},
						Negative:  kong.Bool(true),
					},
					{
						Workspace: kong.String("default"),
						Endpoint:  kong.String("/rbac/*/*/*/*/*"),
						Actions:   []*string{kong.String("read"), kong.String("delete"), kong.String("create"), kong.String("update")},
						Negative:  kong.Bool(true),
					},
					{
						Workspace: kong.String("default"),
						Endpoint:  kong.String("/workspaces/default"),
						Actions:   []*string{kong.String("read"), kong.String("update")},
						Negative:  kong.Bool(false),
					},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runWhen(t, "enterprise", ">=2.7.0 <3.0.0")
			setup(t)

			require.NoError(t, sync(context.Background(), tc.kongFile, "--rbac-resources-only"))
			testKongState(t, client, false, false, tc.expectedState, nil)
		})
	}
}

func Test_Sync_RBAC_3x(t *testing.T) {
	// setup stage
	client, err := getTestClient()
	require.NoError(t, err)

	tests := []struct {
		name          string
		kongFile      string
		expectedState utils.KongRawState
	}{
		{
			name:     "rbac",
			kongFile: "testdata/sync/xxx-rbac-endpoint-permissions/kong3x.yaml",
			expectedState: utils.KongRawState{
				RBACRoles: []*kong.RBACRole{
					{
						Name:    kong.String("workspace-portal-admin"),
						Comment: kong.String("Full access to Dev Portal related endpoints in the workspace"),
					},
				},
				RBACEndpointPermissions: []*kong.RBACEndpointPermission{
					{
						Workspace: kong.String("default"),
						Endpoint:  kong.String("/developers"),
						Actions:   []*string{kong.String("read"), kong.String("delete"), kong.String("create"), kong.String("update")},
						Negative:  kong.Bool(false),
					},
					{
						Workspace: kong.String("default"),
						Endpoint:  kong.String("/developers/*"),
						Actions:   []*string{kong.String("read"), kong.String("delete"), kong.String("create"), kong.String("update")},
						Negative:  kong.Bool(false),
					},
					{
						Workspace: kong.String("default"),
						Endpoint:  kong.String("/files"),
						Actions:   []*string{kong.String("read"), kong.String("delete"), kong.String("create"), kong.String("update")},
						Negative:  kong.Bool(false),
					},
					{
						Workspace: kong.String("default"),
						Endpoint:  kong.String("/files/*"),
						Actions:   []*string{kong.String("read"), kong.String("delete"), kong.String("create"), kong.String("update")},
						Negative:  kong.Bool(false),
					},
					{
						Workspace: kong.String("default"),
						Endpoint:  kong.String("/kong"),
						Actions:   []*string{kong.String("read"), kong.String("delete"), kong.String("create"), kong.String("update")},
						Negative:  kong.Bool(false),
					},
					{
						Workspace: kong.String("default"),
						Endpoint:  kong.String("/rbac/*"),
						Actions:   []*string{kong.String("read"), kong.String("delete"), kong.String("create"), kong.String("update")},
						Negative:  kong.Bool(true),
					},
					{
						Workspace: kong.String("default"),
						Endpoint:  kong.String("/rbac/*/*"),
						Actions:   []*string{kong.String("read"), kong.String("delete"), kong.String("create"), kong.String("update")},
						Negative:  kong.Bool(true),
					},
					{
						Workspace: kong.String("default"),
						Endpoint:  kong.String("/rbac/*/*/*"),
						Actions:   []*string{kong.String("read"), kong.String("delete"), kong.String("create"), kong.String("update")},
						Negative:  kong.Bool(true),
					},
					{
						Workspace: kong.String("default"),
						Endpoint:  kong.String("/rbac/*/*/*/*"),
						Actions:   []*string{kong.String("read"), kong.String("delete"), kong.String("create"), kong.String("update")},
						Negative:  kong.Bool(true),
					},
					{
						Workspace: kong.String("default"),
						Endpoint:  kong.String("/rbac/*/*/*/*/*"),
						Actions:   []*string{kong.String("read"), kong.String("delete"), kong.String("create"), kong.String("update")},
						Negative:  kong.Bool(true),
					},
					{
						Workspace: kong.String("default"),
						Endpoint:  kong.String("/workspaces/default"),
						Actions:   []*string{kong.String("read"), kong.String("update")},
						Negative:  kong.Bool(false),
					},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runWhen(t, "enterprise", ">=3.0.0")
			setup(t)

			require.NoError(t, sync(context.Background(), tc.kongFile, "--rbac-resources-only"))
			testKongState(t, client, false, false, tc.expectedState, nil)
		})
	}
}

func Test_Sync_Create_Route_With_Service_Name_Reference_2x(t *testing.T) {
	// setup stage
	client, err := getTestClient()
	require.NoError(t, err)

	tests := []struct {
		name          string
		kongFile      string
		expectedState utils.KongRawState
	}{
		{
			name:     "create a route with a service name reference",
			kongFile: "testdata/sync/010-create-route-with-service-name-reference/kong.yaml",
			expectedState: utils.KongRawState{
				Services: svc1_207,
				Routes:   route1_20x,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runWhen(t, "kong", ">=2.7.0 <3.0.0")
			setup(t)

			require.NoError(t, sync(context.Background(), tc.kongFile))
			testKongState(t, client, false, false, tc.expectedState, nil)
		})
	}
}

func Test_Sync_Create_Route_With_Service_Name_Reference_3x(t *testing.T) {
	// setup stage
	client, err := getTestClient()
	require.NoError(t, err)

	tests := []struct {
		name          string
		kongFile      string
		expectedState utils.KongRawState
	}{
		{
			name:     "create a route with a service name reference",
			kongFile: "testdata/sync/010-create-route-with-service-name-reference/kong3x.yaml",
			expectedState: utils.KongRawState{
				Services: svc1_207,
				Routes:   route1_20x,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runWhen(t, "kong", ">=2.7.0 <3.0.0")
			setup(t)

			require.NoError(t, sync(context.Background(), tc.kongFile))
			testKongState(t, client, false, false, tc.expectedState, nil)
		})
	}
}

// test scope:
//   - 1.x.x
//   - 2.x.x
func Test_Sync_PluginsOnEntitiesTill_3_0_0(t *testing.T) {
	// setup stage
	client, err := getTestClient()
	require.NoError(t, err)

	tests := []struct {
		name          string
		kongFile      string
		expectedState utils.KongRawState
	}{
		{
			name:     "create plugins on services, routes and consumers",
			kongFile: "testdata/sync/xxx-plugins-on-entities/kong.yaml",
			expectedState: utils.KongRawState{
				Services:  svc1_207,
				Routes:    route1_20x,
				Plugins:   plugin_on_entities,
				Consumers: consumer,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runWhen(t, "kong", ">=2.8.0 <3.0.0")
			setup(t)

			require.NoError(t, sync(context.Background(), tc.kongFile))
			testKongState(t, client, false, false, tc.expectedState, nil)
		})
	}
}

// test scope:
//   - 3.0.0+
func Test_Sync_PluginsOnEntitiesFrom_3_0_0(t *testing.T) {
	// setup stage
	client, err := getTestClient()
	require.NoError(t, err)

	tests := []struct {
		name          string
		kongFile      string
		expectedState utils.KongRawState
		runWhen       string
	}{
		{
			name:     "create plugins on services, routes and consumers <3.8.0",
			kongFile: "testdata/sync/xxx-plugins-on-entities/kong.yaml",
			expectedState: utils.KongRawState{
				Services:  svc1_207,
				Routes:    route1_20x,
				Plugins:   plugin_on_entities3x,
				Consumers: consumer,
			},
			runWhen: ">=3.0.0 <3.8.0",
		},
		{
			name:     "create plugins on services, routes and consumers >=3.8.0 <3.10.0",
			kongFile: "testdata/sync/xxx-plugins-on-entities/kong.yaml",
			expectedState: utils.KongRawState{
				Services:  svc1_207,
				Routes:    route1_20x,
				Plugins:   plugin_on_entities381x,
				Consumers: consumer,
			},
			runWhen: ">=3.8.0 <3.10.0",
		},
		{
			name:     "create plugins on services, routes and consumers >=3.10.0",
			kongFile: "testdata/sync/xxx-plugins-on-entities/kong.yaml",
			expectedState: utils.KongRawState{
				Services:  svc1_207,
				Routes:    route1_20x,
				Plugins:   plugin_on_entities310x,
				Consumers: consumer,
			},
			runWhen: ">=3.10.0",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runWhen(t, "kong", tc.runWhen)
			setup(t)

			require.NoError(t, sync(context.Background(), tc.kongFile))
			testKongState(t, client, false, false, tc.expectedState, nil)
		})
	}
}

// test scope:
//   - konnect
func Test_Sync_PluginsOnEntities_Konnect(t *testing.T) {
	// setup stage
	client, err := getTestClient()
	require.NoError(t, err)

	tests := []struct {
		name          string
		kongFile      string
		expectedState utils.KongRawState
	}{
		{
			name:     "create plugins on services, routes and consumers",
			kongFile: "testdata/sync/xxx-plugins-on-entities/kong.yaml",
			expectedState: utils.KongRawState{
				Services:  svc1_207,
				Routes:    route1_20x,
				Plugins:   plugin_on_entitiesKonnect,
				Consumers: consumer,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runWhenKonnect(t)
			setup(t)

			require.NoError(t, sync(context.Background(), tc.kongFile))
			testKongState(t, client, true, false, tc.expectedState, nil)
		})
	}
}

// test scope:
//   - 3.0.0+
func Test_Sync_PluginOrdering(t *testing.T) {
	// setup stage
	client, err := getTestClient()
	require.NoError(t, err)

	tests := []struct {
		name            string
		kongFile        string
		initialKongFile string
		expectedState   utils.KongRawState
	}{
		{
			name:     "create a plugin with ordering",
			kongFile: "testdata/sync/011-plugin-ordering/kong.yaml",
			expectedState: utils.KongRawState{
				Plugins: []*kong.Plugin{
					{
						Name: kong.String("request-termination"),
						Protocols: []*string{
							kong.String("grpc"),
							kong.String("grpcs"),
							kong.String("http"),
							kong.String("https"),
						},
						Enabled: kong.Bool(true),
						Config: kong.Configuration{
							"status_code":  float64(200),
							"echo":         false,
							"content_type": nil,
							"body":         nil,
							"message":      nil,
							"trigger":      nil,
						},
						Ordering: &kong.PluginOrdering{
							Before: kong.PluginOrderingPhase{
								"access": []string{"basic-auth"},
							},
						},
					},
				},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runWhen(t, "enterprise", ">=3.0.0")
			setup(t)

			require.NoError(t, sync(context.Background(), tc.kongFile))
			testKongState(t, client, false, false, tc.expectedState, nil)
		})
	}
}

// test scope:
//   - 3.x
func Test_Sync_Unsupported_Formats(t *testing.T) {
	tests := []struct {
		name          string
		kongFile      string
		expectedError error
	}{
		{
			name:     "creates a service",
			kongFile: "testdata/sync/001-create-a-service/kong.yaml",
			expectedError: errors.New(
				"cannot apply '1.1' config format version to Kong version 3.0 or above.\n" +
					utils.UpgradeMessage),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runWhen(t, "kong", ">=3.0.0")
			setup(t)

			err := sync(context.Background(), tc.kongFile)
			assert.Equal(t, tc.expectedError, err)
		})
	}
}

var (
	goodCACertPEM = []byte(`-----BEGIN CERTIFICATE-----
MIIE6DCCAtACCQCjgi452nKnUDANBgkqhkiG9w0BAQsFADA2MQswCQYDVQQGEwJV
UzETMBEGA1UECAwKQ2FsaWZvcm5pYTESMBAGA1UEAwwJbG9jYWxob3N0MB4XDTIy
MTAwNDE4NTEyOFoXDTMyMTAwMTE4NTEyOFowNjELMAkGA1UEBhMCVVMxEzARBgNV
BAgMCkNhbGlmb3JuaWExEjAQBgNVBAMMCWxvY2FsaG9zdDCCAiIwDQYJKoZIhvcN
AQEBBQADggIPADCCAgoCggIBALUwleXMo+CxQFvgtmJbWHO4k3YBJwzWqcr2xWn+
vgeoLiKFDQC11F/nnWNKkPZyilLeJda5c9YEVaA9IW6/PZhxQ430RM53EJHoiIPB
B9j7BHGzsvWYHEkjXvGQWeD3mR4TAkoCVTfPAjBji/SL+WvLpgPW5hKRVuedD8ja
cTvkNfk6u2TwPYGgekh9+wS9zcEQs4OwsEiQxmi3Z8if1m1uD09tjqAHb0klPEzM
64tPvlzJrIcH3Z5iF+B9qr91PCQJVYOCjGWlUgPULaqIoTVtY+AnaNnNcol0LM/i
oq7uD0JbeyIFDFMDJVqZwDf/zowzLLlP8Hkok4M8JTefXvB0puQoxmGwOAhwlA0G
KF5etrmhg+dOb+f3nWdgbyjPEytyOeMOOA/4Lb8dHRlf9JnEc4DJqwRVPM9BMeUu
9ZlrSWvURRk8nUZfkjTstLqO2aeubfOvb+tDKUq5Ue2B+AFs0ETLy3bds8TU9syV
5Kl+tIwek2TXzc7afvmeCDoRunAx5nVhmW8dpGhknOmJM0GxOi5s2tiu8/3T9XdH
WcH/GMrocZrkhvzkZccSLYoo1jcDn9LwxHVr/BZ43NymjVa6T3QRTta4Kg5wWpfS
yXi4gIW7VJM12CmNfSDEXqhF03+fjFzoWH+YfBK/9GgUMNjnXWIL9PgFFOBomwEL
tv5zAgMBAAEwDQYJKoZIhvcNAQELBQADggIBAKH8eUGgH/OSS3mHB3Gqv1m2Ea04
Cs03KNEt1weelcHIBWVnPp+jGcSIIfMBnDFAwgxtBKhwptJ9ZKXIzjh7YFxbOT01
NU+KQ6tD+NFDf+SAUC4AWV9Cam63JIaCVNDoo5UjVMlssnng7NefM1q2+ucoP+gs
+bvUCTJcp3FZsq8aUI9Rka575HqRhl/8kyhcwICCgT5UHQJvCQYrInJ0Faem6dr0
tHw+PZ1bo6qB7uxBjK9kyu7dK/vEKliUGM4/MXMDKIc5qXUs47wPLbjxvKsuDglK
KftgUWNYRxx9Bf9ylbjd+ayo3+1Lb9cbvdZnh0UHN6677NvXlWNheCmeysLGQHtm
5H6iIhZ75r6QuC7m6hBSJYtLU3fsQECrmaS/+xBGoSSZjacciO7b7qjQdWOfQREn
7vc5eu0N+CJkp8t3SsyQP6v2Su3ILeTt2EWrmmE4K7SYlJe1HrUVj0AWUwzLa6+Z
+Dx16p3M0RBdFMGNNhLqvG3WRfE5c5md34Aq/C5ePjN7pQGmJhI6weowuX9wCrnh
nJJJRfqyJvqgnVBZ6IawNcOyIofITZHlYVKuaDB1odzWCDNEvFftgJvH0MnO7OY9
Pb9hILPoCy+91jQAVh6Z/ghIcZKHV+N6zV3uS3t5vCejhCNK8mUPSOwAeDf3Bq5r
wQPXd0DdsYGmXVIh
-----END CERTIFICATE-----`)

	badCACertPEM = []byte(`-----BEGIN CERTIFICATE-----
MIIDkzCCAnugAwIBAgIUYGc07pbHSjOBPreXh7OcNT2+sD4wDQYJKoZIhvcNAQEL
BQAwWTELMAkGA1UEBhMCVVMxCzAJBgNVBAgMAkNBMRUwEwYDVQQKDAxZb2xvNDIs
IEluYy4xJjAkBgNVBAMMHVlvbG80MiBzZWxmLXNpZ25lZCB0ZXN0aW5nIENBMB4X
DTIyMDMyOTE5NDczM1oXDTMyMDMyNjE5NDczM1owWTELMAkGA1UEBhMCVVMxCzAJ
BgNVBAgMAkNBMRUwEwYDVQQKDAxZb2xvNDIsIEluYy4xJjAkBgNVBAMMHVlvbG80
MiBzZWxmLXNpZ25lZCB0ZXN0aW5nIENBMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8A
MIIBCgKCAQEAvnhTgdJALnuLKDA0ZUZRVMqcaaC+qvfJkiEFGYwX2ZJiFtzU65F/
sB2L0ToFqY4tmMVlOmiSZFnRLDZecmQDbbNwc3wtNikmxIOzx4qR4kbRP8DDdyIf
gaNmGCuaXTM5+FYy2iNBn6CeibIjqdErQlAbFLwQs5t3mLsjii2U4cyvfRtO+0RV
HdJ6Np5LsVziN0c5gVIesIrrbxLcOjtXDzwd/w/j5NXqL/OwD5EBH2vqd3QKKX4t
s83BLl2EsbUse47VAImavrwDhmV6S/p/NuJHqjJ6dIbXLYxNS7g26ijcrXxvNhiu
YoZTykSgdI3BXMNAm1ahP/BtJPZpU7CVdQIDAQABo1MwUTAdBgNVHQ4EFgQUe1WZ
fMfZQ9QIJIttwTmcrnl40ccwHwYDVR0jBBgwFoAUe1WZfMfZQ9QIJIttwTmcrnl4
0ccwDwYDVR0TAQH/BAUwAwEB/zANBgkqhkiG9w0BAQsFAAOCAQEAs4Z8VYbvEs93
haTHdbbaKk0V6xAL/Q8I8GitK9E8cgf8C5rwwn+wU/Gf39dtMUlnW8uxyzRPx53u
CAAcJAWkabT+xwrlrqjO68H3MgIAwgWA5yZC+qW7ECA8xYEK6DzEHIaOpagJdKcL
IaZr/qTJlEQClvwDs4x/BpHRB5XbmJs86GqEB7XWAm+T2L8DluHAXvek+welF4Xo
fQtLlNS/vqTDqPxkSbJhFv1L7/4gdwfAz51wH/iL7AG/ubFEtoGZPK9YCJ40yTWz
8XrUoqUC+2WIZdtmo6dFFJcLfQg4ARJZjaK6lmxJun3iRMZjKJdQKm/NEKz4y9kA
u8S6yNlu2Q==
-----END CERTIFICATE-----`)
)

// test scope:
//   - 3.0.0+
//
// This test does two things:
// 1. makes sure decK can correctly configure a Vault entity
// 2. makes sure secrets management works as expected end-to-end
//
// Specifically, for (2) we make use of:
// - a Service and a Route to verify the overall flow works end-to-end
// - a Certificate with secret references
// - an {env} Vault using 'MY_SECRET_' as env variables prefix
//
// The Kong EE instance running in the CI includes the MY_SECRET_CERT
// and MY_SECRET_KEY env variables storing cert/key signed with `caCert`.
// These variables are pulled into the {env} Vault after decK deploy
// the configuration.
//
// After the `deck sync` and the configuration verification step,
// an HTTPS client is created using the `caCert` used to sign the
// deployed certificate, and then a GET is performed to test the
// proxy functionality, which should return a 200.
func Test_Sync_Vault(t *testing.T) {
	// setup stage
	client, err := getTestClient()
	require.NoError(t, err)

	tests := []struct {
		name            string
		kongFile        string
		initialKongFile string
		expectedState   utils.KongRawState
	}{
		{
			name:     "create an SSL service/route using an ENV vault",
			kongFile: "testdata/sync/012-vaults/kong3x.yaml",
			expectedState: utils.KongRawState{
				Vaults: []*kong.Vault{
					{
						Name:        kong.String("env"),
						Prefix:      kong.String("my-env-vault"),
						Description: kong.String("ENV vault for secrets"),
						Config: kong.Configuration{
							"prefix": "MY_SECRET_",
						},
					},
				},
				Services: []*kong.Service{
					{
						ID:             kong.String("58076db2-28b6-423b-ba39-a797193017f7"),
						Name:           kong.String("svc1"),
						ConnectTimeout: kong.Int(60000),
						Host:           kong.String("httpbin.org"),
						Port:           kong.Int(80),
						Path:           kong.String("/status/200"),
						Protocol:       kong.String("http"),
						ReadTimeout:    kong.Int(60000),
						Retries:        kong.Int(5),
						WriteTimeout:   kong.Int(60000),
						Tags:           nil,
						Enabled:        kong.Bool(true),
					},
				},
				Routes: route1_20x,
				Certificates: []*kong.Certificate{
					{
						ID:   kong.String("13c562a1-191c-4464-9b18-e5222b46035b"),
						Cert: kong.String("{vault://my-env-vault/cert}"),
						Key:  kong.String("{vault://my-env-vault/key}"),
					},
				},
				SNIs: []*kong.SNI{
					{
						Name: kong.String("localhost"),
						Certificate: &kong.Certificate{
							ID: kong.String("13c562a1-191c-4464-9b18-e5222b46035b"),
						},
					},
				},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runWhen(t, "enterprise", ">=3.0.0")
			setup(t)

			require.NoError(t, sync(context.Background(), tc.kongFile))
			testKongState(t, client, false, false, tc.expectedState, nil)

			// Kong proxy may need a bit to be ready.
			time.Sleep(time.Second * 5)

			// build simple http client
			client := &http.Client{}

			// use simple http client with https should result
			// in a failure due missing certificate.
			_, err := client.Get("https://localhost:8443/r1")
			require.Error(t, err)

			// use transport with wrong CA cert this should result
			// in a failure due to unknown authority.
			badCACertPool := x509.NewCertPool()
			badCACertPool.AppendCertsFromPEM(badCACertPEM)

			client = &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						RootCAs:    badCACertPool,
						ClientAuth: tls.RequireAndVerifyClientCert,
					},
				},
			}

			_, err = client.Get("https://localhost:8443/r1")
			require.Error(t, err)

			// use transport with good CA cert should pass
			// if referenced secrets are resolved correctly
			// using the ENV vault.
			goodCACertPool := x509.NewCertPool()
			goodCACertPool.AppendCertsFromPEM(goodCACertPEM)

			client = &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						RootCAs:    goodCACertPool,
						ClientAuth: tls.RequireAndVerifyClientCert,
					},
				},
			}

			res, err := client.Get("https://localhost:8443/r1")
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, res.StatusCode)
		})
	}
}

// test scope:
//   - 2.8.x
func Test_Sync_UpdateUsernameInConsumerWithCustomID(t *testing.T) {
	// setup stage
	client, err := getTestClient()
	require.NoError(t, err)

	tests := []struct {
		name            string
		kongFile        string
		kongFileInitial string
		expectedState   utils.KongRawState
	}{
		{
			name:            "update username on a consumer with custom_id",
			kongFile:        "testdata/sync/013-update-username-consumer-with-custom-id/kong.yaml",
			kongFileInitial: "testdata/sync/013-update-username-consumer-with-custom-id/kong-initial.yaml",
			expectedState: utils.KongRawState{
				Consumers: []*kong.Consumer{
					{
						Username: kong.String("test_new"),
						CustomID: kong.String("custom_test"),
					},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runWhen(t, "kong", ">=2.8.0 <3.0.0")
			setup(t)

			// set up initial state
			require.NoError(t, sync(context.Background(), tc.kongFileInitial))
			// update with desired final state
			require.NoError(t, sync(context.Background(), tc.kongFile))
			testKongState(t, client, false, false, tc.expectedState, nil)
		})
	}
}

// test scope:
//   - 2.8.x
func Test_Sync_UpdateConsumerWithCustomID(t *testing.T) {
	// setup stage
	client, err := getTestClient()
	require.NoError(t, err)

	tests := []struct {
		name            string
		kongFile        string
		kongFileInitial string
		expectedState   utils.KongRawState
	}{
		{
			name:            "update username on a consumer with custom_id",
			kongFile:        "testdata/sync/014-update-consumer-with-custom-id/kong.yaml",
			kongFileInitial: "testdata/sync/014-update-consumer-with-custom-id/kong-initial.yaml",
			expectedState: utils.KongRawState{
				Consumers: []*kong.Consumer{
					{
						Username: kong.String("test"),
						CustomID: kong.String("new_custom_test"),
					},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runWhen(t, "kong", ">=2.8.0 <3.0.0")
			setup(t)

			// set up initial state
			require.NoError(t, sync(context.Background(), tc.kongFileInitial))
			// update with desired final state
			require.NoError(t, sync(context.Background(), tc.kongFile))
			testKongState(t, client, false, false, tc.expectedState, nil)
		})
	}
}

// test scope:
//   - 3.x
func Test_Sync_UpdateUsernameInConsumerWithCustomID_3x(t *testing.T) {
	// setup stage
	client, err := getTestClient()
	require.NoError(t, err)

	tests := []struct {
		name            string
		kongFile        string
		kongFileInitial string
		expectedState   utils.KongRawState
	}{
		{
			name:            "update username on a consumer with custom_id",
			kongFile:        "testdata/sync/013-update-username-consumer-with-custom-id/kong3x.yaml",
			kongFileInitial: "testdata/sync/013-update-username-consumer-with-custom-id/kong3x-initial.yaml",
			expectedState: utils.KongRawState{
				Consumers: []*kong.Consumer{
					{
						Username: kong.String("test_new"),
						CustomID: kong.String("custom_test"),
					},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runWhen(t, "kong", ">=3.0.0")
			setup(t)

			// set up initial state
			require.NoError(t, sync(context.Background(), tc.kongFileInitial))
			// update with desired final state
			require.NoError(t, sync(context.Background(), tc.kongFile))
			testKongState(t, client, false, false, tc.expectedState, nil)
		})
	}
}

// test scope:
//   - konnect
func Test_Sync_UpdateUsernameInConsumerWithCustomID_Konnect(t *testing.T) {
	// setup stage
	client, err := getTestClient()
	require.NoError(t, err)

	tests := []struct {
		name            string
		kongFile        string
		kongFileInitial string
		expectedState   utils.KongRawState
	}{
		{
			name:            "update username on a consumer with custom_id",
			kongFile:        "testdata/sync/013-update-username-consumer-with-custom-id/kong3x.yaml",
			kongFileInitial: "testdata/sync/013-update-username-consumer-with-custom-id/kong3x-initial.yaml",
			expectedState: utils.KongRawState{
				Consumers: []*kong.Consumer{
					{
						Username: kong.String("test_new"),
						CustomID: kong.String("custom_test"),
					},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runWhen(t, "konnect", "")
			setup(t)

			// set up initial state
			require.NoError(t, sync(context.Background(), tc.kongFileInitial))
			// update with desired final state
			require.NoError(t, sync(context.Background(), tc.kongFile))
			testKongState(t, client, true, false, tc.expectedState, nil)
		})
	}
}

// test scope:
//   - 3.x
func Test_Sync_UpdateConsumerWithCustomID_3x(t *testing.T) {
	// setup stage
	client, err := getTestClient()
	require.NoError(t, err)

	tests := []struct {
		name            string
		kongFile        string
		kongFileInitial string
		expectedState   utils.KongRawState
	}{
		{
			name:            "update username on a consumer with custom_id",
			kongFile:        "testdata/sync/014-update-consumer-with-custom-id/kong3x.yaml",
			kongFileInitial: "testdata/sync/014-update-consumer-with-custom-id/kong3x-initial.yaml",
			expectedState: utils.KongRawState{
				Consumers: []*kong.Consumer{
					{
						Username: kong.String("test_consumer_3x"),
						CustomID: kong.String("test_consumer_3x_custom_test"),
					},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runWhen(t, "kong", ">=3.0.0")
			setup(t)

			// set up initial state
			require.NoError(t, sync(context.Background(), tc.kongFileInitial))
			// update with desired final state
			require.NoError(t, sync(context.Background(), tc.kongFile))
			testKongState(t, client, false, false, tc.expectedState, nil)
		})
	}
}

// test scope:
//   - konnect
func Test_Sync_UpdateConsumerWithCustomID_Konnect(t *testing.T) {
	// setup stage
	client, err := getTestClient()
	require.NoError(t, err)

	tests := []struct {
		name            string
		kongFile        string
		kongFileInitial string
		expectedState   utils.KongRawState
	}{
		{
			name:            "update username on a consumer with custom_id",
			kongFile:        "testdata/sync/014-update-consumer-with-custom-id/kong3x.yaml",
			kongFileInitial: "testdata/sync/014-update-consumer-with-custom-id/kong3x-initial.yaml",
			expectedState: utils.KongRawState{
				Consumers: []*kong.Consumer{
					{
						Username: kong.String("test_consumer_3x"),
						CustomID: kong.String("test_consumer_3x_custom_test"),
					},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runWhen(t, "konnect", "")
			setup(t)

			// set up initial state
			require.NoError(t, sync(context.Background(), tc.kongFileInitial))
			// update with desired final state
			require.NoError(t, sync(context.Background(), tc.kongFile))
			testKongState(t, client, true, false, tc.expectedState, nil)
		})
	}
}

// test scope:
//   - 2.7+
func Test_Sync_ConsumerGroupsTill30(t *testing.T) {
	client, err := getTestClient()
	require.NoError(t, err)
	tests := []struct {
		name          string
		kongFile      string
		expectedState utils.KongRawState
	}{
		{
			name:     "creates consumer groups",
			kongFile: "testdata/sync/015-consumer-groups/kong.yaml",
			expectedState: utils.KongRawState{
				Consumers:      consumerGroupsConsumers,
				ConsumerGroups: consumerGroups,
			},
		},
		{
			name:     "creates consumer groups and plugin",
			kongFile: "testdata/sync/016-consumer-groups-and-plugins/kong.yaml",
			expectedState: utils.KongRawState{
				Consumers:      consumerGroupsConsumers,
				ConsumerGroups: consumerGroupsWithRLA,
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runWhen(t, "enterprise", ">=2.7.0 <3.0.0")
			setup(t)

			require.NoError(t, sync(context.Background(), tc.kongFile))
			testKongState(t, client, false, false, tc.expectedState, nil)
		})
	}
}

// test scope:
//   - 3.1
func Test_Sync_ConsumerGroups_31(t *testing.T) {
	client, err := getTestClient()
	require.NoError(t, err)
	tests := []struct {
		name            string
		kongFile        string
		kongFileInitial string
		expectedState   utils.KongRawState
	}{
		{
			name:            "creates consumer groups",
			kongFile:        "testdata/sync/015-consumer-groups/kong3x.yaml",
			kongFileInitial: "testdata/sync/015-consumer-groups/kong3x-initial.yaml",
			expectedState: utils.KongRawState{
				Consumers:      consumerGroupsConsumers,
				ConsumerGroups: consumerGroupsWithTags,
			},
		},
		{
			name:            "creates consumer groups and plugin",
			kongFile:        "testdata/sync/016-consumer-groups-and-plugins/kong3x.yaml",
			kongFileInitial: "testdata/sync/016-consumer-groups-and-plugins/kong3x-initial.yaml",
			expectedState: utils.KongRawState{
				Consumers:      consumerGroupsConsumers,
				ConsumerGroups: consumerGroupsWithTagsAndRLA,
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runWhen(t, "enterprise", "==3.1.0")
			setup(t)

			// set up initial state
			require.NoError(t, sync(context.Background(), tc.kongFileInitial))
			// update with desired final state
			require.NoError(t, sync(context.Background(), tc.kongFile))

			testKongState(t, client, false, false, tc.expectedState, nil)
		})
	}
}

// This test has 2 goals:
//   - make sure consumer groups and their related properties
//     can be configured correctly in Kong
//   - the actual consumer groups functionality works once set
//
// This is achieved via configuring:
// - 3 consumers:
//   - 1 belonging to Gold Consumer Group
//   - 1 belonging to Silver Consumer Group
//   - 1 not belonging to any Consumer Group
//
// - 3 key-auths, one for each consumer
// - 1 global key-auth plugin
// - 1 global RLA plugin
// - 2 consumer group
// - 2 RLA override, 1 for each consumer group
// - 1 service pointing to mockbin.org
// - 1 route proxying the above service
//
// Once the configuration is verified to be matching in Kong,
// we then check whether the override is correctly applied: consumers
// not belonging to the consumer group should be limited to 5 requests
// every 30s, while consumers belonging to the 'gold' and 'silver' consumer groups
// should be allowed to run respectively 10 and 7 requests in the same timeframe.
// In order to make sure this is the case, we run requests in a loop
// for all consumers and then check at what point they start to receive 429.
func Test_Sync_ConsumerGroupsRLAFrom31(t *testing.T) {
	const (
		maxGoldRequestsNumber    = 10
		maxSilverRequestsNumber  = 7
		maxRegularRequestsNumber = 5
	)
	client, err := getTestClient()
	require.NoError(t, err)
	require.NoError(t, err)
	tests := []struct {
		name          string
		kongFile      string
		expectedState utils.KongRawState
	}{
		{
			name:     "creates consumer groups application",
			kongFile: "testdata/sync/017-consumer-groups-rla-application/kong3x.yaml",
			expectedState: utils.KongRawState{
				Consumers:      consumerGroupsConsumers,
				ConsumerGroups: consumerGroupsWithRLAApp,
				Plugins:        consumerGroupAppPlugins,
				Services:       svc1_207,
				Routes:         route1_20x,
				KeyAuths: []*kong.KeyAuth{
					{
						Consumer: &kong.Consumer{
							ID: kong.String("87095815-5395-454e-8c18-a11c9bc0ef04"),
						},
						Key: kong.String("i-am-special"),
					},
					{
						Consumer: &kong.Consumer{
							ID: kong.String("5a5b9369-baeb-4faa-a902-c40ccdc2928e"),
						},
						Key: kong.String("i-am-not-so-special"),
					},
					{
						Consumer: &kong.Consumer{
							ID: kong.String("e894ea9e-ad08-4acf-a960-5a23aa7701c7"),
						},
						Key: kong.String("i-am-just-average"),
					},
				},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runWhen(t, "enterprise", ">=3.0.0 <3.1.0")
			setup(t)

			require.NoError(t, sync(context.Background(), tc.kongFile))
			testKongState(t, client, false, false, tc.expectedState, nil)

			// Kong proxy may need a bit to be ready.
			time.Sleep(time.Second * 10)

			// build simple http client
			client := &http.Client{}

			// test 'foo' consumer (part of 'gold' group)
			req, err := http.NewRequest("GET", "http://localhost:8000/r1", nil)
			require.NoError(t, err)
			req.Header.Add("apikey", "i-am-special")
			n := 0
			for n < 11 {
				resp, err := client.Do(req)
				require.NoError(t, err)
				defer resp.Body.Close()
				if resp.StatusCode == http.StatusTooManyRequests {
					break
				}
				n++
			}
			assert.Equal(t, maxGoldRequestsNumber, n)

			// test 'bar' consumer (part of 'silver' group)
			req, err = http.NewRequest("GET", "http://localhost:8000/r1", nil)
			require.NoError(t, err)
			req.Header.Add("apikey", "i-am-not-so-special")
			n = 0
			for n < 11 {
				resp, err := client.Do(req)
				require.NoError(t, err)
				defer resp.Body.Close()
				if resp.StatusCode == http.StatusTooManyRequests {
					break
				}
				n++
			}
			assert.Equal(t, maxSilverRequestsNumber, n)

			// test 'baz' consumer (not part of any group)
			req, err = http.NewRequest("GET", "http://localhost:8000/r1", nil)
			require.NoError(t, err)
			req.Header.Add("apikey", "i-am-just-average")
			n = 0
			for n < 11 {
				resp, err := client.Do(req)
				require.NoError(t, err)
				defer resp.Body.Close()
				if resp.StatusCode == http.StatusTooManyRequests {
					break
				}
				n++
			}
			assert.Equal(t, maxRegularRequestsNumber, n)
		})
	}
}

// test scope:
//   - konnect
func Test_Sync_ConsumerGroupsKonnect(t *testing.T) {
	client, err := getTestClient()
	require.NoError(t, err)
	tests := []struct {
		name            string
		kongFile        string
		kongFileInitial string
		expectedState   utils.KongRawState
	}{
		{
			name:            "creates consumer groups",
			kongFile:        "testdata/sync/015-consumer-groups/kong3x.yaml",
			kongFileInitial: "testdata/sync/015-consumer-groups/kong3x-initial.yaml",
			expectedState: utils.KongRawState{
				Consumers:      consumerGroupsConsumers,
				ConsumerGroups: consumerGroupsWithTags,
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runWhen(t, "konnect", "")
			setup(t)

			// set up initial state
			require.NoError(t, sync(context.Background(), tc.kongFileInitial))
			// update with desired final state
			require.NoError(t, sync(context.Background(), tc.kongFile))

			testKongState(t, client, true, false, tc.expectedState, nil)
		})
	}
}

// test scope:
//   - 3.2.0+
func Test_Sync_PluginInstanceName(t *testing.T) {
	// setup stage
	client, err := getTestClient()
	require.NoError(t, err)

	tests := []struct {
		name            string
		kongFile        string
		initialKongFile string
		expectedState   utils.KongRawState
	}{
		{
			name:     "create a plugin with instance_name",
			kongFile: "testdata/sync/018-plugin-instance_name/kong-with-instance_name.yaml",
			expectedState: utils.KongRawState{
				Plugins: []*kong.Plugin{
					{
						Name:         kong.String("request-termination"),
						InstanceName: kong.String("my-plugin"),
						Protocols: []*string{
							kong.String("grpc"),
							kong.String("grpcs"),
							kong.String("http"),
							kong.String("https"),
						},
						Enabled: kong.Bool(true),
						Config: kong.Configuration{
							"status_code":  float64(200),
							"echo":         false,
							"content_type": nil,
							"body":         nil,
							"message":      nil,
							"trigger":      nil,
						},
					},
				},
			},
		},
		{
			name:     "create a plugin without instance_name",
			kongFile: "testdata/sync/018-plugin-instance_name/kong-without-instance_name.yaml",
			expectedState: utils.KongRawState{
				Plugins: []*kong.Plugin{
					{
						Name: kong.String("request-termination"),
						Protocols: []*string{
							kong.String("grpc"),
							kong.String("grpcs"),
							kong.String("http"),
							kong.String("https"),
						},
						Enabled: kong.Bool(true),
						Config: kong.Configuration{
							"status_code":  float64(200),
							"echo":         false,
							"content_type": nil,
							"body":         nil,
							"message":      nil,
							"trigger":      nil,
						},
					},
				},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runWhen(t, "kong", ">=3.2.0")
			setup(t)

			require.NoError(t, sync(context.Background(), tc.kongFile))
			testKongState(t, client, false, false, tc.expectedState, nil)
		})
	}
}

// test scope:
//   - konnect
func Test_Sync_PluginInstanceName_Konnect(t *testing.T) {
	// setup stage
	client, err := getTestClient()
	require.NoError(t, err)

	tests := []struct {
		name            string
		kongFile        string
		initialKongFile string
		expectedState   utils.KongRawState
	}{
		{
			name:     "create a plugin with instance_name",
			kongFile: "testdata/sync/018-plugin-instance_name/kong-with-instance_name.yaml",
			expectedState: utils.KongRawState{
				Plugins: []*kong.Plugin{
					{
						Name:         kong.String("request-termination"),
						InstanceName: kong.String("my-plugin"),
						Protocols: []*string{
							kong.String("grpc"),
							kong.String("grpcs"),
							kong.String("http"),
							kong.String("https"),
						},
						Enabled: kong.Bool(true),
						Config: kong.Configuration{
							"status_code":  float64(200),
							"echo":         false,
							"content_type": nil,
							"body":         nil,
							"message":      nil,
							"trigger":      nil,
						},
					},
				},
			},
		},
		{
			name:     "create a plugin without instance_name",
			kongFile: "testdata/sync/018-plugin-instance_name/kong-without-instance_name.yaml",
			expectedState: utils.KongRawState{
				Plugins: []*kong.Plugin{
					{
						Name: kong.String("request-termination"),
						Protocols: []*string{
							kong.String("grpc"),
							kong.String("grpcs"),
							kong.String("http"),
							kong.String("https"),
						},
						Enabled: kong.Bool(true),
						Config: kong.Configuration{
							"status_code":  float64(200),
							"echo":         false,
							"content_type": nil,
							"body":         nil,
							"message":      nil,
							"trigger":      nil,
						},
					},
				},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runWhen(t, "konnect", "")
			setup(t)

			require.NoError(t, sync(context.Background(), tc.kongFile))
			testKongState(t, client, true, false, tc.expectedState, nil)
		})
	}
}

// test scope:
//   - 3.2.x
//   - 3.3.x
func Test_Sync_SkipConsumers(t *testing.T) {
	// setup stage
	client, err := getTestClient()
	require.NoError(t, err)

	tests := []struct {
		name          string
		kongFile      string
		skipConsumers bool
		expectedState utils.KongRawState
	}{
		{
			name:     "skip-consumers successfully",
			kongFile: "testdata/sync/019-skip-consumers/kong3x.yaml",
			expectedState: utils.KongRawState{
				Services: svc1_207,
			},
			skipConsumers: true,
		},
		{
			name:     "do not skip consumers successfully",
			kongFile: "testdata/sync/019-skip-consumers/kong3x.yaml",
			expectedState: utils.KongRawState{
				Services:       svc1_207,
				Consumers:      consumerGroupsConsumers,
				ConsumerGroups: consumerGroupsWithTagsAndRLA,
			},
			skipConsumers: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runWhen(t, "enterprise", ">=3.2.0 <3.4.0")
			setup(t)

			if tc.skipConsumers {
				require.NoError(t, sync(context.Background(), tc.kongFile, "--skip-consumers"))
			} else {
				require.NoError(t, sync(context.Background(), tc.kongFile))
			}
			testKongState(t, client, false, false, tc.expectedState, nil)
		})
	}
}

// test scope:
//   - 3.4.x
func Test_Sync_SkipConsumers_34x(t *testing.T) {
	runWhen(t, "enterprise", ">=3.4.0 <3.5.0")
	// setup stage
	client, err := getTestClient()
	require.NoError(t, err)

	tests := []struct {
		name          string
		kongFile      string
		skipConsumers bool
		expectedState utils.KongRawState
	}{
		{
			name:     "skip-consumers successfully",
			kongFile: "testdata/sync/019-skip-consumers/kong34.yaml",
			expectedState: utils.KongRawState{
				Services: svc1_207,
			},
			skipConsumers: true,
		},
		{
			name:     "do not skip consumers successfully",
			kongFile: "testdata/sync/019-skip-consumers/kong34.yaml",
			expectedState: utils.KongRawState{
				Services:  svc1_207,
				Consumers: consumerGroupsConsumers,
				ConsumerGroups: []*kong.ConsumerGroupObject{
					{
						ConsumerGroup: &kong.ConsumerGroup{
							Name: kong.String("silver"),
							Tags: kong.StringSlice("tag1", "tag3"),
						},
						Consumers: []*kong.Consumer{
							{
								Username: kong.String("bar"),
							},
						},
					},
					{
						ConsumerGroup: &kong.ConsumerGroup{
							Name: kong.String("gold"),
							Tags: kong.StringSlice("tag1", "tag2"),
						},
						Consumers: []*kong.Consumer{
							{
								Username: kong.String("foo"),
							},
						},
					},
				},
				Plugins: []*kong.Plugin{
					{
						Name: kong.String("rate-limiting-advanced"),
						ConsumerGroup: &kong.ConsumerGroup{
							ID: kong.String("77e6691d-67c0-446a-9401-27be2b141aae"),
						},
						Config: kong.Configuration{
							"consumer_groups":         nil,
							"dictionary_name":         string("kong_rate_limiting_counters"),
							"disable_penalty":         bool(false),
							"enforce_consumer_groups": bool(false),
							"error_code":              float64(429),
							"error_message":           string("API rate limit exceeded"),
							"header_name":             nil,
							"hide_client_headers":     bool(false),
							"identifier":              string("consumer"),
							"limit":                   []any{float64(10)},
							"namespace":               string("gold"),
							"path":                    nil,
							"redis": map[string]any{
								"cluster_addresses":   nil,
								"connect_timeout":     nil,
								"database":            float64(0),
								"host":                nil,
								"keepalive_backlog":   nil,
								"keepalive_pool_size": float64(30),
								"password":            nil,
								"port":                nil,
								"read_timeout":        nil,
								"send_timeout":        nil,
								"sentinel_addresses":  nil,
								"sentinel_master":     nil,
								"sentinel_password":   nil,
								"sentinel_role":       nil,
								"sentinel_username":   nil,
								"server_name":         nil,
								"ssl":                 false,
								"ssl_verify":          false,
								"timeout":             float64(2000),
								"username":            nil,
							},
							"retry_after_jitter_max": float64(1),
							"strategy":               string("local"),
							"sync_rate":              float64(-1),
							"window_size":            []any{float64(60)},
							"window_type":            string("sliding"),
						},
						Enabled:   kong.Bool(true),
						Protocols: []*string{kong.String("grpc"), kong.String("grpcs"), kong.String("http"), kong.String("https")},
					},
					{
						Name: kong.String("rate-limiting-advanced"),
						ConsumerGroup: &kong.ConsumerGroup{
							ID: kong.String("5bcbd3a7-030b-4310-bd1d-2721ff85d236"),
						},
						Config: kong.Configuration{
							"consumer_groups":         nil,
							"dictionary_name":         string("kong_rate_limiting_counters"),
							"disable_penalty":         bool(false),
							"enforce_consumer_groups": bool(false),
							"error_code":              float64(429),
							"error_message":           string("API rate limit exceeded"),
							"header_name":             nil,
							"hide_client_headers":     bool(false),
							"identifier":              string("consumer"),
							"limit":                   []any{float64(7)},
							"namespace":               string("silver"),
							"path":                    nil,
							"redis": map[string]any{
								"cluster_addresses":   nil,
								"connect_timeout":     nil,
								"database":            float64(0),
								"host":                nil,
								"keepalive_backlog":   nil,
								"keepalive_pool_size": float64(30),
								"password":            nil,
								"port":                nil,
								"read_timeout":        nil,
								"send_timeout":        nil,
								"sentinel_addresses":  nil,
								"sentinel_master":     nil,
								"sentinel_password":   nil,
								"sentinel_role":       nil,
								"sentinel_username":   nil,
								"server_name":         nil,
								"ssl":                 false,
								"ssl_verify":          false,
								"timeout":             float64(2000),
								"username":            nil,
							},
							"retry_after_jitter_max": float64(1),
							"strategy":               string("local"),
							"sync_rate":              float64(-1),
							"window_size":            []any{float64(60)},
							"window_type":            string("sliding"),
						},
						Enabled:   kong.Bool(true),
						Protocols: []*string{kong.String("grpc"), kong.String("grpcs"), kong.String("http"), kong.String("https")},
					},
					{
						Name: kong.String("rate-limiting-advanced"),
						Consumer: &kong.Consumer{
							ID: kong.String("416b038a-fd00-45fd-a5a2-a74bf70017fa"),
						},
						Config: kong.Configuration{
							"consumer_groups":         nil,
							"dictionary_name":         string("kong_rate_limiting_counters"),
							"disable_penalty":         bool(false),
							"enforce_consumer_groups": bool(false),
							"error_code":              float64(429),
							"error_message":           string("API rate limit exceeded"),
							"header_name":             nil,
							"hide_client_headers":     bool(false),
							"identifier":              string("consumer"),
							"limit":                   []any{float64(10)},
							"namespace":               string("foo"),
							"path":                    nil,
							"redis": map[string]any{
								"cluster_addresses":   nil,
								"connect_timeout":     nil,
								"database":            float64(0),
								"host":                nil,
								"keepalive_backlog":   nil,
								"keepalive_pool_size": float64(30),
								"password":            nil,
								"port":                nil,
								"read_timeout":        nil,
								"send_timeout":        nil,
								"sentinel_addresses":  nil,
								"sentinel_master":     nil,
								"sentinel_password":   nil,
								"sentinel_role":       nil,
								"sentinel_username":   nil,
								"server_name":         nil,
								"ssl":                 false,
								"ssl_verify":          false,
								"timeout":             float64(2000),
								"username":            nil,
							},
							"retry_after_jitter_max": float64(1),
							"strategy":               string("local"),
							"sync_rate":              float64(-1),
							"window_size":            []any{float64(60)},
							"window_type":            string("sliding"),
						},
						Enabled:   kong.Bool(true),
						Protocols: []*string{kong.String("grpc"), kong.String("grpcs"), kong.String("http"), kong.String("https")},
					},
				},
			},
			skipConsumers: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			setup(t)

			if tc.skipConsumers {
				require.NoError(t, sync(context.Background(), tc.kongFile, "--skip-consumers"))
			} else {
				require.NoError(t, sync(context.Background(), tc.kongFile))
			}
			testKongState(t, client, false, false, tc.expectedState, nil)
		})
	}
}

// test scope:
//   - konnect
func Test_Sync_SkipConsumers_Konnect(t *testing.T) {
	runWhenKonnect(t)
	// setup stage
	client, err := getTestClient()
	require.NoError(t, err)

	tests := []struct {
		name          string
		kongFile      string
		skipConsumers bool
		expectedState utils.KongRawState
	}{
		{
			name:     "skip-consumers successfully",
			kongFile: "testdata/sync/019-skip-consumers/kong34.yaml",
			expectedState: utils.KongRawState{
				Services: svc1_207,
			},
			skipConsumers: true,
		},
		{
			name:     "do not skip consumers successfully",
			kongFile: "testdata/sync/019-skip-consumers/kong34.yaml",
			expectedState: utils.KongRawState{
				Services:  svc1_207,
				Consumers: consumerGroupsConsumers,
				ConsumerGroups: []*kong.ConsumerGroupObject{
					{
						ConsumerGroup: &kong.ConsumerGroup{
							Name: kong.String("silver"),
							Tags: kong.StringSlice("tag1", "tag3"),
						},
						Consumers: []*kong.Consumer{
							{
								Username: kong.String("bar"),
							},
						},
					},
					{
						ConsumerGroup: &kong.ConsumerGroup{
							Name: kong.String("gold"),
							Tags: kong.StringSlice("tag1", "tag2"),
						},
						Consumers: []*kong.Consumer{
							{
								Username: kong.String("foo"),
							},
						},
					},
				},
				Plugins: []*kong.Plugin{
					{
						Name: kong.String("rate-limiting-advanced"),
						ConsumerGroup: &kong.ConsumerGroup{
							ID: kong.String("77e6691d-67c0-446a-9401-27be2b141aae"),
						},
						Config: kong.Configuration{
							"consumer_groups":         nil,
							"dictionary_name":         string("kong_rate_limiting_counters"),
							"disable_penalty":         bool(false),
							"enforce_consumer_groups": bool(false),
							"error_code":              float64(429),
							"error_message":           string("API rate limit exceeded"),
							"header_name":             nil,
							"hide_client_headers":     bool(false),
							"identifier":              string("consumer"),
							"limit":                   []any{float64(10)},
							"namespace":               string("gold"),
							"path":                    nil,
							"redis": map[string]any{
								"cluster_addresses":   nil,
								"connect_timeout":     nil,
								"database":            float64(0),
								"host":                nil,
								"keepalive_backlog":   nil,
								"keepalive_pool_size": float64(30),
								"password":            nil,
								"port":                nil,
								"read_timeout":        nil,
								"send_timeout":        nil,
								"sentinel_addresses":  nil,
								"sentinel_master":     nil,
								"sentinel_password":   nil,
								"sentinel_role":       nil,
								"sentinel_username":   nil,
								"server_name":         nil,
								"ssl":                 false,
								"ssl_verify":          false,
								"timeout":             float64(2000),
								"username":            nil,
							},
							"retry_after_jitter_max": float64(1),
							"strategy":               string("local"),
							"sync_rate":              nil,
							"window_size":            []any{float64(60)},
							"window_type":            string("sliding"),
						},
						Enabled:   kong.Bool(true),
						Protocols: []*string{kong.String("grpc"), kong.String("grpcs"), kong.String("http"), kong.String("https")},
					},
					{
						Name: kong.String("rate-limiting-advanced"),
						ConsumerGroup: &kong.ConsumerGroup{
							ID: kong.String("5bcbd3a7-030b-4310-bd1d-2721ff85d236"),
						},
						Config: kong.Configuration{
							"consumer_groups":         nil,
							"dictionary_name":         string("kong_rate_limiting_counters"),
							"disable_penalty":         bool(false),
							"enforce_consumer_groups": bool(false),
							"error_code":              float64(429),
							"error_message":           string("API rate limit exceeded"),
							"header_name":             nil,
							"hide_client_headers":     bool(false),
							"identifier":              string("consumer"),
							"limit":                   []any{float64(7)},
							"namespace":               string("silver"),
							"path":                    nil,
							"redis": map[string]any{
								"cluster_addresses":   nil,
								"connect_timeout":     nil,
								"database":            float64(0),
								"host":                nil,
								"keepalive_backlog":   nil,
								"keepalive_pool_size": float64(30),
								"password":            nil,
								"port":                nil,
								"read_timeout":        nil,
								"send_timeout":        nil,
								"sentinel_addresses":  nil,
								"sentinel_master":     nil,
								"sentinel_password":   nil,
								"sentinel_role":       nil,
								"sentinel_username":   nil,
								"server_name":         nil,
								"ssl":                 false,
								"ssl_verify":          false,
								"timeout":             float64(2000),
								"username":            nil,
							},
							"retry_after_jitter_max": float64(1),
							"strategy":               string("local"),
							"sync_rate":              nil,
							"window_size":            []any{float64(60)},
							"window_type":            string("sliding"),
						},
						Enabled:   kong.Bool(true),
						Protocols: []*string{kong.String("grpc"), kong.String("grpcs"), kong.String("http"), kong.String("https")},
					},
				},
			},
			skipConsumers: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runWhen(t, "enterprise", ">=3.2.0")
			setup(t)

			if tc.skipConsumers {
				require.NoError(t, sync(context.Background(), tc.kongFile, "--skip-consumers"))
			} else {
				require.NoError(t, sync(context.Background(), tc.kongFile))
			}
			testKongState(t, client, false, false, tc.expectedState, nil)
		})
	}
}

// In the tests we're concerned only with the IDs and names of the entities
// we'll ignore other fields when comparing states.
var ignoreFieldsIrrelevantForIDsTests = []cmp.Option{
	cmpopts.IgnoreFields(
		kong.Plugin{},
		"Config",
		"Protocols",
		"Enabled",
	),
	cmpopts.IgnoreFields(
		kong.Service{},
		"ConnectTimeout",
		"Enabled",
		"Host",
		"Port",
		"Protocol",
		"ReadTimeout",
		"WriteTimeout",
		"Retries",
	),
	cmpopts.IgnoreFields(
		kong.Route{},
		"Paths",
		"PathHandling",
		"PreserveHost",
		"Protocols",
		"RegexPriority",
		"StripPath",
		"HTTPSRedirectStatusCode",
		"Sources",
		"Destinations",
		"RequestBuffering",
		"ResponseBuffering",
	),
}

// test scope:
//   - 3.0.0+
func Test_Sync_ChangingIDsWhileKeepingNames(t *testing.T) {
	runWhen(t, "kong", ">=3.0.0")

	client, err := getTestClient()
	require.NoError(t, err)

	// These are the IDs that should be present in Kong after the second sync in all cases.
	var (
		expectedServiceID  = kong.String("98076db2-28b6-423b-ba39-a797193017f7")
		expectedRouteID    = kong.String("97b6a97e-f3f7-4c47-857a-7464cb9e202b")
		expectedConsumerID = kong.String("9a1e49a8-2536-41fa-a4e9-605bf218a4fa")
	)

	// These are the entities that should be present in Kong after the second sync in all cases.
	var (
		expectedService = &kong.Service{
			Name: kong.String("s1"),
			ID:   expectedServiceID,
		}

		expectedRoute = &kong.Route{
			Name: kong.String("r1"),
			ID:   expectedRouteID,
			Service: &kong.Service{
				ID: expectedServiceID,
			},
		}

		expectedConsumer = &kong.Consumer{
			Username: kong.String("c1"),
			ID:       expectedConsumerID,
		}

		expectedPlugins = []*kong.Plugin{
			{
				Name: kong.String("rate-limiting"),
				Route: &kong.Route{
					ID: expectedRouteID,
				},
			},
			{
				Name: kong.String("rate-limiting"),
				Service: &kong.Service{
					ID: expectedServiceID,
				},
			},
			{
				Name: kong.String("rate-limiting"),
				Consumer: &kong.Consumer{
					ID: expectedConsumerID,
				},
			},
		}
	)

	testCases := []struct {
		name         string
		beforeConfig string
	}{
		{
			name:         "all entities have the same names, but different IDs",
			beforeConfig: "testdata/sync/020-same-names-altered-ids/1-before.yaml",
		},
		{
			name:         "service and consumer changed IDs, route did not",
			beforeConfig: "testdata/sync/020-same-names-altered-ids/2-before.yaml",
		},
		{
			name:         "route and consumer changed IDs, service did not",
			beforeConfig: "testdata/sync/020-same-names-altered-ids/3-before.yaml",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			setup(t)

			// First, create the entities with the original IDs.
			err = sync(context.Background(), tc.beforeConfig)
			require.NoError(t, err)

			// Then, sync again with the same names, but different IDs.
			err = sync(context.Background(), "testdata/sync/020-same-names-altered-ids/desired.yaml")
			require.NoError(t, err)

			// Finally, check that the all entities exist and have the expected IDs.
			testKongState(t, client, false, false, utils.KongRawState{
				Services:  []*kong.Service{expectedService},
				Routes:    []*kong.Route{expectedRoute},
				Consumers: []*kong.Consumer{expectedConsumer},
				Plugins:   expectedPlugins,
			}, ignoreFieldsIrrelevantForIDsTests)
		})
	}
}

// test scope:
//   - konnect
func Test_Sync_ChangingIDsWhileKeepingNames_Konnect(t *testing.T) {
	runWhen(t, "konnect", "")

	client, err := getTestClient()
	require.NoError(t, err)

	// These are the IDs that should be present in Kong after the second sync in all cases.
	var (
		expectedServiceID  = kong.String("98076db2-28b6-423b-ba39-a797193017f7")
		expectedRouteID    = kong.String("97b6a97e-f3f7-4c47-857a-7464cb9e202b")
		expectedConsumerID = kong.String("9a1e49a8-2536-41fa-a4e9-605bf218a4fa")
	)

	// These are the entities that should be present in Kong after the second sync in all cases.
	var (
		expectedService = &kong.Service{
			Name: kong.String("s1"),
			ID:   expectedServiceID,
		}

		expectedRoute = &kong.Route{
			Name: kong.String("r1"),
			ID:   expectedRouteID,
			Service: &kong.Service{
				ID: expectedServiceID,
			},
		}

		expectedConsumer = &kong.Consumer{
			Username: kong.String("c1"),
			ID:       expectedConsumerID,
		}

		expectedPlugins = []*kong.Plugin{
			{
				Name: kong.String("rate-limiting"),
				Route: &kong.Route{
					ID: expectedRouteID,
				},
			},
			{
				Name: kong.String("rate-limiting"),
				Service: &kong.Service{
					ID: expectedServiceID,
				},
			},
			{
				Name: kong.String("rate-limiting"),
				Consumer: &kong.Consumer{
					ID: expectedConsumerID,
				},
			},
		}
	)

	testCases := []struct {
		name         string
		beforeConfig string
	}{
		{
			name:         "all entities have the same names, but different IDs",
			beforeConfig: "testdata/sync/020-same-names-altered-ids/1-before.yaml",
		},
		{
			name:         "service and consumer changed IDs, route did not",
			beforeConfig: "testdata/sync/020-same-names-altered-ids/2-before.yaml",
		},
		{
			name:         "route and consumer changed IDs, service did not",
			beforeConfig: "testdata/sync/020-same-names-altered-ids/3-before.yaml",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			setup(t)

			// First, create the entities with the original IDs.
			err = sync(context.Background(), tc.beforeConfig)
			require.NoError(t, err)

			// Then, sync again with the same names, but different IDs.
			err = sync(context.Background(), "testdata/sync/020-same-names-altered-ids/desired.yaml")
			require.NoError(t, err)

			// Finally, check that the all entities exist and have the expected IDs.
			testKongState(t, client, true, false, utils.KongRawState{
				Services:  []*kong.Service{expectedService},
				Routes:    []*kong.Route{expectedRoute},
				Consumers: []*kong.Consumer{expectedConsumer},
				Plugins:   expectedPlugins,
			}, ignoreFieldsIrrelevantForIDsTests)
		})
	}
}

// test scope:
//   - 3.0.0+
func Test_Sync_UpdateWithExplicitIDs(t *testing.T) {
	runWhen(t, "kong", ">=3.0.0")
	setup(t)

	client, err := getTestClient()
	require.NoError(t, err)

	const (
		beforeConfig = "testdata/sync/021-update-with-explicit-ids/before.yaml"
		afterConfig  = "testdata/sync/021-update-with-explicit-ids/after.yaml"
	)

	// First, create entities with IDs assigned explicitly.
	err = sync(context.Background(), beforeConfig)
	require.NoError(t, err)

	// Then, sync again, adding tags to every entity just to trigger an update.
	err = sync(context.Background(), afterConfig)
	require.NoError(t, err)

	// Finally, verify that the update was successful.
	testKongState(t, client, false, false, utils.KongRawState{
		Services: []*kong.Service{
			{
				Name: kong.String("s1"),
				ID:   kong.String("c75a775b-3a32-4b73-8e05-f68169c23941"),
				Tags: kong.StringSlice("after"),
			},
		},
		Routes: []*kong.Route{
			{
				Name: kong.String("r1"),
				ID:   kong.String("97b6a97e-f3f7-4c47-857a-7464cb9e202b"),
				Tags: kong.StringSlice("after"),
				Service: &kong.Service{
					ID: kong.String("c75a775b-3a32-4b73-8e05-f68169c23941"),
				},
			},
		},
		Consumers: []*kong.Consumer{
			{
				Username: kong.String("c1"),
				Tags:     kong.StringSlice("after"),
			},
		},
	}, ignoreFieldsIrrelevantForIDsTests)
}

// test scope:
//   - konnect
func Test_Sync_UpdateWithExplicitIDs_Konnect(t *testing.T) {
	runWhen(t, "konnect", "")
	setup(t)

	client, err := getTestClient()
	require.NoError(t, err)

	const (
		beforeConfig = "testdata/sync/021-update-with-explicit-ids/before.yaml"
		afterConfig  = "testdata/sync/021-update-with-explicit-ids/after.yaml"
	)

	// First, create entities with IDs assigned explicitly.
	err = sync(context.Background(), beforeConfig)
	require.NoError(t, err)

	// Then, sync again, adding tags to every entity just to trigger an update.
	err = sync(context.Background(), afterConfig)
	require.NoError(t, err)

	// Finally, verify that the update was successful.
	testKongState(t, client, true, false, utils.KongRawState{
		Services: []*kong.Service{
			{
				Name: kong.String("s1"),
				ID:   kong.String("c75a775b-3a32-4b73-8e05-f68169c23941"),
				Tags: kong.StringSlice("after"),
			},
		},
		Routes: []*kong.Route{
			{
				Name: kong.String("r1"),
				ID:   kong.String("97b6a97e-f3f7-4c47-857a-7464cb9e202b"),
				Tags: kong.StringSlice("after"),
				Service: &kong.Service{
					ID: kong.String("c75a775b-3a32-4b73-8e05-f68169c23941"),
				},
			},
		},
		Consumers: []*kong.Consumer{
			{
				Username: kong.String("c1"),
				Tags:     kong.StringSlice("after"),
			},
		},
	}, ignoreFieldsIrrelevantForIDsTests)
}

// test scope:
//   - 3.0.0+
func Test_Sync_UpdateWithExplicitIDsWithNoNames(t *testing.T) {
	runWhen(t, "kong", ">=3.0.0")
	setup(t)

	client, err := getTestClient()
	require.NoError(t, err)

	const (
		beforeConfig = "testdata/sync/022-update-with-explicit-ids-with-no-names/before.yaml"
		afterConfig  = "testdata/sync/022-update-with-explicit-ids-with-no-names/after.yaml"
	)

	// First, create entities with IDs assigned explicitly.
	require.NoError(t, sync(context.Background(), beforeConfig))

	// Then, sync again, adding tags to every entity just to trigger an update.
	require.NoError(t, sync(context.Background(), afterConfig))

	// Finally, verify that the update was successful.
	testKongState(t, client, false, false, utils.KongRawState{
		Services: []*kong.Service{
			{
				ID:   kong.String("c75a775b-3a32-4b73-8e05-f68169c23941"),
				Tags: kong.StringSlice("after"),
			},
		},
		Routes: []*kong.Route{
			{
				ID:   kong.String("97b6a97e-f3f7-4c47-857a-7464cb9e202b"),
				Tags: kong.StringSlice("after"),
				Service: &kong.Service{
					ID: kong.String("c75a775b-3a32-4b73-8e05-f68169c23941"),
				},
			},
		},
	}, ignoreFieldsIrrelevantForIDsTests)
}

// test scope:
//   - konnect
func Test_Sync_UpdateWithExplicitIDsWithNoNames_Konnect(t *testing.T) {
	runWhen(t, "konnect", "")
	setup(t)

	client, err := getTestClient()
	require.NoError(t, err)

	const (
		beforeConfig = "testdata/sync/022-update-with-explicit-ids-with-no-names/before.yaml"
		afterConfig  = "testdata/sync/022-update-with-explicit-ids-with-no-names/after.yaml"
	)

	// First, create entities with IDs assigned explicitly.
	require.NoError(t, sync(context.Background(), beforeConfig))

	// Then, sync again, adding tags to every entity just to trigger an update.
	require.NoError(t, sync(context.Background(), afterConfig))

	// Finally, verify that the update was successful.
	testKongState(t, client, true, false, utils.KongRawState{
		Services: []*kong.Service{
			{
				ID:   kong.String("c75a775b-3a32-4b73-8e05-f68169c23941"),
				Tags: kong.StringSlice("after"),
			},
		},
		Routes: []*kong.Route{
			{
				ID:   kong.String("97b6a97e-f3f7-4c47-857a-7464cb9e202b"),
				Tags: kong.StringSlice("after"),
				Service: &kong.Service{
					ID: kong.String("c75a775b-3a32-4b73-8e05-f68169c23941"),
				},
			},
		},
	}, ignoreFieldsIrrelevantForIDsTests)
}

// test scope:
//   - 3.0.0+
func Test_Sync_CreateCertificateWithSNIs(t *testing.T) {
	runWhen(t, "kong", ">=3.0.0")
	setup(t)

	client, err := getTestClient()
	require.NoError(t, err)

	err = sync(context.Background(), "testdata/sync/023-create-and-update-certificate-with-snis/initial.yaml")
	require.NoError(t, err)

	// To ignore noise, we ignore the Key and Cert fields because they are not relevant for this test.
	ignoredFields := []cmp.Option{
		cmpopts.IgnoreFields(
			kong.Certificate{},
			"Key",
			"Cert",
		),
	}

	testKongState(t, client, false, false, utils.KongRawState{
		Certificates: []*kong.Certificate{
			{
				ID:   kong.String("c75a775b-3a32-4b73-8e05-f68169c23941"),
				Tags: kong.StringSlice("before"),
			},
		},
		SNIs: []*kong.SNI{
			{
				Name: kong.String("example.com"),
				Certificate: &kong.Certificate{
					ID: kong.String("c75a775b-3a32-4b73-8e05-f68169c23941"),
				},
			},
		},
	}, ignoredFields)

	err = sync(context.Background(), "testdata/sync/023-create-and-update-certificate-with-snis/update.yaml")
	require.NoError(t, err)

	testKongState(t, client, false, false, utils.KongRawState{
		Certificates: []*kong.Certificate{
			{
				ID:   kong.String("c75a775b-3a32-4b73-8e05-f68169c23941"),
				Tags: kong.StringSlice("after"), // Tag should be updated.
			},
		},
		SNIs: []*kong.SNI{
			{
				Name: kong.String("example.com"),
				Certificate: &kong.Certificate{
					ID: kong.String("c75a775b-3a32-4b73-8e05-f68169c23941"),
				},
			},
		},
	}, ignoredFields)
}

// test scope:
//   - konnect
func Test_Sync_CreateCertificateWithSNIs_Konnect(t *testing.T) {
	runWhen(t, "konnect", "")
	setup(t)

	client, err := getTestClient()
	require.NoError(t, err)

	err = sync(context.Background(), "testdata/sync/023-create-and-update-certificate-with-snis/initial.yaml")
	require.NoError(t, err)

	// To ignore noise, we ignore the Key and Cert fields because they are not relevant for this test.
	ignoredFields := []cmp.Option{
		cmpopts.IgnoreFields(
			kong.Certificate{},
			"Key",
			"Cert",
		),
	}

	testKongState(t, client, true, false, utils.KongRawState{
		Certificates: []*kong.Certificate{
			{
				ID:   kong.String("c75a775b-3a32-4b73-8e05-f68169c23941"),
				Tags: kong.StringSlice("before"),
			},
		},
		SNIs: []*kong.SNI{
			{
				Name: kong.String("example.com"),
				Certificate: &kong.Certificate{
					ID: kong.String("c75a775b-3a32-4b73-8e05-f68169c23941"),
				},
			},
		},
	}, ignoredFields)

	err = sync(context.Background(), "testdata/sync/023-create-and-update-certificate-with-snis/update.yaml")
	require.NoError(t, err)

	testKongState(t, client, true, false, utils.KongRawState{
		Certificates: []*kong.Certificate{
			{
				ID:   kong.String("c75a775b-3a32-4b73-8e05-f68169c23941"),
				Tags: kong.StringSlice("after"), // Tag should be updated.
			},
		},
		SNIs: []*kong.SNI{
			{
				Name: kong.String("example.com"),
				Certificate: &kong.Certificate{
					ID: kong.String("c75a775b-3a32-4b73-8e05-f68169c23941"),
				},
			},
		},
	}, ignoredFields)
}

// test scope:
//   - 3.0.0+
func Test_Sync_ConsumersWithCustomIDAndOrUsername(t *testing.T) {
	runWhen(t, "kong", ">=3.0.0")
	setup(t)

	client, err := getTestClient()
	require.NoError(t, err)

	err = sync(context.Background(), "testdata/sync/024-consumers-with-custom_id-and-username/kong3x.yaml")
	require.NoError(t, err)

	testKongState(t, client, false, false, utils.KongRawState{
		Consumers: []*kong.Consumer{
			{
				ID:       kong.String("ce49186d-7670-445d-a218-897631b29ada"),
				Username: kong.String("Foo"),
				CustomID: kong.String("foo"),
			},
			{
				ID:       kong.String("7820f383-7b77-4fcc-af7f-14ff3e256693"),
				Username: kong.String("foo"),
				CustomID: kong.String("bar"),
			},
			{
				ID:       kong.String("18c62c3c-12cc-429a-8e5a-57f2c3691a6b"),
				CustomID: kong.String("custom_id_only"),
			},
			{
				ID:       kong.String("8ef278c9-48c1-43e1-b665-e9bc18fab4c8"),
				Username: kong.String("username_only"),
			},
		},
	}, nil)

	err = sync(context.Background(), "testdata/sync/024-consumers-with-custom_id-and-username/kong3x-reverse-order.yaml")
	require.NoError(t, err)

	testKongState(t, client, false, false, utils.KongRawState{
		Consumers: []*kong.Consumer{
			{
				Username: kong.String("TestUser"),
			},
			{
				Username: kong.String("OtherUser"),
				CustomID: kong.String("TestUser"),
			},
		},
	}, nil)
}

// test scope:
//   - konnect
func Test_Sync_ConsumersWithCustomIDAndOrUsername_Konnect(t *testing.T) {
	runWhen(t, "konnect", "")
	setup(t)

	client, err := getTestClient()
	require.NoError(t, err)

	err = sync(context.Background(), "testdata/sync/024-consumers-with-custom_id-and-username/kong3x.yaml")
	require.NoError(t, err)

	testKongState(t, client, true, false, utils.KongRawState{
		Consumers: []*kong.Consumer{
			{
				ID:       kong.String("ce49186d-7670-445d-a218-897631b29ada"),
				Username: kong.String("Foo"),
				CustomID: kong.String("foo"),
			},
			{
				ID:       kong.String("7820f383-7b77-4fcc-af7f-14ff3e256693"),
				Username: kong.String("foo"),
				CustomID: kong.String("bar"),
			},
			{
				ID:       kong.String("18c62c3c-12cc-429a-8e5a-57f2c3691a6b"),
				CustomID: kong.String("custom_id_only"),
			},
			{
				ID:       kong.String("8ef278c9-48c1-43e1-b665-e9bc18fab4c8"),
				Username: kong.String("username_only"),
			},
		},
	}, nil)

	err = sync(context.Background(), "testdata/sync/024-consumers-with-custom_id-and-username/kong3x-reverse-order.yaml")
	require.NoError(t, err)

	testKongState(t, client, true, false, utils.KongRawState{
		Consumers: []*kong.Consumer{
			{
				Username: kong.String("TestUser"),
			},
			{
				Username: kong.String("OtherUser"),
				CustomID: kong.String("TestUser"),
			},
		},
	}, nil)
}

// This test has 2 goals:
//   - make sure consumer groups scoped plugins can be configured correctly in Kong
//   - the actual consumer groups functionality works once set
//
// This is achieved via configuring:
// - 3 consumers:
//   - 1 belonging to Gold Consumer Group
//   - 1 belonging to Silver Consumer Group
//   - 1 not belonging to any Consumer Group
//
// - 3 key-auths, one for each consumer
// - 1 global key-auth plugin
// - 2 consumer group
// - 1 global RLA plugin
// - 2 RLA plugins, scoped to the related consumer groups
// - 1 service pointing to mockbin.org
// - 1 route proxying the above service
//
// Once the configuration is verified to be matching in Kong,
// we then check whether the specific RLA configuration is correctly applied: consumers
// not belonging to the consumer group should be limited to 5 requests
// every 30s, while consumers belonging to the 'gold' and 'silver' consumer groups
// should be allowed to run respectively 10 and 7 requests in the same timeframe.
// In order to make sure this is the case, we run requests in a loop
// for all consumers and then check at what point they start to receive 429.
func Test_Sync_ConsumerGroupsScopedPlugins(t *testing.T) {
	const (
		maxGoldRequestsNumber    = 10
		maxSilverRequestsNumber  = 7
		maxRegularRequestsNumber = 5
	)
	client, err := getTestClient()
	require.NoError(t, err)
	tests := []struct {
		name          string
		kongFile      string
		expectedState utils.KongRawState
	}{
		{
			name:     "creates consumer groups scoped plugins",
			kongFile: "testdata/sync/025-consumer-groups-scoped-plugins/kong3x.yaml",
			expectedState: utils.KongRawState{
				Consumers: consumerGroupsConsumers,
				ConsumerGroups: []*kong.ConsumerGroupObject{
					{
						ConsumerGroup: &kong.ConsumerGroup{
							Name: kong.String("silver"),
						},
						Consumers: []*kong.Consumer{
							{
								Username: kong.String("bar"),
							},
						},
					},
					{
						ConsumerGroup: &kong.ConsumerGroup{
							Name: kong.String("gold"),
						},
						Consumers: []*kong.Consumer{
							{
								Username: kong.String("foo"),
							},
						},
					},
				},
				Plugins:  consumerGroupScopedPlugins,
				Services: svc1_207,
				Routes:   route1_20x,
				KeyAuths: []*kong.KeyAuth{
					{
						Consumer: &kong.Consumer{
							ID: kong.String("87095815-5395-454e-8c18-a11c9bc0ef04"),
						},
						Key: kong.String("i-am-special"),
					},
					{
						Consumer: &kong.Consumer{
							ID: kong.String("5a5b9369-baeb-4faa-a902-c40ccdc2928e"),
						},
						Key: kong.String("i-am-not-so-special"),
					},
					{
						Consumer: &kong.Consumer{
							ID: kong.String("e894ea9e-ad08-4acf-a960-5a23aa7701c7"),
						},
						Key: kong.String("i-am-just-average"),
					},
				},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runWhen(t, "enterprise", ">=3.4.0 <3.5.0")
			setup(t)

			require.NoError(t, sync(context.Background(), tc.kongFile))
			testKongState(t, client, false, false, tc.expectedState, nil)

			// Kong proxy may need a bit to be ready.
			time.Sleep(time.Second * 10)

			// build simple http client
			client := &http.Client{}

			// test 'foo' consumer (part of 'gold' group)
			req, err := http.NewRequest("GET", "http://localhost:8000/r1", nil)
			require.NoError(t, err)
			req.Header.Add("apikey", "i-am-special")
			n := 0
			for n < 11 {
				resp, err := client.Do(req)
				require.NoError(t, err)
				defer resp.Body.Close()
				if resp.StatusCode == http.StatusTooManyRequests {
					break
				}
				n++
			}
			assert.Equal(t, maxGoldRequestsNumber, n)

			// test 'bar' consumer (part of 'silver' group)
			req, err = http.NewRequest("GET", "http://localhost:8000/r1", nil)
			require.NoError(t, err)
			req.Header.Add("apikey", "i-am-not-so-special")
			n = 0
			for n < 11 {
				resp, err := client.Do(req)
				require.NoError(t, err)
				defer resp.Body.Close()
				if resp.StatusCode == http.StatusTooManyRequests {
					break
				}
				n++
			}
			assert.Equal(t, maxSilverRequestsNumber, n)

			// test 'baz' consumer (not part of any group)
			req, err = http.NewRequest("GET", "http://localhost:8000/r1", nil)
			require.NoError(t, err)
			req.Header.Add("apikey", "i-am-just-average")
			n = 0
			for n < 11 {
				resp, err := client.Do(req)
				require.NoError(t, err)
				defer resp.Body.Close()
				if resp.StatusCode == http.StatusTooManyRequests {
					break
				}
				n++
			}
			assert.Equal(t, maxRegularRequestsNumber, n)
		})
	}
}

func Test_Sync_ConsumerGroupsScopedPlugins_After360(t *testing.T) {
	const (
		maxGoldRequestsNumber    = 10
		maxSilverRequestsNumber  = 7
		maxRegularRequestsNumber = 5
	)
	client, err := getTestClient()
	require.NoError(t, err)

	tests := []struct {
		name          string
		runWhen       string
		kongFile      string
		expectedState utils.KongRawState
	}{
		{
			name:     "creates consumer groups scoped plugins",
			runWhen:  ">=3.6.0 <3.7.0",
			kongFile: "testdata/sync/025-consumer-groups-scoped-plugins/kong3x.yaml",
			expectedState: utils.KongRawState{
				Consumers: consumerGroupsConsumers,
				ConsumerGroups: []*kong.ConsumerGroupObject{
					{
						ConsumerGroup: &kong.ConsumerGroup{
							Name: kong.String("silver"),
						},
						Consumers: []*kong.Consumer{
							{
								Username: kong.String("bar"),
							},
						},
					},
					{
						ConsumerGroup: &kong.ConsumerGroup{
							Name: kong.String("gold"),
						},
						Consumers: []*kong.Consumer{
							{
								Username: kong.String("foo"),
							},
						},
					},
				},
				Plugins:  consumerGroupScopedPlugins36x,
				Services: svc1_207,
				Routes:   route1_20x,
				KeyAuths: []*kong.KeyAuth{
					{
						Consumer: &kong.Consumer{
							ID: kong.String("87095815-5395-454e-8c18-a11c9bc0ef04"),
						},
						Key: kong.String("i-am-special"),
					},
					{
						Consumer: &kong.Consumer{
							ID: kong.String("5a5b9369-baeb-4faa-a902-c40ccdc2928e"),
						},
						Key: kong.String("i-am-not-so-special"),
					},
					{
						Consumer: &kong.Consumer{
							ID: kong.String("e894ea9e-ad08-4acf-a960-5a23aa7701c7"),
						},
						Key: kong.String("i-am-just-average"),
					},
				},
			},
		},
		{
			name:     "creates consumer groups scoped plugins",
			runWhen:  ">=3.7.0 <3.8.0",
			kongFile: "testdata/sync/025-consumer-groups-scoped-plugins/kong3x.yaml",
			expectedState: utils.KongRawState{
				Consumers: consumerGroupsConsumers,
				ConsumerGroups: []*kong.ConsumerGroupObject{
					{
						ConsumerGroup: &kong.ConsumerGroup{
							Name: kong.String("silver"),
						},
						Consumers: []*kong.Consumer{
							{
								Username: kong.String("bar"),
							},
						},
					},
					{
						ConsumerGroup: &kong.ConsumerGroup{
							Name: kong.String("gold"),
						},
						Consumers: []*kong.Consumer{
							{
								Username: kong.String("foo"),
							},
						},
					},
				},
				Plugins:  consumerGroupScopedPlugins37x,
				Services: svc1_207,
				Routes:   route1_20x,
				KeyAuths: []*kong.KeyAuth{
					{
						Consumer: &kong.Consumer{
							ID: kong.String("87095815-5395-454e-8c18-a11c9bc0ef04"),
						},
						Key: kong.String("i-am-special"),
					},
					{
						Consumer: &kong.Consumer{
							ID: kong.String("5a5b9369-baeb-4faa-a902-c40ccdc2928e"),
						},
						Key: kong.String("i-am-not-so-special"),
					},
					{
						Consumer: &kong.Consumer{
							ID: kong.String("e894ea9e-ad08-4acf-a960-5a23aa7701c7"),
						},
						Key: kong.String("i-am-just-average"),
					},
				},
			},
		},
		{
			name:     "creates consumer groups scoped plugins",
			runWhen:  ">=3.8.0 <3.9.0",
			kongFile: "testdata/sync/025-consumer-groups-scoped-plugins/kong3x.yaml",
			expectedState: utils.KongRawState{
				Consumers: consumerGroupsConsumers,
				ConsumerGroups: []*kong.ConsumerGroupObject{
					{
						ConsumerGroup: &kong.ConsumerGroup{
							Name: kong.String("silver"),
						},
						Consumers: []*kong.Consumer{
							{
								Username: kong.String("bar"),
							},
						},
					},
					{
						ConsumerGroup: &kong.ConsumerGroup{
							Name: kong.String("gold"),
						},
						Consumers: []*kong.Consumer{
							{
								Username: kong.String("foo"),
							},
						},
					},
				},
				Plugins:  consumerGroupScopedPlugins381x,
				Services: svc1_207,
				Routes:   route1_20x,
				KeyAuths: []*kong.KeyAuth{
					{
						Consumer: &kong.Consumer{
							ID: kong.String("87095815-5395-454e-8c18-a11c9bc0ef04"),
						},
						Key: kong.String("i-am-special"),
					},
					{
						Consumer: &kong.Consumer{
							ID: kong.String("5a5b9369-baeb-4faa-a902-c40ccdc2928e"),
						},
						Key: kong.String("i-am-not-so-special"),
					},
					{
						Consumer: &kong.Consumer{
							ID: kong.String("e894ea9e-ad08-4acf-a960-5a23aa7701c7"),
						},
						Key: kong.String("i-am-just-average"),
					},
				},
			},
		},
		{
			name:     "creates consumer groups scoped plugins",
			runWhen:  ">=3.9.0 <3.10.0",
			kongFile: "testdata/sync/025-consumer-groups-scoped-plugins/kong3x.yaml",
			expectedState: utils.KongRawState{
				Consumers: consumerGroupsConsumers,
				ConsumerGroups: []*kong.ConsumerGroupObject{
					{
						ConsumerGroup: &kong.ConsumerGroup{
							Name: kong.String("silver"),
						},
						Consumers: []*kong.Consumer{
							{
								Username: kong.String("bar"),
							},
						},
					},
					{
						ConsumerGroup: &kong.ConsumerGroup{
							Name: kong.String("gold"),
						},
						Consumers: []*kong.Consumer{
							{
								Username: kong.String("foo"),
							},
						},
					},
				},
				Plugins:  consumerGroupScopedPlugins390x,
				Services: svc1_207,
				Routes:   route1_20x,
				KeyAuths: []*kong.KeyAuth{
					{
						Consumer: &kong.Consumer{
							ID: kong.String("87095815-5395-454e-8c18-a11c9bc0ef04"),
						},
						Key: kong.String("i-am-special"),
					},
					{
						Consumer: &kong.Consumer{
							ID: kong.String("5a5b9369-baeb-4faa-a902-c40ccdc2928e"),
						},
						Key: kong.String("i-am-not-so-special"),
					},
					{
						Consumer: &kong.Consumer{
							ID: kong.String("e894ea9e-ad08-4acf-a960-5a23aa7701c7"),
						},
						Key: kong.String("i-am-just-average"),
					},
				},
			},
		},
		{
			name:     "creates consumer groups scoped plugins",
			runWhen:  ">=3.10.0",
			kongFile: "testdata/sync/025-consumer-groups-scoped-plugins/kong3x.yaml",
			expectedState: utils.KongRawState{
				Consumers: consumerGroupsConsumers,
				ConsumerGroups: []*kong.ConsumerGroupObject{
					{
						ConsumerGroup: &kong.ConsumerGroup{
							Name: kong.String("silver"),
						},
						Consumers: []*kong.Consumer{
							{
								Username: kong.String("bar"),
							},
						},
					},
					{
						ConsumerGroup: &kong.ConsumerGroup{
							Name: kong.String("gold"),
						},
						Consumers: []*kong.Consumer{
							{
								Username: kong.String("foo"),
							},
						},
					},
				},
				Plugins:  consumerGroupScopedPlugins310x,
				Services: svc1_207,
				Routes:   route1_20x,
				KeyAuths: []*kong.KeyAuth{
					{
						Consumer: &kong.Consumer{
							ID: kong.String("87095815-5395-454e-8c18-a11c9bc0ef04"),
						},
						Key: kong.String("i-am-special"),
					},
					{
						Consumer: &kong.Consumer{
							ID: kong.String("5a5b9369-baeb-4faa-a902-c40ccdc2928e"),
						},
						Key: kong.String("i-am-not-so-special"),
					},
					{
						Consumer: &kong.Consumer{
							ID: kong.String("e894ea9e-ad08-4acf-a960-5a23aa7701c7"),
						},
						Key: kong.String("i-am-just-average"),
					},
				},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name+"/"+tc.runWhen, func(t *testing.T) {
			runWhen(t, "enterprise", tc.runWhen)

			setup(t)

			require.NoError(t, sync(context.Background(), tc.kongFile))
			testKongState(t, client, false, false, tc.expectedState, nil)

			// Kong proxy may need a bit to be ready.
			time.Sleep(time.Second * 10)

			// build simple http client
			client := &http.Client{}

			// test 'foo' consumer (part of 'gold' group)
			req, err := http.NewRequest("GET", "http://localhost:8000/r1", nil)
			require.NoError(t, err)
			req.Header.Add("apikey", "i-am-special")
			n := 0
			for n < 11 {
				resp, err := client.Do(req)
				require.NoError(t, err)
				defer resp.Body.Close()
				if resp.StatusCode == http.StatusTooManyRequests {
					break
				}
				n++
			}
			assert.Equal(t, maxGoldRequestsNumber, n)

			// test 'bar' consumer (part of 'silver' group)
			req, err = http.NewRequest("GET", "http://localhost:8000/r1", nil)
			require.NoError(t, err)
			req.Header.Add("apikey", "i-am-not-so-special")
			n = 0
			for n < 11 {
				resp, err := client.Do(req)
				require.NoError(t, err)
				defer resp.Body.Close()
				if resp.StatusCode == http.StatusTooManyRequests {
					break
				}
				n++
			}
			assert.Equal(t, maxSilverRequestsNumber, n)

			// test 'baz' consumer (not part of any group)
			req, err = http.NewRequest("GET", "http://localhost:8000/r1", nil)
			require.NoError(t, err)
			req.Header.Add("apikey", "i-am-just-average")
			n = 0
			for n < 11 {
				resp, err := client.Do(req)
				require.NoError(t, err)
				defer resp.Body.Close()
				if resp.StatusCode == http.StatusTooManyRequests {
					break
				}
				n++
			}
			assert.Equal(t, maxRegularRequestsNumber, n)
		})
	}
}

// test scope:
//   - > 3.4.0
func Test_Sync_ConsumerGroupsScopedPlugins_Post340(t *testing.T) {
	tests := []struct {
		name          string
		kongFile      string
		expectedError error
	}{
		{
			name:          "attempt to create deprecated consumer groups configuration with Kong version >= 3.4.0 fails",
			kongFile:      "testdata/sync/017-consumer-groups-rla-application/kong3x.yaml",
			expectedError: fmt.Errorf("building state: %w", utils.ErrorConsumerGroupUpgrade),
		},
		{
			name:     "empty deprecated consumer groups configuration fields do not fail with Kong version >= 3.4.0",
			kongFile: "testdata/sync/017-consumer-groups-rla-application/kong3x-empty-application.yaml",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runWhen(t, "enterprise", ">=3.4.0")
			setup(t)

			err := sync(context.Background(), tc.kongFile)
			if tc.expectedError == nil {
				require.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError.Error())
			}
		})
	}
}

func Test_Sync_ConsumerGroupsScopedPluginsKonnect(t *testing.T) {
	client, err := getTestClient()
	require.NoError(t, err)

	ignoreFields := []cmp.Option{
		cmp.FilterValues(func(x, y interface{}) bool {
			_, okX := x.(map[string]interface{})
			_, okY := y.(map[string]interface{})
			return okX && okY
		}, cmpopts.IgnoreMapEntries(func(key string, _ interface{}) bool {
			return key == "redis"
		})),
	}

	tests := []struct {
		name          string
		kongFile      string
		expectedState utils.KongRawState
	}{
		{
			name:     "creates consumer groups scoped plugins",
			kongFile: "testdata/sync/025-consumer-groups-scoped-plugins/kong3x.yaml",
			expectedState: utils.KongRawState{
				Consumers: consumerGroupsConsumers,
				ConsumerGroups: []*kong.ConsumerGroupObject{
					{
						ConsumerGroup: &kong.ConsumerGroup{
							Name: kong.String("silver"),
						},
						Consumers: []*kong.Consumer{
							{
								Username: kong.String("bar"),
							},
						},
					},
					{
						ConsumerGroup: &kong.ConsumerGroup{
							Name: kong.String("gold"),
						},
						Consumers: []*kong.Consumer{
							{
								Username: kong.String("foo"),
							},
						},
					},
				},
				Plugins:  consumerGroupScopedPluginsKonnect,
				Services: svc1_207,
				Routes:   route1_20x,
				KeyAuths: []*kong.KeyAuth{
					{
						Consumer: &kong.Consumer{
							ID: kong.String("87095815-5395-454e-8c18-a11c9bc0ef04"),
						},
						Key: kong.String("i-am-special"),
					},
					{
						Consumer: &kong.Consumer{
							ID: kong.String("5a5b9369-baeb-4faa-a902-c40ccdc2928e"),
						},
						Key: kong.String("i-am-not-so-special"),
					},
					{
						Consumer: &kong.Consumer{
							ID: kong.String("e894ea9e-ad08-4acf-a960-5a23aa7701c7"),
						},
						Key: kong.String("i-am-just-average"),
					},
				},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runWhenKonnect(t)
			setup(t)

			require.NoError(t, sync(context.Background(), tc.kongFile))
			testKongState(t, client, true, false, tc.expectedState, ignoreFields)
		})
	}
}

// test scope:
//   - konnect
func Test_Sync_KonnectRename(t *testing.T) {
	// setup stage
	tests := []struct {
		name             string
		controlPlaneName string
		runtimeGroupName string
		kongFile         string
		flags            []string
		expectedState    utils.KongRawState
	}{
		{
			name:     "konnect-runtime-group-name flag - default",
			kongFile: "testdata/sync/026-konnect-rename/default.yaml",
			flags:    []string{"--konnect-runtime-group-name", "default"},
			expectedState: utils.KongRawState{
				Services: defaultCPService,
			},
		},
		{
			name:     "konnect-control-plane-name flag - default",
			kongFile: "testdata/sync/026-konnect-rename/default.yaml",
			flags:    []string{"--konnect-control-plane-name", "default"},
			expectedState: utils.KongRawState{
				Services: defaultCPService,
			},
		},
		{
			name:             "konnect-runtime-group-name flag - test",
			runtimeGroupName: "test",
			kongFile:         "testdata/sync/026-konnect-rename/test.yaml",
			flags:            []string{"--konnect-runtime-group-name", "test"},
			expectedState: utils.KongRawState{
				Services: testCPService,
			},
		},
		{
			name:             "konnect-control-plane-name flag - test",
			controlPlaneName: "test",
			kongFile:         "testdata/sync/026-konnect-rename/test.yaml",
			flags:            []string{"--konnect-control-plane-name", "test"},
			expectedState: utils.KongRawState{
				Services: testCPService,
			},
		},
		{
			name:     "konnect.runtime_group_name - default",
			kongFile: "testdata/sync/026-konnect-rename/konnect_default_rg.yaml",
			expectedState: utils.KongRawState{
				Services: defaultCPService,
			},
		},
		{
			name:     "konnect.control_plane_name - default",
			kongFile: "testdata/sync/026-konnect-rename/konnect_default_cp.yaml",
			expectedState: utils.KongRawState{
				Services: defaultCPService,
			},
		},
		{
			name:             "konnect.runtime_group_name - test",
			runtimeGroupName: "test",
			kongFile:         "testdata/sync/026-konnect-rename/konnect_test_rg.yaml",
			expectedState: utils.KongRawState{
				Services: testCPService,
			},
		},
		{
			name:             "konnect.control_plane_name - test",
			controlPlaneName: "test",
			kongFile:         "testdata/sync/026-konnect-rename/konnect_test_cp.yaml",
			expectedState: utils.KongRawState{
				Services: testCPService,
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runWhenKonnect(t)
			setup(t)
			if tc.controlPlaneName != "" {
				t.Setenv("DECK_KONNECT_CONTROL_PLANE_NAME", tc.controlPlaneName)
				t.Cleanup(func() {
					reset(t, "--konnect-control-plane-name", tc.controlPlaneName)
				})
			} else if tc.runtimeGroupName != "" {
				t.Setenv("DECK_KONNECT_RUNTIME_GROUP_NAME", tc.runtimeGroupName)
				t.Cleanup(func() {
					reset(t, "--konnect-runtime-group-name", tc.runtimeGroupName)
				})
			}
			client, err := getTestClient()
			require.NoError(t, err)

			require.NoError(t, sync(context.Background(), tc.kongFile, tc.flags...))
			testKongState(t, client, true, false, tc.expectedState, nil)
		})
	}
}

func Test_Sync_KonnectRenameErrors(t *testing.T) {
	tests := []struct {
		name          string
		kongFile      string
		flags         []string
		expectedError error
	}{
		{
			name:     "different runtime group names fail",
			kongFile: "testdata/sync/026-konnect-rename/konnect_default_cp.yaml",
			flags:    []string{"--konnect-runtime-group-name", "rg1"},
			expectedError: errors.New(`warning: control plane 'rg1' specified via ` +
				`--konnect-[control-plane|runtime-group]-name flag is different from 'default' found in state file(s)`),
		},
		{
			name:     "different runtime group names fail",
			kongFile: "testdata/sync/026-konnect-rename/konnect_default_rg.yaml",
			flags:    []string{"--konnect-runtime-group-name", "rg1"},
			expectedError: errors.New(`warning: control plane 'rg1' specified via ` +
				`--konnect-[control-plane|runtime-group]-name flag is different from 'default' found in state file(s)`),
		},
		{
			name:     "different control plane names fail",
			kongFile: "testdata/sync/026-konnect-rename/konnect_default_cp.yaml",
			flags:    []string{"--konnect-control-plane-name", "cp1"},
			expectedError: errors.New(`warning: control plane 'cp1' specified via ` +
				`--konnect-[control-plane|runtime-group]-name flag is different from 'default' found in state file(s)`),
		},
		{
			name:     "different control plane names fail",
			kongFile: "testdata/sync/026-konnect-rename/konnect_default_rg.yaml",
			flags:    []string{"--konnect-control-plane-name", "cp1"},
			expectedError: errors.New(`warning: control plane 'cp1' specified via ` +
				`--konnect-[control-plane|runtime-group]-name flag is different from 'default' found in state file(s)`),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := sync(context.Background(), tc.kongFile, tc.flags...)
			assert.Equal(t, tc.expectedError, err)
		})
	}
}

// test scope:
//   - 3.0.0+
func Test_Sync_DoNotUpdateCreatedAt(t *testing.T) {
	runWhen(t, "kong", ">=3.0.0")
	setup(t)

	client, err := getTestClient()
	require.NoError(t, err)

	const (
		oldConfig = "testdata/sync/027-created-at/old.yaml"
		newConfig = "testdata/sync/027-created-at/new.yaml"
	)

	// provision entities
	require.NoError(t, sync(context.Background(), oldConfig))

	// get the current state
	ctx := context.Background()
	oldKongState, err := deckDump.Get(ctx, client, deckDump.Config{})
	require.NoError(t, err)

	// update entities
	time.Sleep(time.Second)
	require.NoError(t, sync(context.Background(), newConfig))

	// get the new state
	newKongState, err := deckDump.Get(ctx, client, deckDump.Config{})
	require.NoError(t, err)

	// verify that the created_at have not changed across deployments
	require.Equal(t, oldKongState.Services[0].CreatedAt, newKongState.Services[0].CreatedAt)
	require.Equal(t, oldKongState.Routes[0].CreatedAt, newKongState.Routes[0].CreatedAt)
	require.Equal(t, oldKongState.Plugins[0].CreatedAt, newKongState.Plugins[0].CreatedAt)
	require.Equal(t, oldKongState.Consumers[0].CreatedAt, newKongState.Consumers[0].CreatedAt)

	// verify that the updated_at have changed across deployments
	require.NotEqual(t, oldKongState.Services[0].UpdatedAt, newKongState.Services[0].UpdatedAt)
	require.NotEqual(t, oldKongState.Routes[0].UpdatedAt, newKongState.Routes[0].UpdatedAt)
	// plugins do not have an updated_at field
	// consumers do not have an updated_at field
}

// Test_Sync_LookupConsumerTags tests that existing behavior when referencing
// consumers from plugins is preserved:
// - if a referenced consumer is not present in the state file, the sync fails
// - if a referenced consumer is present in the state file, the sync succeeds
//
// This test also tests that the new behavior is implemented correctly:
//   - if a referenced consumer is not present in the state file, but is present
//     in Kong when using the new lookup selector tags, the sync succeeds
//   - if a referenced consumer is not present in the state file and neither in
//     Kong when using the new lookup selector tags, the sync fails
func Test_Sync_LookupConsumerTags(t *testing.T) {
	runWhen(t, "enterprise", ">=3.0.0")
	setup(t)

	// test that reference to non-existing consumer fails.
	pluginsNoLookupStateFile := "testdata/sync/029-lookup-tags-consumers/plugins_no_lookup.yaml"
	err := sync(context.Background(), pluginsNoLookupStateFile)
	require.Error(t, err)
	require.EqualError(t, err, "building state: consumer foo for plugin rate-limiting-advanced: entity not found")

	// test that reference to existing local consumer succeeds.
	pluginsAndConsumersStateFile := "testdata/sync/029-lookup-tags-consumers/plugins_and_consumers.yaml"
	require.NoError(t, sync(context.Background(), pluginsAndConsumersStateFile))
	reset(t)

	// test that reference to existing global consumer succeeds via lookup tags.
	globalConsumersStateFile := "testdata/sync/029-lookup-tags-consumers/global_consumers.yaml"
	require.NoError(t, sync(context.Background(), globalConsumersStateFile))
	// sync plugins with lookup reference to global consumers.
	pluginsLookupStateFile := "testdata/sync/029-lookup-tags-consumers/plugins_lookup.yaml"
	require.NoError(t, sync(context.Background(), pluginsLookupStateFile))
	reset(t)

	// test that reference to non-existing global consumer fails via lookup tags.
	require.Error(t, sync(context.Background(), pluginsLookupStateFile))
	require.EqualError(t, err, "building state: consumer foo for plugin rate-limiting-advanced: entity not found")
}

// test scope:
//   - 3.0.0+
func Test_Sync_ConsumerGroupConsumersWithCustomID(t *testing.T) {
	runWhen(t, "enterprise", ">=3.0.0")
	setup(t)

	client, err := getTestClient()
	require.NoError(t, err)

	expectedState := utils.KongRawState{
		ConsumerGroups: []*kong.ConsumerGroupObject{
			{
				ConsumerGroup: &kong.ConsumerGroup{
					ID:   kong.String("48df7cd3-1cd0-4e53-af73-8f57f257be18"),
					Name: kong.String("cg1"),
				},
				Consumers: []*kong.Consumer{
					{
						ID:       kong.String("bcb296c3-22bb-46f6-99c8-4828af750b77"),
						CustomID: kong.String("foo"),
					},
				},
			},
			{
				ConsumerGroup: &kong.ConsumerGroup{
					ID:   kong.String("1a81dc83-5329-4666-8ae7-8a966e62d076"),
					Name: kong.String("cg2"),
				},
				Consumers: []*kong.Consumer{
					{
						ID:       kong.String("562bf5c7-a7d9-4338-84dd-2c1064fb7f67"),
						Username: kong.String("foo"),
					},
				},
			},
			{
				ConsumerGroup: &kong.ConsumerGroup{
					ID:   kong.String("d140f9cc-227e-4872-8b0b-639f6922dfb0"),
					Name: kong.String("cg3"),
				},
				Consumers: []*kong.Consumer{
					{
						ID:       kong.String("7906968b-cd89-4a87-8dda-94678e7106b2"),
						Username: kong.String("bar"),
						CustomID: kong.String("custom_bar"),
					},
				},
			},
		},
		Consumers: []*kong.Consumer{
			{
				ID:       kong.String("bcb296c3-22bb-46f6-99c8-4828af750b77"),
				CustomID: kong.String("foo"),
			},
			{
				ID:       kong.String("562bf5c7-a7d9-4338-84dd-2c1064fb7f67"),
				Username: kong.String("foo"),
			},
			{
				ID:       kong.String("7906968b-cd89-4a87-8dda-94678e7106b2"),
				Username: kong.String("bar"),
				CustomID: kong.String("custom_bar"),
			},
		},
	}
	require.NoError(t, sync(context.Background(), "testdata/sync/028-consumer-group-consumers-custom_id/kong.yaml"))
	testKongState(t, client, false, false, expectedState, nil)
}

// test scope:
//   - konnect
func Test_Sync_ConsumerGroupConsumersWithCustomID_Konnect(t *testing.T) {
	t.Setenv("DECK_KONNECT_CONTROL_PLANE_NAME", "default")
	runWhen(t, "konnect", "")
	setup(t)

	client, err := getTestClient()
	require.NoError(t, err)

	expectedState := utils.KongRawState{
		ConsumerGroups: []*kong.ConsumerGroupObject{
			{
				ConsumerGroup: &kong.ConsumerGroup{
					ID:   kong.String("48df7cd3-1cd0-4e53-af73-8f57f257be18"),
					Name: kong.String("cg1"),
				},
				Consumers: []*kong.Consumer{
					{
						ID:       kong.String("bcb296c3-22bb-46f6-99c8-4828af750b77"),
						CustomID: kong.String("foo"),
					},
				},
			},
			{
				ConsumerGroup: &kong.ConsumerGroup{
					ID:   kong.String("1a81dc83-5329-4666-8ae7-8a966e62d076"),
					Name: kong.String("cg2"),
				},
				Consumers: []*kong.Consumer{
					{
						ID:       kong.String("562bf5c7-a7d9-4338-84dd-2c1064fb7f67"),
						Username: kong.String("foo"),
					},
				},
			},
			{
				ConsumerGroup: &kong.ConsumerGroup{
					ID:   kong.String("d140f9cc-227e-4872-8b0b-639f6922dfb0"),
					Name: kong.String("cg3"),
				},
				Consumers: []*kong.Consumer{
					{
						ID:       kong.String("7906968b-cd89-4a87-8dda-94678e7106b2"),
						Username: kong.String("bar"),
						CustomID: kong.String("custom_bar"),
					},
				},
			},
		},
		Consumers: []*kong.Consumer{
			{
				ID:       kong.String("bcb296c3-22bb-46f6-99c8-4828af750b77"),
				CustomID: kong.String("foo"),
			},
			{
				ID:       kong.String("562bf5c7-a7d9-4338-84dd-2c1064fb7f67"),
				Username: kong.String("foo"),
			},
			{
				ID:       kong.String("7906968b-cd89-4a87-8dda-94678e7106b2"),
				Username: kong.String("bar"),
				CustomID: kong.String("custom_bar"),
			},
		},
	}
	require.NoError(t, sync(context.Background(), "testdata/sync/028-consumer-group-consumers-custom_id/kong.yaml"))
	testKongState(t, client, true, false, expectedState, nil)
}

// Test_Sync_LookupServicesTags tests that existing behavior when referencing
// services from plugins is preserved:
// - if a referenced service is not present in the state file, the sync fails
// - if a referenced service is present in the state file, the sync succeeds
//
// This test also tests that the new behavior is implemented correctly:
//   - if a referenced service is not present in the state file, but is present
//     in Kong when using the new lookup selector tags, the sync succeeds
//   - if a referenced service is not present in the state file and neither in
//     Kong when using the new lookup selector tags, the sync fails
func Test_Sync_LookupServicesTags(t *testing.T) {
	runWhen(t, "enterprise", ">=3.0.0")
	setup(t)

	// test that reference to non-existing service fails.
	pluginsNoLookupServiceStateFile := "testdata/sync/035-lookup-tags-services/plugins_no_lookup.yaml"
	err := sync(context.Background(), pluginsNoLookupServiceStateFile)
	require.Error(t, err)
	require.EqualError(t, err, "building state: service foo for plugin rate-limiting-advanced: entity not found")

	// test that reference to existing local service succeeds.
	pluginsAndServicesStateFile := "testdata/sync/035-lookup-tags-services/plugins_and_services.yaml"
	require.NoError(t, sync(context.Background(), pluginsAndServicesStateFile))
	reset(t)

	// test that reference to existing global service succeeds via lookup tags.
	globalServicesStateFile := "testdata/sync/035-lookup-tags-services/global_services.yaml"
	require.NoError(t, sync(context.Background(), globalServicesStateFile))

	// sync plugins with lookup reference to global services.
	pluginsLookupServiceStateFile := "testdata/sync/035-lookup-tags-services/plugins_lookup.yaml"
	require.NoError(t, sync(context.Background(), pluginsLookupServiceStateFile))
	reset(t)

	// test that reference to non-existing global service fails via lookup tags.
	require.Error(t, sync(context.Background(), pluginsLookupServiceStateFile))
	require.EqualError(t, err, "building state: service foo for plugin rate-limiting-advanced: entity not found")
}

// Test_Sync_LookupRoutesTags tests that existing behavior when referencing
// routes from plugins is preserved:
// - if a referenced route is not present in the state file, the sync fails
// - if a referenced route is present in the state file, the sync succeeds
//
// This test also tests that the new behavior is implemented correctly:
//   - if a referenced route is not present in the state file, but is present
//     in Kong when using the new lookup selector tags, the sync succeeds
//   - if a referenced route is not present in the state file and neither in
//     Kong when using the new lookup selector tags, the sync fails
func Test_Sync_LookupRoutesTags(t *testing.T) {
	runWhen(t, "enterprise", ">=3.0.0")
	setup(t)

	// test that reference to non-existing route fails.
	pluginsNoLookupStateFile := "testdata/sync/030-lookup-tags-routes/plugins_no_lookup.yaml"
	err := sync(context.Background(), pluginsNoLookupStateFile)
	require.Error(t, err)
	require.EqualError(t, err, "building state: route foo for plugin rate-limiting-advanced: entity not found")

	// test that reference to existing local route succeeds.
	pluginsAndRoutesStateFile := "testdata/sync/030-lookup-tags-routes/plugins_and_routes.yaml"
	require.NoError(t, sync(context.Background(), pluginsAndRoutesStateFile))
	reset(t)

	// test that reference to existing global route succeeds via lookup tags.
	globalRoutesStateFile := "testdata/sync/030-lookup-tags-routes/global_routes.yaml"
	require.NoError(t, sync(context.Background(), globalRoutesStateFile))
	// sync plugins with lookup reference to global routes.
	pluginsLookupStateFile := "testdata/sync/030-lookup-tags-routes/plugins_lookup.yaml"
	require.NoError(t, sync(context.Background(), pluginsLookupStateFile))
	reset(t)

	// test that reference to non-existing global route fails via lookup tags.
	require.Error(t, sync(context.Background(), pluginsLookupStateFile))
	require.EqualError(t, err, "building state: route foo for plugin rate-limiting-advanced: entity not found")
}

// Test_Sync_LookupConsumerGroupsTags tests that existing behavior when referencing
// consumer groups from plugins is preserved:
// - if a referenced service/route and consumer group are not present in the state file, the sync fails
// - if a referenced service/route and consumer group are present in the state file, the sync succeeds
//
// This test also tests that the new behavior is implemented correctly:
//   - if a referenced service/route/consumer or consumer group is not present in the state file, but is present
//     in Kong when using the new lookup selector tags, the sync succeeds
//   - if a referenced service/route or consumer group is not present in the state file and neither in
//     Kong when using the new lookup selector tags, the sync fails
func Test_Sync_LookupConsumerGroupsTags(t *testing.T) {
	runWhen(t, "enterprise", ">=3.0.0")
	setup(t)

	// test that reference to non-existing service fails.
	pluginsNoLookupServiceStateFile := "testdata/sync/034-lookup-tags-consumerGroups/plugins_no_lookup_service.yaml"
	errNoService := sync(context.Background(), pluginsNoLookupServiceStateFile)
	require.Error(t, errNoService)
	require.EqualError(t, errNoService, "building state: service foo for plugin rate-limiting-advanced: entity not found")

	// test that reference to non-existing route fails.
	pluginsNoLookupRouteStateFile := "testdata/sync/034-lookup-tags-consumerGroups/plugins_no_lookup_route.yaml"
	errNoRoute := sync(context.Background(), pluginsNoLookupRouteStateFile)
	require.Error(t, errNoRoute)
	require.EqualError(t, errNoRoute, "building state: route bar for plugin rate-limiting-advanced: entity not found")

	// test that reference to non-existing consumer group fails.
	pluginsNoLookupConsumergroupStateFile := "testdata/sync/034-lookup-tags-consumerGroups/plugins_no_lookup_consumerGroup.yaml" //nolint:lll
	errNoConsumerGroup := sync(context.Background(), pluginsNoLookupConsumergroupStateFile)
	require.Error(t, errNoConsumerGroup)
	require.EqualError(
		t,
		errNoConsumerGroup,
		"building state: consumer-group foo2 for plugin rate-limiting-advanced: entity not found",
	)

	// test that reference to existing local service and consumer group succeeds.
	pluginsAndEntitiesConsumersGroupsStateFile := "testdata/sync/034-lookup-tags-consumerGroups/plugins_and_entities_consumerGroups.yaml" //nolint:lll
	require.NoError(t, sync(context.Background(), pluginsAndEntitiesConsumersGroupsStateFile))
	reset(t)

	// test that reference to existing global service and consumer group succeeds via lookup tags.
	globalEntitiesConsumerGroupsStateFile := "testdata/sync/034-lookup-tags-consumerGroups/global_entities_consumerGroups.yaml" //nolint:lll
	require.NoError(t, sync(context.Background(), globalEntitiesConsumerGroupsStateFile))
	// sync plugins with lookup reference to global service.
	pluginsServiceLookupStateFile := "testdata/sync/034-lookup-tags-consumerGroups/plugins_service_lookup.yaml"
	require.NoError(t, sync(context.Background(), pluginsServiceLookupStateFile))
	// sync plugins with lookup reference to global route.
	pluginsRouteLookupStateFile := "testdata/sync/034-lookup-tags-consumerGroups/plugins_route_lookup.yaml"
	require.NoError(t, sync(context.Background(), pluginsRouteLookupStateFile))
	reset(t)

	// test that reference to non-existing global service or consumer groups fails via lookup tags.
	errServiceNoReference := sync(context.Background(), pluginsServiceLookupStateFile)
	require.Error(t, errServiceNoReference)
	require.EqualError(
		t,
		errServiceNoReference,
		"building state: service foo for plugin rate-limiting-advanced: entity not found",
	)

	// test that reference to non-existing global route or consumer groups fails via lookup tags.
	errRouteNoReference := sync(context.Background(), pluginsRouteLookupStateFile)
	require.Error(t, errRouteNoReference)
	require.EqualError(
		t,
		errRouteNoReference,
		"building state: route bar for plugin rate-limiting-advanced: entity not found",
	)
}

// test scope:
//   - 3.5.0+
func Test_Sync_ConsumerGroupConsumerFromUpstream(t *testing.T) {
	runWhen(t, "enterprise", ">=3.4.0")
	setup(t)

	client, err := getTestClient()
	require.NoError(t, err)

	expectedState := utils.KongRawState{
		ConsumerGroups: []*kong.ConsumerGroupObject{
			{
				ConsumerGroup: &kong.ConsumerGroup{
					ID:   kong.String("c0f6c818-470c-4df7-8515-c8e904765fcc"),
					Name: kong.String("group-1"),
					Tags: kong.StringSlice("project:the-project", "managed-by:deck"),
				},
				Consumers: []*kong.Consumer{
					{
						ID:       kong.String("97cab250-1b0a-4119-aa2e-0756e8931034"),
						Username: kong.String("consumer-1"),
						Tags:     kong.StringSlice("project:the-project", "managed-by:the-background-process"),
					},
				},
			},
		},
		Consumers: []*kong.Consumer{
			{
				ID:       kong.String("97cab250-1b0a-4119-aa2e-0756e8931034"),
				Username: kong.String("consumer-1"),
				Tags:     kong.StringSlice("project:the-project", "managed-by:the-background-process"),
			},
		},
	}

	// simulate the following scenario:
	// - a consumer-group defined with a set of tags, ideally managed by decK
	// - a consumer defined with another set of tags, ideally managed by an external process
	// - the consumer -> consumer-group relationship, ideally managed by an external process
	require.NoError(t, sync(context.Background(), "testdata/sync/031-consumer-group-consumers-from-upstream/initial.yaml"))
	testKongState(t, client, false, false, expectedState, nil)

	// referencing the relationship in a file without the consumer would still work
	// if default_lookup_tags are defined to pull consumers from upstream.
	require.NoError(t, sync(context.Background(), "testdata/sync/031-consumer-group-consumers-from-upstream/consumer-groups.yaml")) //nolint:lll
	testKongState(t, client, false, false, expectedState, nil)
}

// test scope:
//   - konnect
func Test_Sync_ConsumerGroupConsumerFromUpstream_Konnect(t *testing.T) {
	t.Setenv("DECK_KONNECT_CONTROL_PLANE_NAME", "default")
	runWhen(t, "konnect", "")
	setup(t)

	client, err := getTestClient()
	require.NoError(t, err)

	expectedState := utils.KongRawState{
		ConsumerGroups: []*kong.ConsumerGroupObject{
			{
				ConsumerGroup: &kong.ConsumerGroup{
					ID:   kong.String("c0f6c818-470c-4df7-8515-c8e904765fcc"),
					Name: kong.String("group-1"),
					Tags: kong.StringSlice("project:the-project", "managed-by:deck"),
				},
				Consumers: []*kong.Consumer{
					{
						ID:       kong.String("97cab250-1b0a-4119-aa2e-0756e8931034"),
						Username: kong.String("consumer-1"),
						Tags:     kong.StringSlice("project:the-project", "managed-by:the-background-process"),
					},
				},
			},
		},
		Consumers: []*kong.Consumer{
			{
				ID:       kong.String("97cab250-1b0a-4119-aa2e-0756e8931034"),
				Username: kong.String("consumer-1"),
				Tags:     kong.StringSlice("project:the-project", "managed-by:the-background-process"),
			},
		},
	}

	// simulate the following scenario:
	// - a consumer-group defined with a set of tags, ideally managed by decK
	// - a consumer defined with another set of tags, ideally managed by an external process
	// - the consumer -> consumer-group relationship, ideally managed by an external process
	require.NoError(t, sync(context.Background(), "testdata/sync/031-consumer-group-consumers-from-upstream/initial.yaml"))
	testKongState(t, client, true, false, expectedState, nil)

	// referencing the relationship in a file without the consumer would still work
	// if default_lookup_tags are defined to pull consumers from upstream.
	require.NoError(t, sync(context.Background(), "testdata/sync/031-consumer-group-consumers-from-upstream/consumer-groups.yaml")) //nolint:lll
	testKongState(t, client, true, false, expectedState, nil)
}

// test scope:
//   - 3.6.0+
func Test_Sync_ConsumerGroupConsumerWithTags(t *testing.T) {
	runWhen(t, "enterprise", ">=3.6.0")
	setup(t)

	client, err := getTestClient()
	require.NoError(t, err)

	expectedState := utils.KongRawState{
		ConsumerGroups: []*kong.ConsumerGroupObject{
			{
				ConsumerGroup: &kong.ConsumerGroup{
					ID:   kong.String("c0f6c818-470c-4df7-8515-c8e904765fcc"),
					Name: kong.String("group-1"),
					Tags: kong.StringSlice("project:the-project", "managed-by:deck"),
				},
				Consumers: []*kong.Consumer{
					{
						ID:       kong.String("97cab250-1b0a-4119-aa2e-0756e8931034"),
						Username: kong.String("consumer-1"),
					},
				},
			},
		},
		Consumers: []*kong.Consumer{
			{
				ID:       kong.String("97cab250-1b0a-4119-aa2e-0756e8931034"),
				Username: kong.String("consumer-1"),
			},
		},
	}

	// simulate the following scenario:
	// - a consumer-group defined with a set of tags, ideally managed by decK
	// - a consumer and the consumer -> consumer-group managed externally
	require.NoError(t, sync(context.Background(), "testdata/sync/032-consumer-group-consumers-with-tags/initial.yaml"))

	// create the consumer
	_, err = client.Consumers.Create(
		context.Background(),
		&kong.Consumer{
			ID:       kong.String("97cab250-1b0a-4119-aa2e-0756e8931034"),
			Username: kong.String("consumer-1"),
		},
	)
	require.NoError(t, err)

	// map the consumer into the consumer-group
	_, err = client.ConsumerGroupConsumers.Create(
		context.Background(),
		kong.String("c0f6c818-470c-4df7-8515-c8e904765fcc"),
		kong.String("97cab250-1b0a-4119-aa2e-0756e8931034"),
	)
	require.NoError(t, err)

	// re-sync again
	require.NoError(t, sync(context.Background(), "testdata/sync/032-consumer-group-consumers-with-tags/initial.yaml"))
	testKongState(t, client, false, false, expectedState, nil)
}

func Test_Sync_FilterChains(t *testing.T) {
	t.Skip("Skipping test till wasm/filter-chains issue is not resolved at gateway level")
	runWhen(t, "kong", ">=3.4.0")
	setup(t)

	client, err := getTestClient()
	require.NoError(t, err)

	service := kong.Service{
		ID:             kong.String("58076db2-28b6-423b-ba39-a797193017f7"),
		Name:           kong.String("test"),
		ConnectTimeout: kong.Int(60000),
		Enabled:        kong.Bool(true),
		Host:           kong.String("test"),
		Port:           kong.Int(8080),
		Protocol:       kong.String("http"),
		ReadTimeout:    kong.Int(60000),
		Retries:        kong.Int(5),
		WriteTimeout:   kong.Int(60000),
	}

	route := kong.Route{
		ID:                      kong.String("37fc74bd-bac6-4bce-bd54-6ec4d341c1c1"),
		Name:                    kong.String("r1"),
		Paths:                   kong.StringSlice("/r1"),
		PathHandling:            kong.String("v0"),
		PreserveHost:            kong.Bool(false),
		Protocols:               kong.StringSlice("http", "https"),
		RegexPriority:           new(int),
		StripPath:               kong.Bool(true),
		HTTPSRedirectStatusCode: kong.Int(426),
		RequestBuffering:        kong.Bool(true),
		ResponseBuffering:       kong.Bool(true),
		Service: &kong.Service{
			ID: service.ID,
		},
	}

	serviceChain := kong.FilterChain{
		Name:    kong.String("service"),
		ID:      kong.String("d3ed9313-acf2-4982-af3c-83d9b52795c3"),
		Enabled: kong.Bool(true),
		Service: &kong.Service{
			ID: service.ID,
		},
		Filters: []*kong.Filter{
			{
				Name:    kong.String("response_transformer"),
				Enabled: kong.Bool(true),
				Config:  kong.JSONRawMessage(`"{\n  \"add\": {\n    \"headers\": [\n      \"x-service:test\"\n    ]\n  }\n}\n"`),
			},
		},
	}

	routeChain := kong.FilterChain{
		Name:    kong.String("route"),
		ID:      kong.String("7b95fe94-df9b-421a-8a7b-c4d8fddbf363"),
		Enabled: kong.Bool(true),
		Route: &kong.Route{
			ID: route.ID,
		},
		Filters: []*kong.Filter{
			{
				Name:    kong.String("response_transformer"),
				Enabled: kong.Bool(true),
				Config:  kong.JSONRawMessage(`"{\n  \"add\": {\n    \"headers\": [\n      \"x-route:test\"\n    ]\n  }\n}\n"`),
			},
		},
	}

	tests := []struct {
		name        string
		version     string
		createFile  string
		createState func(*utils.KongRawState)
		updateFile  string
		updateState func(*utils.KongRawState)
		deleteFile  string
		deleteState func(*utils.KongRawState)
	}{
		// in Kong 3.4.x, filter configurations must always be strings
		{
			name:       "kong 3.4.x",
			version:    "<3.5.0",
			createFile: "testdata/sync/033-filter-chains/create-3.4.x.yaml",
			createState: func(state *utils.KongRawState) {
				state.FilterChains = []*kong.FilterChain{
					routeChain.DeepCopy(),
					serviceChain.DeepCopy(),
				}
			},
			updateFile: "testdata/sync/033-filter-chains/update-3.4.x.yaml",
			updateState: func(state *utils.KongRawState) {
				state.FilterChains[0].Filters[0].Enabled = kong.Bool(false)
				cfg := kong.JSONRawMessage(`"{\n  \"add\": {\n    \"headers\": [\n      \"x-service:CHANGED\"\n    ]\n  }\n}\n"`)
				state.FilterChains[1].Filters[0].Config = cfg
			},
			deleteFile: "testdata/sync/033-filter-chains/delete-3.4.x.yaml",
			deleteState: func(state *utils.KongRawState) {
				state.FilterChains = []*kong.FilterChain{
					state.FilterChains[0],
				}
			},
		},
		// kong 3.5.0 introduced optional JSON[Schema] support for filter configurations
		{
			name:       "kong >=3.5",
			version:    ">=3.5.0",
			createFile: "testdata/sync/033-filter-chains/create.yaml",
			createState: func(state *utils.KongRawState) {
				rc := routeChain.DeepCopy()
				rc.Filters[0].Config = kong.JSONRawMessage(`{"add":{"headers":["x-route:test"]}}`)

				sc := serviceChain.DeepCopy()
				sc.Filters[0].Config = kong.JSONRawMessage(`{"add":{"headers":["x-service:test"]}}`)

				state.FilterChains = []*kong.FilterChain{
					rc,
					sc,
				}
			},
			updateFile: "testdata/sync/033-filter-chains/update.yaml",
			updateState: func(state *utils.KongRawState) {
				state.FilterChains[0].Filters[0].Enabled = kong.Bool(false)
				state.FilterChains[1].Filters[0].Config = kong.JSONRawMessage(`{"add":{"headers":["x-service:CHANGED"]}}`)
			},
			deleteFile: "testdata/sync/033-filter-chains/delete.yaml",
			deleteState: func(state *utils.KongRawState) {
				state.FilterChains = []*kong.FilterChain{
					state.FilterChains[0],
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runWhen(t, "kong", tc.version)

			state := utils.KongRawState{
				Services: []*kong.Service{&service},
				Routes:   []*kong.Route{&route},
			}

			require.NoError(t, sync(context.Background(), "testdata/sync/033-filter-chains/init.yaml"))

			testKongState(t, client, false, false, state, nil)

			require.NoError(t, sync(context.Background(), tc.createFile))
			tc.createState(&state)
			testKongState(t, client, false, false, state, nil)

			require.NoError(t, sync(context.Background(), tc.updateFile))
			tc.updateState(&state)
			testKongState(t, client, false, false, state, nil)

			require.NoError(t, sync(context.Background(), tc.deleteFile))
			tc.deleteState(&state)
			testKongState(t, client, false, false, state, nil)
		})
	}
}

func Test_Sync_FilterChainsUnsupported(t *testing.T) {
	t.Skip("Skipping test till wasm/filter-chains issue is not resolved at gateway level")
	runWhen(t, "kong", "<3.4.0")
	setup(t)
	require.NoError(t, sync(context.Background(), "testdata/sync/033-filter-chains/init.yaml"))
	require.Error(t, sync(context.Background(), "testdata/sync/033-filter-chains/create.yaml"))
}

func Test_Sync_DegraphqlRoutes(t *testing.T) {
	runWhen(t, "enterprise", ">=3.0.0")
	setup(t)

	client, err := getTestClient()
	require.NoError(t, err)

	ctx := context.Background()
	dumpConfig := deckDump.Config{CustomEntityTypes: []string{"degraphql_routes"}}

	t.Run("create degraphql route", func(t *testing.T) {
		require.NoError(t, sync(ctx, "testdata/sync/036-degraphql-routes/kong.yaml"))

		newState, err := fetchCurrentState(ctx, client, dumpConfig, t)
		require.NoError(t, err)

		degraphqlRoutes, err := newState.DegraphqlRoutes.GetAll()
		require.NoError(t, err)

		require.Len(t, degraphqlRoutes, 1)

		d := degraphqlRoutes[0]
		assert.Equal(t, "/foo", *d.URI)
		assert.Equal(t, "query{ foo { bar } }", *d.Query)

		expectedMethods := kong.StringSlice("GET")
		assert.Equal(t, expectedMethods, d.Methods)
	})

	t.Run("create degraphql route - complex query", func(t *testing.T) {
		require.NoError(t, sync(ctx, "testdata/sync/036-degraphql-routes/kong-complex-query.yaml"))

		newState, err := fetchCurrentState(ctx, client, dumpConfig, t)
		require.NoError(t, err)

		degraphqlRoutes, err := newState.DegraphqlRoutes.GetAll()
		require.NoError(t, err)

		require.Len(t, degraphqlRoutes, 1)

		d := degraphqlRoutes[0]

		assert.Equal(t, "/search/posts", *d.URI)
		expectedQuery := kong.String(complexQueryForDegraphqlRoute)
		assert.Equal(t, expectedQuery, d.Query)

		expectedMethods := kong.StringSlice("POST", "GET")
		assert.Equal(t, expectedMethods, d.Methods)
	})
}

func Test_Sync_DegraphqlRoutes_Konnect(t *testing.T) {
	t.Setenv("DECK_KONNECT_CONTROL_PLANE_NAME", "default")
	runWhen(t, "konnect", "")
	setup(t)

	client, err := getTestClient()
	require.NoError(t, err)

	ctx := context.Background()
	dumpConfig := deckDump.Config{CustomEntityTypes: []string{"degraphql_routes"}}

	t.Run("create degraphql route", func(t *testing.T) {
		require.NoError(t, sync(ctx, "testdata/sync/036-degraphql-routes/kong.yaml"))

		newState, err := fetchCurrentState(ctx, client, dumpConfig, t)
		require.NoError(t, err)

		degraphqlRoutes, err := newState.DegraphqlRoutes.GetAll()
		require.NoError(t, err)

		require.Len(t, degraphqlRoutes, 1)

		d := degraphqlRoutes[0]
		assert.Equal(t, "/foo", *d.URI)
		assert.Equal(t, "query{ foo { bar } }", *d.Query)

		expectedMethods := kong.StringSlice("GET")
		assert.Equal(t, expectedMethods, d.Methods)
	})

	t.Run("create degraphql route - complex query", func(t *testing.T) {
		require.NoError(t, sync(ctx, "testdata/sync/036-degraphql-routes/kong-complex-query.yaml"))

		newState, err := fetchCurrentState(ctx, client, dumpConfig, t)
		require.NoError(t, err)

		degraphqlRoutes, err := newState.DegraphqlRoutes.GetAll()
		require.NoError(t, err)

		require.Len(t, degraphqlRoutes, 1)

		d := degraphqlRoutes[0]

		assert.Equal(t, "/search/posts", *d.URI)
		expectedQuery := kong.String(complexQueryForDegraphqlRoute)
		assert.Equal(t, expectedQuery, d.Query)

		expectedMethods := kong.StringSlice("POST", "GET")
		assert.Equal(t, expectedMethods, d.Methods)
	})
}

func Test_Sync_CustomEntitiesFake(t *testing.T) {
	runWhenEnterpriseOrKonnect(t, ">=3.0.0")
	setup(t)
	ctx := context.Background()
	err := sync(ctx, "testdata/sync/036-degraphql-routes/kong-fake.yaml")
	require.Error(t, err)
	assert.ErrorContains(t, err, "unknown entity type: fake-entity")
}

// test scope:
// Kong enterprise: >=3.4.0 <3.8.0
func Test_Sync_ConsumerGroupPlugin_Policy_Overrides_34x(t *testing.T) {
	runWhen(t, "enterprise", ">=3.4.0 <3.8.0")
	setup(t)

	client, err := getTestClient()
	require.NoError(t, err)

	ctx := context.Background()

	expectedState := utils.KongRawState{
		Consumers: []*kong.Consumer{
			{
				Username: kong.String("foo"),
			},
		},
		ConsumerGroups: []*kong.ConsumerGroupObject{
			{
				ConsumerGroup: &kong.ConsumerGroup{
					Name: kong.String("test-group"),
				},
				Consumers: []*kong.Consumer{
					{
						Username: kong.String("foo"),
					},
				},
				Plugins: []*kong.ConsumerGroupPlugin{
					{
						Name: kong.String("rate-limiting-advanced"),
						Config: kong.Configuration{
							"limit":                  []any{float64(100)},
							"retry_after_jitter_max": float64(0),
							"window_size":            []any{float64(60)},
							"window_type":            string("fixed"),
						},
						ConsumerGroup: &kong.ConsumerGroup{
							ID: kong.String("776df478-37a9-4660-81b4-b4e951d76340"),
						},
					},
				},
			},
		},
		Plugins: []*kong.Plugin{
			{
				Name:    kong.String("rate-limiting-advanced"),
				Enabled: kong.Bool(true),
				Protocols: kong.StringSlice(
					"grpc",
					"grpcs",
					"http",
					"https",
				),
				Config: kong.Configuration{
					"consumer_groups":         []any{string("test-group")},
					"dictionary_name":         string("kong_rate_limiting_counters"),
					"disable_penalty":         bool(false),
					"enforce_consumer_groups": bool(true),
					"error_code":              float64(429),
					"error_message":           "API rate limit exceeded",
					"header_name":             nil,
					"hide_client_headers":     false,
					"identifier":              string("consumer"),
					"limit":                   []any{float64(10)},
					"namespace":               string("ZEz47TWgUrv01HenyQBQa8io06MWsp0L"),
					"path":                    nil,
					"redis": map[string]any{
						"cluster_addresses":   nil,
						"connect_timeout":     float64(2000),
						"database":            float64(0),
						"host":                string("127.0.0.1"),
						"keepalive_backlog":   nil,
						"keepalive_pool_size": float64(256),
						"password":            nil,
						"port":                float64(6379),
						"read_timeout":        float64(2000),
						"send_timeout":        float64(2000),
						"sentinel_addresses":  nil,
						"sentinel_master":     nil,
						"sentinel_password":   nil,
						"sentinel_role":       nil,
						"sentinel_username":   nil,
						"server_name":         nil,
						"ssl":                 bool(false),
						"ssl_verify":          bool(false),
						"timeout":             float64(2000),
						"username":            nil,
					},
					"retry_after_jitter_max": float64(0),
					"strategy":               string("redis"),
					"sync_rate":              float64(10),
					"window_size":            []any{float64(60)},
					"window_type":            string("fixed"),
				},
			},
		},
	}

	t.Run("consumer-group policy overrides with info block in yaml file", func(t *testing.T) {
		require.NoError(t, sync(ctx, "testdata/sync/037-consumer-group-policy-overrides/kong34x.yaml"))

		testKongState(t, client, false, true, expectedState, nil)
	})

	t.Run("consumer-group policy overrides --consumer-group-policy-overrides = true", func(t *testing.T) {
		require.NoError(t, sync(ctx, "testdata/sync/037-consumer-group-policy-overrides/kong34x-no-info.yaml",
			"--consumer-group-policy-overrides"))

		testKongState(t, client, false, true, expectedState, nil)
	})

	t.Run("consumer-group policy overrides --consumer-group-policy-overrides = false", func(t *testing.T) {
		err := sync(ctx, "testdata/sync/037-consumer-group-policy-overrides/kong34x-no-info.yaml")
		require.Error(t, err)
		assert.ErrorContains(t, err, errorConsumerGroupPolicies)
	})

	t.Run("checking for valid json output while using --consumer-group-policy-overrides = true", func(t *testing.T) {
		// overwrite default standard output
		r, w, err := os.Pipe()
		require.NoError(t, err)
		color.Output = w

		require.NoError(t, sync(ctx, "testdata/sync/037-consumer-group-policy-overrides/kong34x-no-info.yaml",
			"--consumer-group-policy-overrides", "--json-output"))

		// read command output
		w.Close()
		out, err := io.ReadAll(r)
		require.NoError(t, err)
		require.NotNil(t, out)

		var js interface{}
		assert.NoError(t, json.Unmarshal(out, &js), "JSON validation failed")
	})

	t.Run("checking for valid json output while using consumer-group-policy-overrides in info block", func(t *testing.T) {
		// overwrite default standard output
		r, w, err := os.Pipe()
		require.NoError(t, err)
		color.Output = w

		require.NoError(t, sync(ctx, "testdata/sync/037-consumer-group-policy-overrides/kong34x.yaml",
			"--json-output"))

		// read command output
		w.Close()
		out, err := io.ReadAll(r)
		require.NoError(t, err)
		require.NotNil(t, out)

		var js interface{}
		assert.NoError(t, json.Unmarshal(out, &js), "JSON validation failed")
	})
}

// test scope:
// Kong enterprise: >=3.8.0 <3.9.0
func Test_Sync_ConsumerGroupPlugin_Policy_Overrides_38x(t *testing.T) {
	runWhen(t, "enterprise", ">=3.8.0 <3.9.0")
	setup(t)

	client, err := getTestClient()
	require.NoError(t, err)

	ctx := context.Background()

	expectedState := utils.KongRawState{
		Consumers: []*kong.Consumer{
			{
				Username: kong.String("foo"),
			},
		},
		ConsumerGroups: []*kong.ConsumerGroupObject{
			{
				ConsumerGroup: &kong.ConsumerGroup{
					Name: kong.String("test-group"),
				},
				Consumers: []*kong.Consumer{
					{
						Username: kong.String("foo"),
					},
				},
				Plugins: []*kong.ConsumerGroupPlugin{
					{
						Name: kong.String("rate-limiting-advanced"),
						Config: kong.Configuration{
							"limit":                  []any{float64(100)},
							"retry_after_jitter_max": float64(0),
							"window_size":            []any{float64(60)},
							"window_type":            string("fixed"),
						},
						ConsumerGroup: &kong.ConsumerGroup{
							ID: kong.String("776df478-37a9-4660-81b4-b4e951d76340"),
						},
					},
				},
			},
		},
		Plugins: []*kong.Plugin{
			{
				Name:    kong.String("rate-limiting-advanced"),
				Enabled: kong.Bool(true),
				Protocols: kong.StringSlice(
					"grpc",
					"grpcs",
					"http",
					"https",
				),
				Config: kong.Configuration{
					"consumer_groups":         []any{string("test-group")},
					"dictionary_name":         string("kong_rate_limiting_counters"),
					"disable_penalty":         bool(false),
					"enforce_consumer_groups": bool(true),
					"error_code":              float64(429),
					"error_message":           "API rate limit exceeded",
					"header_name":             nil,
					"hide_client_headers":     false,
					"identifier":              string("consumer"),
					"limit":                   []any{float64(10)},
					"namespace":               string("ZEz47TWgUrv01HenyQBQa8io06MWsp0L"),
					"path":                    nil,
					"redis": map[string]any{
						"cluster_addresses":        nil,
						"cluster_max_redirections": float64(5),
						"cluster_nodes":            nil,
						"connect_timeout":          float64(2000),
						"connection_is_proxied":    bool(false),
						"database":                 float64(0),
						"host":                     string("127.0.0.1"),
						"keepalive_backlog":        nil,
						"keepalive_pool_size":      float64(256),
						"password":                 nil,
						"port":                     float64(6379),
						"read_timeout":             float64(2000),
						"send_timeout":             float64(2000),
						"sentinel_addresses":       nil,
						"sentinel_master":          nil,
						"sentinel_nodes":           nil,
						"sentinel_password":        nil,
						"sentinel_role":            nil,
						"sentinel_username":        nil,
						"server_name":              nil,
						"ssl":                      bool(false),
						"ssl_verify":               bool(false),
						"timeout":                  float64(2000),
						"username":                 nil,
					},
					"retry_after_jitter_max": float64(0),
					"strategy":               string("redis"),
					"sync_rate":              float64(10),
					"window_size":            []any{float64(60)},
					"window_type":            string("fixed"),
				},
			},
		},
	}

	t.Run("consumer-group policy overrides with info block in yaml file", func(t *testing.T) {
		require.NoError(t, sync(ctx, "testdata/sync/037-consumer-group-policy-overrides/kong38x.yaml"))

		testKongState(t, client, false, true, expectedState, nil)
	})

	t.Run("consumer-group policy overrides --consumer-group-policy-overrides = true", func(t *testing.T) {
		require.NoError(t, sync(ctx, "testdata/sync/037-consumer-group-policy-overrides/kong38x-no-info.yaml",
			"--consumer-group-policy-overrides"))

		testKongState(t, client, false, true, expectedState, nil)
	})

	t.Run("consumer-group policy overrides --consumer-group-policy-overrides = false", func(t *testing.T) {
		err := sync(ctx, "testdata/sync/037-consumer-group-policy-overrides/kong38x-no-info.yaml")
		require.Error(t, err)
		assert.ErrorContains(t, err, errorConsumerGroupPolicies)
	})

	t.Run("checking for valid json output while using --consumer-group-policy-overrides = true", func(t *testing.T) {
		// overwrite default standard output
		r, w, err := os.Pipe()
		require.NoError(t, err)
		color.Output = w

		require.NoError(t, sync(ctx, "testdata/sync/037-consumer-group-policy-overrides/kong38x-no-info.yaml",
			"--consumer-group-policy-overrides", "--json-output"))

		// read command output
		w.Close()
		out, err := io.ReadAll(r)
		require.NoError(t, err)
		require.NotNil(t, out)

		var js interface{}
		assert.NoError(t, json.Unmarshal(out, &js), "JSON validation failed")
	})

	t.Run("checking for valid json output while using consumer-group-policy-overrides in info block", func(t *testing.T) {
		// overwrite default standard output
		r, w, err := os.Pipe()
		require.NoError(t, err)
		color.Output = w

		require.NoError(t, sync(ctx, "testdata/sync/037-consumer-group-policy-overrides/kong38x.yaml",
			"--json-output"))

		// read command output
		w.Close()
		out, err := io.ReadAll(r)
		require.NoError(t, err)
		require.NotNil(t, out)

		var js interface{}
		assert.NoError(t, json.Unmarshal(out, &js), "JSON validation failed")
	})
}

// test scope:
// Kong enterprise: >=3.9.0
func Test_Sync_ConsumerGroupPlugin_Policy_Overrides_39x(t *testing.T) {
	runWhen(t, "enterprise", ">=3.9.0")
	setup(t)

	client, err := getTestClient()
	require.NoError(t, err)

	ctx := context.Background()

	expectedState := utils.KongRawState{
		Consumers: []*kong.Consumer{
			{
				Username: kong.String("foo"),
			},
		},
		ConsumerGroups: []*kong.ConsumerGroupObject{
			{
				ConsumerGroup: &kong.ConsumerGroup{
					Name: kong.String("test-group"),
				},
				Consumers: []*kong.Consumer{
					{
						Username: kong.String("foo"),
					},
				},
				Plugins: []*kong.ConsumerGroupPlugin{
					{
						Name: kong.String("rate-limiting-advanced"),
						Config: kong.Configuration{
							"limit":                  []any{float64(100)},
							"retry_after_jitter_max": float64(0),
							"window_size":            []any{float64(60)},
							"window_type":            string("fixed"),
						},
						ConsumerGroup: &kong.ConsumerGroup{
							ID: kong.String("776df478-37a9-4660-81b4-b4e951d76340"),
						},
					},
				},
			},
		},
		Plugins: []*kong.Plugin{
			{
				Name:    kong.String("rate-limiting-advanced"),
				Enabled: kong.Bool(true),
				Protocols: kong.StringSlice(
					"grpc",
					"grpcs",
					"http",
					"https",
				),
				Config: kong.Configuration{
					"compound_identifier":     nil,
					"consumer_groups":         []any{string("test-group")},
					"dictionary_name":         string("kong_rate_limiting_counters"),
					"disable_penalty":         bool(false),
					"enforce_consumer_groups": bool(true),
					"error_code":              float64(429),
					"error_message":           "API rate limit exceeded",
					"header_name":             nil,
					"hide_client_headers":     false,
					"identifier":              string("consumer"),
					"limit":                   []any{float64(10)},
					"lock_dictionary_name":    string("kong_locks"),
					"namespace":               string("ZEz47TWgUrv01HenyQBQa8io06MWsp0L"),
					"path":                    nil,
					"redis": map[string]any{
						"cluster_addresses":        nil,
						"cluster_max_redirections": float64(5),
						"cluster_nodes":            nil,
						"connect_timeout":          float64(2000),
						"connection_is_proxied":    bool(false),
						"database":                 float64(0),
						"host":                     string("127.0.0.1"),
						"keepalive_backlog":        nil,
						"keepalive_pool_size":      float64(256),
						"password":                 nil,
						"port":                     float64(6379),
						"read_timeout":             float64(2000),
						"redis_proxy_type":         nil,
						"send_timeout":             float64(2000),
						"sentinel_addresses":       nil,
						"sentinel_master":          nil,
						"sentinel_nodes":           nil,
						"sentinel_password":        nil,
						"sentinel_role":            nil,
						"sentinel_username":        nil,
						"server_name":              nil,
						"ssl":                      bool(false),
						"ssl_verify":               bool(false),
						"timeout":                  float64(2000),
						"username":                 nil,
					},
					"retry_after_jitter_max": float64(0),
					"strategy":               string("redis"),
					"sync_rate":              float64(10),
					"window_size":            []any{float64(60)},
					"window_type":            string("fixed"),
				},
			},
		},
	}

	t.Run("consumer-group policy overrides with info block in yaml file", func(t *testing.T) {
		require.NoError(t, sync(ctx, "testdata/sync/037-consumer-group-policy-overrides/kong39x.yaml"))

		testKongState(t, client, false, true, expectedState, nil)
	})

	t.Run("consumer-group policy overrides --consumer-group-policy-overrides = true", func(t *testing.T) {
		require.NoError(t, sync(ctx, "testdata/sync/037-consumer-group-policy-overrides/kong39x-no-info.yaml",
			"--consumer-group-policy-overrides"))

		testKongState(t, client, false, true, expectedState, nil)
	})

	t.Run("consumer-group policy overrides --consumer-group-policy-overrides = false", func(t *testing.T) {
		err := sync(ctx, "testdata/sync/037-consumer-group-policy-overrides/kong39x-no-info.yaml")
		require.Error(t, err)
		assert.ErrorContains(t, err, errorConsumerGroupPolicies)
	})

	t.Run("checking for valid json output while using --consumer-group-policy-overrides = true", func(t *testing.T) {
		// overwrite default standard output
		r, w, err := os.Pipe()
		require.NoError(t, err)
		color.Output = w

		require.NoError(t, sync(ctx, "testdata/sync/037-consumer-group-policy-overrides/kong39x-no-info.yaml",
			"--consumer-group-policy-overrides", "--json-output"))

		// read command output
		w.Close()
		out, err := io.ReadAll(r)
		require.NoError(t, err)
		require.NotNil(t, out)

		var js interface{}
		assert.NoError(t, json.Unmarshal(out, &js), "JSON validation failed")
	})

	t.Run("checking for valid json output while using consumer-group-policy-overrides in info block", func(t *testing.T) {
		// overwrite default standard output
		r, w, err := os.Pipe()
		require.NoError(t, err)
		color.Output = w

		require.NoError(t, sync(ctx, "testdata/sync/037-consumer-group-policy-overrides/kong39x.yaml",
			"--json-output"))

		// read command output
		w.Close()
		out, err := io.ReadAll(r)
		require.NoError(t, err)
		require.NotNil(t, out)

		var js interface{}
		assert.NoError(t, json.Unmarshal(out, &js), "JSON validation failed")
	})
}

func Test_Sync_SkipConsumersWithConsumerGroups(t *testing.T) {
	runWhen(t, "enterprise", ">=3.0.0")
	setup(t)

	ctx := context.Background()

	t.Run("--skip-consumers-with-consumer-groups flag set", func(t *testing.T) {
		// Ensure that sync goes through without any errors
		// We aren't checking for expected state as that is not the prime motive of the test
		// We just want to ensure that sync runs without errors in case consumers are synced separately.
		err := sync(ctx, "testdata/sync/038-skip-consumers-with-cgs/base.yaml", "--skip-consumers-with-consumer-groups")
		require.NoError(t, err)

		err = sync(ctx, "testdata/sync/038-skip-consumers-with-cgs/consumers.yaml")
		require.NoError(t, err)

		// second sync to ensure it goes through successfully too without any errors
		err = sync(ctx, "testdata/sync/038-skip-consumers-with-cgs/base.yaml", "--skip-consumers-with-consumer-groups")
		require.NoError(t, err)
	})

	t.Run("--skip-consumers-with-consumer-groups flag set with consumers.group present in file", func(t *testing.T) {
		err := sync(ctx, "testdata/sync/038-skip-consumers-with-cgs/consumers.yaml", "--skip-consumers-with-consumer-groups")
		require.Error(t, err)
		assert.ErrorContains(t, err, "can not use --skip-consumers-with-consumer-groups while adding consumers.groups")
	})
}

func Test_Sync_SkipConsumersWithConsumerGroups_Konnect(t *testing.T) {
	runWhen(t, "konnect", "")
	setup(t)

	ctx := context.Background()

	t.Run("--skip-consumers-with-consumer-groups flag set", func(t *testing.T) {
		err := sync(ctx, "testdata/sync/038-skip-consumers-with-cgs/base.yaml", "--skip-consumers-with-consumer-groups")
		require.Error(t, err)
		assert.ErrorContains(t, err, "the flag --skip-consumers-with-consumer-groups can not be used with Konnect")
	})
}

func Test_Sync_Partials_Plugins(t *testing.T) {
	runWhenEnterpriseOrKonnect(t, ">=3.10.0")

	client, err := getTestClient()
	require.NoError(t, err)

	ctx := context.Background()

	dumpConfig := deckDump.Config{}

	partialConfig := kong.Configuration{
		"cluster_max_redirections": float64(5),
		"cluster_nodes":            nil,
		"connect_timeout":          float64(2000),
		"connection_is_proxied":    bool(false),
		"database":                 float64(0),
		"host":                     string("127.0.0.1"),
		"keepalive_backlog":        nil,
		"keepalive_pool_size":      float64(256),
		"password":                 nil,
		"port":                     float64(6379),
		"read_timeout":             float64(3001),
		"send_timeout":             float64(2004),
		"sentinel_master":          nil,
		"sentinel_nodes":           nil,
		"sentinel_password":        nil,
		"sentinel_role":            nil,
		"sentinel_username":        nil,
		"server_name":              nil,
		"ssl":                      bool(false),
		"ssl_verify":               bool(false),
		"username":                 nil,
	}

	t.Run("create a partial and link to a plugin via name", func(t *testing.T) {
		require.NoError(t, sync(ctx, "testdata/sync/039-partials/kong.yaml"))
		t.Cleanup(func() {
			reset(t)
		})

		newState, err := fetchCurrentState(ctx, client, dumpConfig, t)
		require.NoError(t, err)

		// check for partial
		partials, err := newState.Partials.GetAll()
		require.NoError(t, err)
		require.NotNil(t, partials)

		require.Len(t, partials, 1)
		assert.Equal(t, "my-ee-partial", *partials[0].Name)
		assert.Equal(t, "redis-ee", *partials[0].Type)
		assert.IsType(t, kong.Configuration{}, partials[0].Config)
		assert.Equal(t, partialConfig, partials[0].Config)

		// check for plugin
		plugins, err := newState.Plugins.GetAll()
		require.NoError(t, err)
		require.NotNil(t, plugins)
		require.Len(t, plugins, 1)
		assert.Equal(t, "rate-limiting-advanced", *plugins[0].Name)
		assert.IsType(t, []*kong.PartialLink{}, plugins[0].Partials)
		require.Len(t, plugins[0].Partials, 1)
		assert.Equal(t, *partials[0].ID, *plugins[0].Partials[0].ID)
		assert.Equal(t, "config.redis", *plugins[0].Partials[0].Path)
	})

	t.Run("partial id is preserved if passed and linking can be done via id", func(t *testing.T) {
		require.NoError(t, sync(ctx, "testdata/sync/039-partials/kong-ids.yaml"))
		t.Cleanup(func() {
			reset(t)
		})

		newState, err := fetchCurrentState(ctx, client, dumpConfig, t)
		require.NoError(t, err)

		// check for partial
		partials, err := newState.Partials.GetAll()
		require.NoError(t, err)
		require.NotNil(t, partials)

		require.Len(t, partials, 1)
		assert.Equal(t, "13dc230d-d65e-439a-9f05-9fd71abfee4d", *partials[0].ID)
		assert.Equal(t, "my-ee-partial", *partials[0].Name)
		assert.Equal(t, "redis-ee", *partials[0].Type)
		assert.IsType(t, kong.Configuration{}, partials[0].Config)
		assert.Equal(t, partialConfig, partials[0].Config)

		// check for plugin
		plugins, err := newState.Plugins.GetAll()
		require.NoError(t, err)
		require.NotNil(t, plugins)
		require.Len(t, plugins, 1)
		assert.Equal(t, "rate-limiting-advanced", *plugins[0].Name)
		assert.IsType(t, []*kong.PartialLink{}, plugins[0].Partials)
		require.Len(t, plugins[0].Partials, 1)
		assert.Equal(t, "13dc230d-d65e-439a-9f05-9fd71abfee4d", *plugins[0].Partials[0].ID)
		assert.Equal(t, "config.redis", *plugins[0].Partials[0].Path)
	})

	t.Run("linking to a plugin fails in case of non-existent partial", func(t *testing.T) {
		err := sync(ctx, "testdata/sync/039-partials/kong-wrong.yaml")
		require.Error(t, err)
		assert.ErrorContains(t, err, "partial non-existent-partial for plugin rate-limiting-advanced: entity not found")
	})
}

func Test_Sync_Partials(t *testing.T) {
	runWhenEnterpriseOrKonnect(t, ">=3.10.0")

	ctx := context.Background()

	t.Run("create partials", func(t *testing.T) {
		err := sync(ctx, "testdata/sync/039-partials/kong-partials.yaml")
		t.Cleanup(func() {
			reset(t)
		})
		require.NoError(t, err)
	})

	t.Run("creating a partial errors out if no type is provided", func(t *testing.T) {
		err := sync(ctx, "testdata/sync/039-partials/kong-partials-no-type.yaml")
		require.Error(t, err)
		assert.ErrorContains(t, err, "type is required")
	})

	t.Run("creating a partial works even if no name or id is provided", func(t *testing.T) {
		err := sync(ctx, "testdata/sync/039-partials/kong-partials-no-name.yaml")
		t.Cleanup(func() {
			reset(t)
		})
		require.NoError(t, err)
	})

	t.Run("partial updates work without errors", func(t *testing.T) {
		err := sync(ctx, "testdata/sync/039-partials/kong.yaml")
		t.Cleanup(func() {
			reset(t)
		})
		require.NoError(t, err)

		err = sync(ctx, "testdata/sync/039-partials/kong-update.yaml")
		require.NoError(t, err)
	})
}

func Test_Sync_Consumers_Default_Lookup_Tag(t *testing.T) {
	runWhenEnterpriseOrKonnect(t, ">=2.8.0")

	ctx := context.Background()

	t.Run("no errors occur in case of subsequent syncs with distributed config and defaultLookupTags for consumer-group",
		func(t *testing.T) {
			// sync consumer-group file first
			err := sync(ctx, "testdata/sync/015-consumer-groups/kong-cg.yaml")
			t.Cleanup(func() {
				reset(t)
			})
			require.NoError(t, err)

			// sync consumers file
			err = sync(ctx, "testdata/sync/015-consumer-groups/kong-consumers.yaml")
			require.NoError(t, err)

			// sync again
			err = sync(ctx, "testdata/sync/015-consumer-groups/kong-consumers.yaml")
			require.NoError(t, err)
		})

	t.Run("no errors occur in case of with distributed config when consumer is not tagged but consumer-group is",
		func(t *testing.T) {
			// sync consumer-group file first
			err := sync(ctx, "testdata/sync/015-consumer-groups/kong-cg.yaml")
			t.Cleanup(func() {
				reset(t)
			})
			require.NoError(t, err)

			// sync consumers file
			err = sync(ctx, "testdata/sync/015-consumer-groups/kong-consumers-no-tag.yaml")
			require.NoError(t, err)

			// sync again
			err = sync(ctx, "testdata/sync/015-consumer-groups/kong-consumers-no-tag.yaml")
			require.NoError(t, err)
		})
}

// test scope:
//
//   - >=2.8.0
//   - konnect
//   - enterprise
func Test_Sync_Avoid_Overwrite_On_Select_Tag_Mismatch_With_ID(t *testing.T) {
	setup(t)

	ctx := context.Background()

	tests := []struct {
		name             string
		initialStateFile string
		targetStateFile  string
		errorExpected    string
	}{
		{
			name:             "certificate",
			initialStateFile: "testdata/sync/040-avoid-entity-rewrite-with-id-on-select-tag-mismatch/certificate-initial.yaml",
			targetStateFile:  "testdata/sync/040-avoid-entity-rewrite-with-id-on-select-tag-mismatch/certificate-final.yaml",
			errorExpected:    "error: certificate with ID 13c562a1-191c-4464-9b18-e5222b46035a already exists",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			reset(t)
			err := sync(ctx, tc.initialStateFile)
			require.NoError(t, err)

			err = sync(ctx, tc.targetStateFile, "--select-tag", "after")
			require.Error(t, err)
			assert.ErrorContains(t, err, tc.errorExpected)
		})
	}
}

// test scope:
//
// - >=2.8.0 <3.0.0
func Test_Sync_Plugins_Nested_Foreign_Keys(t *testing.T) {
	runWhen(t, "kong", ">=2.8.0 <3.0.0")
	setup(t)

	ctx := context.Background()

	tests := []struct {
		name      string
		stateFile string
	}{
		{
			name:      "plugins with consumer reference - via name",
			stateFile: "testdata/sync/041-plugins-nested-foreign-keys/1_1/kong-consumers-via-names.yaml",
		},
		{
			name:      "plugins with consumer reference - via id",
			stateFile: "testdata/sync/041-plugins-nested-foreign-keys/1_1/kong-consumers-via-ids.yaml",
		},
		{
			name:      "plugins with route reference - via name",
			stateFile: "testdata/sync/041-plugins-nested-foreign-keys/1_1/kong-routes-via-names.yaml",
		},
		{
			name:      "plugins with route reference - via id",
			stateFile: "testdata/sync/041-plugins-nested-foreign-keys/1_1/kong-routes-via-ids.yaml",
		},
		{
			name:      "plugins with service reference - via name",
			stateFile: "testdata/sync/041-plugins-nested-foreign-keys/1_1/kong-services-via-ids.yaml",
		},
		{
			name:      "plugins with service reference - via id",
			stateFile: "testdata/sync/041-plugins-nested-foreign-keys/1_1/kong-services-via-ids.yaml",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			reset(t)
			err := sync(ctx, tc.stateFile)
			require.NoError(t, err)

			// re-sync with no error
			err = sync(ctx, tc.stateFile)
			require.NoError(t, err)
		})
	}
}

// test scope:
//
// - >=3.0.0
// - konnect
func Test_Sync_Plugins_Nested_Foreign_Keys_3x(t *testing.T) {
	runWhenKongOrKonnect(t, ">=3.0.0")
	setup(t)

	ctx := context.Background()

	tests := []struct {
		name      string
		stateFile string
	}{
		{
			name:      "plugins with consumer reference - via name",
			stateFile: "testdata/sync/041-plugins-nested-foreign-keys/kong3x-consumers-via-names.yaml",
		},
		{
			name:      "plugins with consumer reference - via id",
			stateFile: "testdata/sync/041-plugins-nested-foreign-keys/kong3x-consumers-via-ids.yaml",
		},
		{
			name:      "plugins with route reference - via name",
			stateFile: "testdata/sync/041-plugins-nested-foreign-keys/kong3x-routes-via-names.yaml",
		},
		{
			name:      "plugins with route reference - via id",
			stateFile: "testdata/sync/041-plugins-nested-foreign-keys/kong3x-routes-via-ids.yaml",
		},
		{
			name:      "plugins with service reference - via name",
			stateFile: "testdata/sync/041-plugins-nested-foreign-keys/kong3x-services-via-ids.yaml",
		},
		{
			name:      "plugins with service reference - via id",
			stateFile: "testdata/sync/041-plugins-nested-foreign-keys/kong3x-services-via-ids.yaml",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			reset(t)
			err := sync(ctx, tc.stateFile)
			require.NoError(t, err)

			// re-sync with no error
			err = sync(ctx, tc.stateFile)
			require.NoError(t, err)
		})
	}
}

// test scope:
//
// - >=3.0.0+enterprise
// - konnect
func Test_Sync_Plugins_Nested_Foreign_Keys_EE_3x(t *testing.T) {
	// prior versions don't support a consumer_group foreign key with a value
	runWhenEnterpriseOrKonnect(t, ">=3.6.0")
	setup(t)

	ctx := context.Background()

	tests := []struct {
		name      string
		stateFile string
	}{
		{
			name:      "plugins with consumer-group reference - via name",
			stateFile: "testdata/sync/041-plugins-nested-foreign-keys/kong3x-consumer-groups-via-names.yaml",
		},
		{
			name:      "plugins with consumer-group reference - via id",
			stateFile: "testdata/sync/041-plugins-nested-foreign-keys/kong3x-consumer-groups-via-ids.yaml",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			reset(t)
			err := sync(ctx, tc.stateFile)
			require.NoError(t, err)

			// re-sync with no error
			err = sync(ctx, tc.stateFile)
			require.NoError(t, err)
		})
	}
}

// test scope:
//
// - >=2.8.0
func Test_Sync_Scoped_Plugins_Check_Conflicts(t *testing.T) {
	runWhenKongOrKonnect(t, ">=2.8.0")
	setup(t)

	ctx := context.Background()

	stateFiles := []string{
		"testdata/sync/042-scoped-plugins/plugins1.yaml",
		"testdata/sync/042-scoped-plugins/plugins2.yaml",
	}

	tests := []struct {
		name       string
		multiSync  bool
		emptyState bool
	}{
		{
			name:       "sync plugin files one by one, existing state: empty",
			multiSync:  false,
			emptyState: true,
		},
		{
			name:       "sync plugin files one by one, existing state: non-empty",
			multiSync:  false,
			emptyState: false,
		},
		{
			name:       "sync plugin files together, existing state: empty",
			multiSync:  true,
			emptyState: true,
		},
		{
			name:       "sync plugin files together, existing state: non-empty",
			multiSync:  true,
			emptyState: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			reset(t)
			if !tc.emptyState {
				// syncing first state file to create initial state
				err := sync(ctx, stateFiles[0])
				require.NoError(t, err)
			}

			if tc.multiSync {
				err := multiFileSync(ctx, stateFiles)
				require.NoError(t, err)

				// re-sync with no error
				err = multiFileSync(ctx, stateFiles)
				require.NoError(t, err)

				return
			}

			for _, s := range stateFiles {
				err := sync(ctx, s)
				require.NoError(t, err)

				// re-sync with no error
				err = sync(ctx, s)
				require.NoError(t, err)
			}
		})
	}
}

// test scope:
//
// - >=3.1.0
func Test_Sync_KeysAndKeySets(t *testing.T) {
	runWhen(t, "kong", ">=3.1.0")
	setup(t)

	client, err := getTestClient()
	require.NoError(t, err)
	ctx := context.Background()

	tests := []struct {
		name          string
		kongFile      string
		expectedState utils.KongRawState
	}{
		{
			name:     "creates keys and key_sets",
			kongFile: "testdata/sync/043-keys-and-key_sets/kong.yaml",
			expectedState: utils.KongRawState{
				Keys: []*kong.Key{
					{
						ID:   kong.String("f21a7073-1183-4b1c-bd87-4d5b8b18eeb4"),
						Name: kong.String("foo"),
						KID:  kong.String("vsR8NCNV_1_LB06LqudGa2r-T0y4Z6VQVYue9IQz6A4"),
						Set: &kong.KeySet{
							ID: kong.String("d46b0e15-ffbc-4b15-ad92-09ef67935453"),
						},
						JWK: kong.String("{\"kty\": \"RSA\", \"kid\": \"vsR8NCNV_1_LB06LqudGa2r-T0y4Z6VQVYue9IQz6A4\", \"n\": \"v2KAzzfruqctVHaE9WSCWIg1xAhMwxTIK-i56WNqPtpWBo9AqxcVea8NyVctEjUNq_mix5CklNy3ru7ARh7rBG_LU65fzs4fY_uYalul3QZSnr61Gj-cTUB3Gy4PhA63yXCbYRR3gDy6WR_wfis1MS61j0R_AjgXuVufmmC0F7R9qSWfR8ft0CbQgemEHY3ddKeW7T7fKv1jnRwYAkl5B_xtvxRFIYT-uR9NNftixNpUIW7q8qvOH7D9icXOg4_wIVxTRe5QiRYwEFoUbV1V9bFtu5FLal0vZnLaWwg5tA6enhzBpxJNdrS0v1RcPpyeNP-9r3cUDGmeftwz9v95UQ\", \"e\": \"AQAB\", \"alg\": \"A256GCM\"}"), //nolint:lll
					},
					{
						ID:   kong.String("d7cef208-23c3-46f8-94e8-fa1eddf43f0a"),
						Name: kong.String("baz"),
						KID:  kong.String("IiI4ffge7LZXPztrZVOt26zgRt0EPsWPaxAmwhbJhDQ"),
						Set: &kong.KeySet{
							ID: kong.String("d46b0e15-ffbc-4b15-ad92-09ef67935453"),
						},
						JWK: kong.String("{\n      \"kty\": \"RSA\",\n      \"kid\": \"IiI4ffge7LZXPztrZVOt26zgRt0EPsWPaxAmwhbJhDQ\",\n      \"use\": \"sig\",\n      \"alg\": \"RS256\",\n      \"e\": \"AQAB\",\n      \"n\": \"1Sn1X_y-RUzGna0hR00Wu64ZtY5N5BVzpRIby9wQ5EZVyWL9DRhU5PXqM3Y5gzgUVEQu548qQcMKOfs46PhOQudz-HPbwKWzcJCDUeNQsxdAEhW1uJR0EEV_SGJ-jTuKGqoEQc7bNrmhyXBMIeMkTeE_-ys75iiwvNjYphiOhsokC_vRTf_7TOPTe1UQasgxEVSLlTsen0vtK_FXcpbwdxZt02IysICcX5TcWX_XBuFP4cpwI9AS3M-imc01awc1t7FE5UWp62H5Ro2S5V9YwdxSjf4lX87AxYmawaWAjyO595XLuIXA3qt8-irzbCeglR1-cTB7a4I7_AclDmYrpw\"\n  }"), //nolint:lll
					},
					{
						ID:   kong.String("03ad4618-82bb-4375-b9d1-edeefced868d"),
						Name: kong.String("my-pem-key"),
						KID:  kong.String("my-pem-key"),
						Set: &kong.KeySet{
							ID: kong.String("d46b0e15-ffbc-4b15-ad92-09ef67935453"),
						},
						PEM: &kong.PEM{
							PublicKey:  kong.String("-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAqvxMU4LTcHBYmCuLMhMP\nDWlZdcNRXuJkw26MRjLBxXjnPAyDolmuFFMIqPDlSaJkkzu2tn7m9p8KB90wLiMC\nIbDjseruCO+7EaIRY4d6RdpE+XowCjJu7SbC2CqWBAzKkO7WWAunO3KOsQRk1NEK\nI51CoZ26LPYQvjIGIY2/pPxq0Ydl9dyURqVfmTywni1WeScgdEZXuy9WIcobqBST\n8vV5Q5HJsZNFLR7Fy61+HHfnQiWIYyi6h8QRT+Css9y5KbH7KuN6tnb94UZaOmHl\nYeoHcP/CqviZnQOf5804qcVpPKbsGU8jupTriiJZU3a8f59eHV0ybI4ORXYgDSWd\nFQIDAQAB\n-----END PUBLIC KEY-----"),                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               //nolint:lll
							PrivateKey: kong.String("-----BEGIN RSA PRIVATE KEY-----\nMIIEpAIBAAKCAQEAqvxMU4LTcHBYmCuLMhMPDWlZdcNRXuJkw26MRjLBxXjnPAyD\nolmuFFMIqPDlSaJkkzu2tn7m9p8KB90wLiMCIbDjseruCO+7EaIRY4d6RdpE+Xow\nCjJu7SbC2CqWBAzKkO7WWAunO3KOsQRk1NEKI51CoZ26LPYQvjIGIY2/pPxq0Ydl\n9dyURqVfmTywni1WeScgdEZXuy9WIcobqBST8vV5Q5HJsZNFLR7Fy61+HHfnQiWI\nYyi6h8QRT+Css9y5KbH7KuN6tnb94UZaOmHlYeoHcP/CqviZnQOf5804qcVpPKbs\nGU8jupTriiJZU3a8f59eHV0ybI4ORXYgDSWdFQIDAQABAoIBAEOOqAGfATe9y+Nj\n4P2J9jqQU15qK65XuQRWm2npCBKj8IkTULdGw7cYD6XgeFedqCtcPpbgkRUERYxR\n4oV4I5F4OJ7FegNh5QHUjRZMIw2Sbgo8Mtr0jkt5MycBvIAhJbAaDep/wDWGz8Y1\nPDmx1lW3/umoTjURjA/5594+CWiABYzuIi4WprWe4pIKqSKOMHnCYVAD243mwJ7y\nvsatO3LRKYfLw74ifCYhWNBHaZwfw+OO2P5Ku0AGhY4StOLCHobJ8/KkkmkTlYzv\nrcF4cVdvpBfdTEQed0oD7u3xfnp3GpNU3wZFsZJRSVXouhroaMC7en4uMc+5yguW\nqrPIoEkCgYEAxm1UllY9rRfGV6884hdBFKDjE825BC1VlqcRIUEB4CpJvUF/6+A3\ngx5c4nKDJAFQMrWpr4jOcq3iLiWnJ73e80b+JpWFODdt16g2KCOINs1j8vf2U6Og\nx+Vo8vHek/Uomz1n5W0oXrJ4VedHl9NYa8r/YrVXd4k4WcaA0TXmMhMCgYEA3Jit\nzrEmrQIrLK66RgXF2RafA5c3atRHWBb5ddnGk0bV90cfsTsaDMDvpy7ZYgojBNpw\n7U6AYzqnPro6cHEginV97BFb6oetMvOWvljUob+tpnYOofgwk2hw7PeChViX7iS9\nujgTygi8ZIc2G0r7xntH+v6WHKp4yNQiCAyfGTcCgYAYKgZMDJKUOrn3wapraiON\nzI36wmnOnWq33v6SCyWcU+oI9yoJ4pNAD3mGRiW8Q8CtfDv+2W0ywAQ0VHeHunKl\nM7cNodXIY8+nnJ+Dwdf7vIV4eEPyKZIR5dkjBNtzLz7TsOWvJdzts1Q+Od0ZGy7A\naccyER1mvDo1jJvxXlv7KwKBgQDDBK9TdUVt2eb1X5sJ4HyiiN8XO44ggX55IAZ1\n64skFJGARH5+HnPPJpo3wLEpfTCsT7lZ8faKwwWr7NNRKJHOFkS2eDo8QqoZy0NP\nEBUa0evgp6oUAuheyQxcUgwver0GKbEZeg30pHh4nxh0VHv1YnOmL3/h48tYMEHN\nv+q/TQKBgQCXQmN8cY2K7UfZJ6BYEdguQZS5XISFbLNkG8wXQX9vFiF8TuSWawDN\nTrRHVDGwoMGWxjZBLCsitA6zwrMLJZs4RuetKHFou7MiDQ69YGdfNRlRvD5QCJDc\nY0ICsYjI7VM89Qj/41WQyRHYHm7E9key3avMGdbYtxdc0Ku4LnD4zg==\n-----END RSA PRIVATE KEY-----"), //nolint:lll
						},
					},
				},
				KeySets: []*kong.KeySet{
					{
						Name: kong.String("bar"),
						ID:   kong.String("d46b0e15-ffbc-4b15-ad92-09ef67935453"),
					},
				},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			reset(t)
			err := sync(ctx, tc.kongFile)
			require.NoError(t, err)

			testKongState(t, client, false, false, tc.expectedState, nil)
		})
	}
}

// test scope:
//
// - konnect
// Separate test function is added as Konnect has different
// ordering and spacing for keys.
func Test_Sync_KeysAndKeySets_Konnect(t *testing.T) {
	runWhenKonnect(t)
	setup(t)

	client, err := getTestClient()
	require.NoError(t, err)
	ctx := context.Background()

	tests := []struct {
		name          string
		kongFile      string
		expectedState utils.KongRawState
	}{
		{
			name:     "creates keys and key_sets",
			kongFile: "testdata/sync/043-keys-and-key_sets/kong.yaml",
			expectedState: utils.KongRawState{
				Keys: []*kong.Key{
					{
						ID:   kong.String("f21a7073-1183-4b1c-bd87-4d5b8b18eeb4"),
						Name: kong.String("foo"),
						KID:  kong.String("vsR8NCNV_1_LB06LqudGa2r-T0y4Z6VQVYue9IQz6A4"),
						Set: &kong.KeySet{
							ID: kong.String("d46b0e15-ffbc-4b15-ad92-09ef67935453"),
						},
						JWK: kong.String("{\"kid\":\"vsR8NCNV_1_LB06LqudGa2r-T0y4Z6VQVYue9IQz6A4\",\"kty\":\"RSA\",\"alg\":\"A256GCM\",\"n\":\"v2KAzzfruqctVHaE9WSCWIg1xAhMwxTIK-i56WNqPtpWBo9AqxcVea8NyVctEjUNq_mix5CklNy3ru7ARh7rBG_LU65fzs4fY_uYalul3QZSnr61Gj-cTUB3Gy4PhA63yXCbYRR3gDy6WR_wfis1MS61j0R_AjgXuVufmmC0F7R9qSWfR8ft0CbQgemEHY3ddKeW7T7fKv1jnRwYAkl5B_xtvxRFIYT-uR9NNftixNpUIW7q8qvOH7D9icXOg4_wIVxTRe5QiRYwEFoUbV1V9bFtu5FLal0vZnLaWwg5tA6enhzBpxJNdrS0v1RcPpyeNP-9r3cUDGmeftwz9v95UQ\",\"e\":\"AQAB\"}"), //nolint:lll
					},
					{
						ID:   kong.String("d7cef208-23c3-46f8-94e8-fa1eddf43f0a"),
						Name: kong.String("baz"),
						KID:  kong.String("IiI4ffge7LZXPztrZVOt26zgRt0EPsWPaxAmwhbJhDQ"),
						Set: &kong.KeySet{
							ID: kong.String("d46b0e15-ffbc-4b15-ad92-09ef67935453"),
						},
						JWK: kong.String("{\"kid\":\"IiI4ffge7LZXPztrZVOt26zgRt0EPsWPaxAmwhbJhDQ\",\"kty\":\"RSA\",\"use\":\"sig\",\"alg\":\"RS256\",\"n\":\"1Sn1X_y-RUzGna0hR00Wu64ZtY5N5BVzpRIby9wQ5EZVyWL9DRhU5PXqM3Y5gzgUVEQu548qQcMKOfs46PhOQudz-HPbwKWzcJCDUeNQsxdAEhW1uJR0EEV_SGJ-jTuKGqoEQc7bNrmhyXBMIeMkTeE_-ys75iiwvNjYphiOhsokC_vRTf_7TOPTe1UQasgxEVSLlTsen0vtK_FXcpbwdxZt02IysICcX5TcWX_XBuFP4cpwI9AS3M-imc01awc1t7FE5UWp62H5Ro2S5V9YwdxSjf4lX87AxYmawaWAjyO595XLuIXA3qt8-irzbCeglR1-cTB7a4I7_AclDmYrpw\",\"e\":\"AQAB\"}"), //nolint:lll
					},
					{
						ID:   kong.String("03ad4618-82bb-4375-b9d1-edeefced868d"),
						Name: kong.String("my-pem-key"),
						KID:  kong.String("my-pem-key"),
						Set: &kong.KeySet{
							ID: kong.String("d46b0e15-ffbc-4b15-ad92-09ef67935453"),
						},
						PEM: &kong.PEM{
							PublicKey:  kong.String("-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAqvxMU4LTcHBYmCuLMhMP\nDWlZdcNRXuJkw26MRjLBxXjnPAyDolmuFFMIqPDlSaJkkzu2tn7m9p8KB90wLiMC\nIbDjseruCO+7EaIRY4d6RdpE+XowCjJu7SbC2CqWBAzKkO7WWAunO3KOsQRk1NEK\nI51CoZ26LPYQvjIGIY2/pPxq0Ydl9dyURqVfmTywni1WeScgdEZXuy9WIcobqBST\n8vV5Q5HJsZNFLR7Fy61+HHfnQiWIYyi6h8QRT+Css9y5KbH7KuN6tnb94UZaOmHl\nYeoHcP/CqviZnQOf5804qcVpPKbsGU8jupTriiJZU3a8f59eHV0ybI4ORXYgDSWd\nFQIDAQAB\n-----END PUBLIC KEY-----"),                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               //nolint:lll
							PrivateKey: kong.String("-----BEGIN RSA PRIVATE KEY-----\nMIIEpAIBAAKCAQEAqvxMU4LTcHBYmCuLMhMPDWlZdcNRXuJkw26MRjLBxXjnPAyD\nolmuFFMIqPDlSaJkkzu2tn7m9p8KB90wLiMCIbDjseruCO+7EaIRY4d6RdpE+Xow\nCjJu7SbC2CqWBAzKkO7WWAunO3KOsQRk1NEKI51CoZ26LPYQvjIGIY2/pPxq0Ydl\n9dyURqVfmTywni1WeScgdEZXuy9WIcobqBST8vV5Q5HJsZNFLR7Fy61+HHfnQiWI\nYyi6h8QRT+Css9y5KbH7KuN6tnb94UZaOmHlYeoHcP/CqviZnQOf5804qcVpPKbs\nGU8jupTriiJZU3a8f59eHV0ybI4ORXYgDSWdFQIDAQABAoIBAEOOqAGfATe9y+Nj\n4P2J9jqQU15qK65XuQRWm2npCBKj8IkTULdGw7cYD6XgeFedqCtcPpbgkRUERYxR\n4oV4I5F4OJ7FegNh5QHUjRZMIw2Sbgo8Mtr0jkt5MycBvIAhJbAaDep/wDWGz8Y1\nPDmx1lW3/umoTjURjA/5594+CWiABYzuIi4WprWe4pIKqSKOMHnCYVAD243mwJ7y\nvsatO3LRKYfLw74ifCYhWNBHaZwfw+OO2P5Ku0AGhY4StOLCHobJ8/KkkmkTlYzv\nrcF4cVdvpBfdTEQed0oD7u3xfnp3GpNU3wZFsZJRSVXouhroaMC7en4uMc+5yguW\nqrPIoEkCgYEAxm1UllY9rRfGV6884hdBFKDjE825BC1VlqcRIUEB4CpJvUF/6+A3\ngx5c4nKDJAFQMrWpr4jOcq3iLiWnJ73e80b+JpWFODdt16g2KCOINs1j8vf2U6Og\nx+Vo8vHek/Uomz1n5W0oXrJ4VedHl9NYa8r/YrVXd4k4WcaA0TXmMhMCgYEA3Jit\nzrEmrQIrLK66RgXF2RafA5c3atRHWBb5ddnGk0bV90cfsTsaDMDvpy7ZYgojBNpw\n7U6AYzqnPro6cHEginV97BFb6oetMvOWvljUob+tpnYOofgwk2hw7PeChViX7iS9\nujgTygi8ZIc2G0r7xntH+v6WHKp4yNQiCAyfGTcCgYAYKgZMDJKUOrn3wapraiON\nzI36wmnOnWq33v6SCyWcU+oI9yoJ4pNAD3mGRiW8Q8CtfDv+2W0ywAQ0VHeHunKl\nM7cNodXIY8+nnJ+Dwdf7vIV4eEPyKZIR5dkjBNtzLz7TsOWvJdzts1Q+Od0ZGy7A\naccyER1mvDo1jJvxXlv7KwKBgQDDBK9TdUVt2eb1X5sJ4HyiiN8XO44ggX55IAZ1\n64skFJGARH5+HnPPJpo3wLEpfTCsT7lZ8faKwwWr7NNRKJHOFkS2eDo8QqoZy0NP\nEBUa0evgp6oUAuheyQxcUgwver0GKbEZeg30pHh4nxh0VHv1YnOmL3/h48tYMEHN\nv+q/TQKBgQCXQmN8cY2K7UfZJ6BYEdguQZS5XISFbLNkG8wXQX9vFiF8TuSWawDN\nTrRHVDGwoMGWxjZBLCsitA6zwrMLJZs4RuetKHFou7MiDQ69YGdfNRlRvD5QCJDc\nY0ICsYjI7VM89Qj/41WQyRHYHm7E9key3avMGdbYtxdc0Ku4LnD4zg==\n-----END RSA PRIVATE KEY-----"), //nolint:lll
						},
					},
				},
				KeySets: []*kong.KeySet{
					{
						Name: kong.String("bar"),
						ID:   kong.String("d46b0e15-ffbc-4b15-ad92-09ef67935453"),
					},
				},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			reset(t)
			err := sync(ctx, tc.kongFile)
			require.NoError(t, err)

			testKongState(t, client, true, false, tc.expectedState, nil)
		})
	}
}

// test scope:
//
// - >=3.10.0+enterprise
func Test_Sync_Partials_Tagging(t *testing.T) {
	runWhen(t, "enterprise", ">=3.10.0")
	setup(t)

	client, err := getTestClient()
	require.NoError(t, err)
	ctx := context.Background()

	expectedStatePostSync := utils.KongRawState{
		Partials: []*kong.Partial{
			{
				ID:   kong.String("13dc230d-d65e-439a-9f05-9fd71abfee4d"),
				Name: kong.String("redis-ee-common"),
				Type: kong.String("redis-ee"),
				Config: kong.Configuration{
					"cluster_max_redirections": float64(5),
					"cluster_nodes":            nil,
					"connect_timeout":          float64(2000),
					"connection_is_proxied":    bool(false),
					"database":                 float64(0),
					"host":                     string("127.0.0.1"),
					"keepalive_backlog":        nil,
					"keepalive_pool_size":      float64(256),
					"password":                 nil,
					"port":                     float64(6379),
					"read_timeout":             float64(3001),
					"send_timeout":             float64(2004),
					"sentinel_master":          nil,
					"sentinel_nodes":           nil,
					"sentinel_password":        nil,
					"sentinel_role":            nil,
					"sentinel_username":        nil,
					"server_name":              nil,
					"ssl":                      bool(false),
					"ssl_verify":               bool(false),
					"username":                 nil,
				},
				Tags: kong.StringSlice("redis-partials"),
			},
			{
				ID:   kong.String("b426adc7-7f11-4cda-a862-112ddabae9ef"),
				Name: kong.String("redis-ee-sentinel"),
				Type: kong.String("redis-ee"),
				Config: kong.Configuration{
					"cluster_max_redirections": float64(5),
					"cluster_nodes":            nil,
					"connect_timeout":          float64(2000),
					"connection_is_proxied":    bool(false),
					"database":                 float64(0),
					"host":                     string("127.0.0.1"),
					"keepalive_backlog":        nil,
					"keepalive_pool_size":      float64(256),
					"password":                 nil,
					"port":                     float64(6379),
					"read_timeout":             float64(2000),
					"send_timeout":             float64(2000),
					"sentinel_master":          string("mymaster"),
					"sentinel_nodes": []any{
						map[string]any{"host": string("redis-node-0"), "port": float64(26379)},
						map[string]any{"host": string("redis-node-1"), "port": float64(26379)},
						map[string]any{"host": string("redis-node-2"), "port": float64(26379)},
					},
					"sentinel_password": nil,
					"sentinel_role":     string("master"),
					"sentinel_username": nil,
					"server_name":       nil,
					"ssl":               bool(false),
					"ssl_verify":        bool(false),
					"username":          nil,
				},
				Tags: kong.StringSlice("redis-partials"),
			},
		},
		Services: []*kong.Service{
			{
				ConnectTimeout: kong.Int(60000),
				Enabled:        kong.Bool(true),
				Host:           kong.String("httpbin.konghq.com"),
				ID:             kong.String("ccb2e714-8398-4167-bf3f-049e1242483b"),
				Name:           kong.String("httpbin-1"),
				Path:           kong.String("/anything"),
				Port:           kong.Int(443),
				Protocol:       kong.String("https"),
				ReadTimeout:    kong.Int(60000),
				Retries:        kong.Int(5),
				WriteTimeout:   kong.Int(60000),
				Tags:           kong.StringSlice("api:partials-test-1"),
			},
		},
		Plugins: []*kong.Plugin{
			{
				ID:   kong.String("82c27e99-b1de-4772-aa60-4caa86c0480d"),
				Name: kong.String("rate-limiting-advanced"),
				Config: kong.Configuration{
					"compound_identifier":     nil,
					"consumer_groups":         nil,
					"dictionary_name":         string("kong_rate_limiting_counters"),
					"disable_penalty":         bool(false),
					"enforce_consumer_groups": bool(false),
					"error_code":              float64(429),
					"error_message":           string("API rate limit exceeded"),
					"header_name":             nil,
					"hide_client_headers":     bool(false),
					"identifier":              string("consumer"),
					"limit":                   []any{float64(10)},
					"lock_dictionary_name":    string("kong_locks"),
					"namespace":               string("test-ns"),
					"path":                    nil,
					"retry_after_jitter_max":  float64(0),
					"strategy":                string("local"),
					"sync_rate":               float64(-1),
					"window_size":             []any{float64(60)},
					"window_type":             string("fixed"),
				},
				Enabled:   kong.Bool(true),
				Protocols: kong.StringSlice("grpc", "grpcs", "http", "https"),
				Partials: []*kong.PartialLink{
					{
						Partial: &kong.Partial{
							ID:   kong.String("13dc230d-d65e-439a-9f05-9fd71abfee4d"),
							Name: kong.String("redis-ee-common"),
						},
						Path: kong.String("config.redis"),
					},
				},
			},
			{
				ID:   kong.String("88e5442f-5ff6-49ab-b4d7-ce41735cc2e0"),
				Name: kong.String("rate-limiting-advanced"),
				Service: &kong.Service{
					ID: kong.String("5167329f-b331-48d0-801a-0a045a7e8bce"),
				},
				Config: kong.Configuration{
					"compound_identifier":     nil,
					"consumer_groups":         nil,
					"dictionary_name":         string("kong_rate_limiting_counters"),
					"disable_penalty":         bool(false),
					"enforce_consumer_groups": bool(false),
					"error_code":              float64(429),
					"error_message":           string("API rate limit exceeded"),
					"header_name":             nil,
					"hide_client_headers":     bool(false),
					"identifier":              string("ip"),
					"limit":                   []any{float64(10000)},
					"lock_dictionary_name":    string("kong_locks"),
					"namespace":               string("testns"),
					"path":                    nil,
					"retry_after_jitter_max":  float64(0),
					"strategy":                string("redis"),
					"sync_rate":               float64(2),
					"window_size":             []any{float64(30)},
					"window_type":             string("sliding"),
				},
				Enabled:   kong.Bool(true),
				Protocols: kong.StringSlice("grpc", "grpcs", "http", "https"),
				Tags:      kong.StringSlice("api:partials-test-1"),
				Partials: []*kong.PartialLink{
					{
						Partial: &kong.Partial{
							ID:   kong.String("b426adc7-7f11-4cda-a862-112ddabae9ef"),
							Name: kong.String("redis-ee-sentinel"),
						},
						Path: kong.String("config.redis"),
					},
				},
			},
		},
	}

	ignoreFields := []cmp.Option{
		cmpopts.IgnoreFields(kong.Service{}, "ID"),
	}

	tests := []struct {
		name          string
		kongFile      string
		errorExpected bool
		errorString   string
	}{
		{
			name:          "sync partials with default lookup tags - via names",
			kongFile:      "testdata/sync/044-partials-tagging/partial-lookup-tags-names.yaml",
			errorExpected: false,
		},
		{
			name:          "sync partials with default lookup tags - via ids",
			kongFile:      "testdata/sync/044-partials-tagging/partial-lookup-tags-ids.yaml",
			errorExpected: false,
		},
		{
			name:          "syncing partials with default lookup tags errors out with wrong tags",
			kongFile:      "testdata/sync/044-partials-tagging/partial-lookup-tags-wrong.yaml",
			errorExpected: true,
			errorString:   "partial redis-ee-common for plugin rate-limiting-advanced: entity not found",
		},
		{
			name:          "syncing partials with default lookup tags errors out with wrong names",
			kongFile:      "testdata/sync/044-partials-tagging/partial-lookup-tags-wrong-names.yaml",
			errorExpected: true,
			errorString:   "partial fake-name for plugin rate-limiting-advanced: entity not found",
		},
		{
			name:          "syncing partials with default lookup tags errors out with wrong ids",
			kongFile:      "testdata/sync/044-partials-tagging/partial-lookup-tags-wrong-ids.yaml",
			errorExpected: true,
			errorString:   "partial fake-id-1234 for plugin rate-limiting-advanced: entity not found",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			reset(t)

			// syncing partials
			err := sync(ctx, "testdata/sync/044-partials-tagging/partials.yaml")
			require.NoError(t, err)

			// syncing the kong file with partial lookup tags
			err = sync(ctx, tc.kongFile)

			if tc.errorExpected {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.errorString)
				return
			}

			require.NoError(t, err)
			testKongState(t, client, false, false, expectedStatePostSync, ignoreFields)
		})
	}
}

// test scope:
//
// - konnect
func Test_Sync_Partials_Tagging_Konnect(t *testing.T) {
	setDefaultKonnectControlPlane(t)
	runWhenKonnect(t)
	setup(t)

	client, err := getTestClient()
	require.NoError(t, err)
	ctx := context.Background()

	expectedStatePostSync := utils.KongRawState{
		Partials: []*kong.Partial{
			{
				ID:   kong.String("13dc230d-d65e-439a-9f05-9fd71abfee4d"),
				Name: kong.String("redis-ee-common"),
				Type: kong.String("redis-ee"),
				Config: kong.Configuration{
					"cluster_max_redirections": float64(5),
					"cluster_nodes":            nil,
					"connect_timeout":          float64(2000),
					"connection_is_proxied":    bool(false),
					"database":                 float64(0),
					"host":                     string("127.0.0.1"),
					"keepalive_backlog":        nil,
					"keepalive_pool_size":      float64(256),
					"password":                 nil,
					"port":                     float64(6379),
					"read_timeout":             float64(3001),
					"send_timeout":             float64(2004),
					"sentinel_master":          nil,
					"sentinel_nodes":           nil,
					"sentinel_password":        nil,
					"sentinel_role":            nil,
					"sentinel_username":        nil,
					"server_name":              nil,
					"ssl":                      bool(false),
					"ssl_verify":               bool(false),
					"username":                 nil,
				},
				Tags: kong.StringSlice("redis-partials"),
			},
			{
				ID:   kong.String("b426adc7-7f11-4cda-a862-112ddabae9ef"),
				Name: kong.String("redis-ee-sentinel"),
				Type: kong.String("redis-ee"),
				Config: kong.Configuration{
					"cluster_max_redirections": float64(5),
					"cluster_nodes":            nil,
					"connect_timeout":          float64(2000),
					"connection_is_proxied":    bool(false),
					"database":                 float64(0),
					"host":                     string("127.0.0.1"),
					"keepalive_backlog":        nil,
					"keepalive_pool_size":      float64(256),
					"password":                 nil,
					"port":                     float64(6379),
					"read_timeout":             float64(2000),
					"send_timeout":             float64(2000),
					"sentinel_master":          string("mymaster"),
					"sentinel_nodes": []any{
						map[string]any{"host": string("redis-node-0"), "port": float64(26379)},
						map[string]any{"host": string("redis-node-1"), "port": float64(26379)},
						map[string]any{"host": string("redis-node-2"), "port": float64(26379)},
					},
					"sentinel_password": nil,
					"sentinel_role":     string("master"),
					"sentinel_username": nil,
					"server_name":       nil,
					"ssl":               bool(false),
					"ssl_verify":        bool(false),
					"username":          nil,
				},
				Tags: kong.StringSlice("redis-partials"),
			},
		},
		Services: []*kong.Service{
			{
				ConnectTimeout: kong.Int(60000),
				Enabled:        kong.Bool(true),
				Host:           kong.String("httpbin.konghq.com"),
				ID:             kong.String("ccb2e714-8398-4167-bf3f-049e1242483b"),
				Name:           kong.String("httpbin-1"),
				Path:           kong.String("/anything"),
				Port:           kong.Int(443),
				Protocol:       kong.String("https"),
				ReadTimeout:    kong.Int(60000),
				Retries:        kong.Int(5),
				WriteTimeout:   kong.Int(60000),
				Tags:           kong.StringSlice("api:partials-test-1"),
			},
		},
		Plugins: []*kong.Plugin{
			{
				ID:   kong.String("82c27e99-b1de-4772-aa60-4caa86c0480d"),
				Name: kong.String("rate-limiting-advanced"),
				Config: kong.Configuration{
					"compound_identifier":     nil,
					"consumer_groups":         nil,
					"dictionary_name":         string("kong_rate_limiting_counters"),
					"disable_penalty":         bool(false),
					"enforce_consumer_groups": bool(false),
					"error_code":              float64(429),
					"error_message":           string("API rate limit exceeded"),
					"header_name":             nil,
					"hide_client_headers":     bool(false),
					"identifier":              string("consumer"),
					"limit":                   []any{float64(10)},
					"lock_dictionary_name":    string("kong_locks"),
					"namespace":               string("test-ns"),
					"path":                    nil,
					"retry_after_jitter_max":  float64(0),
					"strategy":                string("local"),
					"sync_rate":               float64(-1),
					"window_size":             []any{float64(60)},
					"window_type":             string("fixed"),
				},
				Enabled:   kong.Bool(true),
				Protocols: kong.StringSlice("grpc", "grpcs", "http", "https"),
				Partials: []*kong.PartialLink{
					{
						Partial: &kong.Partial{
							ID: kong.String("13dc230d-d65e-439a-9f05-9fd71abfee4d"),
						},
						Path: kong.String("config.redis"),
					},
				},
			},
			{
				ID:   kong.String("88e5442f-5ff6-49ab-b4d7-ce41735cc2e0"),
				Name: kong.String("rate-limiting-advanced"),
				Service: &kong.Service{
					ID: kong.String("5167329f-b331-48d0-801a-0a045a7e8bce"),
				},
				Config: kong.Configuration{
					"compound_identifier":     nil,
					"consumer_groups":         nil,
					"dictionary_name":         string("kong_rate_limiting_counters"),
					"disable_penalty":         bool(false),
					"enforce_consumer_groups": bool(false),
					"error_code":              float64(429),
					"error_message":           string("API rate limit exceeded"),
					"header_name":             nil,
					"hide_client_headers":     bool(false),
					"identifier":              string("ip"),
					"limit":                   []any{float64(10000)},
					"lock_dictionary_name":    string("kong_locks"),
					"namespace":               string("testns"),
					"path":                    nil,
					"retry_after_jitter_max":  float64(0),
					"strategy":                string("redis"),
					"sync_rate":               float64(2),
					"window_size":             []any{float64(30)},
					"window_type":             string("sliding"),
				},
				Enabled:   kong.Bool(true),
				Protocols: kong.StringSlice("grpc", "grpcs", "http", "https"),
				Tags:      kong.StringSlice("api:partials-test-1"),
				Partials: []*kong.PartialLink{
					{
						Partial: &kong.Partial{
							ID: kong.String("b426adc7-7f11-4cda-a862-112ddabae9ef"),
						},
						Path: kong.String("config.redis"),
					},
				},
			},
		},
	}

	ignoreFields := []cmp.Option{
		cmpopts.IgnoreFields(kong.Service{}, "ID"),
	}

	tests := []struct {
		name          string
		kongFile      string
		errorExpected bool
		errorString   string
	}{
		{
			name:          "sync partials with default lookup tags - via names",
			kongFile:      "testdata/sync/044-partials-tagging/partial-lookup-tags-names.yaml",
			errorExpected: false,
		},
		{
			name:          "sync partials with default lookup tags - via ids",
			kongFile:      "testdata/sync/044-partials-tagging/partial-lookup-tags-ids.yaml",
			errorExpected: false,
		},
		{
			name:          "syncing partials with default lookup tags errors out with wrong tags",
			kongFile:      "testdata/sync/044-partials-tagging/partial-lookup-tags-wrong.yaml",
			errorExpected: true,
			errorString:   "partial redis-ee-common for plugin rate-limiting-advanced: entity not found",
		},
		{
			name:          "syncing partials with default lookup tags errors out with wrong names",
			kongFile:      "testdata/sync/044-partials-tagging/partial-lookup-tags-wrong-names.yaml",
			errorExpected: true,
			errorString:   "partial fake-name for plugin rate-limiting-advanced: entity not found",
		},
		{
			name:          "syncing partials with default lookup tags errors out with wrong ids",
			kongFile:      "testdata/sync/044-partials-tagging/partial-lookup-tags-wrong-ids.yaml",
			errorExpected: true,
			errorString:   "partial fake-id-1234 for plugin rate-limiting-advanced: entity not found",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			reset(t)

			// syncing partials
			err := sync(ctx, "testdata/sync/044-partials-tagging/partials.yaml")
			require.NoError(t, err)

			// syncing the kong file with partial lookup tags
			err = sync(ctx, tc.kongFile)

			if tc.errorExpected {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.errorString)
				return
			}

			require.NoError(t, err)
			testKongState(t, client, true, false, expectedStatePostSync, ignoreFields)
		})
	}
}
