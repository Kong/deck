package kong2tf

import (
	"github.com/kong/go-database-reconciler/pkg/file"
)

type ITerraformBuilder interface {
	buildControlPlaneVar(*string)
	buildServices(*file.Content, *string)
	buildRoutes(*file.Content, *string)
	buildGlobalPlugins(*file.Content, *string)
	buildConsumers(*file.Content, *string, bool)
	buildConsumerGroups(*file.Content, *string)
	buildUpstreams(*file.Content, *string)
	buildCACertificates(*file.Content, *string)
	buildCertificates(*file.Content, *string)
	buildVaults(*file.Content, *string)
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

func (d *Director) builTerraformResources(
	content *file.Content,
	generateImportsForControlPlaneID *string,
	ignoreCredentialChanges bool,
) string {
	d.builder.buildControlPlaneVar(generateImportsForControlPlaneID)
	d.builder.buildGlobalPlugins(content, generateImportsForControlPlaneID)
	d.builder.buildServices(content, generateImportsForControlPlaneID)
	d.builder.buildUpstreams(content, generateImportsForControlPlaneID)
	d.builder.buildRoutes(content, generateImportsForControlPlaneID)
	d.builder.buildConsumers(content, generateImportsForControlPlaneID, ignoreCredentialChanges)
	d.builder.buildConsumerGroups(content, generateImportsForControlPlaneID)
	d.builder.buildCACertificates(content, generateImportsForControlPlaneID)
	d.builder.buildCertificates(content, generateImportsForControlPlaneID)
	d.builder.buildVaults(content, generateImportsForControlPlaneID)
	return d.builder.getContent()
}
