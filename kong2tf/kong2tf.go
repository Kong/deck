package kong2tf

import "github.com/kong/go-database-reconciler/pkg/file"

type Provider string

const (
	ProviderKonnect     Provider = "konnect"
	ProviderKongGateway Provider = "kong-gateway"
)

type TfConfig struct {
	ControlPlaneID string
}

func Convert(inputContent *file.Content, generateImports *string, ignoreCredentialChanges bool) (string, error) {
	return ConvertWithProvider(inputContent, generateImports, ignoreCredentialChanges, ProviderKonnect)
}

func ConvertWithProvider(
	inputContent *file.Content,
	generateImports *string,
	ignoreCredentialChanges bool,
	provider Provider,
) (string, error) {
	builder := getTerraformBuilder(provider)
	director := newDirector(builder)
	return director.builTerraformResources(inputContent, generateImports, ignoreCredentialChanges), nil
}
