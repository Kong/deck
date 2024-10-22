package validate

import (
	"errors"
	"fmt"
	"testing"

	"github.com/kong/go-database-reconciler/pkg/dump"
	"github.com/kong/go-database-reconciler/pkg/file"
	"github.com/kong/go-kong/kong"
	"github.com/stretchr/testify/assert"
)

func Test_KonnectCompatibility(t *testing.T) {
	tests := []struct {
		name       string
		content    *file.Content
		dumpConfig dump.Config
		expected   []error
	}{
		{
			name: "version invalid",
			content: &file.Content{
				FormatVersion: "2.9",
				Workspace:     "test",
				Konnect: &file.Konnect{
					RuntimeGroupName: "s",
					ControlPlaneName: "s",
				},
			},
			dumpConfig: dump.Config{},
			expected: []error{
				errors.New(errWorkspace),
				errors.New(errBadVersion),
			},
		},
		{
			name: "no konnect",
			content: &file.Content{
				FormatVersion: "3.1",
			},
			dumpConfig: dump.Config{},
			expected: []error{
				errors.New(errKonnect),
			},
		},
		{
			name: "incompatible service plugin",
			content: &file.Content{
				FormatVersion: "3.1",
				Konnect: &file.Konnect{
					RuntimeGroupName: "s",
					ControlPlaneName: "s",
				},
				Services: []file.FService{
					{Plugins: []*file.FPlugin{
						{
							Plugin: kong.Plugin{
								Name:    kong.String("oauth2"),
								Enabled: kong.Bool(true),
								Config:  kong.Configuration{"config": "config"},
							},
						},
					}},
				},
			},
			dumpConfig: dump.Config{},
			expected: []error{
				fmt.Errorf(errPluginIncompatible, "oauth2"),
			},
		},
		{
			name: "incompatible service route plugins",
			content: &file.Content{
				FormatVersion: "3.1",
				Konnect: &file.Konnect{
					RuntimeGroupName: "s",
					ControlPlaneName: "s",
				},
				Services: []file.FService{
					{Routes: []*file.FRoute{
						{
							Plugins: []*file.FPlugin{
								{
									Plugin: kong.Plugin{
										Name:    kong.String("oauth2"),
										Enabled: kong.Bool(true),
										Config:  kong.Configuration{"config": "config"},
									},
								},
								{
									Plugin: kong.Plugin{
										Name:    kong.String("key-auth-enc"),
										Enabled: kong.Bool(true),
										Config:  kong.Configuration{"config": "config"},
									},
								},
							},
						},
					}},
				},
			},
			dumpConfig: dump.Config{},
			expected: []error{
				fmt.Errorf(errPluginIncompatible, "oauth2"),
				fmt.Errorf("[%s] keys are automatically encrypted in Konnect, use the key auth plugin instead", "key-auth-enc"),
			},
		},
		{
			name: "incompatible top-level and consumer-group plugins",
			content: &file.Content{
				FormatVersion: "3.1",
				Konnect: &file.Konnect{
					RuntimeGroupName: "s",
					ControlPlaneName: "s",
				},
				Plugins: []file.FPlugin{
					{
						Plugin: kong.Plugin{
							Name:    kong.String("response-ratelimiting"),
							Enabled: kong.Bool(true),
							Config:  kong.Configuration{"strategy": "cluster"},
						},
					},
				},
				ConsumerGroups: []file.FConsumerGroupObject{
					{
						Plugins: []*kong.ConsumerGroupPlugin{
							{
								Name:   kong.String("key-auth-enc"),
								Config: kong.Configuration{"config": "config"},
							},
						},
					},
				},
			},
			dumpConfig: dump.Config{},
			expected: []error{
				fmt.Errorf(errPluginNoCluster, "response-ratelimiting"),
				fmt.Errorf("[%s] keys are automatically encrypted in Konnect, use the key auth plugin instead", "key-auth-enc"),
			},
		},
		{
			name: "no konnect info in file, but passed via cli flag",
			content: &file.Content{
				FormatVersion: "3.1",
			},
			dumpConfig: dump.Config{
				KonnectControlPlane: "default",
			},
			expected: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := KonnectCompatibility(tt.content, tt.dumpConfig)
			assert.Equal(t, tt.expected, errs)
		})
	}
}
