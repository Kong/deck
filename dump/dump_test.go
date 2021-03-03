package dump

import "testing"

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
