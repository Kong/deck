package kong2tf

import (
	"github.com/kong/go-database-reconciler/pkg/file"
)

type ITerraformBuilder interface {
	buildServices(*file.Content)
	buildRoutes(*file.Content)
	buildGlobalPlugins(*file.Content)
	buildConsumers(*file.Content)
	buildConsumerGroups(*file.Content)
	buildUpstreams(*file.Content)
	buildCACertificates(*file.Content)
	buildCertificates(*file.Content)
	buildVaults(*file.Content)
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

func (d *Director) builTerraformResources(content *file.Content) string {
	d.builder.buildGlobalPlugins(content)
	d.builder.buildServices(content)
	d.builder.buildUpstreams(content)
	d.builder.buildRoutes(content)
	d.builder.buildConsumers(content)
	d.builder.buildConsumerGroups(content)
	d.builder.buildCACertificates(content)
	d.builder.buildCertificates(content)
	d.builder.buildVaults(content)
	return d.builder.getContent()
}
