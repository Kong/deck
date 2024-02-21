package kong2kic

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/gosimple/slug"
	"github.com/kong/go-database-reconciler/pkg/file"
	kicv1 "github.com/kong/kubernetes-ingress-controller/v3/pkg/apis/configuration/v1"
	kicv1beta1 "github.com/kong/kubernetes-ingress-controller/v3/pkg/apis/configuration/v1beta1"
	k8scorev1 "k8s.io/api/core/v1"
	k8snetv1 "k8s.io/api/networking/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	k8sgwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

// Builder + Director design pattern to create kubernetes manifests based on:
// 1 - Kong custom resource definitions
// 2 - Kong annotations
// 3 - Kubernetes Gateway spec
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

const (
	KICV3GATEWAY             = "KICV3_GATEWAY"
	KICV3INGRESS             = "KICV3_INGRESS"
	KICV2GATEWAY             = "KICV2_GATEWAY"
	KICV2INGRESS             = "KICV2_INGRESS"
	KICAPIVersion            = "configuration.konghq.com/v1"
	KICAPIVersionV1Beta1     = "configuration.konghq.com/v1beta1"
	GatewayAPIVersionV1Beta1 = "gateway.networking.k8s.io/v1beta1"
	GatewayAPIVersionV1      = "gateway.networking.k8s.io/v1"
	KongPluginKind           = "KongPlugin"
	SecretKind               = "Secret"
	IngressKind              = "KongIngress"
	UpstreamPolicyKind       = "KongUpstreamPolicy"
	IngressClass             = "kubernetes.io/ingress.class"
)

// ClassName is set by the CLI flag --class-name
var ClassName = "kong"

// targetKICVersionAPI is KIC v3.x Gateway API by default.
// Can be overridden by CLI flags.
var targetKICVersionAPI = KICV3GATEWAY

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

type KICv3IngressAPIBuilder struct {
	kicContent *KICContent
}

func newKICv3IngressAPIBuilder() *KICv3IngressAPIBuilder {
	return &KICv3IngressAPIBuilder{
		kicContent: &KICContent{},
	}
}

func (b *KICv3IngressAPIBuilder) buildServices(content *file.Content) {
	err := populateKICServicesWithAnnotations(content, b.kicContent)
	if err != nil {
		log.Fatal(err)
	}
}

func (b *KICv3IngressAPIBuilder) buildRoutes(content *file.Content) {
	err := populateKICIngressesWithAnnotations(content, b.kicContent)
	if err != nil {
		log.Fatal(err)
	}
}

func (b *KICv3IngressAPIBuilder) buildGlobalPlugins(content *file.Content) {
	err := populateKICKongClusterPlugins(content, b.kicContent)
	if err != nil {
		log.Fatal(err)
	}
}

func (b *KICv3IngressAPIBuilder) buildConsumers(content *file.Content) {
	err := populateKICConsumers(content, b.kicContent)
	if err != nil {
		log.Fatal(err)
	}
}

func (b *KICv3IngressAPIBuilder) buildConsumerGroups(content *file.Content) {
	err := populateKICConsumerGroups(content, b.kicContent)
	if err != nil {
		log.Fatal(err)
	}
}

func (b *KICv3IngressAPIBuilder) buildCACertificates(content *file.Content) {
	populateKICCACertificate(content, b.kicContent)
}

func (b *KICv3IngressAPIBuilder) buildCertificates(content *file.Content) {
	populateKICCertificates(content, b.kicContent)
}

func (b *KICv3IngressAPIBuilder) getContent() *KICContent {
	return b.kicContent
}

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
	err := populateKICIngressesWithGatewayAPI(content, b.kicContent)
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

////////////////////
/// End of Builder + Director
////////////////////

func MarshalKongToKIC(content *file.Content, builderType string, format string) ([]byte, error) {
	targetKICVersionAPI = builderType
	kicContent := convertKongToKIC(content, builderType)
	return kicContent.marshalKICContentToFormat(format)
}

func convertKongToKIC(content *file.Content, builderType string) *KICContent {
	builder := getBuilder(builderType)
	director := newDirector(builder)
	return director.buildManifests(content)
}

// utility function to make sure that objectmeta.name is always
// compatible with kubernetes naming conventions.
func calculateSlug(input string) string {
	// Use the slug library to create a slug
	slugStr := slug.Make(input)

	// Replace underscores with dashes
	slugStr = strings.ReplaceAll(slugStr, "_", "-")

	// If the resulting string has more than 63 characters
	if len(slugStr) > 63 {
		// Calculate the sha256 sum of the string
		hash := sha256.Sum256([]byte(slugStr))

		// Truncate the slug to 53 characters
		slugStr = slugStr[:53]

		// Replace the last 10 characters with the first 10 characters of the sha256 sum
		slugStr = slugStr[:len(slugStr)-10] + fmt.Sprintf("%x", hash)[:10]
	}

	return slugStr
}

/////
// Functions valid for both custom resources and annotations based manifests
/////

func populateKICCACertificate(content *file.Content, file *KICContent) {
	// iterate content.CACertificates and copy them into k8scorev1.Secret, then add them to kicContent.Secrets
	for _, caCert := range content.CACertificates {
		digest := sha256.Sum256([]byte(*caCert.Cert))
		var (
			secret     k8scorev1.Secret
			secretName = "ca-cert-" + fmt.Sprintf("%x", digest)
		)
		secret.TypeMeta.APIVersion = "v1"
		secret.TypeMeta.Kind = SecretKind
		secret.Type = "generic"
		secret.ObjectMeta.Name = calculateSlug(secretName)
		secret.ObjectMeta.Annotations = map[string]string{IngressClass: ClassName}
		secret.StringData = make(map[string]string)
		secret.StringData["ca.crt"] = *caCert.Cert
		if caCert.CertDigest != nil {
			secret.StringData["ca.digest"] = *caCert.CertDigest
		}

		file.Secrets = append(file.Secrets, secret)
	}
}

func populateKICCertificates(content *file.Content, file *KICContent) {
	// iterate content.Certificates and copy them into k8scorev1.Secret, then add them to kicContent.Secrets
	for _, cert := range content.Certificates {
		digest := sha256.Sum256([]byte(*cert.Cert))
		var (
			secret     k8scorev1.Secret
			secretName = "cert-" + fmt.Sprintf("%x", digest)
		)
		secret.TypeMeta.APIVersion = "v1"
		secret.TypeMeta.Kind = SecretKind
		secret.Type = "kubernetes.io/tls"
		secret.ObjectMeta.Name = calculateSlug(secretName)
		secret.ObjectMeta.Annotations = map[string]string{IngressClass: ClassName}
		secret.Data = make(map[string][]byte)
		secret.Data["tls.crt"] = []byte(*cert.Cert)
		secret.Data["tls.key"] = []byte(*cert.Key)
		// what to do with SNIs?

		file.Secrets = append(file.Secrets, secret)
	}
}

func populateKICKongClusterPlugins(content *file.Content, file *KICContent) error {
	// Global Plugins map to KongClusterPlugins
	// iterate content.Plugins and copy them into kicv1.KongPlugin manifests
	// add the kicv1.KongPlugin to the KICContent.KongClusterPlugins slice
	for _, plugin := range content.Plugins {
		// skip this plugin instance if it is a kongconsumergroup plugin.
		// It is a kongconsumergroup plugin if it has a consumer_group property
		if plugin.ConsumerGroup != nil {
			continue
		}
		var kongPlugin kicv1.KongClusterPlugin
		kongPlugin.APIVersion = KICAPIVersion
		kongPlugin.Kind = "KongClusterPlugin"
		kongPlugin.ObjectMeta.Annotations = map[string]string{IngressClass: ClassName}
		if plugin.Name != nil {
			kongPlugin.PluginName = *plugin.Name
			kongPlugin.ObjectMeta.Name = calculateSlug(*plugin.Name)
		} else {
			log.Println("Global Plugin name is empty. This is not recommended." +
				"Please, provide a name for the plugin before generating Kong Ingress Controller manifests.")
		}

		// transform the plugin config from map[string]interface{} to apiextensionsv1.JSON
		var configJSON apiextensionsv1.JSON
		var err error
		configJSON.Raw, err = json.Marshal(plugin.Config)
		if err != nil {
			return err
		}
		kongPlugin.Config = configJSON
		file.KongClusterPlugins = append(file.KongClusterPlugins, kongPlugin)
	}
	return nil
}

func populateKICConsumers(content *file.Content, file *KICContent) error {
	// Iterate Kong Consumers and copy them into KongConsumer
	for i := range content.Consumers {
		consumer := content.Consumers[i]

		var kongConsumer kicv1.KongConsumer
		kongConsumer.APIVersion = KICAPIVersion
		kongConsumer.Kind = "KongConsumer"
		kongConsumer.ObjectMeta.Name = calculateSlug(*consumer.Username)
		kongConsumer.ObjectMeta.Annotations = map[string]string{IngressClass: ClassName}
		kongConsumer.Username = *consumer.Username
		if consumer.CustomID != nil {
			kongConsumer.CustomID = *consumer.CustomID
		}

		populateKICKeyAuthSecrets(&consumer, &kongConsumer, file)
		populateKICHMACSecrets(&consumer, &kongConsumer, file)
		populateKICJWTAuthSecrets(&consumer, &kongConsumer, file)
		populateKICBasicAuthSecrets(&consumer, &kongConsumer, file)
		populateKICOAuth2CredSecrets(&consumer, &kongConsumer, file)
		populateKICACLGroupSecrets(&consumer, &kongConsumer, file)
		populateKICMTLSAuthSecrets(&consumer, &kongConsumer, file)

		// for each consumer.plugin, create a KongPlugin and a plugin annotation in the kongConsumer
		// to link the plugin
		for _, plugin := range consumer.Plugins {
			var kongPlugin kicv1.KongPlugin
			kongPlugin.APIVersion = KICAPIVersion
			kongPlugin.Kind = KongPluginKind
			kongPlugin.ObjectMeta.Annotations = map[string]string{IngressClass: ClassName}
			if plugin.Name != nil {
				kongPlugin.PluginName = *plugin.Name
				kongPlugin.ObjectMeta.Name = calculateSlug(*consumer.Username + "-" + *plugin.Name)
			} else {
				log.Println("Plugin name is empty. This is not recommended." +
					"Please, provide a name for the plugin before generating Kong Ingress Controller manifests.")
			}

			// transform the plugin config from map[string]interface{} to apiextensionsv1.JSON
			var configJSON apiextensionsv1.JSON
			var err error
			configJSON.Raw, err = json.Marshal(plugin.Config)
			if err != nil {
				return err
			}
			kongPlugin.Config = configJSON
			file.KongPlugins = append(file.KongPlugins, kongPlugin)

			if kongConsumer.ObjectMeta.Annotations["konghq.com/plugins"] == "" {
				kongConsumer.ObjectMeta.Annotations["konghq.com/plugins"] = kongPlugin.ObjectMeta.Name
			} else {
				annotations := kongConsumer.ObjectMeta.Annotations["konghq.com/plugins"] + "," + kongPlugin.ObjectMeta.Name
				kongConsumer.ObjectMeta.Annotations["konghq.com/plugins"] = annotations
			}
		}

		file.KongConsumers = append(file.KongConsumers, kongConsumer)
	}

	return nil
}

func populateKICMTLSAuthSecrets(consumer *file.FConsumer, kongConsumer *kicv1.KongConsumer, file *KICContent) {
	// iterate consumer.MTLSAuths and copy them into k8scorev1.Secret, then add them to kicContent.Secrets
	for _, mtlsAuth := range consumer.MTLSAuths {
		var secret k8scorev1.Secret
		secretName := "mtls-auth-" + *consumer.Username
		secret.TypeMeta.APIVersion = "v1"
		secret.TypeMeta.Kind = SecretKind
		secret.Type = "Opaque"
		secret.ObjectMeta.Name = calculateSlug(secretName)
		secret.ObjectMeta.Annotations = map[string]string{IngressClass: ClassName}
		secret.StringData = make(map[string]string)
		if targetKICVersionAPI == KICV3GATEWAY || targetKICVersionAPI == KICV3INGRESS {
			// for KIC v3, use the konghq.com/credential label to identify the credential type
			secret.ObjectMeta.Labels = map[string]string{"konghq.com/credential": "mtls-auth"}
		} else {
			// for KIC v2, use the kongCredType field to identify the credential type
			secret.StringData["kongCredType"] = "mtls-auth"
		}

		if mtlsAuth.SubjectName != nil {
			secret.StringData["subject_name"] = *mtlsAuth.SubjectName
		}

		if mtlsAuth.ID != nil {
			secret.StringData["id"] = *mtlsAuth.ID
		}

		if mtlsAuth.CACertificate != nil && mtlsAuth.CACertificate.Cert != nil {
			secret.StringData["ca_certificate"] = *mtlsAuth.CACertificate.Cert
		}

		// add the secret name to the kongConsumer.credentials
		kongConsumer.Credentials = append(kongConsumer.Credentials, secretName)

		file.Secrets = append(file.Secrets, secret)
	}
}

func populateKICACLGroupSecrets(consumer *file.FConsumer, kongConsumer *kicv1.KongConsumer, file *KICContent) {
	// iterate consumer.ACLGroups and copy them into k8scorev1.Secret, then add them to kicContent.Secrets
	for _, aclGroup := range consumer.ACLGroups {
		var secret k8scorev1.Secret
		secretName := "acl-group-" + *consumer.Username
		secret.TypeMeta.APIVersion = "v1"
		secret.TypeMeta.Kind = SecretKind
		secret.ObjectMeta.Name = calculateSlug(secretName)
		secret.ObjectMeta.Annotations = map[string]string{IngressClass: ClassName}
		secret.StringData = make(map[string]string)

		if targetKICVersionAPI == KICV3GATEWAY || targetKICVersionAPI == KICV3INGRESS {
			// for KIC v3, use the konghq.com/credential label to identify the credential type
			secret.ObjectMeta.Labels = map[string]string{"konghq.com/credential": "acl"}
		} else {
			// for KIC v2, use the kongCredType field to identify the credential type
			secret.StringData["kongCredType"] = "acl"
		}

		if aclGroup.Group != nil {
			secret.StringData["group"] = *aclGroup.Group
		}

		// add the secret name to the kongConsumer.credentials
		kongConsumer.Credentials = append(kongConsumer.Credentials, secretName)

		file.Secrets = append(file.Secrets, secret)
	}
}

func populateKICOAuth2CredSecrets(consumer *file.FConsumer, kongConsumer *kicv1.KongConsumer, file *KICContent) {
	// iterate consumer.OAuth2Creds and copy them into k8scorev1.Secret, then add them to kicContent.Secrets
	for _, oauth2Cred := range consumer.Oauth2Creds {
		var secret k8scorev1.Secret
		secretName := "oauth2cred-" + *consumer.Username
		secret.TypeMeta.APIVersion = "v1"
		secret.TypeMeta.Kind = SecretKind
		secret.ObjectMeta.Name = calculateSlug(secretName)
		secret.ObjectMeta.Annotations = map[string]string{IngressClass: ClassName}
		secret.StringData = make(map[string]string)

		if targetKICVersionAPI == KICV3GATEWAY || targetKICVersionAPI == KICV3INGRESS {
			// for KIC v3, use the konghq.com/credential label to identify the credential type
			secret.ObjectMeta.Labels = map[string]string{"konghq.com/credential": "oauth2"}
		} else {
			// for KIC v2, use the kongCredType field to identify the credential type
			secret.StringData["kongCredType"] = "oauth2"
		}

		if oauth2Cred.Name != nil {
			secret.StringData["name"] = *oauth2Cred.Name
		}

		if oauth2Cred.ClientID != nil {
			secret.StringData["client_id"] = *oauth2Cred.ClientID
		}

		if oauth2Cred.ClientSecret != nil {
			secret.StringData["client_secret"] = *oauth2Cred.ClientSecret
		}

		if oauth2Cred.ClientType != nil {
			secret.StringData["client_type"] = *oauth2Cred.ClientType
		}

		if oauth2Cred.HashSecret != nil {
			secret.StringData["hash_secret"] = strconv.FormatBool(*oauth2Cred.HashSecret)
		}

		// add the secret name to the kongConsumer.credentials
		kongConsumer.Credentials = append(kongConsumer.Credentials, secretName)

		file.Secrets = append(file.Secrets, secret)
	}
}

func populateKICBasicAuthSecrets(consumer *file.FConsumer, kongConsumer *kicv1.KongConsumer, file *KICContent) {
	// iterate consumer.BasicAuths and copy them into k8scorev1.Secret, then add them to kicContent.Secrets
	for _, basicAuth := range consumer.BasicAuths {
		var secret k8scorev1.Secret
		secretName := "basic-auth-" + *consumer.Username
		secret.TypeMeta.APIVersion = "v1"
		secret.TypeMeta.Kind = SecretKind
		secret.ObjectMeta.Name = calculateSlug(secretName)
		secret.ObjectMeta.Annotations = map[string]string{IngressClass: ClassName}
		secret.StringData = make(map[string]string)

		if targetKICVersionAPI == KICV3GATEWAY || targetKICVersionAPI == KICV3INGRESS {
			// for KIC v3, use the konghq.com/credential label to identify the credential type
			secret.ObjectMeta.Labels = map[string]string{"konghq.com/credential": "basic-auth"}
		} else {
			// for KIC v2, use the kongCredType field to identify the credential type
			secret.StringData["kongCredType"] = "basic-auth"
		}

		if basicAuth.Username != nil {
			secret.StringData["username"] = *basicAuth.Username
		}
		if basicAuth.Password != nil {
			secret.StringData["password"] = *basicAuth.Password
		}

		// add the secret name to the kongConsumer.credentials
		kongConsumer.Credentials = append(kongConsumer.Credentials, secretName)

		file.Secrets = append(file.Secrets, secret)
	}
}

func populateKICJWTAuthSecrets(consumer *file.FConsumer, kongConsumer *kicv1.KongConsumer, file *KICContent) {
	// iterate consumer.JWTAuths and copy them into k8scorev1.Secret, then add them to kicContent.Secrets
	for _, jwtAuth := range consumer.JWTAuths {
		var secret k8scorev1.Secret
		secretName := "jwt-auth-" + *consumer.Username
		secret.TypeMeta.APIVersion = "v1"
		secret.TypeMeta.Kind = SecretKind
		secret.ObjectMeta.Name = calculateSlug(secretName)
		secret.ObjectMeta.Annotations = map[string]string{IngressClass: ClassName}
		secret.StringData = make(map[string]string)

		if targetKICVersionAPI == KICV3GATEWAY || targetKICVersionAPI == KICV3INGRESS {
			// for KIC v3, use the konghq.com/credential label to identify the credential type
			secret.ObjectMeta.Labels = map[string]string{"konghq.com/credential": "jwt"}
		} else {
			// for KIC v2, use the kongCredType field to identify the credential type
			secret.StringData["kongCredType"] = "jwt"
		}

		// only do the following assignments if not null
		if jwtAuth.Key != nil {
			secret.StringData["key"] = *jwtAuth.Key
		}

		if jwtAuth.Algorithm != nil {
			secret.StringData["algorithm"] = *jwtAuth.Algorithm
		}

		if jwtAuth.RSAPublicKey != nil {
			secret.StringData["rsa_public_key"] = *jwtAuth.RSAPublicKey
		}

		if jwtAuth.Secret != nil {
			secret.StringData["secret"] = *jwtAuth.Secret
		}

		// add the secret name to the kongConsumer.credentials
		kongConsumer.Credentials = append(kongConsumer.Credentials, secretName)

		file.Secrets = append(file.Secrets, secret)
	}
}

func populateKICHMACSecrets(consumer *file.FConsumer, kongConsumer *kicv1.KongConsumer, file *KICContent) {
	// iterate consumer.HMACAuths and copy them into k8scorev1.Secret, then add them to kicContent.Secrets
	for _, hmacAuth := range consumer.HMACAuths {
		var secret k8scorev1.Secret
		secretName := "hmac-auth-" + *consumer.Username
		secret.TypeMeta.APIVersion = "v1"
		secret.TypeMeta.Kind = SecretKind
		secret.ObjectMeta.Name = calculateSlug(secretName)
		secret.ObjectMeta.Annotations = map[string]string{IngressClass: ClassName}
		secret.StringData = make(map[string]string)

		if targetKICVersionAPI == KICV3GATEWAY || targetKICVersionAPI == KICV3INGRESS {
			// for KIC v3, use the konghq.com/credential label to identify the credential type
			secret.ObjectMeta.Labels = map[string]string{"konghq.com/credential": "hmac-auth"}
		} else {
			// for KIC v2, use the kongCredType field to identify the credential type
			secret.StringData["kongCredType"] = "hmac-auth"
		}

		if hmacAuth.Username != nil {
			secret.StringData["username"] = *hmacAuth.Username
		}

		if hmacAuth.Secret != nil {
			secret.StringData["secret"] = *hmacAuth.Secret
		}

		// add the secret name to the kongConsumer.credentials
		kongConsumer.Credentials = append(kongConsumer.Credentials, secretName)

		file.Secrets = append(file.Secrets, secret)
	}
}

func populateKICKeyAuthSecrets(consumer *file.FConsumer, kongConsumer *kicv1.KongConsumer, file *KICContent) {
	// iterate consumer.KeyAuths and copy them into k8scorev1.Secret, then add them to kicContent.Secrets
	// add the secret name to the kongConsumer.credentials
	for _, keyAuth := range consumer.KeyAuths {
		var secret k8scorev1.Secret
		secretName := "key-auth-" + *consumer.Username
		secret.TypeMeta.APIVersion = "v1"
		secret.TypeMeta.Kind = SecretKind
		secret.ObjectMeta.Name = calculateSlug(secretName)
		secret.ObjectMeta.Annotations = map[string]string{IngressClass: ClassName}
		secret.StringData = make(map[string]string)

		if targetKICVersionAPI == KICV3GATEWAY || targetKICVersionAPI == KICV3INGRESS {
			// for KIC v3, use the konghq.com/credential label to identify the credential type
			secret.ObjectMeta.Labels = map[string]string{"konghq.com/credential": "key-auth"}
		} else {
			// for KIC v2, use the kongCredType field to identify the credential type
			secret.StringData["kongCredType"] = "key-auth"
		}

		if keyAuth.Key != nil {
			secret.StringData["key"] = *keyAuth.Key
		}

		kongConsumer.Credentials = append(kongConsumer.Credentials, secretName)

		file.Secrets = append(file.Secrets, secret)

	}
}

func populateKICUpstreamPolicy(
	content *file.Content,
	service *file.FService,
	k8sservice *k8scorev1.Service,
	kicContent *KICContent,
) {
	if content.Upstreams != nil {
		var kongUpstreamPolicy kicv1beta1.KongUpstreamPolicy
		kongUpstreamPolicy.APIVersion = KICAPIVersionV1Beta1
		kongUpstreamPolicy.Kind = UpstreamPolicyKind
		if service.Name != nil {
			kongUpstreamPolicy.ObjectMeta.Name = calculateSlug(*service.Name + "-upstream")
		} else {
			log.Println("Service name is empty. This is not recommended." +
				"Please, provide a name for the service before generating Kong Ingress Controller manifests.")
		}

		k8sservice.ObjectMeta.Annotations["konghq.com/upstream-policy"] = kongUpstreamPolicy.ObjectMeta.Name

		// Find the upstream (if any) whose name matches the service host and copy the upstream
		// into kongUpstreamPolicy. Append the kongUpstreamPolicy to kicContent.KongUpstreamPolicies.
		found := false
		for _, upstream := range content.Upstreams {
			if upstream.Name != nil && strings.EqualFold(*upstream.Name, *service.Host) {
				found = true
				var threshold int
				if upstream.Healthchecks != nil && upstream.Healthchecks.Threshold != nil {
					threshold = int(*upstream.Healthchecks.Threshold)
				}
				var activeHealthyHTTPStatuses []kicv1beta1.HTTPStatus
				var activeUnhealthyHTTPStatuses []kicv1beta1.HTTPStatus
				var passiveHealthyHTTPStatuses []kicv1beta1.HTTPStatus
				var passiveUnhealthyHTTPStatuses []kicv1beta1.HTTPStatus

				if upstream.Healthchecks != nil &&
					upstream.Healthchecks.Active != nil &&
					upstream.Healthchecks.Active.Healthy != nil {
					activeHealthyHTTPStatuses = make([]kicv1beta1.HTTPStatus, len(upstream.Healthchecks.Active.Healthy.HTTPStatuses))
					for i, status := range upstream.Healthchecks.Active.Healthy.HTTPStatuses {
						activeHealthyHTTPStatuses[i] = kicv1beta1.HTTPStatus(status)
					}
				}

				if upstream.Healthchecks != nil &&
					upstream.Healthchecks.Active != nil &&
					upstream.Healthchecks.Active.Unhealthy != nil {
					activeUnhealthyHTTPStatuses = make([]kicv1beta1.HTTPStatus,
						len(upstream.Healthchecks.Active.Unhealthy.HTTPStatuses))
					for i, status := range upstream.Healthchecks.Active.Unhealthy.HTTPStatuses {
						activeUnhealthyHTTPStatuses[i] = kicv1beta1.HTTPStatus(status)
					}
				}

				if upstream.Healthchecks != nil &&
					upstream.Healthchecks.Passive != nil &&
					upstream.Healthchecks.Passive.Healthy != nil {
					passiveHealthyHTTPStatuses = make([]kicv1beta1.HTTPStatus, len(upstream.Healthchecks.Passive.Healthy.HTTPStatuses))
					for i, status := range upstream.Healthchecks.Passive.Healthy.HTTPStatuses {
						passiveHealthyHTTPStatuses[i] = kicv1beta1.HTTPStatus(status)
					}
				}

				if upstream.Healthchecks != nil &&
					upstream.Healthchecks.Passive != nil &&
					upstream.Healthchecks.Passive.Unhealthy != nil {
					passiveUnhealthyHTTPStatuses = make([]kicv1beta1.HTTPStatus,
						len(upstream.Healthchecks.Passive.Unhealthy.HTTPStatuses))
					for i, status := range upstream.Healthchecks.Passive.Unhealthy.HTTPStatuses {
						passiveUnhealthyHTTPStatuses[i] = kicv1beta1.HTTPStatus(status)
					}
				}

				// populeate kongUpstreamPolicy.Spec with the
				// non-nil attributes in upstream.
				if upstream.Algorithm != nil {
					kongUpstreamPolicy.Spec.Algorithm = upstream.Algorithm
				}
				if upstream.Slots != nil {
					kongUpstreamPolicy.Spec.Slots = upstream.Slots
				}

				if upstream.HashOn != nil && upstream.Algorithm != nil && *upstream.Algorithm == "consistent-hashing" {
					kongUpstreamPolicy.Spec.HashOn = &kicv1beta1.KongUpstreamHash{
						Input:      (*kicv1beta1.HashInput)(upstream.HashOn),
						Header:     upstream.HashOnHeader,
						Cookie:     upstream.HashOnCookie,
						CookiePath: upstream.HashOnCookiePath,
						QueryArg:   upstream.HashOnQueryArg,
						URICapture: upstream.HashOnURICapture,
					}
				}
				if upstream.HashFallback != nil && upstream.Algorithm != nil && *upstream.Algorithm == "consistent-hashing" {
					kongUpstreamPolicy.Spec.HashOnFallback = &kicv1beta1.KongUpstreamHash{
						Input:      (*kicv1beta1.HashInput)(upstream.HashFallback),
						Header:     upstream.HashFallbackHeader,
						QueryArg:   upstream.HashFallbackQueryArg,
						URICapture: upstream.HashFallbackURICapture,
					}
				}
				if upstream.Healthchecks != nil {
					kongUpstreamPolicy.Spec.Healthchecks = &kicv1beta1.KongUpstreamHealthcheck{
						Threshold: &threshold,
						Active: &kicv1beta1.KongUpstreamActiveHealthcheck{
							Type:                   upstream.Healthchecks.Active.Type,
							Concurrency:            upstream.Healthchecks.Active.Concurrency,
							HTTPPath:               upstream.Healthchecks.Active.HTTPPath,
							HTTPSSNI:               upstream.Healthchecks.Active.HTTPSSni,
							HTTPSVerifyCertificate: upstream.Healthchecks.Active.HTTPSVerifyCertificate,
							Timeout:                upstream.Healthchecks.Active.Timeout,
							Headers:                upstream.Healthchecks.Active.Headers,
							Healthy: &kicv1beta1.KongUpstreamHealthcheckHealthy{
								Interval:     upstream.Healthchecks.Active.Healthy.Interval,
								Successes:    upstream.Healthchecks.Active.Healthy.Successes,
								HTTPStatuses: activeHealthyHTTPStatuses,
							},
							Unhealthy: &kicv1beta1.KongUpstreamHealthcheckUnhealthy{
								HTTPFailures: upstream.Healthchecks.Active.Unhealthy.HTTPFailures,
								TCPFailures:  upstream.Healthchecks.Active.Unhealthy.TCPFailures,
								Timeouts:     upstream.Healthchecks.Active.Unhealthy.Timeouts,
								Interval:     upstream.Healthchecks.Active.Unhealthy.Interval,
								HTTPStatuses: activeUnhealthyHTTPStatuses,
							},
						},
						Passive: &kicv1beta1.KongUpstreamPassiveHealthcheck{
							Type: upstream.Healthchecks.Passive.Type,
							Healthy: &kicv1beta1.KongUpstreamHealthcheckHealthy{
								HTTPStatuses: passiveHealthyHTTPStatuses,
								Interval:     upstream.Healthchecks.Passive.Healthy.Interval,
								Successes:    upstream.Healthchecks.Passive.Healthy.Successes,
							},
							Unhealthy: &kicv1beta1.KongUpstreamHealthcheckUnhealthy{
								HTTPFailures: upstream.Healthchecks.Passive.Unhealthy.HTTPFailures,
								HTTPStatuses: passiveUnhealthyHTTPStatuses,
								TCPFailures:  upstream.Healthchecks.Passive.Unhealthy.TCPFailures,
								Timeouts:     upstream.Healthchecks.Passive.Unhealthy.Timeouts,
								Interval:     upstream.Healthchecks.Passive.Unhealthy.Interval,
							},
						},
					}
				}
			}
			if found {
				kicContent.KongUpstreamPolicies = append(kicContent.KongUpstreamPolicies, kongUpstreamPolicy)
			}
		}
	}
}

func populateKICUpstream(
	content *file.Content,
	service *file.FService,
	k8sservice *k8scorev1.Service,
	kicContent *KICContent,
) {
	// add Kong specific configuration to the k8s service via a KongIngress resource

	if content.Upstreams != nil {
		var kongIngress kicv1.KongIngress
		kongIngress.APIVersion = KICAPIVersion
		kongIngress.Kind = IngressKind
		if service.Name != nil {
			kongIngress.ObjectMeta.Name = calculateSlug(*service.Name + "-upstream")
		} else {
			log.Println("Service name is empty. This is not recommended." +
				"Please, provide a name for the service before generating Kong Ingress Controller manifests.")
		}
		kongIngress.ObjectMeta.Annotations = map[string]string{IngressClass: ClassName}

		// add an annotation to the k8sservice to link this kongIngress to it
		k8sservice.ObjectMeta.Annotations["konghq.com/override"] = kongIngress.ObjectMeta.Name

		// Find the upstream (if any) whose name matches the service host and copy the upstream
		// into a kicv1.KongIngress resource. Append the kicv1.KongIngress to kicContent.KongIngresses.
		found := false
		for _, upstream := range content.Upstreams {
			if upstream.Name != nil && strings.EqualFold(*upstream.Name, *service.Host) {
				found = true
				kongIngress.Upstream = &kicv1.KongIngressUpstream{
					HostHeader:             upstream.HostHeader,
					Algorithm:              upstream.Algorithm,
					Slots:                  upstream.Slots,
					Healthchecks:           upstream.Healthchecks,
					HashOn:                 upstream.HashOn,
					HashFallback:           upstream.HashFallback,
					HashOnHeader:           upstream.HashOnHeader,
					HashFallbackHeader:     upstream.HashFallbackHeader,
					HashOnCookie:           upstream.HashOnCookie,
					HashOnCookiePath:       upstream.HashOnCookiePath,
					HashOnQueryArg:         upstream.HashOnQueryArg,
					HashFallbackQueryArg:   upstream.HashFallbackQueryArg,
					HashOnURICapture:       upstream.HashOnURICapture,
					HashFallbackURICapture: upstream.HashFallbackURICapture,
				}
			}
		}
		if found {
			kicContent.KongIngresses = append(kicContent.KongIngresses, kongIngress)
		}
	}
}

func addPluginsToService(service file.FService, k8sService k8scorev1.Service, kicContent *KICContent) error {
	for _, plugin := range service.Plugins {
		var kongPlugin kicv1.KongPlugin
		kongPlugin.APIVersion = KICAPIVersion
		kongPlugin.Kind = KongPluginKind
		if plugin.Name != nil && service.Name != nil {
			kongPlugin.ObjectMeta.Name = calculateSlug(*service.Name + "-" + *plugin.Name)
		} else {
			log.Println("Service name or plugin name is empty. This is not recommended." +
				"Please, provide a name for the service and the plugin before generating Kong Ingress Controller manifests.")
		}
		kongPlugin.ObjectMeta.Annotations = map[string]string{IngressClass: ClassName}
		kongPlugin.PluginName = *plugin.Name

		var configJSON apiextensionsv1.JSON
		var err error
		configJSON.Raw, err = json.Marshal(plugin.Config)
		if err != nil {
			return err
		}
		kongPlugin.Config = configJSON

		if k8sService.ObjectMeta.Annotations["konghq.com/plugins"] == "" {
			k8sService.ObjectMeta.Annotations["konghq.com/plugins"] = kongPlugin.ObjectMeta.Name
		} else {
			annotations := k8sService.ObjectMeta.Annotations["konghq.com/plugins"] + "," + kongPlugin.ObjectMeta.Name
			k8sService.ObjectMeta.Annotations["konghq.com/plugins"] = annotations
		}

		kicContent.KongPlugins = append(kicContent.KongPlugins, kongPlugin)
	}
	return nil
}

func addPluginsToRoute(
	service file.FService,
	route *file.FRoute,
	ingress k8snetv1.Ingress,
	kicContent *KICContent,
) error {
	for _, plugin := range route.Plugins {
		var kongPlugin kicv1.KongPlugin
		kongPlugin.APIVersion = KICAPIVersion
		kongPlugin.Kind = KongPluginKind
		if plugin.Name != nil && route.Name != nil && service.Name != nil {
			kongPlugin.ObjectMeta.Name = calculateSlug(*service.Name + "-" + *route.Name + "-" + *plugin.Name)
		} else {
			log.Println("Service name, route name or plugin name is empty. This is not recommended." +
				"Please, provide a name for the service, route and the plugin before generating Kong Ingress Controller manifests.")
		}
		kongPlugin.ObjectMeta.Annotations = map[string]string{IngressClass: ClassName}
		kongPlugin.PluginName = *plugin.Name

		var configJSON apiextensionsv1.JSON
		var err error
		configJSON.Raw, err = json.Marshal(plugin.Config)
		if err != nil {
			return err
		}
		kongPlugin.Config = configJSON

		if ingress.ObjectMeta.Annotations["konghq.com/plugins"] == "" {
			ingress.ObjectMeta.Annotations["konghq.com/plugins"] = kongPlugin.ObjectMeta.Name
		} else {
			annotations := ingress.ObjectMeta.Annotations["konghq.com/plugins"] + "," + kongPlugin.ObjectMeta.Name
			ingress.ObjectMeta.Annotations["konghq.com/plugins"] = annotations
		}

		kicContent.KongPlugins = append(kicContent.KongPlugins, kongPlugin)
	}
	return nil
}

func populateKICConsumerGroups(content *file.Content, kicContent *KICContent) error {
	// iterate over the consumer groups and create a KongConsumerGroup for each one
	for _, consumerGroup := range content.ConsumerGroups {
		var kongConsumerGroup kicv1beta1.KongConsumerGroup
		kongConsumerGroup.APIVersion = "configuration.konghq.com/v1beta1"
		kongConsumerGroup.Kind = "KongConsumerGroup"
		if consumerGroup.Name != nil {
			kongConsumerGroup.ObjectMeta.Name = calculateSlug(*consumerGroup.Name)
		} else {
			log.Println("Consumer group name is empty. This is not recommended." +
				"Please, provide a name for the consumer group before generating Kong Ingress Controller manifests.")
		}
		kongConsumerGroup.ObjectMeta.Annotations = map[string]string{IngressClass: ClassName}
		kongConsumerGroup.Name = *consumerGroup.Name

		// Iterate over the consumers in consumerGroup and
		// find the KongConsumer with the same username in kicContent.KongConsumers
		// and add it to the KongConsumerGroup
		for _, consumer := range consumerGroup.Consumers {
			for idx := range kicContent.KongConsumers {
				if kicContent.KongConsumers[idx].Username == *consumer.Username {
					if kicContent.KongConsumers[idx].ConsumerGroups == nil {
						kicContent.KongConsumers[idx].ConsumerGroups = make([]string, 0)
					}
					consumerGroups := append(kicContent.KongConsumers[idx].ConsumerGroups, *consumerGroup.Name)
					kicContent.KongConsumers[idx].ConsumerGroups = consumerGroups
				}
			}
		}

		// for each consumerGroup plugin, create a KongPlugin and a plugin annotation in the kongConsumerGroup
		// to link the plugin. Consumer group plugins are "global plugins" with a consumer_group property
		for _, plugin := range content.Plugins {
			if plugin.ConsumerGroup.ID != nil && *plugin.ConsumerGroup.ID == *consumerGroup.Name {
				var kongPlugin kicv1.KongPlugin
				kongPlugin.APIVersion = KICAPIVersion
				kongPlugin.Kind = KongPluginKind
				kongPlugin.ObjectMeta.Annotations = map[string]string{IngressClass: ClassName}
				if plugin.Name != nil {
					kongPlugin.PluginName = *consumerGroup.Name + "-" + *plugin.Name
					kongPlugin.ObjectMeta.Name = calculateSlug(*consumerGroup.Name + "-" + *plugin.Name)
				} else {
					log.Println("Plugin name is empty. This is not recommended." +
						"Please, provide a name for the plugin before generating Kong Ingress Controller manifests.")
				}

				// transform the plugin config from map[string]interface{} to apiextensionsv1.JSON
				var configJSON apiextensionsv1.JSON
				var err error
				configJSON.Raw, err = json.Marshal(plugin.Config)
				if err != nil {
					return err
				}
				kongPlugin.Config = configJSON
				kicContent.KongPlugins = append(kicContent.KongPlugins, kongPlugin)

				if kongConsumerGroup.ObjectMeta.Annotations["konghq.com/plugins"] == "" {
					kongConsumerGroup.ObjectMeta.Annotations["konghq.com/plugins"] = kongPlugin.ObjectMeta.Name
				} else {
					annotations := kongConsumerGroup.ObjectMeta.Annotations["konghq.com/plugins"] + "," + kongPlugin.ObjectMeta.Name
					kongConsumerGroup.ObjectMeta.Annotations["konghq.com/plugins"] = annotations
				}
			}
		}

		kicContent.KongConsumerGroups = append(kicContent.KongConsumerGroups, kongConsumerGroup)
	}

	return nil
}

/////
// Functions for ANNOTATION based manifests
/////

func populateKICServicesWithAnnotations(content *file.Content, kicContent *KICContent) error {
	// Iterate Kong Services and create k8s Services,
	// then create KongIngress resources for Kong Service Upstream data.
	// Finally, create KongPlugin resources for each plugin
	// associated with the service.
	for i := range content.Services {
		service := content.Services[i]

		var k8sService k8scorev1.Service
		var protocol k8scorev1.Protocol

		k8sService.TypeMeta.APIVersion = "v1"
		k8sService.TypeMeta.Kind = "Service"
		if service.Name != nil {
			k8sService.ObjectMeta.Name = calculateSlug(*service.Name)
		} else {
			log.Println("Service name is empty. This is not recommended." +
				"Please, provide a name for the service before generating Kong Ingress Controller manifests.")
		}
		k8sService.ObjectMeta.Annotations = make(map[string]string)

		// default TCP unless service.Protocol is equal to k8scorev1.ProtocolUDP
		if service.Protocol != nil && k8scorev1.Protocol(strings.ToUpper(*service.Protocol)) == k8scorev1.ProtocolUDP {
			protocol = k8scorev1.ProtocolUDP
		} else {
			protocol = k8scorev1.ProtocolTCP
		}

		if service.Port != nil {
			sPort := k8scorev1.ServicePort{
				Protocol:   protocol,
				Port:       int32(*service.Port),
				TargetPort: intstr.IntOrString{IntVal: int32(*service.Port)},
			}
			k8sService.Spec.Ports = append(k8sService.Spec.Ports, sPort)
		}

		if service.Name != nil {
			k8sService.Spec.Selector = map[string]string{"app": *service.Name}
		} else {
			log.Println("Service without a name is not recommended")
		}

		// add konghq.com/read-timeout annotation if service.ReadTimeout is not nil
		if service.ReadTimeout != nil {
			k8sService.ObjectMeta.Annotations["konghq.com/read-timeout"] = strconv.Itoa(*service.ReadTimeout)
		}

		// add konghq.com/write-timeout annotation if service.WriteTimeout is not nil
		if service.WriteTimeout != nil {
			k8sService.ObjectMeta.Annotations["konghq.com/write-timeout"] = strconv.Itoa(*service.WriteTimeout)
		}

		// add konghq.com/connect-timeout annotation if service.ConnectTimeout is not nil
		if service.ConnectTimeout != nil {
			k8sService.ObjectMeta.Annotations["konghq.com/connect-timeout"] = strconv.Itoa(*service.ConnectTimeout)
		}

		// add konghq.com/protocol annotation if service.Protocol is not nil
		if service.Protocol != nil {
			k8sService.ObjectMeta.Annotations["konghq.com/protocol"] = *service.Protocol
		}

		// add konghq.com/path annotation if service.Path is not nil
		if service.Path != nil {
			k8sService.ObjectMeta.Annotations["konghq.com/path"] = *service.Path
		}

		// add konghq.com/retries annotation if service.Retries is not nil
		if service.Retries != nil {
			k8sService.ObjectMeta.Annotations["konghq.com/retries"] = strconv.Itoa(*service.Retries)
		}

		if targetKICVersionAPI == KICV3GATEWAY || targetKICVersionAPI == KICV3INGRESS {
			// Use KongUpstreamPolicy for KICv3
			populateKICUpstreamPolicy(content, &service, &k8sService, kicContent)
		} else {
			// Use KongIngress for KICv2
			populateKICUpstream(content, &service, &k8sService, kicContent)
		}

		// iterate over the plugins for this service, create a KongPlugin for each one and add an annotation to the service
		// transform the plugin config from map[string]interface{} to apiextensionsv1.JSON
		// create a plugins annotation in the k8sservice to link the plugin to it
		err := addPluginsToService(service, k8sService, kicContent)
		if err != nil {
			return err
		}

		kicContent.Services = append(kicContent.Services, k8sService)

	}
	return nil
}

func populateKICIngressesWithAnnotations(content *file.Content, kicContent *KICContent) error {
	// For each route under each service create one ingress.
	// If the route has multiple hosts, create the relevant host declarations in the ingress
	// and under each hosts create a declaration for each path in the route.
	// If the route has no hosts, create a declaration in the ingress for each path in the route.
	// Map additional fields in the route to annotations in the ingress.
	// If the route has plugins, create a KongPlugin for each one and add an annotation to the ingress
	// to link the plugin to it.
	for _, service := range content.Services {
		for _, route := range service.Routes {
			// save all ingresses we create for this route so we can then
			// assign them the plugins defined for the route
			var (
				k8sIngress           k8snetv1.Ingress
				pathTypeImplSpecific = k8snetv1.PathTypeImplementationSpecific
			)

			k8sIngress.TypeMeta.APIVersion = "networking.k8s.io/v1"
			k8sIngress.TypeMeta.Kind = "Ingress"
			if service.Name != nil && route.Name != nil {
				k8sIngress.ObjectMeta.Name = calculateSlug(*service.Name + "-" + *route.Name)
			} else {
				log.Println("Service name or route name is empty. This is not recommended." +
					"Please, provide a name for the service and the route before generating Kong Ingress Controller manifests.")
			}
			ingressClassName := ClassName
			k8sIngress.Spec.IngressClassName = &ingressClassName
			k8sIngress.ObjectMeta.Annotations = make(map[string]string)

			// add konghq.com/protocols annotation if route.Protocols is not nil
			if route.Protocols != nil {
				var protocols string
				for _, protocol := range route.Protocols {
					if protocols == "" {
						protocols = *protocol
					} else {
						protocols = protocols + "," + *protocol
					}
				}
				k8sIngress.ObjectMeta.Annotations["konghq.com/protocols"] = protocols
			}

			// add konghq.com/strip-path annotation if route.StripPath is not nil
			if route.StripPath != nil {
				k8sIngress.ObjectMeta.Annotations["konghq.com/strip-path"] = strconv.FormatBool(*route.StripPath)
			}

			// add konghq.com/preserve-host annotation if route.PreserveHost is not nil
			if route.PreserveHost != nil {
				k8sIngress.ObjectMeta.Annotations["konghq.com/preserve-host"] = strconv.FormatBool(*route.PreserveHost)
			}

			// add konghq.com/regex-priority annotation if route.RegexPriority is not nil
			if route.RegexPriority != nil {
				k8sIngress.ObjectMeta.Annotations["konghq.com/regex-priority"] = strconv.Itoa(*route.RegexPriority)
			}

			// add konghq.com/https-redirect-status-code annotation if route.HTTPSRedirectStatusCode is not nil
			if route.HTTPSRedirectStatusCode != nil {
				value := strconv.Itoa(*route.HTTPSRedirectStatusCode)
				k8sIngress.ObjectMeta.Annotations["konghq.com/https-redirect-status-code"] = value
			}

			// add konghq.com/headers.* annotation if route.Headers is not nil
			if route.Headers != nil {
				for key, value := range route.Headers {
					k8sIngress.ObjectMeta.Annotations["konghq.com/headers."+key] = strings.Join(value, ",")
				}
			}

			// add konghq.com/path-handling annotation if route.PathHandling is not nil
			if route.PathHandling != nil {
				k8sIngress.ObjectMeta.Annotations["konghq.com/path-handling"] = *route.PathHandling
			}

			// add konghq.com/snis annotation if route.SNIs is not nil
			if route.SNIs != nil {
				var snis string
				for _, sni := range route.SNIs {
					if snis == "" {
						snis = *sni
					} else {
						snis = snis + "," + *sni
					}
				}
				k8sIngress.ObjectMeta.Annotations["konghq.com/snis"] = snis
			}

			// add konghq.com/request-buffering annotation if route.RequestBuffering is not nil
			if route.RequestBuffering != nil {
				k8sIngress.ObjectMeta.Annotations["konghq.com/request-buffering"] = strconv.FormatBool(*route.RequestBuffering)
			}

			// add konghq.com/response-buffering annotation if route.ResponseBuffering is not nil
			if route.ResponseBuffering != nil {
				k8sIngress.ObjectMeta.Annotations["konghq.com/response-buffering"] = strconv.FormatBool(*route.ResponseBuffering)
			}

			// add konghq.com/methods annotation if route.Methods is not nil
			if route.Methods != nil {
				var methods string
				for _, method := range route.Methods {
					if methods == "" {
						methods = *method
					} else {
						methods = methods + "," + *method
					}
				}
				k8sIngress.ObjectMeta.Annotations["konghq.com/methods"] = methods
			}

			if len(route.Hosts) == 0 {
				// iterate route.Paths and create a k8sIngress.Spec.Rules for each one.
				// If service.Port is not nil, add it to the k8sIngress.Spec.Rules
				ingressRule := k8snetv1.IngressRule{
					IngressRuleValue: k8snetv1.IngressRuleValue{
						HTTP: &k8snetv1.HTTPIngressRuleValue{
							Paths: []k8snetv1.HTTPIngressPath{},
						},
					},
				}
				for _, path := range route.Paths {
					// if path starts with ~ then add / to the beginning of the path
					// see: https://docs.konghq.com/kubernetes-ingress-controller/latest/guides/upgrade-kong-3x/#update-ingr
					//                                             ess-regular-expression-paths-for-kong-3x-compatibility
					sCopy := *path
					if strings.HasPrefix(*path, "~") {
						sCopy = "/" + *path
					}
					if service.Port != nil {
						ingressRule.IngressRuleValue.HTTP.Paths = append(ingressRule.IngressRuleValue.HTTP.Paths,
							k8snetv1.HTTPIngressPath{
								Path:     sCopy,
								PathType: &pathTypeImplSpecific,
								Backend: k8snetv1.IngressBackend{
									Service: &k8snetv1.IngressServiceBackend{
										Name: *service.Name,
										Port: k8snetv1.ServiceBackendPort{
											Number: int32(*service.Port),
										},
									},
								},
							})
					} else {
						ingressRule.IngressRuleValue.HTTP.Paths = append(ingressRule.IngressRuleValue.HTTP.Paths,
							k8snetv1.HTTPIngressPath{
								Path:     sCopy,
								PathType: &pathTypeImplSpecific,
								Backend: k8snetv1.IngressBackend{
									Service: &k8snetv1.IngressServiceBackend{
										Name: *service.Name,
									},
								},
							})
					}
				}
				k8sIngress.Spec.Rules = append(k8sIngress.Spec.Rules, ingressRule)
			} else {
				// Iterate route.Hosts and create a k8sIngress.Spec.Rules for each one.
				// For each host, iterate route.Paths and add a k8snetv1.HTTPIngressPath to the k8sIngress.Spec.Rules.
				// If service.Port is not nil, add it to the k8sIngress.Spec.Rules
				for _, host := range route.Hosts {
					ingressRule := k8snetv1.IngressRule{
						Host: *host,
						IngressRuleValue: k8snetv1.IngressRuleValue{
							HTTP: &k8snetv1.HTTPIngressRuleValue{
								Paths: []k8snetv1.HTTPIngressPath{},
							},
						},
					}
					for _, path := range route.Paths {
						// if path starts with ~ then add / to the beginning of the path
						// see: https://docs.konghq.com/kubernetes-ingress-controller/latest/guides/upgrade-kong-3x/#update-ingr
						//                                             ess-regular-expression-paths-for-kong-3x-compatibility
						sCopy := *path
						if strings.HasPrefix(*path, "~") {
							sCopy = "/" + *path
						}
						if service.Port != nil {
							ingressRule.IngressRuleValue.HTTP.Paths = append(ingressRule.IngressRuleValue.HTTP.Paths,
								k8snetv1.HTTPIngressPath{
									Path:     sCopy,
									PathType: &pathTypeImplSpecific,
									Backend: k8snetv1.IngressBackend{
										Service: &k8snetv1.IngressServiceBackend{
											Name: *service.Name,
											Port: k8snetv1.ServiceBackendPort{
												Number: int32(*service.Port),
											},
										},
									},
								})
						} else {
							ingressRule.IngressRuleValue.HTTP.Paths = append(ingressRule.IngressRuleValue.HTTP.Paths,
								k8snetv1.HTTPIngressPath{
									Path:     sCopy,
									PathType: &pathTypeImplSpecific,
									Backend: k8snetv1.IngressBackend{
										Service: &k8snetv1.IngressServiceBackend{
											Name: *service.Name,
										},
									},
								})
						}
					}
					k8sIngress.Spec.Rules = append(k8sIngress.Spec.Rules, ingressRule)
				}
			}

			err := addPluginsToRoute(service, route, k8sIngress, kicContent)
			if err != nil {
				return err
			}
			kicContent.Ingresses = append(kicContent.Ingresses, k8sIngress)
		}
	}
	return nil
}

// ///
// Functions for GATEWAY API based manifests
// ///
func populateKICIngressesWithGatewayAPI(content *file.Content, kicContent *KICContent) error {
	for _, service := range content.Services {
		for _, route := range service.Routes {
			var httpRoute k8sgwapiv1.HTTPRoute
			httpRoute.Kind = "HTTPRoute"
			if targetKICVersionAPI == KICV3GATEWAY {
				httpRoute.APIVersion = GatewayAPIVersionV1
			} else {
				httpRoute.APIVersion = GatewayAPIVersionV1Beta1
			}
			if service.Name != nil && route.Name != nil {
				httpRoute.ObjectMeta.Name = calculateSlug(*service.Name + "-" + *route.Name)
			} else {
				log.Println("Service name or route name is empty. This is not recommended." +
					"Please, provide a name for the service and the route before generating HTTPRoute manifests.")
			}
			httpRoute.ObjectMeta.Annotations = make(map[string]string)

			// add konghq.com/preserve-host annotation if route.PreserveHost is not nil
			if route.PreserveHost != nil {
				httpRoute.ObjectMeta.Annotations["konghq.com/preserve-host"] = strconv.FormatBool(*route.PreserveHost)
			}

			// add konghq.com/strip-path annotation if route.StripPath is not nil
			if route.StripPath != nil {
				httpRoute.ObjectMeta.Annotations["konghq.com/strip-path"] = strconv.FormatBool(*route.StripPath)
			}

			// add konghq.com/https-redirect-status-code annotation if route.HTTPSRedirectStatusCode is not nil
			if route.HTTPSRedirectStatusCode != nil {
				value := strconv.Itoa(*route.HTTPSRedirectStatusCode)
				httpRoute.ObjectMeta.Annotations["konghq.com/https-redirect-status-code"] = value
			}

			// add konghq.com/regex-priority annotation if route.RegexPriority is not nil
			if route.RegexPriority != nil {
				httpRoute.ObjectMeta.Annotations["konghq.com/regex-priority"] = strconv.Itoa(*route.RegexPriority)
			}

			// add konghq.com/path-handling annotation if route.PathHandling is not nil
			if route.PathHandling != nil {
				httpRoute.ObjectMeta.Annotations["konghq.com/path-handling"] = *route.PathHandling
			}

			// add konghq.com/snis annotation if route.SNIs is not nil
			if route.SNIs != nil {
				var snis string
				for _, sni := range route.SNIs {
					if snis == "" {
						snis = *sni
					} else {
						snis = snis + "," + *sni
					}
				}
				httpRoute.ObjectMeta.Annotations["konghq.com/snis"] = snis
			}

			// add konghq.com/request-buffering annotation if route.RequestBuffering is not nil
			if route.RequestBuffering != nil {
				httpRoute.ObjectMeta.Annotations["konghq.com/request-buffering"] = strconv.FormatBool(*route.RequestBuffering)
			}

			// add konghq.com/response-buffering annotation if route.ResponseBuffering is not nil
			if route.ResponseBuffering != nil {
				httpRoute.ObjectMeta.Annotations["konghq.com/response-buffering"] = strconv.FormatBool(*route.ResponseBuffering)
			}

			// if route.hosts is not nil, add them to the httpRoute
			if route.Hosts != nil {
				for _, host := range route.Hosts {
					httpRoute.Spec.Hostnames = append(httpRoute.Spec.Hostnames, k8sgwapiv1.Hostname(*host))
				}
			}

			// add kong as the spec.parentRef.name
			httpRoute.Spec.ParentRefs = append(httpRoute.Spec.ParentRefs, k8sgwapiv1.ParentReference{
				Name: k8sgwapiv1.ObjectName(ClassName),
			})

			// add service details to HTTPBackendRef
			portNumber := k8sgwapiv1.PortNumber(*service.Port)
			backendRef := k8sgwapiv1.BackendRef{
				BackendObjectReference: k8sgwapiv1.BackendObjectReference{
					Name: k8sgwapiv1.ObjectName(*service.Name),
					Port: &portNumber,
				},
			}

			var httpHeaderMatch []k8sgwapiv1.HTTPHeaderMatch
			headerMatchExact := k8sgwapiv1.HeaderMatchExact
			headerMatchRegex := k8sgwapiv1.HeaderMatchRegularExpression
			// if route.Headers is not nil, add them to the httpHeaderMatch
			if route.Headers != nil {
				for key, values := range route.Headers {
					// if values has only one value and that value starts with
					// the special prefix ~*, the value is interpreted as a regular expression.
					if len(values) == 1 && strings.HasPrefix(values[0], "~*") {
						httpHeaderMatch = append(httpHeaderMatch, k8sgwapiv1.HTTPHeaderMatch{
							Name:  k8sgwapiv1.HTTPHeaderName(key),
							Value: values[0][2:],
							Type:  &headerMatchRegex,
						})
					} else {
						// if multiple values are present, add them as comma separated values
						// if only one value is present, add it as a single value
						var value string
						if len(values) > 1 {
							value = strings.Join(values, ",")
						} else {
							value = values[0]
						}
						httpHeaderMatch = append(httpHeaderMatch, k8sgwapiv1.HTTPHeaderMatch{
							Name:  k8sgwapiv1.HTTPHeaderName(key),
							Value: value,
							Type:  &headerMatchExact,
						})
					}
				}
			}

			// If path is not nil, then for each path, for each method, add a httpRouteRule
			// to the httpRoute
			if route.Paths != nil {
				for _, path := range route.Paths {
					var httpPathMatch k8sgwapiv1.HTTPPathMatch
					pathMatchRegex := k8sgwapiv1.PathMatchRegularExpression
					pathMatchPrefix := k8sgwapiv1.PathMatchPathPrefix

					if strings.HasPrefix(*path, "~") {
						httpPathMatch.Type = &pathMatchRegex
						regexPath := (*path)[1:]
						httpPathMatch.Value = &regexPath
					} else {
						httpPathMatch.Type = &pathMatchPrefix
						httpPathMatch.Value = path
					}

					for _, method := range route.Methods {
						httpMethod := k8sgwapiv1.HTTPMethod(*method)
						httpRoute.Spec.Rules = append(httpRoute.Spec.Rules, k8sgwapiv1.HTTPRouteRule{
							Matches: []k8sgwapiv1.HTTPRouteMatch{
								{
									Path:    &httpPathMatch,
									Method:  &httpMethod,
									Headers: httpHeaderMatch,
								},
							},
							BackendRefs: []k8sgwapiv1.HTTPBackendRef{
								{
									BackendRef: backendRef,
								},
							},
						})
					}
				}
			} else {
				// If path is nil, then for each method, add a httpRouteRule
				// to the httpRoute with headers and no path
				for _, method := range route.Methods {
					httpMethod := k8sgwapiv1.HTTPMethod(*method)
					httpRoute.Spec.Rules = append(httpRoute.Spec.Rules, k8sgwapiv1.HTTPRouteRule{
						Matches: []k8sgwapiv1.HTTPRouteMatch{
							{
								Method:  &httpMethod,
								Headers: httpHeaderMatch,
							},
						},
						BackendRefs: []k8sgwapiv1.HTTPBackendRef{
							{
								BackendRef: backendRef,
							},
						},
					})
				}
			}
			err := addPluginsToGatewayAPIRoute(service, route, httpRoute, kicContent)
			if err != nil {
				return err
			}
			kicContent.HTTPRoutes = append(kicContent.HTTPRoutes, httpRoute)
		}
	}
	return nil
}

func addPluginsToGatewayAPIRoute(
	service file.FService, route *file.FRoute, httpRoute k8sgwapiv1.HTTPRoute, kicContent *KICContent,
) error {
	for _, plugin := range route.Plugins {
		var kongPlugin kicv1.KongPlugin
		kongPlugin.APIVersion = KICAPIVersion
		kongPlugin.Kind = KongPluginKind
		if plugin.Name != nil && route.Name != nil && service.Name != nil {
			kongPlugin.ObjectMeta.Name = calculateSlug(*service.Name + "-" + *route.Name + "-" + *plugin.Name)
		} else {
			log.Println("Service name, route name or plugin name is empty. This is not recommended." +
				"Please, provide a name for the service, route and the plugin before generating Kong Ingress Controller manifests.")
		}
		kongPlugin.ObjectMeta.Annotations = map[string]string{IngressClass: ClassName}
		kongPlugin.PluginName = *plugin.Name

		var configJSON apiextensionsv1.JSON
		var err error
		configJSON.Raw, err = json.Marshal(plugin.Config)
		if err != nil {
			return err
		}
		kongPlugin.Config = configJSON

		// add plugins as extensionRef under filters for every rule
		for i := range httpRoute.Spec.Rules {
			httpRoute.Spec.Rules[i].Filters = append(httpRoute.Spec.Rules[i].Filters, k8sgwapiv1.HTTPRouteFilter{
				ExtensionRef: &k8sgwapiv1.LocalObjectReference{
					Name:  k8sgwapiv1.ObjectName(kongPlugin.ObjectMeta.Name),
					Kind:  KongPluginKind,
					Group: "configuration.konghq.com",
				},
				Type: k8sgwapiv1.HTTPRouteFilterExtensionRef,
			})
		}

		kicContent.KongPlugins = append(kicContent.KongPlugins, kongPlugin)
	}
	return nil
}
