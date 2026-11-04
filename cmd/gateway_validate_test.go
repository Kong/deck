package cmd

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kong/go-database-reconciler/pkg/schema"
	"github.com/kong/go-kong/kong"
	"github.com/stretchr/testify/require"
)

func TestValidateCmd_InvalidParallelism(t *testing.T) {
	cmd := newValidateCmd(false, true)
	validateParallelism = 0
	err := cmd.PreRunE(cmd, []string{"-"})
	require.Error(t, err)
	require.EqualError(t, err, "--parallelism cannot be less than 1, got 0")
}

func TestValidateCmd_ValidParallelism(t *testing.T) {
	cmd := newValidateCmd(false, true)
	validateParallelism = 10
	err := cmd.PreRunE(cmd, []string{"-"})
	require.NoError(t, err)
}

func TestFillPluginAutoFields_ClonedPluginUsesTheSchemaOfItsRef(t *testing.T) {
	var paths []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		if r.URL.Path != "/schemas/plugins/request-transformer" {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"message":"no plugin-schema"}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"fields":[{"config":{"type":"record","fields":[]}}]}`))
	}))
	t.Cleanup(server.Close)
	client, err := kong.NewClient(kong.String(server.URL), server.Client())
	require.NoError(t, err)

	plugins := []*kong.Plugin{
		{Name: kong.String("request-transformer")},
		{Name: kong.String("example-clone")},
	}
	cloned := []*kong.ClonedPluginDefinition{
		{Name: kong.String("example-clone"), Ref: kong.String("request-transformer")},
	}

	require.NoError(t, fillPluginAutoFields(context.Background(), schema.NewRegistry(client, false), plugins, cloned))
	require.Equal(t, []string{"/schemas/plugins/request-transformer"}, paths)

	err = fillPluginAutoFields(context.Background(), schema.NewRegistry(client, false), plugins, nil)
	require.ErrorContains(t, err, "retrieving schema for plugin example-clone")
}
