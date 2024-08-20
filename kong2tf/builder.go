package kong2tf

import (
	"github.com/kong/go-database-reconciler/pkg/file"
)

type ITerraformBuilder interface {
	buildServices(*file.Content, *string, bool)
	buildRoutes(*file.Content, *string, bool)
	buildGlobalPlugins(*file.Content, *string, bool)
	buildConsumers(*file.Content, *string, bool)
	buildConsumerGroups(*file.Content, *string, bool)
	buildUpstreams(*file.Content, *string, bool)
	buildCACertificates(*file.Content, *string, bool)
	buildCertificates(*file.Content, *string, bool)
	buildVaults(*file.Content, *string, bool)
	getContent() string
}

func getTerraformBuilder() ITerraformBuilder {
	return newDefaultTerraformBuilder()
}

type Director struct {
	builder ITerraformBuilder
}

func newDirector(builder ITerraformBuilder) *Director {
	return &Director{
		builder: builder,
	}
}

func (d *Director) builTerraformResources(content *file.Content, generateImportsForControlPlaneID *string, ignoreCredentialChanges bool) string {

	d.builder.buildGlobalPlugins(content, generateImportsForControlPlaneID, ignoreCredentialChanges)
	d.builder.buildServices(content, generateImportsForControlPlaneID, ignoreCredentialChanges)
	d.builder.buildUpstreams(content, generateImportsForControlPlaneID, ignoreCredentialChanges)
	d.builder.buildRoutes(content, generateImportsForControlPlaneID, ignoreCredentialChanges)
	d.builder.buildConsumers(content, generateImportsForControlPlaneID, ignoreCredentialChanges)
	d.builder.buildConsumerGroups(content, generateImportsForControlPlaneID, ignoreCredentialChanges)
	d.builder.buildCACertificates(content, generateImportsForControlPlaneID, ignoreCredentialChanges)
	d.builder.buildCertificates(content, generateImportsForControlPlaneID, ignoreCredentialChanges)
	d.builder.buildVaults(content, generateImportsForControlPlaneID, ignoreCredentialChanges)
	return d.builder.getContent()
}
