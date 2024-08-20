package kong2tf

import "github.com/kong/go-database-reconciler/pkg/file"

type TfConfig struct {
	ControlPlaneID string
}

func Convert(inputContent *file.Content, generateImportsForControlPlaneID *string, ignoreCredentialChanges bool) (string, error) {
	builder := getTerraformBuilder()
	director := newDirector(builder)
	return director.builTerraformResources(inputContent, generateImportsForControlPlaneID, ignoreCredentialChanges), nil
}
