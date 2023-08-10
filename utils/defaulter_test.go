package utils

import (
	"context"
	"reflect"
	"testing"

	"github.com/kong/go-kong/kong"
	"github.com/stretchr/testify/assert"
)

type kongDefaultForTesting struct {
	Service  *kong.Service
	Route    *kong.Route
	Upstream *kong.Upstream
	Target   *kong.Target
}

var kongDefaults = kongDefaultForTesting{
	Service:  &serviceDefaults,
	Route:    &routeDefaults,
	Upstream: &upstreamDefaults,
	Target:   &targetDefaults,
}

var defaulterTestOpts = DefaulterOpts{
	KongDefaults:           kongDefaults,
	DisableDynamicDefaults: false,
}

func TestDefaulter(t *testing.T) {
	assert := assert.New(t)

	var d Defaulter

	assert.NotNil(d.Register(nil))
	assert.NotNil(d.Set(nil))

	assert.Panics(func() {
		d.MustSet(d)
	})

	type Foo struct {
		A string
		B []string
	}
	defaultFoo := &Foo{
		A: "defaultA",
		B: []string{"default1"},
	}
	assert.Nil(d.Register(defaultFoo))

	// sets a default
	var arg Foo
	assert.Nil(d.Set(&arg))
	assert.Equal("defaultA", arg.A)
	assert.Equal([]string{"default1"}, arg.B)

	// doesn't set a default
	arg1 := Foo{
		A: "non-default-value",
	}
	assert.Nil(d.Set(&arg1))
	assert.Equal("non-default-value", arg1.A)

	// errors on an unregistered type
	type Bar struct {
		A string
	}
	assert.NotNil(d.Set(&Bar{}))
	assert.Panics(func() {
		d.MustSet(&Bar{})
	})
}

func TestServiceSetTest(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()
	d, err := GetDefaulter(ctx, defaulterTestOpts)
	assert.NotNil(d)
	assert.Nil(err)

	testCases := []struct {
		desc string
		arg  *kong.Service
		want *kong.Service
	}{
		{
			desc: "empty service",
			arg:  &kong.Service{},
			want: &serviceDefaults,
		},
		{
			desc: "retries can be set to 0",
			arg: &kong.Service{
				Retries: kong.Int(0),
			},
			want: &kong.Service{
				Retries:        kong.Int(0),
				Protocol:       kong.String("http"),
				ConnectTimeout: kong.Int(60000),
				WriteTimeout:   kong.Int(60000),
				ReadTimeout:    kong.Int(60000),
			},
		},
		{
			desc: "timeout value value is not overridden",
			arg: &kong.Service{
				WriteTimeout: kong.Int(42),
			},
			want: &kong.Service{
				Protocol:       kong.String("http"),
				ConnectTimeout: kong.Int(60000),
				WriteTimeout:   kong.Int(42),
				ReadTimeout:    kong.Int(60000),
			},
		},
		{
			desc: "path value is not overridden",
			arg: &kong.Service{
				Path: kong.String("/foo"),
			},
			want: &kong.Service{
				Protocol:       kong.String("http"),
				Path:           kong.String("/foo"),
				ConnectTimeout: kong.Int(60000),
				WriteTimeout:   kong.Int(60000),
				ReadTimeout:    kong.Int(60000),
			},
		},
		{
			desc: "Name is not reset",
			arg: &kong.Service{
				Name: kong.String("foo"),
				Host: kong.String("example.com"),
				Path: kong.String("/bar"),
			},
			want: &kong.Service{
				Name:           kong.String("foo"),
				Host:           kong.String("example.com"),
				Protocol:       kong.String("http"),
				Path:           kong.String("/bar"),
				ConnectTimeout: kong.Int(60000),
				WriteTimeout:   kong.Int(60000),
				ReadTimeout:    kong.Int(60000),
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			err := d.Set(tC.arg)
			assert.Nil(err)
			assert.Equal(tC.want, tC.arg)
		})
	}
}

func TestRouteSetTest(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()
	d, err := GetDefaulter(ctx, defaulterTestOpts)
	assert.NotNil(d)
	assert.Nil(err)

	testCases := []struct {
		desc string
		arg  *kong.Route
		want *kong.Route
	}{
		{
			desc: "empty route",
			arg:  &kong.Route{},
			want: &routeDefaults,
		},
		{
			desc: "preserve host is not overridden",
			arg: &kong.Route{
				PreserveHost: kong.Bool(true),
			},
			want: &kong.Route{
				PreserveHost:  kong.Bool(true),
				RegexPriority: kong.Int(0),
				StripPath:     kong.Bool(true),
				Protocols:     kong.StringSlice("http", "https"),
			},
		},
		{
			desc: "Protocols is not reset",
			arg: &kong.Route{
				Protocols: kong.StringSlice("http", "tls"),
			},
			want: &kong.Route{
				PreserveHost:  kong.Bool(false),
				RegexPriority: kong.Int(0),
				StripPath:     kong.Bool(true),
				Protocols:     kong.StringSlice("http", "tls"),
			},
		},
		{
			desc: "non-default feilds is not reset",
			arg: &kong.Route{
				Name:      kong.String("foo"),
				Hosts:     kong.StringSlice("1.example.com", "2.example.com"),
				Methods:   kong.StringSlice("GET", "POST"),
				StripPath: kong.Bool(true),
			},
			want: &kong.Route{
				Name:          kong.String("foo"),
				Hosts:         kong.StringSlice("1.example.com", "2.example.com"),
				Methods:       kong.StringSlice("GET", "POST"),
				PreserveHost:  kong.Bool(false),
				RegexPriority: kong.Int(0),
				StripPath:     kong.Bool(true),
				Protocols:     kong.StringSlice("http", "https"),
			},
		},
		{
			desc: "strip-path can be set to false",
			arg: &kong.Route{
				StripPath: kong.Bool(false),
			},
			want: &kong.Route{
				PreserveHost:  kong.Bool(false),
				RegexPriority: kong.Int(0),
				StripPath:     kong.Bool(false),
				Protocols:     kong.StringSlice("http", "https"),
			},
		},
		{
			desc: "strip-path can be set to true",
			arg: &kong.Route{
				StripPath: kong.Bool(true),
			},
			want: &kong.Route{
				PreserveHost:  kong.Bool(false),
				RegexPriority: kong.Int(0),
				StripPath:     kong.Bool(true),
				Protocols:     kong.StringSlice("http", "https"),
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			err := d.Set(tC.arg)
			assert.Nil(err)
			assert.Equal(tC.want, tC.arg)
		})
	}
}

func TestUpstreamSetTest(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()
	d, err := GetDefaulter(ctx, defaulterTestOpts)
	assert.NotNil(d)
	assert.Nil(err)

	testCases := []struct {
		desc string
		arg  *kong.Upstream
		want *kong.Upstream
	}{
		{
			desc: "empty upstream",
			arg:  &kong.Upstream{},
			want: &upstreamDefaults,
		},
		{
			desc: "Healthchecks.Active.Healthy.HTTPStatuses is not overridden",
			arg: &kong.Upstream{
				Healthchecks: &kong.Healthcheck{
					Active: &kong.ActiveHealthcheck{
						Healthy: &kong.Healthy{
							HTTPStatuses: []int{200},
						},
					},
				},
			},
			want: &kong.Upstream{
				Slots: kong.Int(10000),
				Healthchecks: &kong.Healthcheck{
					Active: &kong.ActiveHealthcheck{
						Concurrency: kong.Int(10),
						Healthy: &kong.Healthy{
							HTTPStatuses: []int{200},
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
							HTTPStatuses: []int{429, 404, 500, 501, 502, 503, 504, 505},
							Interval:     kong.Int(0),
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
			},
		},
		{
			desc: "Healthchecks.Active.Healthy.Timeout is not overridden",
			arg: &kong.Upstream{
				Name: kong.String("foo"),
				Healthchecks: &kong.Healthcheck{
					Active: &kong.ActiveHealthcheck{
						Healthy: &kong.Healthy{
							Interval: kong.Int(1),
						},
					},
				},
			},
			want: &kong.Upstream{
				Name:  kong.String("foo"),
				Slots: kong.Int(10000),
				Healthchecks: &kong.Healthcheck{
					Active: &kong.ActiveHealthcheck{
						Concurrency: kong.Int(10),
						Healthy: &kong.Healthy{
							HTTPStatuses: []int{200, 302},
							Interval:     kong.Int(1),
							Successes:    kong.Int(0),
						},
						HTTPPath: kong.String("/"),
						Type:     kong.String("http"),
						Timeout:  kong.Int(1),
						Unhealthy: &kong.Unhealthy{
							HTTPFailures: kong.Int(0),
							TCPFailures:  kong.Int(0),
							Timeouts:     kong.Int(0),
							HTTPStatuses: []int{429, 404, 500, 501, 502, 503, 504, 505},
							Interval:     kong.Int(0),
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
			},
		},
		{
			desc: "Healthchecks.Active.HTTPSVerifyCertificate can be set to false",
			arg: &kong.Upstream{
				Name: kong.String("foo"),
				Healthchecks: &kong.Healthcheck{
					Active: &kong.ActiveHealthcheck{
						Healthy: &kong.Healthy{
							Interval: kong.Int(1),
						},
						HTTPSVerifyCertificate: kong.Bool(false),
					},
				},
			},
			want: &kong.Upstream{
				Name:  kong.String("foo"),
				Slots: kong.Int(10000),
				Healthchecks: &kong.Healthcheck{
					Active: &kong.ActiveHealthcheck{
						Concurrency: kong.Int(10),
						Healthy: &kong.Healthy{
							HTTPStatuses: []int{200, 302},
							Interval:     kong.Int(1),
							Successes:    kong.Int(0),
						},
						HTTPPath:               kong.String("/"),
						HTTPSVerifyCertificate: kong.Bool(false),
						Type:                   kong.String("http"),
						Timeout:                kong.Int(1),
						Unhealthy: &kong.Unhealthy{
							HTTPFailures: kong.Int(0),
							TCPFailures:  kong.Int(0),
							Timeouts:     kong.Int(0),
							HTTPStatuses: []int{429, 404, 500, 501, 502, 503, 504, 505},
							Interval:     kong.Int(0),
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
			},
		},
		{
			desc: "Healthchecks.Active.HTTPSVerifyCertificate can be set to true",
			arg: &kong.Upstream{
				Name: kong.String("foo"),
				Healthchecks: &kong.Healthcheck{
					Active: &kong.ActiveHealthcheck{
						Healthy: &kong.Healthy{
							Interval: kong.Int(1),
						},
						HTTPSVerifyCertificate: kong.Bool(true),
					},
				},
			},
			want: &kong.Upstream{
				Name:  kong.String("foo"),
				Slots: kong.Int(10000),
				Healthchecks: &kong.Healthcheck{
					Active: &kong.ActiveHealthcheck{
						Concurrency: kong.Int(10),
						Healthy: &kong.Healthy{
							HTTPStatuses: []int{200, 302},
							Interval:     kong.Int(1),
							Successes:    kong.Int(0),
						},
						HTTPPath:               kong.String("/"),
						HTTPSVerifyCertificate: kong.Bool(true),
						Type:                   kong.String("http"),
						Timeout:                kong.Int(1),
						Unhealthy: &kong.Unhealthy{
							HTTPFailures: kong.Int(0),
							TCPFailures:  kong.Int(0),
							Timeouts:     kong.Int(0),
							HTTPStatuses: []int{429, 404, 500, 501, 502, 503, 504, 505},
							Interval:     kong.Int(0),
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
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			err := d.Set(tC.arg)
			assert.Nil(err)
			assert.Equal(tC.want, tC.arg)
		})
	}
}

func TestGetDefaulter_Konnect(t *testing.T) {
	assert := assert.New(t)

	testCases := []struct {
		desc string
		opts DefaulterOpts
		want *Defaulter
	}{
		{
			desc: "empty user defaults",
			opts: DefaulterOpts{
				KongDefaults:           &kongDefaultForTesting{},
				DisableDynamicDefaults: true,
			},
			want: &Defaulter{
				service:  &serviceDefaults,
				route:    &routeDefaults,
				upstream: &upstreamDefaults,
				target:   &targetDefaults,
			},
		},
		{
			desc: "user defaults take precedence",
			opts: DefaulterOpts{
				KongDefaults: &kongDefaultForTesting{
					Service: &kong.Service{
						Port:           kong.Int(8080),
						Path:           kong.String("/v1"),
						Protocol:       kong.String("http"),
						ConnectTimeout: kong.Int(defaultTimeout),
						WriteTimeout:   kong.Int(defaultTimeout),
						ReadTimeout:    kong.Int(defaultTimeout),
					},
				},
				DisableDynamicDefaults: true,
			},
			want: &Defaulter{
				service: &kong.Service{
					Port:           kong.Int(8080),
					Path:           kong.String("/v1"),
					Protocol:       kong.String("http"),
					ConnectTimeout: kong.Int(defaultTimeout),
					WriteTimeout:   kong.Int(defaultTimeout),
					ReadTimeout:    kong.Int(defaultTimeout),
				},
				route:    &routeDefaults,
				upstream: &upstreamDefaults,
				target:   &targetDefaults,
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			ctx := context.Background()
			d, err := GetDefaulter(ctx, tc.opts)
			assert.NotNil(d)
			assert.Nil(err)

			if !reflect.DeepEqual(d.service, tc.want.service) {
				assert.Equal(t, tc.want.service, d.service)
			}
			if !reflect.DeepEqual(d.route, tc.want.route) {
				assert.Equal(t, tc.want.route, d.route)
			}
			if !reflect.DeepEqual(d.upstream, tc.want.upstream) {
				assert.Equal(t, tc.want.upstream, d.upstream)
			}
			if !reflect.DeepEqual(d.target, tc.want.target) {
				assert.Equal(t, tc.want.target, d.target)
			}
		})
	}
}

func TestCheckRestrictedFields(t *testing.T) {
	assert := assert.New(t)

	testCases := []struct {
		desc             string
		entity           *kong.Service
		restrictedFields []string
		wantErr          bool
		expectedErr      string
	}{
		{
			desc: "no restricted fields",
			entity: &kong.Service{
				ID:   kong.String("testID"),
				Name: kong.String("testName"),
			},
			restrictedFields: []string{},
		},
		{
			desc: "one restricted fields",
			entity: &kong.Service{
				ID:   kong.String("testID"),
				Name: kong.String("testName"),
			},
			restrictedFields: []string{"ID"},
			wantErr:          true,
			expectedErr:      "cannot have these restricted fields set: id",
		},
		{
			desc: "multiple restricted fields",
			entity: &kong.Service{
				ID:   kong.String("testID"),
				Name: kong.String("testName"),
				Port: kong.Int(80),
			},
			restrictedFields: []string{"ID", "Name", "Port"},
			wantErr:          true,
			expectedErr:      "cannot have these restricted fields set: id, name, port",
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			err := checkEntityDefaults(tC.entity, tC.restrictedFields)
			if (err != nil) != tC.wantErr {
				t.Errorf("got error = %v, expected error = %v", err, tC.wantErr)
			}
			if tC.expectedErr != "" {
				assert.Equal(err.Error(), tC.expectedErr)
			}
		})
	}
}

func TestKongDefaultsRestrictedFields(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	testCases := []struct {
		desc         string
		kongDefaults *kongDefaultForTesting
		wantErr      bool
		expectedErr  string
	}{
		{
			desc: "service no restricted fields",
			kongDefaults: &kongDefaultForTesting{
				Service: &kong.Service{
					Path: kong.String("/v1"),
				},
			},
		},
		{
			desc: "route no restricted fields",
			kongDefaults: &kongDefaultForTesting{
				Route: &kong.Route{
					StripPath: kong.Bool(false),
				},
			},
		},
		{
			desc: "target no restricted fields",
			kongDefaults: &kongDefaultForTesting{
				Target: &kong.Target{
					Weight: kong.Int(42),
				},
			},
		},
		{
			desc: "upstream no restricted fields",
			kongDefaults: &kongDefaultForTesting{
				Upstream: &kong.Upstream{
					HostHeader: kong.String("testHostHeader"),
				},
			},
		},
		{
			desc: "service restricted fields",
			kongDefaults: &kongDefaultForTesting{
				Service: &kong.Service{
					ID:   kong.String("testID"),
					Name: kong.String("testName"),
					Path: kong.String("/v1"),
				},
			},
			wantErr:     true,
			expectedErr: "service defaults cannot have these restricted fields set: id, name",
		},
		{
			desc: "route restricted fields",
			kongDefaults: &kongDefaultForTesting{
				Route: &kong.Route{
					ID:        kong.String("testID"),
					Name:      kong.String("testName"),
					StripPath: kong.Bool(false),
				},
			},
			wantErr:     true,
			expectedErr: "route defaults cannot have these restricted fields set: id, name",
		},
		{
			desc: "target restricted fields",
			kongDefaults: &kongDefaultForTesting{
				Target: &kong.Target{
					ID:     kong.String("testID"),
					Target: kong.String("testTarget"),
				},
			},
			wantErr:     true,
			expectedErr: "target defaults cannot have these restricted fields set: id, target",
		},
		{
			desc: "upstream restricted fields",
			kongDefaults: &kongDefaultForTesting{
				Upstream: &kong.Upstream{
					ID:         kong.String("testID"),
					Name:       kong.String("testName"),
					HostHeader: kong.String("testHostHeader"),
				},
			},
			wantErr:     true,
			expectedErr: "upstream defaults cannot have these restricted fields set: id, name",
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			opts := DefaulterOpts{
				KongDefaults: tC.kongDefaults,
			}
			_, err := GetDefaulter(ctx, opts)
			if (err != nil) != tC.wantErr {
				t.Errorf("got error = %v, expected error = %v", err, tC.wantErr)
			}
			if tC.expectedErr != "" {
				assert.Contains(err.Error(), tC.expectedErr)
			}
		})
	}
}
