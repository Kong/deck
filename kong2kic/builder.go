package kong2kic

import (
	"github.com/kong/go-database-reconciler/pkg/file"
)

type IBuilder interface {
	buildServices(*file.Content)
	buildRoutes(*file.Content)
	buildGlobalPlugins(*file.Content)
	buildConsumers(*file.Content)
	buildConsumerGroups(*file.Content)
	buildCACertificates(*file.Content)
	buildCertificates(*file.Content)
	getContent() *KICContent
}

func getBuilder(builderType string) IBuilder {
	if builderType == KICV3GATEWAY {
		return newKICv3GatewayAPIBuilder()
	} else if builderType == KICV3INGRESS {
		return newKICv3IngressAPIBuilder()
	} else if builderType == KICV2GATEWAY {
		return newKICv2GatewayAPIBuilder()
	} else if builderType == KICV2INGRESS {
		return newKICv2IngressAPIBuilder()
	}
	return nil
}

type Director struct {
	builder IBuilder
}

func newDirector(builder IBuilder) *Director {
	return &Director{
		builder: builder,
	}
}

func (d *Director) buildManifests(content *file.Content) *KICContent {
	d.builder.buildServices(content)
	d.builder.buildRoutes(content)
	d.builder.buildGlobalPlugins(content)
	d.builder.buildConsumers(content)
	d.builder.buildConsumerGroups(content)
	d.builder.buildCACertificates(content)
	d.builder.buildCertificates(content)
	return d.builder.getContent()
}
