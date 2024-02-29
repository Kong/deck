package kong2kic

import (
	"log"

	"github.com/kong/go-database-reconciler/pkg/file"
)

type KICv2IngressAPIBuilder struct {
	kicContent *KICContent
}

func newKICv2IngressAPIBuilder() *KICv2IngressAPIBuilder {
	return &KICv2IngressAPIBuilder{
		kicContent: &KICContent{},
	}
}

func (b *KICv2IngressAPIBuilder) buildServices(content *file.Content) {
	err := populateKICServicesWithAnnotations(content, b.kicContent)
	if err != nil {
		log.Fatal(err)
	}
}

func (b *KICv2IngressAPIBuilder) buildRoutes(content *file.Content) {
	err := populateKICIngressesWithAnnotations(content, b.kicContent)
	if err != nil {
		log.Fatal(err)
	}
}

func (b *KICv2IngressAPIBuilder) buildGlobalPlugins(content *file.Content) {
	err := populateKICKongClusterPlugins(content, b.kicContent)
	if err != nil {
		log.Fatal(err)
	}
}

func (b *KICv2IngressAPIBuilder) buildConsumers(content *file.Content) {
	err := populateKICConsumers(content, b.kicContent)
	if err != nil {
		log.Fatal(err)
	}
}

func (b *KICv2IngressAPIBuilder) buildConsumerGroups(content *file.Content) {
	err := populateKICConsumerGroups(content, b.kicContent)
	if err != nil {
		log.Fatal(err)
	}
}

func (b *KICv2IngressAPIBuilder) buildCACertificates(content *file.Content) {
	populateKICCACertificate(content, b.kicContent)
}

func (b *KICv2IngressAPIBuilder) buildCertificates(content *file.Content) {
	populateKICCertificates(content, b.kicContent)
}

func (b *KICv2IngressAPIBuilder) getContent() *KICContent {
	return b.kicContent
}
