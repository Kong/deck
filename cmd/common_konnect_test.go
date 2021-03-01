package cmd

import (
	"testing"

	"github.com/kong/deck/konnect"
	"github.com/kong/go-kong/kong"
)

func Test_singleOutKongCP(t *testing.T) {
	type args struct {
		controlPlanes []konnect.ControlPlane
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "returns an error when no control-plane is found",
			args: args{
				controlPlanes: []konnect.ControlPlane{
					{
						ID: kong.String("cp-id1"),
						Type: &konnect.ControlPlaneType{
							Name: kong.String("non-kong-ee"),
						},
					},
				},
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "returns an error when multiple control-planes are found",
			args: args{
				controlPlanes: []konnect.ControlPlane{
					{
						ID: kong.String("cp-id1"),
						Type: &konnect.ControlPlaneType{
							Name: kong.String("kong-ee"),
						},
					},
					{
						ID: kong.String("cp-id2"),
						Type: &konnect.ControlPlaneType{
							Name: kong.String("kong-ee"),
						},
					},
					{
						// tests for panics due to nil pointers
					},
				},
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "returns id of a single control-plane",
			args: args{
				controlPlanes: []konnect.ControlPlane{
					{
						ID: kong.String("cp-id1"),
						Type: &konnect.ControlPlaneType{
							Name: kong.String("kong-ee"),
						},
					},
				},
			},
			want:    "cp-id1",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := singleOutKongCP(tt.args.controlPlanes)
			if (err != nil) != tt.wantErr {
				t.Errorf("singleOutKongCP() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("singleOutKongCP() = %v, want %v", got, tt.want)
			}
		})
	}
}
