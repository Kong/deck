package cmd

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kong/go-database-reconciler/pkg/konnect"
	"github.com/kong/go-database-reconciler/pkg/utils"
	"github.com/kong/go-kong/kong"
)

func TestKonnectClientWithMinTimeoutDoesNotRetry(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests++
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	t.Cleanup(server.Close)

	client, err := kong.NewClient(kong.String(server.URL+"/v2/control-planes/cp/core-entities"), server.Client())
	if err != nil {
		t.Fatal(err)
	}
	client, err = konnectClientWithMinTimeout(client, &utils.KonnectConfig{}, 1)
	if err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest(http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := client.DoRAW(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if requests != 1 {
		t.Fatalf("request count = %d, want 1", requests)
	}
}

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
