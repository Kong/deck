package dump

import (
	"reflect"
	"testing"

	"github.com/kong/deck/konnect"
	"github.com/kong/go-kong/kong"
)

func Test_kongServiceIDs(t *testing.T) {
	type args struct {
		cpID      string
		relations []*konnect.ControlPlaneServiceRelation
	}
	tests := []struct {
		name string
		args args
		want map[string]bool
	}{
		{
			name: "returns services belonging to the same control plane",
			args: args{
				cpID: "cp1",
				relations: []*konnect.ControlPlaneServiceRelation{
					{
						ID:                   kong.String("id1"),
						ControlPlaneEntityID: kong.String("kong-svc-1"),
						ControlPlane: &konnect.ControlPlane{
							ID: kong.String("cp1"),
						},
					},
					{
						ID:                   kong.String("id2"),
						ControlPlaneEntityID: kong.String("kong-svc-2"),
						ControlPlane: &konnect.ControlPlane{
							ID: kong.String("cp1"),
						},
					},
				},
			},
			want: map[string]bool{
				"kong-svc-1": true,
				"kong-svc-2": true,
			},
		},
		{
			name: "doesn't panic if relation.ControlPlaneEntityID is nil",
			args: args{
				cpID: "cp1",
				relations: []*konnect.ControlPlaneServiceRelation{
					{
						ID: kong.String("id1"),
						ControlPlane: &konnect.ControlPlane{
							ID: kong.String("cp2"),
						},
					},
				},
			},
			want: map[string]bool{},
		},
		{
			name: "doesn't include a service belonging to a different control plane",
			args: args{
				cpID: "cp1",
				relations: []*konnect.ControlPlaneServiceRelation{
					{
						ID:                   kong.String("id1"),
						ControlPlaneEntityID: kong.String("kong-svc-1"),
						ControlPlane: &konnect.ControlPlane{
							ID: kong.String("cp2"),
						},
					},
				},
			},
			want: map[string]bool{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := kongServiceIDs(tt.args.cpID, tt.args.relations)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("kongServiceIDs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_filterNonKongPackages(t *testing.T) {
	type args struct {
		controlPlaneID string
		packages       []*konnect.ServicePackage
		relations      []*konnect.ControlPlaneServiceRelation
	}
	tests := []struct {
		name string
		args args
		want []*konnect.ServicePackage
	}{
		{
			name: "empty packages and relations returns nil",
			args: args{
				controlPlaneID: "cp1",
				packages:       []*konnect.ServicePackage{},
				relations:      []*konnect.ControlPlaneServiceRelation{},
			},
			want: nil,
		},
		{
			name: "package with no versions is returned in output",
			args: args{
				controlPlaneID: "cp1",
				packages: []*konnect.ServicePackage{
					{
						ID:   kong.String("sp-id1"),
						Name: kong.String("sp-name1"),
					},
				},
			},
			want: []*konnect.ServicePackage{
				{
					ID:   kong.String("sp-id1"),
					Name: kong.String("sp-name1"),
				},
			},
		},
		{
			name: "package with version that belong to a different control-plane is not included in output",
			args: args{
				controlPlaneID: "cp1",
				packages: []*konnect.ServicePackage{
					{
						ID:   kong.String("sp-id1"),
						Name: kong.String("sp-name1"),
						Versions: []konnect.ServiceVersion{
							{
								ID:      kong.String("sv-id1"),
								Version: kong.String("sv-v1"),
								ControlPlaneServiceRelation: &konnect.ControlPlaneServiceRelation{
									ControlPlaneEntityID: kong.String("kong-svc-1"),
								},
							},
						},
					},
				},
				relations: []*konnect.ControlPlaneServiceRelation{
					{
						ID:                   kong.String("id1"),
						ControlPlaneEntityID: kong.String("kong-svc-1"),
						ControlPlane: &konnect.ControlPlane{
							ID: kong.String("cp2"),
						},
					},
				},
			},
			want: nil,
		},
		{
			name: "package with version that belong to same control-plane is included in output",
			args: args{
				controlPlaneID: "cp1",
				packages: []*konnect.ServicePackage{
					{
						ID:   kong.String("sp-id1"),
						Name: kong.String("sp-name1"),
						Versions: []konnect.ServiceVersion{
							{
								ID:      kong.String("sv-id1"),
								Version: kong.String("sv-v1"),
								ControlPlaneServiceRelation: &konnect.ControlPlaneServiceRelation{
									ControlPlaneEntityID: kong.String("kong-svc-1"),
								},
							},
						},
					},
				},
				relations: []*konnect.ControlPlaneServiceRelation{
					{
						ID:                   kong.String("id1"),
						ControlPlaneEntityID: kong.String("kong-svc-1"),
						ControlPlane: &konnect.ControlPlane{
							ID: kong.String("cp1"),
						},
					},
				},
			},
			want: []*konnect.ServicePackage{
				{
					ID:   kong.String("sp-id1"),
					Name: kong.String("sp-name1"),
					Versions: []konnect.ServiceVersion{
						{
							ID:      kong.String("sv-id1"),
							Version: kong.String("sv-v1"),
							ControlPlaneServiceRelation: &konnect.ControlPlaneServiceRelation{
								ControlPlaneEntityID: kong.String("kong-svc-1"),
							},
						},
					},
				},
			},
		},
		{
			name: "package with versions without any implementations is not included",
			args: args{
				controlPlaneID: "cp1",
				packages: []*konnect.ServicePackage{
					{
						ID:   kong.String("sp-id1"),
						Name: kong.String("sp-name1"),
						Versions: []konnect.ServiceVersion{
							{
								ID:      kong.String("sv-id1"),
								Version: kong.String("sv-v1"),
							},
							{
								ID:      kong.String("sv-id2"),
								Version: kong.String("sv-v2"),
							},
						},
					},
				},
				relations: []*konnect.ControlPlaneServiceRelation{},
			},
			want: []*konnect.ServicePackage{
				{
					ID:   kong.String("sp-id1"),
					Name: kong.String("sp-name1"),
					Versions: []konnect.ServiceVersion{
						{
							ID:      kong.String("sv-id1"),
							Version: kong.String("sv-v1"),
						},
						{
							ID:      kong.String("sv-id2"),
							Version: kong.String("sv-v2"),
						},
					},
				},
			},
		},
		{
			name: "package with version's implementation absent from relations is not included",
			args: args{
				controlPlaneID: "cp1",
				packages: []*konnect.ServicePackage{
					{
						ID:   kong.String("sp-id1"),
						Name: kong.String("sp-name1"),
						Versions: []konnect.ServiceVersion{
							{
								ID:      kong.String("sv-id1"),
								Version: kong.String("sv-v1"),
								ControlPlaneServiceRelation: &konnect.ControlPlaneServiceRelation{
									ControlPlaneEntityID: kong.String("kong-svc-1"),
								},
							},
						},
					},
				},
				relations: []*konnect.ControlPlaneServiceRelation{
					{
						ID:                   kong.String("id1"),
						ControlPlaneEntityID: kong.String("kong-svc-42"),
						ControlPlane: &konnect.ControlPlane{
							ID: kong.String("cp1"),
						},
					},
				},
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterNonKongPackages(tt.args.controlPlaneID, tt.args.packages, tt.args.relations)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("filterNonKongPackages() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_excludeKonnectManagedPlugins(t *testing.T) {
	tests := []struct {
		name    string
		plugins []*kong.Plugin
		want    []*kong.Plugin
	}{
		{
			name: "eclude konnect tags",
			plugins: []*kong.Plugin{
				{
					Name: kong.String("rate-limiting"),
					Tags: []*string{kong.String("tag1")},
				},
				{
					Name: kong.String("basic-auth"),
					Tags: []*string{},
				},
				{
					Name: kong.String("key-auth"),
					Tags: []*string{
						kong.String("konnect-app-registration"),
						kong.String("konnect-managed-plugin"),
					},
				},
				{
					Name: kong.String("acl"),
					Tags: []*string{
						kong.String("konnect-app-registration"),
						kong.String("konnect-managed-plugin"),
					},
				},
				{
					Name: kong.String("prometheus"),
					Tags: []*string{
						kong.String("konnect-managed-plugin"),
					},
				},
			},
			want: []*kong.Plugin{
				{
					Name: kong.String("rate-limiting"),
					Tags: []*string{kong.String("tag1")},
				},
				{
					Name: kong.String("basic-auth"),
					Tags: []*string{},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := excludeKonnectManagedPlugins(tt.plugins)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("excludeKonnectManagedPlugins() = %v, want %v", got, tt.want)
			}
		})
	}
}
