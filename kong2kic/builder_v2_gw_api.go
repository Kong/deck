package kong2kic

import (
	"log"

	"github.com/kong/go-database-reconciler/pkg/file"
)

type KICv2GatewayAPIBuilder struct {
	kicContent *KICContent
}

func newKICv2GatewayAPIBuilder() *KICv2GatewayAPIBuilder {
	return &KICv2GatewayAPIBuilder{
		kicContent: &KICContent{},
	}
}

func (b *KICv2GatewayAPIBuilder) buildServices(content *file.Content) {
	err := populateKICServicesWithAnnotations(content, b.kicContent)
	if err != nil {
		log.Fatal(err)
	}
}

func (b *KICv2GatewayAPIBuilder) buildRoutes(content *file.Content) {
	err := populateKICIngressesWithAnnotations(content, b.kicContent)
	if err != nil {
		log.Fatal(err)
	}
}

func (b *KICv2GatewayAPIBuilder) buildGlobalPlugins(content *file.Content) {
	err := populateKICKongClusterPlugins(content, b.kicContent)
	if err != nil {
		log.Fatal(err)
	}
}

func (b *KICv2GatewayAPIBuilder) buildConsumers(content *file.Content) {
	err := populateKICConsumers(content, b.kicContent)
	if err != nil {
		log.Fatal(err)
	}
}

func (b *KICv2GatewayAPIBuilder) buildConsumerGroups(content *file.Content) {
	err := populateKICConsumerGroups(content, b.kicContent)
	if err != nil {
		log.Fatal(err)
	}
}

func (b *KICv2GatewayAPIBuilder) buildCACertificates(content *file.Content) {
	populateKICCACertificate(content, b.kicContent)
}

func (b *KICv2GatewayAPIBuilder) buildCertificates(content *file.Content) {
	populateKICCertificates(content, b.kicContent)
}

func (b *KICv2GatewayAPIBuilder) getContent() *KICContent {
	return b.kicContent
}
