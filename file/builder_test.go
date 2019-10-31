package file

import (
	"encoding/hex"
	"math/rand"
	"os"
	"testing"

	"github.com/hbagdi/deck/state"
	"github.com/hbagdi/deck/utils"
	"github.com/hbagdi/go-kong/kong"
	"github.com/stretchr/testify/assert"
)

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
				ID: kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
			},
		},
	})
	s.BasicAuths.Add(state.BasicAuth{
		BasicAuth: kong.BasicAuth{
			ID:       kong.String("92f4c849-960b-43af-aad3-f307051408d3"),
			Username: kong.String("basic-username"),
			Password: kong.String("basic-password"),
			Consumer: &kong.Consumer{
				ID: kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
			},
		},
	})
	s.JWTAuths.Add(state.JWTAuth{
		JWTAuth: kong.JWTAuth{
			ID:     kong.String("917b9402-1be0-49d2-b482-ca4dccc2054e"),
			Key:    kong.String("jwt-key"),
			Secret: kong.String("jwt-secret"),
			Consumer: &kong.Consumer{
				ID: kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
			},
		},
	})
	s.HMACAuths.Add(state.HMACAuth{
		HMACAuth: kong.HMACAuth{
			ID:       kong.String("e5d81b73-bf9e-42b0-9d68-30a1d791b9c9"),
			Username: kong.String("hmac-username"),
			Secret:   kong.String("hmac-secret"),
			Consumer: &kong.Consumer{
				ID: kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
			},
		},
	})
	s.ACLGroups.Add(state.ACLGroup{
		ACLGroup: kong.ACLGroup{
			ID:    kong.String("b7c9352a-775a-4ba5-9869-98e926a3e6cb"),
			Group: kong.String("foo-group"),
			Consumer: &kong.Consumer{
				ID: kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
			},
		},
	})
	s.Oauth2Creds.Add(state.Oauth2Credential{
		Oauth2Credential: kong.Oauth2Credential{
			ID:       kong.String("4eef5285-3d6a-4f6b-b659-8957a940e2ca"),
			ClientID: kong.String("oauth2-clientid"),
			Name:     kong.String("oauth2-name"),
			Consumer: &kong.Consumer{
				ID: kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
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
						SelectorTags: []string{"tag1"},
					},
					Services: []Service{
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
						Port:           kong.Int(80),
						Retries:        kong.Int(5),
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
					Services: []Service{
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
						Port:           kong.Int(80),
						Retries:        kong.Int(5),
						Protocol:       kong.String("http"),
						ConnectTimeout: kong.Int(60000),
						WriteTimeout:   kong.Int(60000),
						ReadTimeout:    kong.Int(60000),
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
			}
			d, _ := utils.GetKongDefaulter()
			b.defaulter = d
			b.build()
			assert.Equal(tt.want, b.rawState)
		})
	}
}

func Test_stateBuilder_ingestRoutes(t *testing.T) {
	assert := assert.New(t)
	rand.Seed(42)
	type fields struct {
		currentState *state.KongState
	}
	type args struct {
		routes []kong.Route
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
				routes: []kong.Route{
					{
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
				routes: []kong.Route{
					{
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &stateBuilder{
				currentState: tt.fields.currentState,
			}
			b.rawState = &utils.KongRawState{}
			d, _ := utils.GetKongDefaulter()
			b.defaulter = d
			b.intermediate, _ = state.NewKongState()
			if err := b.ingestRoutes(tt.args.routes); (err != nil) != tt.wantErr {
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
			b := &stateBuilder{
				currentState: tt.fields.currentState,
			}
			b.rawState = &utils.KongRawState{}
			d, _ := utils.GetKongDefaulter()
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
		plugins []kong.Plugin
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
				plugins: []kong.Plugin{
					{
						Name: kong.String("foo"),
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
				plugins: []kong.Plugin{
					{
						Name: kong.String("foo"),
					},
					{
						Name: kong.String("bar"),
						Consumer: &kong.Consumer{
							ID: kong.String("f77ca8c7-581d-45a4-a42c-c003234228e1"),
						},
					},
					{
						Name: kong.String("foo"),
						Consumer: &kong.Consumer{
							ID: kong.String("f77ca8c7-581d-45a4-a42c-c003234228e1"),
						},
						Route: &kong.Route{
							ID: kong.String("700bc504-b2b1-4abd-bd38-cec92779659e"),
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
		name    string
		args    args
		wantCID string
		wantRID string
		wantSID string
	}{
		{
			args: args{
				plugin: &kong.Plugin{
					Name: kong.String("foo"),
				},
			},
			wantCID: "",
			wantRID: "",
			wantSID: "",
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
				},
			},
			wantCID: "cID",
			wantRID: "rID",
			wantSID: "sID",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCID, gotRID, gotSID := pluginRelations(tt.args.plugin)
			if gotCID != tt.wantCID {
				t.Errorf("pluginRelations() gotCID = %v, want %v", gotCID, tt.wantCID)
			}
			if gotRID != tt.wantRID {
				t.Errorf("pluginRelations() gotRID = %v, want %v", gotRID, tt.wantRID)
			}
			if gotSID != tt.wantSID {
				t.Errorf("pluginRelations() gotSID = %v, want %v", gotSID, tt.wantSID)
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
					Consumers: []Consumer{
						{
							Consumer: kong.Consumer{
								Username: kong.String("foo"),
							},
						},
					},
					Info: &Info{
						SelectorTags: []string{"tag1"},
					},
				},
				currentState: emptyState(),
			},
			want: &utils.KongRawState{
				Consumers: []*kong.Consumer{
					{
						ID:       kong.String("538c7f96-b164-4f1b-97bb-9f4bb472e89f"),
						Username: kong.String("foo"),
						Tags:     kong.StringSlice("tag1"),
					},
				},
			},
		},
		{
			name: "generates ID for a non-existing credential",
			fields: fields{
				targetContent: &Content{
					Consumers: []Consumer{
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
					Info: &Info{
						SelectorTags: []string{"tag1"},
					},
				},
				currentState: emptyState(),
			},
			want: &utils.KongRawState{
				Consumers: []*kong.Consumer{
					{
						ID:       kong.String("5b1484f2-5209-49d9-b43e-92ba09dd9d52"),
						Username: kong.String("foo"),
						Tags:     kong.StringSlice("tag1"),
					},
				},
				KeyAuths: []*kong.KeyAuth{
					{
						ID:  kong.String("dfd79b4d-7642-4b61-ba0c-9f9f0d3ba55b"),
						Key: kong.String("foo-key"),
						Consumer: &kong.Consumer{
							ID: kong.String("5b1484f2-5209-49d9-b43e-92ba09dd9d52"),
						},
						Tags: kong.StringSlice("tag1"),
					},
				},
				BasicAuths: []*kong.BasicAuth{
					{
						ID:       kong.String("0cc0d614-4c88-4535-841a-cbe0709b0758"),
						Username: kong.String("basic-username"),
						Password: kong.String("basic-password"),
						Consumer: &kong.Consumer{
							ID: kong.String("5b1484f2-5209-49d9-b43e-92ba09dd9d52"),
						},
						Tags: kong.StringSlice("tag1"),
					},
				},
				HMACAuths: []*kong.HMACAuth{
					{
						ID:       kong.String("083f61d3-75bc-42b4-9df4-f91929e18fda"),
						Username: kong.String("hmac-username"),
						Secret:   kong.String("hmac-secret"),
						Consumer: &kong.Consumer{
							ID: kong.String("5b1484f2-5209-49d9-b43e-92ba09dd9d52"),
						},
						Tags: kong.StringSlice("tag1"),
					},
				},
				JWTAuths: []*kong.JWTAuth{
					{
						ID:     kong.String("9e6f82e5-4e74-4e81-a79e-4bbd6fe34cdc"),
						Key:    kong.String("jwt-key"),
						Secret: kong.String("jwt-secret"),
						Consumer: &kong.Consumer{
							ID: kong.String("5b1484f2-5209-49d9-b43e-92ba09dd9d52"),
						},
						Tags: kong.StringSlice("tag1"),
					},
				},
				Oauth2Creds: []*kong.Oauth2Credential{
					{
						ID:       kong.String("ba843ee8-d63e-4c4f-be1c-ebea546d8fac"),
						ClientID: kong.String("oauth2-clientid"),
						Name:     kong.String("oauth2-name"),
						Consumer: &kong.Consumer{
							ID: kong.String("5b1484f2-5209-49d9-b43e-92ba09dd9d52"),
						},
						Tags: kong.StringSlice("tag1"),
					},
				},
				ACLGroups: []*kong.ACLGroup{
					{
						ID:    kong.String("13dd1aac-04ce-4ea2-877c-5579cfa2c78e"),
						Group: kong.String("foo-group"),
						Consumer: &kong.Consumer{
							ID: kong.String("5b1484f2-5209-49d9-b43e-92ba09dd9d52"),
						},
						Tags: kong.StringSlice("tag1"),
					},
				},
			},
		},
		{
			name: "matches ID of an existing consumer",
			fields: fields{
				targetContent: &Content{
					Consumers: []Consumer{
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
					Consumers: []Consumer{
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
						},
					},
					Info: &Info{
						SelectorTags: []string{"tag1"},
					},
				},
				currentState: existingConsumerCredState(),
			},
			want: &utils.KongRawState{
				Consumers: []*kong.Consumer{
					{
						ID:       kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
						Username: kong.String("foo"),
						Tags:     kong.StringSlice("tag1"),
					},
				},
				KeyAuths: []*kong.KeyAuth{
					{
						ID:  kong.String("5f1ef1ea-a2a5-4a1b-adbb-b0d3434013e5"),
						Key: kong.String("foo-apikey"),
						Consumer: &kong.Consumer{
							ID: kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
						},
						Tags: kong.StringSlice("tag1"),
					},
				},
				BasicAuths: []*kong.BasicAuth{
					{
						ID:       kong.String("92f4c849-960b-43af-aad3-f307051408d3"),
						Username: kong.String("basic-username"),
						Password: kong.String("basic-password"),
						Consumer: &kong.Consumer{
							ID: kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
						},
						Tags: kong.StringSlice("tag1"),
					},
				},
				HMACAuths: []*kong.HMACAuth{
					{
						ID:       kong.String("e5d81b73-bf9e-42b0-9d68-30a1d791b9c9"),
						Username: kong.String("hmac-username"),
						Secret:   kong.String("hmac-secret"),
						Consumer: &kong.Consumer{
							ID: kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
						},
						Tags: kong.StringSlice("tag1"),
					},
				},
				JWTAuths: []*kong.JWTAuth{
					{
						ID:     kong.String("917b9402-1be0-49d2-b482-ca4dccc2054e"),
						Key:    kong.String("jwt-key"),
						Secret: kong.String("jwt-secret"),
						Consumer: &kong.Consumer{
							ID: kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
						},
						Tags: kong.StringSlice("tag1"),
					},
				},
				Oauth2Creds: []*kong.Oauth2Credential{
					{
						ID:       kong.String("4eef5285-3d6a-4f6b-b659-8957a940e2ca"),
						ClientID: kong.String("oauth2-clientid"),
						Name:     kong.String("oauth2-name"),
						Consumer: &kong.Consumer{
							ID: kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
						},
						Tags: kong.StringSlice("tag1"),
					},
				},
				ACLGroups: []*kong.ACLGroup{
					{
						ID:    kong.String("b7c9352a-775a-4ba5-9869-98e926a3e6cb"),
						Group: kong.String("foo-group"),
						Consumer: &kong.Consumer{
							ID: kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
						},
						Tags: kong.StringSlice("tag1"),
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
			}
			d, _ := utils.GetKongDefaulter()
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
					Certificates: []Certificate{
						{
							Certificate: kong.Certificate{
								Cert: kong.String("foo"),
								Key:  kong.String("bar"),
							},
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
					Certificates: []Certificate{
						{
							Certificate: kong.Certificate{
								Cert: kong.String("foo"),
								Key:  kong.String("bar"),
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
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &stateBuilder{
				targetContent: tt.fields.targetContent,
				currentState:  tt.fields.currentState,
			}
			d, _ := utils.GetKongDefaulter()
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
					CACertificates: []CACertificate{
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
					CACertificates: []CACertificate{
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
			b := &stateBuilder{
				targetContent: tt.fields.targetContent,
				currentState:  tt.fields.currentState,
			}
			d, _ := utils.GetKongDefaulter()
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
						SelectorTags: []string{"tag1"},
					},
					Upstreams: []Upstream{
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
								HTTPPath:               kong.String("/"),
								HTTPSVerifyCertificate: kong.Bool(true),
								Type:                   kong.String("http"),
								Timeout:                kong.Int(1),
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
						Tags:             kong.StringSlice("tag1"),
					},
				},
			},
		},
		{
			name: "matches ID of an existing service",
			fields: fields{
				targetContent: &Content{
					Upstreams: []Upstream{
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
								HTTPPath:               kong.String("/"),
								HTTPSVerifyCertificate: kong.Bool(true),
								Type:                   kong.String("http"),
								Timeout:                kong.Int(1),
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
			}
			d, _ := utils.GetKongDefaulter()
			b.defaulter = d
			b.build()
			assert.Equal(tt.want, b.rawState)
		})
	}
}
