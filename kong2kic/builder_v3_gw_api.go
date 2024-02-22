package kong2kic

import (
	"log"

	"github.com/kong/go-database-reconciler/pkg/file"
)

type KICv3GatewayAPIBuider struct {
	kicContent *KICContent
}

func newKICv3GatewayAPIBuilder() *KICv3GatewayAPIBuider {
	return &KICv3GatewayAPIBuider{
		kicContent: &KICContent{},
	}
}

func (b *KICv3GatewayAPIBuider) buildServices(content *file.Content) {
	err := populateKICServicesWithAnnotations(content, b.kicContent)
	if err != nil {
		log.Fatal(err)
	}
}

func (b *KICv3GatewayAPIBuider) buildRoutes(content *file.Content) {
	err := populateKICIngressesWithGatewayAPI(content, b.kicContent)
	if err != nil {
		log.Fatal(err)
	}
}

func (b *KICv3GatewayAPIBuider) buildGlobalPlugins(content *file.Content) {
	err := populateKICKongClusterPlugins(content, b.kicContent)
	if err != nil {
		log.Fatal(err)
	}
}

func (b *KICv3GatewayAPIBuider) buildConsumers(content *file.Content) {
	err := populateKICConsumers(content, b.kicContent)
	if err != nil {
		log.Fatal(err)
	}
}

func (b *KICv3GatewayAPIBuider) buildConsumerGroups(content *file.Content) {
	err := populateKICConsumerGroups(content, b.kicContent)
	if err != nil {
		log.Fatal(err)
	}
}

func (b *KICv3GatewayAPIBuider) buildCACertificates(content *file.Content) {
	populateKICCACertificate(content, b.kicContent)
}

func (b *KICv3GatewayAPIBuider) buildCertificates(content *file.Content) {
	populateKICCertificates(content, b.kicContent)
}

func (b *KICv3GatewayAPIBuider) getContent() *KICContent {
	return b.kicContent
}
