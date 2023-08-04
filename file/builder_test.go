package file

import (
	"context"
	"encoding/hex"
	"math/rand"
	"os"
	"reflect"
	"testing"

	"github.com/blang/semver/v4"
	"github.com/kong/deck/konnect"
	"github.com/kong/deck/state"
	"github.com/kong/deck/utils"
	"github.com/kong/go-kong/kong"
	"github.com/stretchr/testify/assert"
)

const (
	defaultTimeout     = 60000
	defaultSlots       = 10000
	defaultWeight      = 100
	defaultConcurrency = 10
)

var kong130Version = semver.MustParse("1.3.0")

var kongDefaults = KongDefaults{
	Service: &kong.Service{
		Protocol:       kong.String("http"),
		ConnectTimeout: kong.Int(defaultTimeout),
		WriteTimeout:   kong.Int(defaultTimeout),
		ReadTimeout:    kong.Int(defaultTimeout),
	},
	Route: &kong.Route{
		PreserveHost:  kong.Bool(false),
		RegexPriority: kong.Int(0),
		StripPath:     kong.Bool(false),
		Protocols:     kong.StringSlice("http", "https"),
	},
	Upstream: &kong.Upstream{
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
	},
	Target: &kong.Target{
		Weight: kong.Int(defaultWeight),
	},
}

var defaulterTestOpts = utils.DefaulterOpts{
	KongDefaults:           kongDefaults,
	DisableDynamicDefaults: false,
}

func emptyState() *state.KongState {
	s, _ := state.NewKongState()
	return s
}

func existingRouteState() *state.KongState {
	s, _ := state.NewKongState()
	s.Routes.Add(state.Route{
		Route: kong.Route{
			ID:   kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
			Name: kong.String("foo"),
		},
	})
	return s
}

func existingServiceState() *state.KongState {
	s, _ := state.NewKongState()
	s.Services.Add(state.Service{
		Service: kong.Service{
			ID:   kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
			Name: kong.String("foo"),
		},
	})
	return s
}

func existingConsumerCredState() *state.KongState {
	s, _ := state.NewKongState()
	s.Consumers.Add(state.Consumer{
		Consumer: kong.Consumer{
			ID:       kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
			Username: kong.String("foo"),
		},
	})
	s.KeyAuths.Add(state.KeyAuth{
		KeyAuth: kong.KeyAuth{
			ID:  kong.String("5f1ef1ea-a2a5-4a1b-adbb-b0d3434013e5"),
			Key: kong.String("foo-apikey"),
			Consumer: &kong.Consumer{
				ID:       kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
				Username: kong.String("foo"),
			},
		},
	})
	s.BasicAuths.Add(state.BasicAuth{
		BasicAuth: kong.BasicAuth{
			ID:       kong.String("92f4c849-960b-43af-aad3-f307051408d3"),
			Username: kong.String("basic-username"),
			Password: kong.String("basic-password"),
			Consumer: &kong.Consumer{
				ID:       kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
				Username: kong.String("foo"),
			},
		},
	})
	s.JWTAuths.Add(state.JWTAuth{
		JWTAuth: kong.JWTAuth{
			ID:     kong.String("917b9402-1be0-49d2-b482-ca4dccc2054e"),
			Key:    kong.String("jwt-key"),
			Secret: kong.String("jwt-secret"),
			Consumer: &kong.Consumer{
				ID:       kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
				Username: kong.String("foo"),
			},
		},
	})
	s.HMACAuths.Add(state.HMACAuth{
		HMACAuth: kong.HMACAuth{
			ID:       kong.String("e5d81b73-bf9e-42b0-9d68-30a1d791b9c9"),
			Username: kong.String("hmac-username"),
			Secret:   kong.String("hmac-secret"),
			Consumer: &kong.Consumer{
				ID:       kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
				Username: kong.String("foo"),
			},
		},
	})
	s.ACLGroups.Add(state.ACLGroup{
		ACLGroup: kong.ACLGroup{
			ID:    kong.String("b7c9352a-775a-4ba5-9869-98e926a3e6cb"),
			Group: kong.String("foo-group"),
			Consumer: &kong.Consumer{
				ID:       kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
				Username: kong.String("foo"),
			},
		},
	})
	s.Oauth2Creds.Add(state.Oauth2Credential{
		Oauth2Credential: kong.Oauth2Credential{
			ID:       kong.String("4eef5285-3d6a-4f6b-b659-8957a940e2ca"),
			ClientID: kong.String("oauth2-clientid"),
			Name:     kong.String("oauth2-name"),
			Consumer: &kong.Consumer{
				ID:       kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
				Username: kong.String("foo"),
			},
		},
	})
	s.MTLSAuths.Add(state.MTLSAuth{
		MTLSAuth: kong.MTLSAuth{
			ID:          kong.String("92f4c829-968b-42af-afd3-f337051508d3"),
			SubjectName: kong.String("test@example.com"),
			Consumer: &kong.Consumer{
				ID:       kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
				Username: kong.String("foo"),
			},
		},
	})
	return s
}

func existingUpstreamState() *state.KongState {
	s, _ := state.NewKongState()
	s.Upstreams.Add(state.Upstream{
		Upstream: kong.Upstream{
			ID:   kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
			Name: kong.String("foo"),
		},
	})
	return s
}

func existingCertificateState() *state.KongState {
	s, _ := state.NewKongState()
	s.Certificates.Add(state.Certificate{
		Certificate: kong.Certificate{
			ID:   kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
			Cert: kong.String("foo"),
			Key:  kong.String("bar"),
		},
	})
	return s
}

func existingCertificateAndSNIState() *state.KongState {
	s, _ := state.NewKongState()
	s.Certificates.Add(state.Certificate{
		Certificate: kong.Certificate{
			ID:   kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
			Cert: kong.String("foo"),
			Key:  kong.String("bar"),
		},
	})
	s.SNIs.Add(state.SNI{
		SNI: kong.SNI{
			ID:   kong.String("a53e9598-3a5e-4c12-a672-71a4cdcf7a47"),
			Name: kong.String("foo.example.com"),
			Certificate: &kong.Certificate{
				ID: kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
			},
		},
	})
	s.SNIs.Add(state.SNI{
		SNI: kong.SNI{
			ID:   kong.String("5f8e6848-4cb9-479a-a27e-860e1a77f875"),
			Name: kong.String("bar.example.com"),
			Certificate: &kong.Certificate{
				ID: kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
			},
		},
	})
	return s
}

func existingCACertificateState() *state.KongState {
	s, _ := state.NewKongState()
	s.CACertificates.Add(state.CACertificate{
		CACertificate: kong.CACertificate{
			ID:   kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
			Cert: kong.String("foo"),
		},
	})
	return s
}

func existingPluginState() *state.KongState {
	s, _ := state.NewKongState()
	s.Plugins.Add(state.Plugin{
		Plugin: kong.Plugin{
			ID:   kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
			Name: kong.String("foo"),
		},
	})
	s.Plugins.Add(state.Plugin{
		Plugin: kong.Plugin{
			ID:   kong.String("f7e64af5-e438-4a9b-8ff8-ec6f5f06dccb"),
			Name: kong.String("bar"),
			Consumer: &kong.Consumer{
				ID: kong.String("f77ca8c7-581d-45a4-a42c-c003234228e1"),
			},
		},
	})
	s.Plugins.Add(state.Plugin{
		Plugin: kong.Plugin{
			ID:   kong.String("53ce0a9c-d518-40ee-b8ab-1ee83a20d382"),
			Name: kong.String("foo"),
			Consumer: &kong.Consumer{
				ID: kong.String("f77ca8c7-581d-45a4-a42c-c003234228e1"),
			},
			Route: &kong.Route{
				ID: kong.String("700bc504-b2b1-4abd-bd38-cec92779659e"),
			},
			ConsumerGroup: &kong.ConsumerGroup{
				ID: kong.String("69ed4618-a653-4b54-8bb6-dc33bd6fe048"),
			},
		},
	})
	return s
}

func existingTargetsState() *state.KongState {
	s, _ := state.NewKongState()
	s.Targets.Add(state.Target{
		Target: kong.Target{
			ID:     kong.String("f7e64af5-e438-4a9b-8ff8-ec6f5f06dccb"),
			Target: kong.String("bar"),
			Upstream: &kong.Upstream{
				ID: kong.String("f77ca8c7-581d-45a4-a42c-c003234228e1"),
			},
		},
	})
	s.Targets.Add(state.Target{
		Target: kong.Target{
			ID:     kong.String("53ce0a9c-d518-40ee-b8ab-1ee83a20d382"),
			Target: kong.String("foo"),
			Upstream: &kong.Upstream{
				ID: kong.String("700bc504-b2b1-4abd-bd38-cec92779659e"),
			},
		},
	})
	return s
}

func existingDocumentState() *state.KongState {
	s, _ := state.NewKongState()
	s.ServicePackages.Add(state.ServicePackage{
		ServicePackage: konnect.ServicePackage{
			ID:   kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
			Name: kong.String("foo"),
		},
	})
	parent, _ := s.ServicePackages.Get("4bfcb11f-c962-4817-83e5-9433cf20b663")
	s.Documents.Add(state.Document{
		Document: konnect.Document{
			ID:        kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
			Path:      kong.String("/foo.md"),
			Published: kong.Bool(true),
			Content:   kong.String("foo"),
			Parent:    parent,
		},
	})
	return s
}

var deterministicUUID = func() *string {
	version := byte(4)
	uuid := make([]byte, 16)
	rand.Read(uuid)

	// Set version
	uuid[6] = (uuid[6] & 0x0f) | (version << 4)

	// Set variant
	uuid[8] = (uuid[8] & 0xbf) | 0x80

	buf := make([]byte, 36)
	var dash byte = '-'
	hex.Encode(buf[0:8], uuid[0:4])
	buf[8] = dash
	hex.Encode(buf[9:13], uuid[4:6])
	buf[13] = dash
	hex.Encode(buf[14:18], uuid[6:8])
	buf[18] = dash
	hex.Encode(buf[19:23], uuid[8:10])
	buf[23] = dash
	hex.Encode(buf[24:], uuid[10:])
	s := string(buf)
	return &s
}

func TestMain(m *testing.M) {
	uuid = deterministicUUID
	os.Exit(m.Run())
}

func Test_stateBuilder_services(t *testing.T) {
	assert := assert.New(t)
	rand.Seed(42)
	type fields struct {
		targetContent *Content
		currentState  *state.KongState
	}
	tests := []struct {
		name   string
		fields fields
		want   *utils.KongRawState
	}{
		{
			name: "matches ID of an existing service",
			fields: fields{
				targetContent: &Content{
					Info: &Info{
						Defaults: kongDefaults,
					},
					Services: []FService{
						{
							Service: kong.Service{
								Name: kong.String("foo"),
							},
						},
					},
				},
				currentState: existingServiceState(),
			},
			want: &utils.KongRawState{
				Services: []*kong.Service{
					{
						ID:             kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
						Name:           kong.String("foo"),
						Protocol:       kong.String("http"),
						ConnectTimeout: kong.Int(60000),
						WriteTimeout:   kong.Int(60000),
						ReadTimeout:    kong.Int(60000),
						Tags:           kong.StringSlice("tag1"),
					},
				},
			},
		},
		{
			name: "process a non-existent service",
			fields: fields{
				targetContent: &Content{
					Info: &Info{
						Defaults: kongDefaults,
					},
					Services: []FService{
						{
							Service: kong.Service{
								Name: kong.String("foo"),
							},
						},
					},
				},
				currentState: emptyState(),
			},
			want: &utils.KongRawState{
				Services: []*kong.Service{
					{
						ID:             kong.String("538c7f96-b164-4f1b-97bb-9f4bb472e89f"),
						Name:           kong.String("foo"),
						Protocol:       kong.String("http"),
						ConnectTimeout: kong.Int(60000),
						WriteTimeout:   kong.Int(60000),
						ReadTimeout:    kong.Int(60000),
						Tags:           kong.StringSlice("tag1"),
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &stateBuilder{
				targetContent: tt.fields.targetContent,
				currentState:  tt.fields.currentState,
				selectTags:    []string{"tag1"},
			}
			b.build()
			assert.Equal(tt.want, b.rawState)
		})
	}
}

func Test_stateBuilder_ingestRoute(t *testing.T) {
	assert := assert.New(t)
	rand.Seed(42)
	type fields struct {
		currentState *state.KongState
	}
	type args struct {
		route FRoute
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantErr   bool
		wantState *utils.KongRawState
	}{
		{
			name: "generates ID for a non-existing route",
			fields: fields{
				currentState: emptyState(),
			},
			args: args{
				route: FRoute{
					Route: kong.Route{
						Name: kong.String("foo"),
					},
				},
			},
			wantErr: false,
			wantState: &utils.KongRawState{
				Routes: []*kong.Route{
					{
						ID:            kong.String("538c7f96-b164-4f1b-97bb-9f4bb472e89f"),
						Name:          kong.String("foo"),
						PreserveHost:  kong.Bool(false),
						RegexPriority: kong.Int(0),
						StripPath:     kong.Bool(false),
						Protocols:     kong.StringSlice("http", "https"),
					},
				},
			},
		},
		{
			name: "matches up IDs of routes correctly",
			fields: fields{
				currentState: existingRouteState(),
			},
			args: args{
				route: FRoute{
					Route: kong.Route{
						Name: kong.String("foo"),
					},
				},
			},
			wantErr: false,
			wantState: &utils.KongRawState{
				Routes: []*kong.Route{
					{
						ID:            kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
						Name:          kong.String("foo"),
						PreserveHost:  kong.Bool(false),
						RegexPriority: kong.Int(0),
						StripPath:     kong.Bool(false),
						Protocols:     kong.StringSlice("http", "https"),
					},
				},
			},
		},
		{
			name: "grpc route has strip_path=false",
			fields: fields{
				currentState: existingRouteState(),
			},
			args: args{
				route: FRoute{
					Route: kong.Route{
						Name:      kong.String("foo"),
						Protocols: kong.StringSlice("grpc"),
					},
				},
			},
			wantErr: false,
			wantState: &utils.KongRawState{
				Routes: []*kong.Route{
					{
						ID:            kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
						Name:          kong.String("foo"),
						PreserveHost:  kong.Bool(false),
						RegexPriority: kong.Int(0),
						StripPath:     kong.Bool(false),
						Protocols:     kong.StringSlice("grpc"),
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			b := &stateBuilder{
				currentState: tt.fields.currentState,
			}
			b.rawState = &utils.KongRawState{}
			d, _ := utils.GetDefaulter(ctx, defaulterTestOpts)
			b.defaulter = d
			b.intermediate, _ = state.NewKongState()
			if err := b.ingestRoute(tt.args.route); (err != nil) != tt.wantErr {
				t.Errorf("stateBuilder.ingestPlugins() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(tt.wantState, b.rawState)
		})
	}
}

func Test_stateBuilder_ingestTargets(t *testing.T) {
	assert := assert.New(t)
	rand.Seed(42)
	type fields struct {
		currentState *state.KongState
	}
	type args struct {
		targets []kong.Target
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantErr   bool
		wantState *utils.KongRawState
	}{
		{
			name: "generates ID for a non-existing target",
			fields: fields{
				currentState: emptyState(),
			},
			args: args{
				targets: []kong.Target{
					{
						Target: kong.String("foo"),
						Upstream: &kong.Upstream{
							ID: kong.String("952ddf37-e815-40b6-b119-5379a3b1f7be"),
						},
					},
				},
			},
			wantErr: false,
			wantState: &utils.KongRawState{
				Targets: []*kong.Target{
					{
						ID:     kong.String("538c7f96-b164-4f1b-97bb-9f4bb472e89f"),
						Target: kong.String("foo"),
						Weight: kong.Int(100),
						Upstream: &kong.Upstream{
							ID: kong.String("952ddf37-e815-40b6-b119-5379a3b1f7be"),
						},
					},
				},
			},
		},
		{
			name: "matches up IDs of Targets correctly",
			fields: fields{
				currentState: existingTargetsState(),
			},
			args: args{
				targets: []kong.Target{
					{
						Target: kong.String("bar"),
						Upstream: &kong.Upstream{
							ID: kong.String("f77ca8c7-581d-45a4-a42c-c003234228e1"),
						},
					},
					{
						Target: kong.String("foo"),
						Upstream: &kong.Upstream{
							ID: kong.String("700bc504-b2b1-4abd-bd38-cec92779659e"),
						},
					},
				},
			},
			wantErr: false,
			wantState: &utils.KongRawState{
				Targets: []*kong.Target{
					{
						ID:     kong.String("f7e64af5-e438-4a9b-8ff8-ec6f5f06dccb"),
						Target: kong.String("bar"),
						Weight: kong.Int(100),
						Upstream: &kong.Upstream{
							ID: kong.String("f77ca8c7-581d-45a4-a42c-c003234228e1"),
						},
					},
					{
						ID:     kong.String("53ce0a9c-d518-40ee-b8ab-1ee83a20d382"),
						Target: kong.String("foo"),
						Weight: kong.Int(100),
						Upstream: &kong.Upstream{
							ID: kong.String("700bc504-b2b1-4abd-bd38-cec92779659e"),
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			b := &stateBuilder{
				currentState: tt.fields.currentState,
			}
			b.rawState = &utils.KongRawState{}
			d, _ := utils.GetDefaulter(ctx, defaulterTestOpts)
			b.defaulter = d
			if err := b.ingestTargets(tt.args.targets); (err != nil) != tt.wantErr {
				t.Errorf("stateBuilder.ingestPlugins() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(tt.wantState, b.rawState)
		})
	}
}

func Test_stateBuilder_ingestPlugins(t *testing.T) {
	assert := assert.New(t)
	rand.Seed(42)
	type fields struct {
		currentState *state.KongState
	}
	type args struct {
		plugins []FPlugin
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantErr   bool
		wantState *utils.KongRawState
	}{
		{
			name: "generates ID for a non-existing plugin",
			fields: fields{
				currentState: emptyState(),
			},
			args: args{
				plugins: []FPlugin{
					{
						Plugin: kong.Plugin{
							Name: kong.String("foo"),
						},
					},
				},
			},
			wantErr: false,
			wantState: &utils.KongRawState{
				Plugins: []*kong.Plugin{
					{
						ID:     kong.String("538c7f96-b164-4f1b-97bb-9f4bb472e89f"),
						Name:   kong.String("foo"),
						Config: kong.Configuration{},
					},
				},
			},
		},
		{
			name: "matches up IDs of plugins correctly",
			fields: fields{
				currentState: existingPluginState(),
			},
			args: args{
				plugins: []FPlugin{
					{
						Plugin: kong.Plugin{
							Name: kong.String("foo"),
						},
					},
					{
						Plugin: kong.Plugin{
							Name: kong.String("bar"),
							Consumer: &kong.Consumer{
								ID: kong.String("f77ca8c7-581d-45a4-a42c-c003234228e1"),
							},
						},
					},
					{
						Plugin: kong.Plugin{
							Name: kong.String("foo"),
							Consumer: &kong.Consumer{
								ID: kong.String("f77ca8c7-581d-45a4-a42c-c003234228e1"),
							},
							Route: &kong.Route{
								ID: kong.String("700bc504-b2b1-4abd-bd38-cec92779659e"),
							},
							ConsumerGroup: &kong.ConsumerGroup{
								ID: kong.String("69ed4618-a653-4b54-8bb6-dc33bd6fe048"),
							},
						},
					},
				},
			},
			wantErr: false,
			wantState: &utils.KongRawState{
				Plugins: []*kong.Plugin{
					{
						ID:     kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
						Name:   kong.String("foo"),
						Config: kong.Configuration{},
					},
					{
						ID:   kong.String("f7e64af5-e438-4a9b-8ff8-ec6f5f06dccb"),
						Name: kong.String("bar"),
						Consumer: &kong.Consumer{
							ID: kong.String("f77ca8c7-581d-45a4-a42c-c003234228e1"),
						},
						Config: kong.Configuration{},
					},
					{
						ID:   kong.String("53ce0a9c-d518-40ee-b8ab-1ee83a20d382"),
						Name: kong.String("foo"),
						Consumer: &kong.Consumer{
							ID: kong.String("f77ca8c7-581d-45a4-a42c-c003234228e1"),
						},
						Route: &kong.Route{
							ID: kong.String("700bc504-b2b1-4abd-bd38-cec92779659e"),
						},
						ConsumerGroup: &kong.ConsumerGroup{
							ID: kong.String("69ed4618-a653-4b54-8bb6-dc33bd6fe048"),
						},
						Config: kong.Configuration{},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &stateBuilder{
				currentState: tt.fields.currentState,
			}
			b.rawState = &utils.KongRawState{}
			if err := b.ingestPlugins(tt.args.plugins); (err != nil) != tt.wantErr {
				t.Errorf("stateBuilder.ingestPlugins() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(tt.wantState, b.rawState)
		})
	}
}

func Test_pluginRelations(t *testing.T) {
	type args struct {
		plugin *kong.Plugin
	}
	tests := []struct {
		name     string
		args     args
		wantCID  string
		wantRID  string
		wantSID  string
		wantCGID string
	}{
		{
			args: args{
				plugin: &kong.Plugin{
					Name: kong.String("foo"),
				},
			},
			wantCID:  "",
			wantRID:  "",
			wantSID:  "",
			wantCGID: "",
		},
		{
			args: args{
				plugin: &kong.Plugin{
					Name: kong.String("foo"),
					Consumer: &kong.Consumer{
						ID: kong.String("cID"),
					},
					Route: &kong.Route{
						ID: kong.String("rID"),
					},
					Service: &kong.Service{
						ID: kong.String("sID"),
					},
					ConsumerGroup: &kong.ConsumerGroup{
						ID: kong.String("cgID"),
					},
				},
			},
			wantCID:  "cID",
			wantRID:  "rID",
			wantSID:  "sID",
			wantCGID: "cgID",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCID, gotRID, gotSID, gotCGID := pluginRelations(tt.args.plugin)
			if gotCID != tt.wantCID {
				t.Errorf("pluginRelations() gotCID = %v, want %v", gotCID, tt.wantCID)
			}
			if gotRID != tt.wantRID {
				t.Errorf("pluginRelations() gotRID = %v, want %v", gotRID, tt.wantRID)
			}
			if gotSID != tt.wantSID {
				t.Errorf("pluginRelations() gotSID = %v, want %v", gotSID, tt.wantSID)
			}
			if gotCGID != tt.wantCGID {
				t.Errorf("pluginRelations() gotCGID = %v, want %v", gotCGID, tt.wantCGID)
			}
		})
	}
}

func Test_stateBuilder_consumers(t *testing.T) {
	assert := assert.New(t)
	rand.Seed(42)
	type fields struct {
		currentState  *state.KongState
		targetContent *Content
		kongVersion   *semver.Version
	}
	tests := []struct {
		name   string
		fields fields
		want   *utils.KongRawState
	}{
		{
			name: "generates ID for a non-existing consumer",
			fields: fields{
				targetContent: &Content{
					Consumers: []FConsumer{
						{
							Consumer: kong.Consumer{
								Username: kong.String("foo"),
							},
						},
					},
					Info: &Info{},
				},
				currentState: emptyState(),
			},
			want: &utils.KongRawState{
				Consumers: []*kong.Consumer{
					{
						ID:       kong.String("538c7f96-b164-4f1b-97bb-9f4bb472e89f"),
						Username: kong.String("foo"),
					},
				},
			},
		},
		{
			name: "generates ID for a non-existing credential",
			fields: fields{
				targetContent: &Content{
					Consumers: []FConsumer{
						{
							Consumer: kong.Consumer{
								Username: kong.String("foo"),
							},
							KeyAuths: []*kong.KeyAuth{
								{
									Key: kong.String("foo-key"),
								},
							},
							BasicAuths: []*kong.BasicAuth{
								{
									Username: kong.String("basic-username"),
									Password: kong.String("basic-password"),
								},
							},
							HMACAuths: []*kong.HMACAuth{
								{
									Username: kong.String("hmac-username"),
									Secret:   kong.String("hmac-secret"),
								},
							},
							JWTAuths: []*kong.JWTAuth{
								{
									Key:    kong.String("jwt-key"),
									Secret: kong.String("jwt-secret"),
								},
							},
							Oauth2Creds: []*kong.Oauth2Credential{
								{
									ClientID: kong.String("oauth2-clientid"),
									Name:     kong.String("oauth2-name"),
								},
							},
							ACLGroups: []*kong.ACLGroup{
								{
									Group: kong.String("foo-group"),
								},
							},
						},
					},
					Info: &Info{},
				},
				currentState: emptyState(),
			},
			want: &utils.KongRawState{
				Consumers: []*kong.Consumer{
					{
						ID:       kong.String("5b1484f2-5209-49d9-b43e-92ba09dd9d52"),
						Username: kong.String("foo"),
					},
				},
				KeyAuths: []*kong.KeyAuth{
					{
						ID:  kong.String("dfd79b4d-7642-4b61-ba0c-9f9f0d3ba55b"),
						Key: kong.String("foo-key"),
						Consumer: &kong.Consumer{
							ID:       kong.String("5b1484f2-5209-49d9-b43e-92ba09dd9d52"),
							Username: kong.String("foo"),
						},
					},
				},
				BasicAuths: []*kong.BasicAuth{
					{
						ID:       kong.String("0cc0d614-4c88-4535-841a-cbe0709b0758"),
						Username: kong.String("basic-username"),
						Password: kong.String("basic-password"),
						Consumer: &kong.Consumer{
							ID:       kong.String("5b1484f2-5209-49d9-b43e-92ba09dd9d52"),
							Username: kong.String("foo"),
						},
					},
				},
				HMACAuths: []*kong.HMACAuth{
					{
						ID:       kong.String("083f61d3-75bc-42b4-9df4-f91929e18fda"),
						Username: kong.String("hmac-username"),
						Secret:   kong.String("hmac-secret"),
						Consumer: &kong.Consumer{
							ID:       kong.String("5b1484f2-5209-49d9-b43e-92ba09dd9d52"),
							Username: kong.String("foo"),
						},
					},
				},
				JWTAuths: []*kong.JWTAuth{
					{
						ID:     kong.String("9e6f82e5-4e74-4e81-a79e-4bbd6fe34cdc"),
						Key:    kong.String("jwt-key"),
						Secret: kong.String("jwt-secret"),
						Consumer: &kong.Consumer{
							ID:       kong.String("5b1484f2-5209-49d9-b43e-92ba09dd9d52"),
							Username: kong.String("foo"),
						},
					},
				},
				Oauth2Creds: []*kong.Oauth2Credential{
					{
						ID:       kong.String("ba843ee8-d63e-4c4f-be1c-ebea546d8fac"),
						ClientID: kong.String("oauth2-clientid"),
						Name:     kong.String("oauth2-name"),
						Consumer: &kong.Consumer{
							ID:       kong.String("5b1484f2-5209-49d9-b43e-92ba09dd9d52"),
							Username: kong.String("foo"),
						},
					},
				},
				ACLGroups: []*kong.ACLGroup{
					{
						ID:    kong.String("13dd1aac-04ce-4ea2-877c-5579cfa2c78e"),
						Group: kong.String("foo-group"),
						Consumer: &kong.Consumer{
							ID:       kong.String("5b1484f2-5209-49d9-b43e-92ba09dd9d52"),
							Username: kong.String("foo"),
						},
					},
				},
				MTLSAuths: nil,
			},
		},
		{
			name: "matches ID of an existing consumer",
			fields: fields{
				targetContent: &Content{
					Consumers: []FConsumer{
						{
							Consumer: kong.Consumer{
								Username: kong.String("foo"),
							},
						},
					},
				},
				currentState: existingConsumerCredState(),
			},
			want: &utils.KongRawState{
				Consumers: []*kong.Consumer{
					{
						ID:       kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
						Username: kong.String("foo"),
					},
				},
			},
		},
		{
			name: "matches ID of an existing credential",
			fields: fields{
				targetContent: &Content{
					Consumers: []FConsumer{
						{
							Consumer: kong.Consumer{
								Username: kong.String("foo"),
							},
							KeyAuths: []*kong.KeyAuth{
								{
									Key: kong.String("foo-apikey"),
								},
							},
							BasicAuths: []*kong.BasicAuth{
								{
									Username: kong.String("basic-username"),
									Password: kong.String("basic-password"),
								},
							},
							HMACAuths: []*kong.HMACAuth{
								{
									Username: kong.String("hmac-username"),
									Secret:   kong.String("hmac-secret"),
								},
							},
							JWTAuths: []*kong.JWTAuth{
								{
									Key:    kong.String("jwt-key"),
									Secret: kong.String("jwt-secret"),
								},
							},
							Oauth2Creds: []*kong.Oauth2Credential{
								{
									ClientID: kong.String("oauth2-clientid"),
									Name:     kong.String("oauth2-name"),
								},
							},
							ACLGroups: []*kong.ACLGroup{
								{
									Group: kong.String("foo-group"),
								},
							},
							MTLSAuths: []*kong.MTLSAuth{
								{
									ID:          kong.String("533c259e-bf71-4d77-99d2-97944c70a6a4"),
									SubjectName: kong.String("test@example.com"),
								},
							},
						},
					},
					Info: &Info{},
				},
				currentState: existingConsumerCredState(),
			},
			want: &utils.KongRawState{
				Consumers: []*kong.Consumer{
					{
						ID:       kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
						Username: kong.String("foo"),
					},
				},
				KeyAuths: []*kong.KeyAuth{
					{
						ID:  kong.String("5f1ef1ea-a2a5-4a1b-adbb-b0d3434013e5"),
						Key: kong.String("foo-apikey"),
						Consumer: &kong.Consumer{
							ID:       kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
							Username: kong.String("foo"),
						},
					},
				},
				BasicAuths: []*kong.BasicAuth{
					{
						ID:       kong.String("92f4c849-960b-43af-aad3-f307051408d3"),
						Username: kong.String("basic-username"),
						Password: kong.String("basic-password"),
						Consumer: &kong.Consumer{
							ID:       kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
							Username: kong.String("foo"),
						},
					},
				},
				HMACAuths: []*kong.HMACAuth{
					{
						ID:       kong.String("e5d81b73-bf9e-42b0-9d68-30a1d791b9c9"),
						Username: kong.String("hmac-username"),
						Secret:   kong.String("hmac-secret"),
						Consumer: &kong.Consumer{
							ID:       kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
							Username: kong.String("foo"),
						},
					},
				},
				JWTAuths: []*kong.JWTAuth{
					{
						ID:     kong.String("917b9402-1be0-49d2-b482-ca4dccc2054e"),
						Key:    kong.String("jwt-key"),
						Secret: kong.String("jwt-secret"),
						Consumer: &kong.Consumer{
							ID:       kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
							Username: kong.String("foo"),
						},
					},
				},
				Oauth2Creds: []*kong.Oauth2Credential{
					{
						ID:       kong.String("4eef5285-3d6a-4f6b-b659-8957a940e2ca"),
						ClientID: kong.String("oauth2-clientid"),
						Name:     kong.String("oauth2-name"),
						Consumer: &kong.Consumer{
							ID:       kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
							Username: kong.String("foo"),
						},
					},
				},
				ACLGroups: []*kong.ACLGroup{
					{
						ID:    kong.String("b7c9352a-775a-4ba5-9869-98e926a3e6cb"),
						Group: kong.String("foo-group"),
						Consumer: &kong.Consumer{
							ID:       kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
							Username: kong.String("foo"),
						},
					},
				},
				MTLSAuths: []*kong.MTLSAuth{
					{
						ID:          kong.String("533c259e-bf71-4d77-99d2-97944c70a6a4"),
						SubjectName: kong.String("test@example.com"),
						Consumer: &kong.Consumer{
							ID:       kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
							Username: kong.String("foo"),
						},
					},
				},
			},
		},
		{
			name: "does not inject tags if Kong version is older than 1.4",
			fields: fields{
				targetContent: &Content{
					Consumers: []FConsumer{
						{
							Consumer: kong.Consumer{
								Username: kong.String("foo"),
							},
							KeyAuths: []*kong.KeyAuth{
								{
									Key: kong.String("foo-apikey"),
								},
							},
							BasicAuths: []*kong.BasicAuth{
								{
									Username: kong.String("basic-username"),
									Password: kong.String("basic-password"),
								},
							},
							HMACAuths: []*kong.HMACAuth{
								{
									Username: kong.String("hmac-username"),
									Secret:   kong.String("hmac-secret"),
								},
							},
							JWTAuths: []*kong.JWTAuth{
								{
									Key:    kong.String("jwt-key"),
									Secret: kong.String("jwt-secret"),
								},
							},
							Oauth2Creds: []*kong.Oauth2Credential{
								{
									ClientID: kong.String("oauth2-clientid"),
									Name:     kong.String("oauth2-name"),
								},
							},
							ACLGroups: []*kong.ACLGroup{
								{
									Group: kong.String("foo-group"),
								},
							},
							MTLSAuths: []*kong.MTLSAuth{
								{
									ID:          kong.String("533c259e-bf71-4d77-99d2-97944c70a6a4"),
									SubjectName: kong.String("test@example.com"),
								},
							},
						},
					},
					Info: &Info{},
				},
				currentState: existingConsumerCredState(),
				kongVersion:  &kong130Version,
			},
			want: &utils.KongRawState{
				Consumers: []*kong.Consumer{
					{
						ID:       kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
						Username: kong.String("foo"),
					},
				},
				KeyAuths: []*kong.KeyAuth{
					{
						ID:  kong.String("5f1ef1ea-a2a5-4a1b-adbb-b0d3434013e5"),
						Key: kong.String("foo-apikey"),
						Consumer: &kong.Consumer{
							ID:       kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
							Username: kong.String("foo"),
						},
					},
				},
				BasicAuths: []*kong.BasicAuth{
					{
						ID:       kong.String("92f4c849-960b-43af-aad3-f307051408d3"),
						Username: kong.String("basic-username"),
						Password: kong.String("basic-password"),
						Consumer: &kong.Consumer{
							ID:       kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
							Username: kong.String("foo"),
						},
					},
				},
				HMACAuths: []*kong.HMACAuth{
					{
						ID:       kong.String("e5d81b73-bf9e-42b0-9d68-30a1d791b9c9"),
						Username: kong.String("hmac-username"),
						Secret:   kong.String("hmac-secret"),
						Consumer: &kong.Consumer{
							ID:       kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
							Username: kong.String("foo"),
						},
					},
				},
				JWTAuths: []*kong.JWTAuth{
					{
						ID:     kong.String("917b9402-1be0-49d2-b482-ca4dccc2054e"),
						Key:    kong.String("jwt-key"),
						Secret: kong.String("jwt-secret"),
						Consumer: &kong.Consumer{
							ID:       kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
							Username: kong.String("foo"),
						},
					},
				},
				Oauth2Creds: []*kong.Oauth2Credential{
					{
						ID:       kong.String("4eef5285-3d6a-4f6b-b659-8957a940e2ca"),
						ClientID: kong.String("oauth2-clientid"),
						Name:     kong.String("oauth2-name"),
						Consumer: &kong.Consumer{
							ID:       kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
							Username: kong.String("foo"),
						},
					},
				},
				ACLGroups: []*kong.ACLGroup{
					{
						ID:    kong.String("b7c9352a-775a-4ba5-9869-98e926a3e6cb"),
						Group: kong.String("foo-group"),
						Consumer: &kong.Consumer{
							ID:       kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
							Username: kong.String("foo"),
						},
					},
				},
				MTLSAuths: []*kong.MTLSAuth{
					{
						ID:          kong.String("533c259e-bf71-4d77-99d2-97944c70a6a4"),
						SubjectName: kong.String("test@example.com"),
						Consumer: &kong.Consumer{
							ID:       kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
							Username: kong.String("foo"),
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			b := &stateBuilder{
				targetContent: tt.fields.targetContent,
				currentState:  tt.fields.currentState,
				kongVersion:   utils.Kong140Version,
			}
			if tt.fields.kongVersion != nil {
				b.kongVersion = *tt.fields.kongVersion
			}
			d, _ := utils.GetDefaulter(ctx, defaulterTestOpts)
			b.defaulter = d
			b.build()
			assert.Equal(tt.want, b.rawState)
		})
	}
}

func Test_stateBuilder_certificates(t *testing.T) {
	assert := assert.New(t)
	rand.Seed(42)
	type fields struct {
		currentState  *state.KongState
		targetContent *Content
	}
	tests := []struct {
		name   string
		fields fields
		want   *utils.KongRawState
	}{
		{
			name: "generates ID for a non-existing certificate",
			fields: fields{
				targetContent: &Content{
					Certificates: []FCertificate{
						{
							Cert: kong.String("foo"),
							Key:  kong.String("bar"),
						},
					},
				},
				currentState: emptyState(),
			},
			want: &utils.KongRawState{
				Certificates: []*kong.Certificate{
					{
						ID:   kong.String("538c7f96-b164-4f1b-97bb-9f4bb472e89f"),
						Cert: kong.String("foo"),
						Key:  kong.String("bar"),
					},
				},
			},
		},
		{
			name: "matches ID of an existing certificate",
			fields: fields{
				targetContent: &Content{
					Certificates: []FCertificate{
						{
							Cert: kong.String("foo"),
							Key:  kong.String("bar"),
						},
					},
				},
				currentState: existingCertificateState(),
			},
			want: &utils.KongRawState{
				Certificates: []*kong.Certificate{
					{
						ID:   kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
						Cert: kong.String("foo"),
						Key:  kong.String("bar"),
					},
				},
			},
		},
		{
			name: "generates ID for SNIs",
			fields: fields{
				targetContent: &Content{
					Certificates: []FCertificate{
						{
							Cert: kong.String("foo"),
							Key:  kong.String("bar"),
							SNIs: []kong.SNI{
								{
									Name: kong.String("foo.example.com"),
								},
								{
									Name: kong.String("bar.example.com"),
								},
							},
						},
					},
				},
				currentState: existingCertificateState(),
			},
			want: &utils.KongRawState{
				Certificates: []*kong.Certificate{
					{
						ID:   kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
						Cert: kong.String("foo"),
						Key:  kong.String("bar"),
					},
				},
				SNIs: []*kong.SNI{
					{
						ID:   kong.String("5b1484f2-5209-49d9-b43e-92ba09dd9d52"),
						Name: kong.String("foo.example.com"),
						Certificate: &kong.Certificate{
							ID: kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
						},
					},
					{
						ID:   kong.String("dfd79b4d-7642-4b61-ba0c-9f9f0d3ba55b"),
						Name: kong.String("bar.example.com"),
						Certificate: &kong.Certificate{
							ID: kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
						},
					},
				},
			},
		},
		{
			name: "matches ID for SNIs",
			fields: fields{
				targetContent: &Content{
					Certificates: []FCertificate{
						{
							Cert: kong.String("foo"),
							Key:  kong.String("bar"),
							SNIs: []kong.SNI{
								{
									Name: kong.String("foo.example.com"),
								},
								{
									Name: kong.String("bar.example.com"),
								},
							},
						},
					},
				},
				currentState: existingCertificateAndSNIState(),
			},
			want: &utils.KongRawState{
				Certificates: []*kong.Certificate{
					{
						ID:   kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
						Cert: kong.String("foo"),
						Key:  kong.String("bar"),
					},
				},
				SNIs: []*kong.SNI{
					{
						ID:   kong.String("a53e9598-3a5e-4c12-a672-71a4cdcf7a47"),
						Name: kong.String("foo.example.com"),
						Certificate: &kong.Certificate{
							ID: kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
						},
					},
					{
						ID:   kong.String("5f8e6848-4cb9-479a-a27e-860e1a77f875"),
						Name: kong.String("bar.example.com"),
						Certificate: &kong.Certificate{
							ID: kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			b := &stateBuilder{
				targetContent: tt.fields.targetContent,
				currentState:  tt.fields.currentState,
			}
			d, _ := utils.GetDefaulter(ctx, defaulterTestOpts)
			b.defaulter = d
			b.build()
			assert.Equal(tt.want, b.rawState)
		})
	}
}

func Test_stateBuilder_caCertificates(t *testing.T) {
	assert := assert.New(t)
	rand.Seed(42)
	type fields struct {
		currentState  *state.KongState
		targetContent *Content
	}
	tests := []struct {
		name   string
		fields fields
		want   *utils.KongRawState
	}{
		{
			name: "generates ID for a non-existing CACertificate",
			fields: fields{
				targetContent: &Content{
					CACertificates: []FCACertificate{
						{
							CACertificate: kong.CACertificate{
								Cert: kong.String("foo"),
							},
						},
					},
				},
				currentState: emptyState(),
			},
			want: &utils.KongRawState{
				CACertificates: []*kong.CACertificate{
					{
						ID:   kong.String("538c7f96-b164-4f1b-97bb-9f4bb472e89f"),
						Cert: kong.String("foo"),
					},
				},
			},
		},
		{
			name: "matches ID of an existing CACertificate",
			fields: fields{
				targetContent: &Content{
					CACertificates: []FCACertificate{
						{
							CACertificate: kong.CACertificate{
								Cert: kong.String("foo"),
							},
						},
					},
				},
				currentState: existingCACertificateState(),
			},
			want: &utils.KongRawState{
				CACertificates: []*kong.CACertificate{
					{
						ID:   kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
						Cert: kong.String("foo"),
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			b := &stateBuilder{
				targetContent: tt.fields.targetContent,
				currentState:  tt.fields.currentState,
			}
			d, _ := utils.GetDefaulter(ctx, defaulterTestOpts)
			b.defaulter = d
			b.build()
			assert.Equal(tt.want, b.rawState)
		})
	}
}

func Test_stateBuilder_upstream(t *testing.T) {
	assert := assert.New(t)
	rand.Seed(42)
	type fields struct {
		targetContent *Content
		currentState  *state.KongState
	}
	tests := []struct {
		name   string
		fields fields
		want   *utils.KongRawState
	}{
		{
			name: "process a non-existent upstream",
			fields: fields{
				targetContent: &Content{
					Info: &Info{
						Defaults: kongDefaults,
					},
					Upstreams: []FUpstream{
						{
							Upstream: kong.Upstream{
								Name:  kong.String("foo"),
								Slots: kong.Int(42),
							},
						},
					},
				},
				currentState: existingServiceState(),
			},
			want: &utils.KongRawState{
				Upstreams: []*kong.Upstream{
					{
						ID:    kong.String("538c7f96-b164-4f1b-97bb-9f4bb472e89f"),
						Name:  kong.String("foo"),
						Slots: kong.Int(42),
						Healthchecks: &kong.Healthcheck{
							Active: &kong.ActiveHealthcheck{
								Concurrency: kong.Int(10),
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
					},
				},
			},
		},
		{
			name: "matches ID of an existing service",
			fields: fields{
				targetContent: &Content{
					Info: &Info{
						Defaults: kongDefaults,
					},
					Upstreams: []FUpstream{
						{
							Upstream: kong.Upstream{
								Name: kong.String("foo"),
							},
						},
					},
				},
				currentState: existingUpstreamState(),
			},
			want: &utils.KongRawState{
				Upstreams: []*kong.Upstream{
					{
						ID:    kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
						Name:  kong.String("foo"),
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
					},
				},
			},
		},
		{
			name: "multiple upstreams are handled correctly",
			fields: fields{
				targetContent: &Content{
					Info: &Info{
						Defaults: kongDefaults,
					},
					Upstreams: []FUpstream{
						{
							Upstream: kong.Upstream{
								Name: kong.String("foo"),
							},
						},
						{
							Upstream: kong.Upstream{
								Name: kong.String("bar"),
							},
						},
					},
				},
				currentState: emptyState(),
			},
			want: &utils.KongRawState{
				Upstreams: []*kong.Upstream{
					{
						ID:    kong.String("5b1484f2-5209-49d9-b43e-92ba09dd9d52"),
						Name:  kong.String("foo"),
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
					},
					{
						ID:    kong.String("dfd79b4d-7642-4b61-ba0c-9f9f0d3ba55b"),
						Name:  kong.String("bar"),
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
					},
				},
			},
		},
		{
			name: "upstream with new 3.0 fields",
			fields: fields{
				targetContent: &Content{
					Info: &Info{
						Defaults: kongDefaults,
					},
					Upstreams: []FUpstream{
						{
							Upstream: kong.Upstream{
								Name:  kong.String("foo"),
								Slots: kong.Int(42),
								// not actually valid configuration, but this only needs to check that these translate
								// into the raw state
								HashOnQueryArg:         kong.String("foo"),
								HashFallbackQueryArg:   kong.String("foo"),
								HashOnURICapture:       kong.String("foo"),
								HashFallbackURICapture: kong.String("foo"),
							},
						},
					},
				},
				currentState: existingServiceState(),
			},
			want: &utils.KongRawState{
				Upstreams: []*kong.Upstream{
					{
						ID:    kong.String("0cc0d614-4c88-4535-841a-cbe0709b0758"),
						Name:  kong.String("foo"),
						Slots: kong.Int(42),
						Healthchecks: &kong.Healthcheck{
							Active: &kong.ActiveHealthcheck{
								Concurrency: kong.Int(10),
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
						HashOn:                 kong.String("none"),
						HashFallback:           kong.String("none"),
						HashOnCookiePath:       kong.String("/"),
						HashOnQueryArg:         kong.String("foo"),
						HashFallbackQueryArg:   kong.String("foo"),
						HashOnURICapture:       kong.String("foo"),
						HashFallbackURICapture: kong.String("foo"),
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			b := &stateBuilder{
				targetContent: tt.fields.targetContent,
				currentState:  tt.fields.currentState,
			}
			d, _ := utils.GetDefaulter(ctx, defaulterTestOpts)
			b.defaulter = d
			b.build()
			assert.Equal(tt.want, b.rawState)
		})
	}
}

func Test_stateBuilder_documents(t *testing.T) {
	assert := assert.New(t)
	rand.Seed(42)
	type fields struct {
		targetContent *Content
		currentState  *state.KongState
	}
	tests := []struct {
		name   string
		fields fields
		want   *utils.KonnectRawState
	}{
		{
			name: "matches ID of an existing document",
			fields: fields{
				targetContent: &Content{
					ServicePackages: []FServicePackage{
						{
							Name: kong.String("foo"),
							Document: &FDocument{
								Path:      kong.String("/foo.md"),
								Published: kong.Bool(true),
								Content:   kong.String("foo"),
							},
						},
					},
				},
				currentState: existingDocumentState(),
			},
			want: &utils.KonnectRawState{
				Documents: []*konnect.Document{
					{
						ID:        kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
						Path:      kong.String("/foo.md"),
						Published: kong.Bool(true),
						Content:   kong.String("foo"),
						Parent: &konnect.ServicePackage{
							ID:   kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
							Name: kong.String("foo"),
						},
					},
				},
				ServicePackages: []*konnect.ServicePackage{
					{
						ID:   kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
						Name: kong.String("foo"),
					},
				},
			},
		},
		{
			name: "process a non-existent document",
			fields: fields{
				targetContent: &Content{
					ServicePackages: []FServicePackage{
						{
							Name: kong.String("bar"),
							Document: &FDocument{
								Path:      kong.String("/bar.md"),
								Published: kong.Bool(true),
								Content:   kong.String("bar"),
							},
						},
					},
				},
				currentState: existingDocumentState(),
			},
			want: &utils.KonnectRawState{
				Documents: []*konnect.Document{
					{
						ID:        kong.String("5b1484f2-5209-49d9-b43e-92ba09dd9d52"),
						Path:      kong.String("/bar.md"),
						Published: kong.Bool(true),
						Content:   kong.String("bar"),
						Parent: &konnect.ServicePackage{
							ID:   kong.String("538c7f96-b164-4f1b-97bb-9f4bb472e89f"),
							Name: kong.String("bar"),
						},
					},
				},
				ServicePackages: []*konnect.ServicePackage{
					{
						ID:   kong.String("538c7f96-b164-4f1b-97bb-9f4bb472e89f"),
						Name: kong.String("bar"),
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			b := &stateBuilder{
				targetContent: tt.fields.targetContent,
				currentState:  tt.fields.currentState,
			}
			d, _ := utils.GetDefaulter(ctx, defaulterTestOpts)
			b.defaulter = d
			b.build()
			assert.Equal(tt.want, b.konnectRawState)
		})
	}
}

func Test_stateBuilder(t *testing.T) {
	assert := assert.New(t)
	type fields struct {
		targetContent *Content
		currentState  *state.KongState
	}
	tests := []struct {
		name   string
		fields fields
		want   *utils.KongRawState
	}{
		{
			name: "end to end test with all entities",
			fields: fields{
				targetContent: &Content{
					Info: &Info{
						Defaults: kongDefaults,
					},
					Services: []FService{
						{
							Service: kong.Service{
								Name: kong.String("foo-service"),
							},
							Routes: []*FRoute{
								{
									Route: kong.Route{
										Name: kong.String("foo-route1"),
									},
								},
								{
									Route: kong.Route{
										ID:   kong.String("d125e79a-297c-414b-bc00-ad3a87be6c2b"),
										Name: kong.String("foo-route2"),
									},
								},
							},
						},
						{
							Service: kong.Service{
								Name: kong.String("bar-service"),
							},
							Routes: []*FRoute{
								{
									Route: kong.Route{
										Name: kong.String("bar-route1"),
									},
								},
								{
									Route: kong.Route{
										Name: kong.String("bar-route2"),
									},
								},
							},
						},
						{
							Service: kong.Service{
								Name: kong.String("large-payload-service"),
							},
							Routes: []*FRoute{
								{
									Route: kong.Route{
										Name:              kong.String("dont-buffer-these"),
										RequestBuffering:  kong.Bool(false),
										ResponseBuffering: kong.Bool(false),
									},
								},
								{
									Route: kong.Route{
										Name:              kong.String("buffer-these"),
										RequestBuffering:  kong.Bool(true),
										ResponseBuffering: kong.Bool(true),
									},
								},
							},
						},
					},
					Upstreams: []FUpstream{
						{
							Upstream: kong.Upstream{
								Name:  kong.String("foo"),
								Slots: kong.Int(42),
							},
						},
					},
				},
				currentState: existingServiceState(),
			},
			want: &utils.KongRawState{
				Services: []*kong.Service{
					{
						ID:             kong.String("538c7f96-b164-4f1b-97bb-9f4bb472e89f"),
						Name:           kong.String("foo-service"),
						Protocol:       kong.String("http"),
						ConnectTimeout: kong.Int(60000),
						WriteTimeout:   kong.Int(60000),
						ReadTimeout:    kong.Int(60000),
					},
					{
						ID:             kong.String("dfd79b4d-7642-4b61-ba0c-9f9f0d3ba55b"),
						Name:           kong.String("bar-service"),
						Protocol:       kong.String("http"),
						ConnectTimeout: kong.Int(60000),
						WriteTimeout:   kong.Int(60000),
						ReadTimeout:    kong.Int(60000),
					},
					{
						ID:             kong.String("9e6f82e5-4e74-4e81-a79e-4bbd6fe34cdc"),
						Name:           kong.String("large-payload-service"),
						Protocol:       kong.String("http"),
						ConnectTimeout: kong.Int(60000),
						WriteTimeout:   kong.Int(60000),
						ReadTimeout:    kong.Int(60000),
					},
				},
				Routes: []*kong.Route{
					{
						ID:            kong.String("5b1484f2-5209-49d9-b43e-92ba09dd9d52"),
						Name:          kong.String("foo-route1"),
						PreserveHost:  kong.Bool(false),
						RegexPriority: kong.Int(0),
						StripPath:     kong.Bool(false),
						Protocols:     kong.StringSlice("http", "https"),
						Service: &kong.Service{
							ID:   kong.String("538c7f96-b164-4f1b-97bb-9f4bb472e89f"),
							Name: kong.String("foo-service"),
						},
					},
					{
						ID:            kong.String("d125e79a-297c-414b-bc00-ad3a87be6c2b"),
						Name:          kong.String("foo-route2"),
						PreserveHost:  kong.Bool(false),
						RegexPriority: kong.Int(0),
						StripPath:     kong.Bool(false),
						Protocols:     kong.StringSlice("http", "https"),
						Service: &kong.Service{
							ID:   kong.String("538c7f96-b164-4f1b-97bb-9f4bb472e89f"),
							Name: kong.String("foo-service"),
						},
					},
					{
						ID:            kong.String("0cc0d614-4c88-4535-841a-cbe0709b0758"),
						Name:          kong.String("bar-route1"),
						PreserveHost:  kong.Bool(false),
						RegexPriority: kong.Int(0),
						StripPath:     kong.Bool(false),
						Protocols:     kong.StringSlice("http", "https"),
						Service: &kong.Service{
							ID:   kong.String("dfd79b4d-7642-4b61-ba0c-9f9f0d3ba55b"),
							Name: kong.String("bar-service"),
						},
					},
					{
						ID:            kong.String("083f61d3-75bc-42b4-9df4-f91929e18fda"),
						Name:          kong.String("bar-route2"),
						PreserveHost:  kong.Bool(false),
						RegexPriority: kong.Int(0),
						StripPath:     kong.Bool(false),
						Protocols:     kong.StringSlice("http", "https"),
						Service: &kong.Service{
							ID:   kong.String("dfd79b4d-7642-4b61-ba0c-9f9f0d3ba55b"),
							Name: kong.String("bar-service"),
						},
					},
					{
						ID:            kong.String("ba843ee8-d63e-4c4f-be1c-ebea546d8fac"),
						Name:          kong.String("dont-buffer-these"),
						PreserveHost:  kong.Bool(false),
						RegexPriority: kong.Int(0),
						StripPath:     kong.Bool(false),
						Protocols:     kong.StringSlice("http", "https"),
						Service: &kong.Service{
							ID:   kong.String("9e6f82e5-4e74-4e81-a79e-4bbd6fe34cdc"),
							Name: kong.String("large-payload-service"),
						},
						RequestBuffering:  kong.Bool(false),
						ResponseBuffering: kong.Bool(false),
					},
					{
						ID:            kong.String("13dd1aac-04ce-4ea2-877c-5579cfa2c78e"),
						Name:          kong.String("buffer-these"),
						PreserveHost:  kong.Bool(false),
						RegexPriority: kong.Int(0),
						StripPath:     kong.Bool(false),
						Protocols:     kong.StringSlice("http", "https"),
						Service: &kong.Service{
							ID:   kong.String("9e6f82e5-4e74-4e81-a79e-4bbd6fe34cdc"),
							Name: kong.String("large-payload-service"),
						},
						RequestBuffering:  kong.Bool(true),
						ResponseBuffering: kong.Bool(true),
					},
				},
				Upstreams: []*kong.Upstream{
					{
						ID:    kong.String("1b0bafae-881b-42a7-9110-8a42ed3c903c"),
						Name:  kong.String("foo"),
						Slots: kong.Int(42),
						Healthchecks: &kong.Healthcheck{
							Active: &kong.ActiveHealthcheck{
								Concurrency: kong.Int(10),
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
					},
				},
			},
		},
		{
			name: "entities with configurable defaults",
			fields: fields{
				targetContent: &Content{
					Info: &Info{
						Defaults: KongDefaults{
							Route: &kong.Route{
								PathHandling:     kong.String("v0"),
								PreserveHost:     kong.Bool(false),
								RegexPriority:    kong.Int(0),
								StripPath:        kong.Bool(false),
								Protocols:        kong.StringSlice("http", "https"),
								RequestBuffering: kong.Bool(false),
							},
							Service: &kong.Service{
								Protocol:       kong.String("https"),
								ConnectTimeout: kong.Int(5000),
								WriteTimeout:   kong.Int(5000),
								ReadTimeout:    kong.Int(5000),
							},
							Upstream: &kong.Upstream{
								Slots: kong.Int(100),
								Healthchecks: &kong.Healthcheck{
									Active: &kong.ActiveHealthcheck{
										Concurrency: kong.Int(5),
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
							},
						},
					},
					Services: []FService{
						{
							Service: kong.Service{
								Name: kong.String("foo-service"),
							},
							Routes: []*FRoute{
								{
									Route: kong.Route{
										Name: kong.String("foo-route1"),
									},
								},
								{
									Route: kong.Route{
										ID:   kong.String("d125e79a-297c-414b-bc00-ad3a87be6c2b"),
										Name: kong.String("foo-route2"),
									},
								},
							},
						},
						{
							Service: kong.Service{
								Name: kong.String("bar-service"),
							},
							Routes: []*FRoute{
								{
									Route: kong.Route{
										Name: kong.String("bar-route1"),
									},
								},
								{
									Route: kong.Route{
										Name: kong.String("bar-route2"),
									},
								},
							},
						},
						{
							Service: kong.Service{
								Name: kong.String("large-payload-service"),
							},
							Routes: []*FRoute{
								{
									Route: kong.Route{
										Name:              kong.String("dont-buffer-these"),
										RequestBuffering:  kong.Bool(false),
										ResponseBuffering: kong.Bool(false),
									},
								},
								{
									Route: kong.Route{
										Name:              kong.String("buffer-these"),
										RequestBuffering:  kong.Bool(true),
										ResponseBuffering: kong.Bool(true),
									},
								},
							},
						},
					},
					Upstreams: []FUpstream{
						{
							Upstream: kong.Upstream{
								Name:  kong.String("foo"),
								Slots: kong.Int(42),
							},
						},
					},
				},
				currentState: existingServiceState(),
			},
			want: &utils.KongRawState{
				Services: []*kong.Service{
					{
						ID:             kong.String("538c7f96-b164-4f1b-97bb-9f4bb472e89f"),
						Name:           kong.String("foo-service"),
						Protocol:       kong.String("https"),
						ConnectTimeout: kong.Int(5000),
						WriteTimeout:   kong.Int(5000),
						ReadTimeout:    kong.Int(5000),
					},
					{
						ID:             kong.String("dfd79b4d-7642-4b61-ba0c-9f9f0d3ba55b"),
						Name:           kong.String("bar-service"),
						Protocol:       kong.String("https"),
						ConnectTimeout: kong.Int(5000),
						WriteTimeout:   kong.Int(5000),
						ReadTimeout:    kong.Int(5000),
					},
					{
						ID:             kong.String("9e6f82e5-4e74-4e81-a79e-4bbd6fe34cdc"),
						Name:           kong.String("large-payload-service"),
						Protocol:       kong.String("https"),
						ConnectTimeout: kong.Int(5000),
						WriteTimeout:   kong.Int(5000),
						ReadTimeout:    kong.Int(5000),
					},
				},
				Routes: []*kong.Route{
					{
						ID:               kong.String("5b1484f2-5209-49d9-b43e-92ba09dd9d52"),
						Name:             kong.String("foo-route1"),
						PreserveHost:     kong.Bool(false),
						RegexPriority:    kong.Int(0),
						StripPath:        kong.Bool(false),
						Protocols:        kong.StringSlice("http", "https"),
						RequestBuffering: kong.Bool(false),
						PathHandling:     kong.String("v0"),
						Service: &kong.Service{
							ID:   kong.String("538c7f96-b164-4f1b-97bb-9f4bb472e89f"),
							Name: kong.String("foo-service"),
						},
					},
					{
						ID:               kong.String("d125e79a-297c-414b-bc00-ad3a87be6c2b"),
						Name:             kong.String("foo-route2"),
						PreserveHost:     kong.Bool(false),
						RegexPriority:    kong.Int(0),
						StripPath:        kong.Bool(false),
						Protocols:        kong.StringSlice("http", "https"),
						RequestBuffering: kong.Bool(false),
						PathHandling:     kong.String("v0"),
						Service: &kong.Service{
							ID:   kong.String("538c7f96-b164-4f1b-97bb-9f4bb472e89f"),
							Name: kong.String("foo-service"),
						},
					},
					{
						ID:               kong.String("0cc0d614-4c88-4535-841a-cbe0709b0758"),
						Name:             kong.String("bar-route1"),
						PreserveHost:     kong.Bool(false),
						RegexPriority:    kong.Int(0),
						StripPath:        kong.Bool(false),
						Protocols:        kong.StringSlice("http", "https"),
						RequestBuffering: kong.Bool(false),
						PathHandling:     kong.String("v0"),
						Service: &kong.Service{
							ID:   kong.String("dfd79b4d-7642-4b61-ba0c-9f9f0d3ba55b"),
							Name: kong.String("bar-service"),
						},
					},
					{
						ID:               kong.String("083f61d3-75bc-42b4-9df4-f91929e18fda"),
						Name:             kong.String("bar-route2"),
						PreserveHost:     kong.Bool(false),
						RegexPriority:    kong.Int(0),
						StripPath:        kong.Bool(false),
						Protocols:        kong.StringSlice("http", "https"),
						RequestBuffering: kong.Bool(false),
						PathHandling:     kong.String("v0"),
						Service: &kong.Service{
							ID:   kong.String("dfd79b4d-7642-4b61-ba0c-9f9f0d3ba55b"),
							Name: kong.String("bar-service"),
						},
					},
					{
						ID:            kong.String("ba843ee8-d63e-4c4f-be1c-ebea546d8fac"),
						Name:          kong.String("dont-buffer-these"),
						PreserveHost:  kong.Bool(false),
						RegexPriority: kong.Int(0),
						StripPath:     kong.Bool(false),
						Protocols:     kong.StringSlice("http", "https"),
						PathHandling:  kong.String("v0"),
						Service: &kong.Service{
							ID:   kong.String("9e6f82e5-4e74-4e81-a79e-4bbd6fe34cdc"),
							Name: kong.String("large-payload-service"),
						},
						RequestBuffering:  kong.Bool(false),
						ResponseBuffering: kong.Bool(false),
					},
					{
						ID:            kong.String("13dd1aac-04ce-4ea2-877c-5579cfa2c78e"),
						Name:          kong.String("buffer-these"),
						PreserveHost:  kong.Bool(false),
						RegexPriority: kong.Int(0),
						StripPath:     kong.Bool(false),
						Protocols:     kong.StringSlice("http", "https"),
						PathHandling:  kong.String("v0"),
						Service: &kong.Service{
							ID:   kong.String("9e6f82e5-4e74-4e81-a79e-4bbd6fe34cdc"),
							Name: kong.String("large-payload-service"),
						},
						RequestBuffering:  kong.Bool(true),
						ResponseBuffering: kong.Bool(true),
					},
				},
				Upstreams: []*kong.Upstream{
					{
						ID:    kong.String("1b0bafae-881b-42a7-9110-8a42ed3c903c"),
						Name:  kong.String("foo"),
						Slots: kong.Int(42),
						Healthchecks: &kong.Healthcheck{
							Active: &kong.ActiveHealthcheck{
								Concurrency: kong.Int(5),
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
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			rand.Seed(42)
			b := &stateBuilder{
				targetContent: tt.fields.targetContent,
				currentState:  tt.fields.currentState,
			}
			d, _ := utils.GetDefaulter(ctx, defaulterTestOpts)
			b.defaulter = d
			b.build()
			assert.Equal(tt.want, b.rawState)
		})
	}
}

func Test_stateBuilder_fillPluginConfig(t *testing.T) {
	type fields struct {
		targetContent *Content
	}
	type args struct {
		plugin *FPlugin
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		result  FPlugin
	}{
		{
			name:    "nil arg throws an error",
			wantErr: true,
		},
		{
			name: "no _plugin_config throws an error",
			fields: fields{
				targetContent: &Content{},
			},
			args: args{
				plugin: &FPlugin{
					ConfigSource: kong.String("foo"),
				},
			},
			wantErr: true,
		},
		{
			name: "no _plugin_config throws an error",
			fields: fields{
				targetContent: &Content{
					PluginConfigs: map[string]kong.Configuration{
						"foo": {
							"k2":  "v3",
							"k3:": "v3",
						},
					},
				},
			},
			args: args{
				plugin: &FPlugin{
					ConfigSource: kong.String("foo"),
					Plugin: kong.Plugin{
						Config: kong.Configuration{
							"k1": "v1",
							"k2": "v2",
						},
					},
				},
			},
			result: FPlugin{
				ConfigSource: kong.String("foo"),
				Plugin: kong.Plugin{
					Config: kong.Configuration{
						"k1":  "v1",
						"k2":  "v2",
						"k3:": "v3",
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &stateBuilder{
				targetContent: tt.fields.targetContent,
			}
			if err := b.fillPluginConfig(tt.args.plugin); (err != nil) != tt.wantErr {
				t.Errorf("stateBuilder.fillPluginConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && !reflect.DeepEqual(tt.result, tt.args.plugin) {
				assert.Equal(t, tt.result, *tt.args.plugin)
			}
		})
	}
}

func Test_getStripPathBasedOnProtocols(t *testing.T) {
	tests := []struct {
		name              string
		route             kong.Route
		wantErr           bool
		expectedStripPath *bool
	}{
		{
			name: "true strip_path and grpc protocols",
			route: kong.Route{
				Protocols: []*string{kong.String("grpc")},
				StripPath: kong.Bool(true),
			},
			wantErr: true,
		},
		{
			name: "true strip_path and grpcs protocol",
			route: kong.Route{
				Protocols: []*string{kong.String("grpcs")},
				StripPath: kong.Bool(true),
			},
			wantErr: true,
		},
		{
			name: "no strip_path and http protocol",
			route: kong.Route{
				Protocols: []*string{kong.String("http")},
			},
			expectedStripPath: nil,
		},
		{
			name: "no strip_path and grpc protocol",
			route: kong.Route{
				Protocols: []*string{kong.String("grpc")},
			},
			expectedStripPath: kong.Bool(false),
		},
		{
			name: "no strip_path and grpcs protocol",
			route: kong.Route{
				Protocols: []*string{kong.String("grpcs")},
			},
			expectedStripPath: kong.Bool(false),
		},
		{
			name: "false strip_path and grpc protocol",
			route: kong.Route{
				Protocols: []*string{kong.String("grpc")},
				StripPath: kong.Bool(false),
			},
			expectedStripPath: kong.Bool(false),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stripPath, err := getStripPathBasedOnProtocols(tt.route)
			if (err != nil) != tt.wantErr {
				t.Errorf("getStripPathBasedOnProtocols() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && tt.expectedStripPath != nil {
				assert.Equal(t, *tt.expectedStripPath, *stripPath)
			} else {
				assert.Equal(t, tt.expectedStripPath, stripPath)
			}
		})
	}
}
