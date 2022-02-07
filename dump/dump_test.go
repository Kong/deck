package dump

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/kong/deck/utils"
	"github.com/kong/go-kong/kong"
)

func Test_validateConfig(t *testing.T) {
	type args struct {
		config Config
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid config for RBAC resources",
			args: args{
				config: Config{
					RBACResourcesOnly: true,
				},
			},
			wantErr: false,
		},
		{
			name: "valid config for proxy resources",
			args: args{
				config: Config{
					SkipConsumers: true,
					SelectorTags:  []string{"foo", "bar"},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid config mixing RBAC and selector tags",
			args: args{
				config: Config{
					SelectorTags:      []string{"foo", "bar"},
					RBACResourcesOnly: true,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid config mixing RBAC and SkipConsumers",
			args: args{
				config: Config{
					SkipConsumers:     true,
					RBACResourcesOnly: true,
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateConfig(tt.args.config); (err != nil) != tt.wantErr {
				t.Errorf("validateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_dedupPluginsConfig(t *testing.T) {
	tests := []struct {
		name     string
		plugins  []*kong.Plugin
		expected map[string]utils.SharedPlugins
	}{
		{
			name: "same plugin same config",
			plugins: []*kong.Plugin{
				{
					Name: kong.String("rate-limiting"),
					Consumer: &kong.Consumer{
						ID: kong.String("c88659d5-f0a4-4ab2-b3e3-56d93ea9af6b"),
					},
					Config: kong.Configuration{
						"minute": 20,
					},
				},
				{
					Name: kong.String("rate-limiting"),
					Consumer: &kong.Consumer{
						ID: kong.String("a49672c8-84e7-4b56-a03a-ef9dccf81d0b"),
					},
					Config: kong.Configuration{
						"minute": 20,
					},
				},
			},
			expected: map[string]utils.SharedPlugins{
				"rate-limiting": {
					"rate-limiting-0": {
						Config: kong.Configuration{
							"minute": 20,
						},
						Consumers: []string{
							"c88659d5-f0a4-4ab2-b3e3-56d93ea9af6b",
							"a49672c8-84e7-4b56-a03a-ef9dccf81d0b",
						},
					},
				},
			},
		},
		{
			name: "same plugin but different configs",
			plugins: []*kong.Plugin{
				{
					Name: kong.String("rate-limiting"),
					Consumer: &kong.Consumer{
						ID: kong.String("c88659d5-f0a4-4ab2-b3e3-56d93ea9af6b"),
					},
					Config: kong.Configuration{
						"minute": 20,
					},
				},
				{
					Name: kong.String("rate-limiting"),
					Consumer: &kong.Consumer{
						ID: kong.String("a49672c8-84e7-4b56-a03a-ef9dccf81d0b"),
					},
					Config: kong.Configuration{
						"minute": 20,
					},
				},
				{
					Name: kong.String("rate-limiting"),
					Consumer: &kong.Consumer{
						ID: kong.String("9bfd9929-81a2-4eab-8212-cbf91a2b8726"),
					},
					Config: kong.Configuration{
						"minute": 30,
					},
				},
				{
					Name: kong.String("rate-limiting"),
					Consumer: &kong.Consumer{
						ID: kong.String("1d1f8ad8-85c0-4d6b-8bc9-767b334533e1"),
					},
					Config: kong.Configuration{
						"minute": 30,
					},
				},
			},
			expected: map[string]utils.SharedPlugins{
				"rate-limiting": {
					"rate-limiting-0": {
						Config: kong.Configuration{
							"minute": 20,
						},
						Consumers: []string{
							"c88659d5-f0a4-4ab2-b3e3-56d93ea9af6b",
							"a49672c8-84e7-4b56-a03a-ef9dccf81d0b",
						},
					},
					"rate-limiting-2": {
						Config: kong.Configuration{
							"minute": 30,
						},
						Consumers: []string{
							"9bfd9929-81a2-4eab-8212-cbf91a2b8726",
							"1d1f8ad8-85c0-4d6b-8bc9-767b334533e1",
						},
					},
				},
			},
		},
		{
			name: "same plugin but no duplicates",
			plugins: []*kong.Plugin{
				{
					Name: kong.String("rate-limiting"),
					Consumer: &kong.Consumer{
						ID: kong.String("c88659d5-f0a4-4ab2-b3e3-56d93ea9af6b"),
					},
					Config: kong.Configuration{
						"minute": 20,
					},
				},
				{
					Name: kong.String("rate-limiting"),
					Consumer: &kong.Consumer{
						ID: kong.String("a49672c8-84e7-4b56-a03a-ef9dccf81d0b"),
					},
					Config: kong.Configuration{
						"minute": 30,
					},
				},
			},
			expected: map[string]utils.SharedPlugins{},
		},
		{
			name: "different plugins but no duplicates",
			plugins: []*kong.Plugin{
				{
					Name: kong.String("rate-limiting"),
					Consumer: &kong.Consumer{
						ID: kong.String("c88659d5-f0a4-4ab2-b3e3-56d93ea9af6b"),
					},
					Config: kong.Configuration{
						"minute": 20,
					},
				},
				{
					Name: kong.String("prometheus"),
					Consumer: &kong.Consumer{
						ID: kong.String("a49672c8-84e7-4b56-a03a-ef9dccf81d0b"),
					},
					Config: kong.Configuration{
						"per_consumer": false,
					},
				},
			},
			expected: map[string]utils.SharedPlugins{},
		},
		{
			name: "different plugins with duplicates",
			plugins: []*kong.Plugin{
				{
					Name: kong.String("rate-limiting"),
					Consumer: &kong.Consumer{
						ID: kong.String("c88659d5-f0a4-4ab2-b3e3-56d93ea9af6b"),
					},
					Config: kong.Configuration{
						"minute": 20,
					},
				},
				{
					Name: kong.String("rate-limiting"),
					Consumer: &kong.Consumer{
						ID: kong.String("a49672c8-84e7-4b56-a03a-ef9dccf81d0b"),
					},
					Config: kong.Configuration{
						"minute": 20,
					},
				},
				{
					Name: kong.String("prometheus"),
					Consumer: &kong.Consumer{
						ID: kong.String("9bfd9929-81a2-4eab-8212-cbf91a2b8726"),
					},
					Config: kong.Configuration{
						"per_consumer": false,
					},
				},
				{
					Name: kong.String("prometheus"),
					Consumer: &kong.Consumer{
						ID: kong.String("1d1f8ad8-85c0-4d6b-8bc9-767b334533e1"),
					},
					Config: kong.Configuration{
						"per_consumer": false,
					},
				},
			},
			expected: map[string]utils.SharedPlugins{
				"rate-limiting": {
					"rate-limiting-0": {
						Config: kong.Configuration{
							"minute": 20,
						},
						Consumers: []string{
							"c88659d5-f0a4-4ab2-b3e3-56d93ea9af6b",
							"a49672c8-84e7-4b56-a03a-ef9dccf81d0b",
						},
					},
				},
				"prometheus": {
					"prometheus-2": {
						Config: kong.Configuration{
							"per_consumer": false,
						},
						Consumers: []string{
							"9bfd9929-81a2-4eab-8212-cbf91a2b8726",
							"1d1f8ad8-85c0-4d6b-8bc9-767b334533e1",
						},
					},
				},
			},
		},
		{
			name: "same plugin and config with multiple parents",
			plugins: []*kong.Plugin{
				{
					Name: kong.String("rate-limiting"),
					Consumer: &kong.Consumer{
						ID: kong.String("c88659d5-f0a4-4ab2-b3e3-56d93ea9af6b"),
					},
					Route: &kong.Route{
						ID: kong.String("ad5f9293-711d-4a17-8c2a-012817866c76"),
					},
					Service: &kong.Service{
						ID: kong.String("bb47e126-e0fb-47e4-9df2-b23bd4cb6fc8"),
					},
					Config: kong.Configuration{
						"minute": 20,
					},
				},
				{
					Name: kong.String("rate-limiting"),
					Consumer: &kong.Consumer{
						ID: kong.String("a49672c8-84e7-4b56-a03a-ef9dccf81d0b"),
					},
					Route: &kong.Route{
						ID: kong.String("ad5f9293-123d-4a17-8c2a-012817866b76"),
					},
					Service: &kong.Service{
						ID: kong.String("bb47e126-e0fb-47e4-9df2-b23bd4123456"),
					},
					Config: kong.Configuration{
						"minute": 20,
					},
				},
			},
			expected: map[string]utils.SharedPlugins{
				"rate-limiting": {
					"rate-limiting-0": {
						Config: kong.Configuration{
							"minute": 20,
						},
						Consumers: []string{
							"c88659d5-f0a4-4ab2-b3e3-56d93ea9af6b",
							"a49672c8-84e7-4b56-a03a-ef9dccf81d0b",
						},
						Routes: []string{
							"ad5f9293-711d-4a17-8c2a-012817866c76",
							"ad5f9293-123d-4a17-8c2a-012817866b76",
						},
						Services: []string{
							"bb47e126-e0fb-47e4-9df2-b23bd4cb6fc8",
							"bb47e126-e0fb-47e4-9df2-b23bd4123456",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := dedupPluginsConfig(tt.plugins)
			if diff := cmp.Diff(&results, &tt.expected); diff != "" {
				t.Errorf(diff)
			}
		})
	}
}

func Test_removeSingleEntries(t *testing.T) {
	tests := []struct {
		name          string
		sharedPlugins map[string]utils.SharedPlugins
		expected      map[string]utils.SharedPlugins
	}{
		{
			name: "shared across multiple consumers",
			sharedPlugins: map[string]utils.SharedPlugins{
				"rate-limiting": {
					"rate-limiting-0": {
						Config: kong.Configuration{
							"minute": 20,
						},
						Consumers: []string{
							"c88659d5-f0a4-4ab2-b3e3-56d93ea9af6b",
							"a49672c8-84e7-4b56-a03a-ef9dccf81d0b",
						},
					},
				},
			},
			expected: map[string]utils.SharedPlugins{
				"rate-limiting": {
					"rate-limiting-0": {
						Config: kong.Configuration{
							"minute": 20,
						},
						Consumers: []string{
							"c88659d5-f0a4-4ab2-b3e3-56d93ea9af6b",
							"a49672c8-84e7-4b56-a03a-ef9dccf81d0b",
						},
					},
				},
			},
		},
		{
			name: "shared across multiple services",
			sharedPlugins: map[string]utils.SharedPlugins{
				"rate-limiting": {
					"rate-limiting-0": {
						Config: kong.Configuration{
							"minute": 20,
						},
						Services: []string{
							"c88659d5-f0a4-4ab2-b3e3-56d93ea9af6b",
							"a49672c8-84e7-4b56-a03a-ef9dccf81d0b",
						},
					},
				},
			},
			expected: map[string]utils.SharedPlugins{
				"rate-limiting": {
					"rate-limiting-0": {
						Config: kong.Configuration{
							"minute": 20,
						},
						Services: []string{
							"c88659d5-f0a4-4ab2-b3e3-56d93ea9af6b",
							"a49672c8-84e7-4b56-a03a-ef9dccf81d0b",
						},
					},
				},
			},
		},
		{
			name: "shared across multiple routes",
			sharedPlugins: map[string]utils.SharedPlugins{
				"rate-limiting": {
					"rate-limiting-0": {
						Config: kong.Configuration{
							"minute": 20,
						},
						Routes: []string{
							"c88659d5-f0a4-4ab2-b3e3-56d93ea9af6b",
							"a49672c8-84e7-4b56-a03a-ef9dccf81d0b",
						},
					},
				},
			},
			expected: map[string]utils.SharedPlugins{
				"rate-limiting": {
					"rate-limiting-0": {
						Config: kong.Configuration{
							"minute": 20,
						},
						Routes: []string{
							"c88659d5-f0a4-4ab2-b3e3-56d93ea9af6b",
							"a49672c8-84e7-4b56-a03a-ef9dccf81d0b",
						},
					},
				},
			},
		},
		{
			name: "not shared",
			sharedPlugins: map[string]utils.SharedPlugins{
				"rate-limiting": {
					"rate-limiting-0": {
						Config: kong.Configuration{
							"minute": 20,
						},
						Routes: []string{
							"c88659d5-f0a4-4ab2-b3e3-56d93ea9af6b",
						},
					},
				},
				"prometheus": {
					"prometheus-0": {
						Config: kong.Configuration{
							"per_consumer": false,
						},
						Routes: []string{
							"a49672c8-84e7-4b56-a03a-ef9dccf81d0b",
						},
					},
				},
			},
			expected: map[string]utils.SharedPlugins{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := removeSingleEntries(tt.sharedPlugins)
			if diff := cmp.Diff(&results, &tt.expected); diff != "" {
				t.Errorf(diff)
			}
		})
	}
}
