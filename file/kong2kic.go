package file

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	kicv1 "github.com/kong/kubernetes-ingress-controller/v2/pkg/apis/configuration/v1"
	kicv1beta1 "github.com/kong/kubernetes-ingress-controller/v2/pkg/apis/configuration/v1beta1"
	k8scorev1 "k8s.io/api/core/v1"
	k8snetv1 "k8s.io/api/networking/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// Builder + Director design pattern to create kubernetes manifests based on:
// 1 - Kong custom resource definitions
// 2 - Kong annotations
// 3 - Kubernetes Gateway spec
type IBuilder interface {
	buildServices(*Content)
	buildRoutes(*Content)
	buildGlobalPlugins(*Content)
	buildConsumers(*Content)
	buildConsumerGroups(*Content)
	buildCACertificates(*Content)
	buildCertificates(*Content)
	getContent() *KICContent
}

const (
	CUSTOMRESOURCE = "CUSTOM_RESOURCE"
	ANNOTATIONS    = "ANNOTATIONS"
	GATEWAY        = "GATEWAY"
	KICAPIVersion  = "configuration.konghq.com/v1"
	KongPluginKind = "KongPlugin"
	SecretKind     = "Secret"
	IngressKind    = "KongIngress"
)

func getBuilder(builderType string) IBuilder {
	if builderType == CUSTOMRESOURCE {
		return newCustomResourceBuilder()
	}

	if builderType == ANNOTATIONS {
		return newAnnotationsBuilder()
	}

	// if builderType == GATEWAY {
	// 	// TODO: implement gateway builder
	// }
	return nil
}

type CustomResourceBuilder struct {
	kicContent *KICContent
}

func newCustomResourceBuilder() *CustomResourceBuilder {
	return &CustomResourceBuilder{
		kicContent: &KICContent{},
	}
}

func (b *CustomResourceBuilder) buildServices(content *Content) {
	err := populateKICServicesWithCustomResources(content, b.kicContent)
	if err != nil {
		log.Fatal(err)
	}
}

func (b *CustomResourceBuilder) buildRoutes(content *Content) {
	err := populateKICIngressesWithCustomResources(content, b.kicContent)
	if err != nil {
		log.Fatal(err)
	}
}

func (b *CustomResourceBuilder) buildGlobalPlugins(content *Content) {
	err := populateKICKongClusterPlugins(content, b.kicContent)
	if err != nil {
		log.Fatal(err)
	}
}

func (b *CustomResourceBuilder) buildConsumers(content *Content) {
	err := populateKICConsumers(content, b.kicContent)
	if err != nil {
		log.Fatal(err)
	}
}

func (b *CustomResourceBuilder) buildConsumerGroups(content *Content) {
	err := populateKICConsumerGroups(content, b.kicContent)
	if err != nil {
		log.Fatal(err)
	}
}

func (b *CustomResourceBuilder) buildCACertificates(content *Content) {
	err := populateKICCACertificate(content, b.kicContent)
	if err != nil {
		log.Fatal(err)
	}
}

func (b *CustomResourceBuilder) buildCertificates(content *Content) {
	err := populateKICCertificates(content, b.kicContent)
	if err != nil {
		log.Fatal(err)
	}
}

func (b *CustomResourceBuilder) getContent() *KICContent {
	return b.kicContent
}

type AnnotationsBuilder struct {
	kicContent *KICContent
}

func newAnnotationsBuilder() *AnnotationsBuilder {
	return &AnnotationsBuilder{
		kicContent: &KICContent{},
	}
}

func (b *AnnotationsBuilder) buildServices(content *Content) {
	err := populateKICServicesWithAnnotations(content, b.kicContent)
	if err != nil {
		log.Fatal(err)
	}
}

func (b *AnnotationsBuilder) buildRoutes(content *Content) {
	err := populateKICIngressesWithAnnotations(content, b.kicContent)
	if err != nil {
		log.Fatal(err)
	}
}

func (b *AnnotationsBuilder) buildGlobalPlugins(content *Content) {
	err := populateKICKongClusterPlugins(content, b.kicContent)
	if err != nil {
		log.Fatal(err)
	}
}

func (b *AnnotationsBuilder) buildConsumers(content *Content) {
	err := populateKICConsumers(content, b.kicContent)
	if err != nil {
		log.Fatal(err)
	}
}

func (b *AnnotationsBuilder) buildConsumerGroups(content *Content) {
	err := populateKICConsumerGroups(content, b.kicContent)
	if err != nil {
		log.Fatal(err)
	}
}

func (b *AnnotationsBuilder) buildCACertificates(content *Content) {
	err := populateKICCACertificate(content, b.kicContent)
	if err != nil {
		log.Fatal(err)
	}
}

func (b *AnnotationsBuilder) buildCertificates(content *Content) {
	err := populateKICCertificates(content, b.kicContent)
	if err != nil {
		log.Fatal(err)
	}
}

func (b *AnnotationsBuilder) getContent() *KICContent {
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

func (d *Director) buildManifests(content *Content) *KICContent {
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

func MarshalKongToKICYaml(content *Content, builderType string) ([]byte, error) {
	kicContent := convertKongToKIC(content, builderType)
	return kicContent.marshalKICContentToYaml()
}

func MarshalKongToKICJson(content *Content, builderType string) ([]byte, error) {
	kicContent := convertKongToKIC(content, builderType)
	return kicContent.marshalKICContentToJSON()
}

func convertKongToKIC(content *Content, builderType string) *KICContent {
	builder := getBuilder(builderType)
	director := newDirector(builder)
	return director.buildManifests(content)
}

/////
// Functions valid for both custom resources and annotations based manifests
/////

func populateKICCACertificate(content *Content, file *KICContent) error {
	// iterate content.CACertificates and copy them into k8scorev1.Secret, then add them to kicContent.Secrets
	for _, caCert := range content.CACertificates {
		var secret k8scorev1.Secret
		digest := sha256.Sum256([]byte(*caCert.Cert))
		var secretName = "ca-cert-" + fmt.Sprintf("%x", digest)
		secret.TypeMeta.APIVersion = "v1"
		secret.TypeMeta.Kind = SecretKind
		secret.Type = "kubernetes.io/tls"
		secret.ObjectMeta.Name = strings.ToLower(secretName)
		secret.ObjectMeta.Annotations = map[string]string{"kubernetes.io/ingress.class": "kong"}
		secret.Data = make(map[string][]byte)
		secret.Data["tls.crt"] = []byte(*caCert.Cert)

		file.Secrets = append(file.Secrets, secret)
	}
	return nil
}

func populateKICCertificates(content *Content, file *KICContent) error {
	// iterate content.Certificates and copy them into k8scorev1.Secret, then add them to kicContent.Secrets
	for _, cert := range content.Certificates {
		var secret k8scorev1.Secret
		digest := sha256.Sum256([]byte(*cert.Cert))
		var secretName = "cert-" + fmt.Sprintf("%x", digest)
		secret.TypeMeta.APIVersion = "v1"
		secret.TypeMeta.Kind = SecretKind
		secret.Type = "kubernetes.io/tls"
		secret.ObjectMeta.Name = strings.ToLower(secretName)
		secret.ObjectMeta.Annotations = map[string]string{"kubernetes.io/ingress.class": "kong"}
		secret.Data = make(map[string][]byte)
		secret.Data["tls.crt"] = []byte(*cert.Cert)
		secret.Data["tls.key"] = []byte(*cert.Key)
		// what to do with SNIs?

		file.Secrets = append(file.Secrets, secret)
	}
	return nil
}

func populateKICKongClusterPlugins(content *Content, file *KICContent) error {
	// Global Plugins map to KongClusterPlugins
	// iterate content.Plugins and copy them into kicv1.KongPlugin manifests
	// add the kicv1.KongPlugin to the KICContent.KongClusterPlugins slice
	for _, plugin := range content.Plugins {
		var kongPlugin kicv1.KongClusterPlugin
		kongPlugin.APIVersion = KICAPIVersion
		kongPlugin.Kind = "KongClusterPlugin"
		if plugin.InstanceName != nil {
			kongPlugin.ObjectMeta.Name = *plugin.InstanceName
		}
		kongPlugin.ObjectMeta.Annotations = map[string]string{"kubernetes.io/ingress.class": "kong"}
		if plugin.Name != nil {
			kongPlugin.PluginName = *plugin.Name
			kongPlugin.ObjectMeta.Name = *plugin.Name
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

func populateKICConsumers(content *Content, file *KICContent) error {
	// Iterate Kong Consumers and copy them into KongConsumer
	for i := range content.Consumers {
		consumer := content.Consumers[i]

		var kongConsumer kicv1.KongConsumer
		kongConsumer.APIVersion = KICAPIVersion
		kongConsumer.Kind = "KongConsumer"
		kongConsumer.ObjectMeta.Name = *consumer.Username
		kongConsumer.ObjectMeta.Annotations = map[string]string{"kubernetes.io/ingress.class": "kong"}
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
			if plugin.InstanceName != nil {
				kongPlugin.ObjectMeta.Name = *plugin.InstanceName
			}
			kongPlugin.ObjectMeta.Annotations = map[string]string{"kubernetes.io/ingress.class": "kong"}
			if plugin.Name != nil {
				kongPlugin.PluginName = *plugin.Name
				kongPlugin.ObjectMeta.Name = *consumer.Username + "-" + *plugin.Name
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
				kongConsumer.ObjectMeta.Annotations["konghq.com/plugins"] = kongPlugin.PluginName
			} else {
				annotations := kongConsumer.ObjectMeta.Annotations["konghq.com/plugins"] + "," + kongPlugin.PluginName
				kongConsumer.ObjectMeta.Annotations["konghq.com/plugins"] = annotations
			}
		}

		file.KongConsumers = append(file.KongConsumers, kongConsumer)
	}

	return nil
}

func populateKICMTLSAuthSecrets(consumer *FConsumer, kongConsumer *kicv1.KongConsumer, file *KICContent) {
	// iterate consumer.MTLSAuths and copy them into k8scorev1.Secret, then add them to kicContent.Secrets
	for _, mtlsAuth := range consumer.MTLSAuths {
		var secret k8scorev1.Secret
		secretName := "mtls-auth-" + *consumer.Username
		secret.TypeMeta.APIVersion = "v1"
		secret.TypeMeta.Kind = SecretKind
		secret.Type = "Opaque"
		secret.ObjectMeta.Name = strings.ToLower(secretName)
		secret.ObjectMeta.Annotations = map[string]string{"kubernetes.io/ingress.class": "kong"}
		secret.StringData = make(map[string]string)
		secret.StringData["kongCredType"] = "mtls-auth"

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

func populateKICACLGroupSecrets(consumer *FConsumer, kongConsumer *kicv1.KongConsumer, file *KICContent) {
	// iterate consumer.ACLGroups and copy them into k8scorev1.Secret, then add them to kicContent.Secrets
	for _, aclGroup := range consumer.ACLGroups {
		var secret k8scorev1.Secret
		secretName := "acl-group-" + *consumer.Username
		secret.TypeMeta.APIVersion = "v1"
		secret.TypeMeta.Kind = SecretKind
		secret.ObjectMeta.Name = strings.ToLower(secretName)
		secret.ObjectMeta.Annotations = map[string]string{"kubernetes.io/ingress.class": "kong"}
		secret.StringData = make(map[string]string)

		secret.StringData["kongCredType"] = "acl"
		if aclGroup.Group != nil {
			secret.StringData["group"] = *aclGroup.Group
		}

		// add the secret name to the kongConsumer.credentials
		kongConsumer.Credentials = append(kongConsumer.Credentials, secretName)

		file.Secrets = append(file.Secrets, secret)
	}
}

func populateKICOAuth2CredSecrets(consumer *FConsumer, kongConsumer *kicv1.KongConsumer, file *KICContent) {
	// iterate consumer.OAuth2Creds and copy them into k8scorev1.Secret, then add them to kicContent.Secrets
	for _, oauth2Cred := range consumer.Oauth2Creds {
		var secret k8scorev1.Secret
		secretName := "oauth2cred-" + *consumer.Username
		secret.TypeMeta.APIVersion = "v1"
		secret.TypeMeta.Kind = SecretKind
		secret.ObjectMeta.Name = strings.ToLower(secretName)
		secret.ObjectMeta.Annotations = map[string]string{"kubernetes.io/ingress.class": "kong"}
		secret.StringData = make(map[string]string)
		secret.StringData["kongCredType"] = "oauth2"

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

func populateKICBasicAuthSecrets(consumer *FConsumer, kongConsumer *kicv1.KongConsumer, file *KICContent) {
	// iterate consumer.BasicAuths and copy them into k8scorev1.Secret, then add them to kicContent.Secrets
	for _, basicAuth := range consumer.BasicAuths {
		var secret k8scorev1.Secret
		secretName := "basic-auth-" + *consumer.Username
		secret.TypeMeta.APIVersion = "v1"
		secret.TypeMeta.Kind = SecretKind
		secret.ObjectMeta.Name = strings.ToLower(secretName)
		secret.ObjectMeta.Annotations = map[string]string{"kubernetes.io/ingress.class": "kong"}
		secret.StringData = make(map[string]string)
		secret.StringData["kongCredType"] = "basic-auth"

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

func populateKICJWTAuthSecrets(consumer *FConsumer, kongConsumer *kicv1.KongConsumer, file *KICContent) {
	// iterate consumer.JWTAuths and copy them into k8scorev1.Secret, then add them to kicContent.Secrets
	for _, jwtAuth := range consumer.JWTAuths {
		var secret k8scorev1.Secret
		secretName := "jwt-auth-" + *consumer.Username
		secret.TypeMeta.APIVersion = "v1"
		secret.TypeMeta.Kind = SecretKind
		secret.ObjectMeta.Name = strings.ToLower(secretName)
		secret.ObjectMeta.Annotations = map[string]string{"kubernetes.io/ingress.class": "kong"}
		secret.StringData = make(map[string]string)
		secret.StringData["kongCredType"] = "jwt"

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

func populateKICHMACSecrets(consumer *FConsumer, kongConsumer *kicv1.KongConsumer, file *KICContent) {
	// iterate consumer.HMACAuths and copy them into k8scorev1.Secret, then add them to kicContent.Secrets
	for _, hmacAuth := range consumer.HMACAuths {
		var secret k8scorev1.Secret
		secretName := "hmac-auth-" + *consumer.Username
		secret.TypeMeta.APIVersion = "v1"
		secret.TypeMeta.Kind = SecretKind
		secret.ObjectMeta.Name = strings.ToLower(secretName)
		secret.ObjectMeta.Annotations = map[string]string{"kubernetes.io/ingress.class": "kong"}
		secret.StringData = make(map[string]string)
		secret.StringData["kongCredType"] = "hmac-auth"

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

func populateKICKeyAuthSecrets(consumer *FConsumer, kongConsumer *kicv1.KongConsumer, file *KICContent) {
	// iterate consumer.KeyAuths and copy them into k8scorev1.Secret, then add them to kicContent.Secrets
	// add the secret name to the kongConsumer.credentials
	for _, keyAuth := range consumer.KeyAuths {
		var secret k8scorev1.Secret
		secretName := "key-auth-" + *consumer.Username
		secret.TypeMeta.APIVersion = "v1"
		secret.TypeMeta.Kind = SecretKind
		secret.ObjectMeta.Name = strings.ToLower(secretName)
		secret.ObjectMeta.Annotations = map[string]string{"kubernetes.io/ingress.class": "kong"}
		secret.StringData = make(map[string]string)
		secret.StringData["kongCredType"] = "key-auth"

		if keyAuth.Key != nil {
			secret.StringData["key"] = *keyAuth.Key
		}

		kongConsumer.Credentials = append(kongConsumer.Credentials, secretName)

		file.Secrets = append(file.Secrets, secret)

	}
}

func populateKICUpstream(content *Content, service *FService, k8sservice *k8scorev1.Service, kicContent *KICContent) {
	// add Kong specific configuration to the k8s service via a KongIngress resource

	if content.Upstreams != nil {
		var kongIngress kicv1.KongIngress
		kongIngress.APIVersion = KICAPIVersion
		kongIngress.Kind = IngressKind
		if service.Name != nil {
			kongIngress.ObjectMeta.Name = *service.Name + "-upstream"
		}
		kongIngress.ObjectMeta.Annotations = map[string]string{"kubernetes.io/ingress.class": "kong"}

		// add an annotation to the k8sservice to link this kongIngress to it
		k8sservice.ObjectMeta.Annotations["konghq.com/override"] = kongIngress.ObjectMeta.Name

		// Find the upstream (if any) whose name matches the service host and copy the upstream
		// into a kicv1.KongIngress resource. Append the kicv1.KongIngress to kicContent.KongIngresses.
		for _, upstream := range content.Upstreams {
			if upstream.Name != nil && strings.EqualFold(*upstream.Name, *service.Host) {
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
		kicContent.KongIngresses = append(kicContent.KongIngresses, kongIngress)
	}
}

func addPluginsToService(service FService, k8sService k8scorev1.Service, kicContent *KICContent) error {
	for _, plugin := range service.Plugins {
		var kongPlugin kicv1.KongPlugin
		kongPlugin.APIVersion = KICAPIVersion
		kongPlugin.Kind = KongPluginKind
		if plugin.Name != nil && service.Name != nil {
			kongPlugin.ObjectMeta.Name = *service.Name + "-" + *plugin.Name
		}
		kongPlugin.ObjectMeta.Annotations = map[string]string{"kubernetes.io/ingress.class": "kong"}
		kongPlugin.PluginName = *plugin.Name

		var configJSON apiextensionsv1.JSON
		var err error
		configJSON.Raw, err = json.Marshal(plugin.Config)
		if err != nil {
			return err
		}
		kongPlugin.Config = configJSON

		if k8sService.ObjectMeta.Annotations["konghq.com/plugins"] == "" {
			k8sService.ObjectMeta.Annotations["konghq.com/plugins"] = kongPlugin.PluginName
		} else {
			annotations := k8sService.ObjectMeta.Annotations["konghq.com/plugins"] + "," + kongPlugin.PluginName
			k8sService.ObjectMeta.Annotations["konghq.com/plugins"] = annotations
		}

		kicContent.KongPlugins = append(kicContent.KongPlugins, kongPlugin)
	}
	return nil
}

func addPluginsToRoute(service FService, route *FRoute, ingress k8snetv1.Ingress, kicContent *KICContent) error {
	for _, plugin := range route.Plugins {
		var kongPlugin kicv1.KongPlugin
		kongPlugin.APIVersion = KICAPIVersion
		kongPlugin.Kind = KongPluginKind
		if plugin.Name != nil && route.Name != nil && service.Name != nil {
			kongPlugin.ObjectMeta.Name = *service.Name + "-" + *route.Name + "-" + *plugin.Name
		}
		kongPlugin.ObjectMeta.Annotations = map[string]string{"kubernetes.io/ingress.class": "kong"}
		kongPlugin.PluginName = *plugin.Name

		var configJSON apiextensionsv1.JSON
		var err error
		configJSON.Raw, err = json.Marshal(plugin.Config)
		if err != nil {
			return err
		}
		kongPlugin.Config = configJSON

		if ingress.ObjectMeta.Annotations["konghq.com/plugins"] == "" {
			ingress.ObjectMeta.Annotations["konghq.com/plugins"] = kongPlugin.PluginName
		} else {
			annotations := ingress.ObjectMeta.Annotations["konghq.com/plugins"] + "," + kongPlugin.PluginName
			ingress.ObjectMeta.Annotations["konghq.com/plugins"] = annotations
		}

		kicContent.KongPlugins = append(kicContent.KongPlugins, kongPlugin)
	}
	return nil
}

func populateKICConsumerGroups(content *Content, kicContent *KICContent) error {
	// iterate over the consumer groups and create a KongConsumerGroup for each one
	for _, consumerGroup := range content.ConsumerGroups {
		var kongConsumerGroup kicv1beta1.KongConsumerGroup
		kongConsumerGroup.APIVersion = KICAPIVersion
		kongConsumerGroup.Kind = "KongConsumerGroup"
		if consumerGroup.Name != nil {
			kongConsumerGroup.ObjectMeta.Name = *consumerGroup.Name
		}
		kongConsumerGroup.ObjectMeta.Annotations = map[string]string{"kubernetes.io/ingress.class": "kong"}
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

		// for each consumerGroup.plugin, create a KongPlugin and a plugin annotation in the kongConsumerGroup
		// to link the plugin
		for _, plugin := range consumerGroup.Plugins {
			var kongPlugin kicv1.KongPlugin
			kongPlugin.APIVersion = KICAPIVersion
			kongPlugin.Kind = KongPluginKind
			kongPlugin.ObjectMeta.Annotations = map[string]string{"kubernetes.io/ingress.class": "kong"}
			if plugin.Name != nil {
				kongPlugin.PluginName = *consumerGroup.Name + "-" + *plugin.Name
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
				kongConsumerGroup.ObjectMeta.Annotations["konghq.com/plugins"] = kongPlugin.PluginName
			} else {
				annotations := kongConsumerGroup.ObjectMeta.Annotations["konghq.com/plugins"] + "," + kongPlugin.PluginName
				kongConsumerGroup.ObjectMeta.Annotations["konghq.com/plugins"] = annotations
			}
		}

		kicContent.KongConsumerGroups = append(kicContent.KongConsumerGroups, kongConsumerGroup)
	}

	return nil
}

/////
// Functions for CUSTOM RESOURCES based manifests
/////

func populateKICServiceProxyAndUpstreamCustomResources(
	content *Content,
	service *FService,
	k8sservice *k8scorev1.Service,
	kicContent *KICContent,
) {
	// add Kong specific configuration to the k8s service via a KongIngress resource

	var kongIngress kicv1.KongIngress
	kongIngress.APIVersion = KICAPIVersion
	kongIngress.Kind = IngressKind
	if service.Name != nil {
		kongIngress.ObjectMeta.Name = *service.Name + "-proxy-upstream"
	}
	kongIngress.ObjectMeta.Annotations = map[string]string{"kubernetes.io/ingress.class": "kong"}

	// add an annotation to the k8sservice to link this kongIngress to it
	k8sservice.ObjectMeta.Annotations["konghq.com/override"] = kongIngress.ObjectMeta.Name

	// proxy attributes from the service to the kongIngress
	kongIngress.Proxy = &kicv1.KongIngressService{
		Protocol:       service.Protocol,
		Path:           service.Path,
		Retries:        service.Retries,
		ConnectTimeout: service.ConnectTimeout,
		WriteTimeout:   service.WriteTimeout,
		ReadTimeout:    service.ReadTimeout,
	}

	// Find the upstream (if any) whose name matches the service host and copy the upstream
	// into a kicv1.KongIngress resource. Append the kicv1.KongIngress to kicContent.KongIngresses.
	for _, upstream := range content.Upstreams {
		if upstream.Name != nil && strings.EqualFold(*upstream.Name, *service.Host) {
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
	kicContent.KongIngresses = append(kicContent.KongIngresses, kongIngress)
}

func populateKICServicesWithCustomResources(content *Content, kicContent *KICContent) error {
	// Iterate Kong Services and create k8s Services,
	// then create KongIngress resources for Kong Service
	// specific configuration and Upstream data.
	// Finally, create KongPlugin resources for each plugin
	// associated with the service.
	for i := range content.Services {
		service := content.Services[i]

		var k8sService k8scorev1.Service
		var protocol k8scorev1.Protocol

		k8sService.TypeMeta.APIVersion = "v1"
		k8sService.TypeMeta.Kind = "Service"
		if service.Name != nil {
			k8sService.ObjectMeta.Name = *service.Name
		} else {
			log.Println("Service without a name is not recommended")
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

		populateKICServiceProxyAndUpstreamCustomResources(content, &service, &k8sService, kicContent)

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

func populateKICIngressesWithCustomResources(content *Content, kicContent *KICContent) error {

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
			var k8sIngress k8snetv1.Ingress
			var pathTypeImplSpecific k8snetv1.PathType = k8snetv1.PathTypeImplementationSpecific

			k8sIngress.TypeMeta.APIVersion = "networking.k8s.io/v1"
			k8sIngress.TypeMeta.Kind = "Ingress"
			if service.Name != nil && route.Name != nil {
				k8sIngress.ObjectMeta.Name = *service.Name + "-" + *route.Name
			}
			ingressClassName := "kong"
			k8sIngress.Spec.IngressClassName = &ingressClassName
			k8sIngress.ObjectMeta.Annotations = make(map[string]string)

			// Create a KongIngress resource and copy Kong specific route data into it
			var kongIngress kicv1.KongIngress
			kongIngress.APIVersion = KICAPIVersion
			kongIngress.Kind = IngressKind
			kongIngress.ObjectMeta.Name = *service.Name + "-" + *route.Name
			kongIngress.ObjectMeta.Annotations = map[string]string{"kubernetes.io/ingress.class": "kong"}

			var kongProtocols []*kicv1.KongProtocol
			for _, protocol := range route.Protocols {
				p := kicv1.KongProtocol(*protocol)
				kongProtocols = append(kongProtocols, &p)
			}

			kongIngress.Route = &kicv1.KongIngressRoute{
				Methods:                 route.Methods,
				Protocols:               kongProtocols,
				StripPath:               route.StripPath,
				PreserveHost:            route.PreserveHost,
				RegexPriority:           route.RegexPriority,
				HTTPSRedirectStatusCode: route.HTTPSRedirectStatusCode,
				Headers:                 route.Headers,
				PathHandling:            route.PathHandling,
				SNIs:                    route.SNIs,
				RequestBuffering:        route.RequestBuffering,
				ResponseBuffering:       route.ResponseBuffering,
			}

			// add an annotation to the k8sIngress to link it to the kongIngress
			k8sIngress.ObjectMeta.Annotations = map[string]string{"konghq.com/override": kongIngress.ObjectMeta.Name}

			if len(route.Hosts) == 0 {
				// iterate route.Paths and create a k8sIngress.Spec.Rules for each one.
				// If service.Port is not nil, add it to the k8sIngress.Spec.Rules
				ingressRule := k8snetv1.IngressRule{
					IngressRuleValue: k8snetv1.IngressRuleValue{
						HTTP: &k8snetv1.HTTPIngressRuleValue{
							Paths: []k8snetv1.HTTPIngressPath{},
						},
					}}
				for _, path := range route.Paths {
					if service.Port != nil {
						ingressRule.IngressRuleValue.HTTP.Paths = append(ingressRule.IngressRuleValue.HTTP.Paths, k8snetv1.HTTPIngressPath{
							Path:     *path,
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
						ingressRule.IngressRuleValue.HTTP.Paths = append(ingressRule.IngressRuleValue.HTTP.Paths, k8snetv1.HTTPIngressPath{
							Path:     *path,
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
						}}
					for _, path := range route.Paths {
						if service.Port != nil {
							ingressRule.IngressRuleValue.HTTP.Paths = append(ingressRule.IngressRuleValue.HTTP.Paths, k8snetv1.HTTPIngressPath{
								Path:     *path,
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
							ingressRule.IngressRuleValue.HTTP.Paths = append(ingressRule.IngressRuleValue.HTTP.Paths, k8snetv1.HTTPIngressPath{
								Path:     *path,
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

			// transform the plugin config from map[string]interface{} to apiextensionsv1.JSON
			// create a plugins annotation in the routeIngresses to link them to this plugin.
			// separate plugins with commas
			err := addPluginsToRoute(service, route, k8sIngress, kicContent)
			if err != nil {
				return err
			}
			kicContent.Ingresses = append(kicContent.Ingresses, k8sIngress)
			kicContent.KongIngresses = append(kicContent.KongIngresses, kongIngress)
		}
	}
	return nil
}

/////
// Functions for ANNOTATION based manifests
/////

func populateKICServicesWithAnnotations(content *Content, kicContent *KICContent) error {
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
			k8sService.ObjectMeta.Name = *service.Name
		} else {
			log.Println("Service without a name is not recommended")
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

		populateKICUpstream(content, &service, &k8sService, kicContent)

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

func populateKICIngressesWithAnnotations(content *Content, kicContent *KICContent) error {

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
			var k8sIngress k8snetv1.Ingress
			var pathTypeImplSpecific k8snetv1.PathType = k8snetv1.PathTypeImplementationSpecific

			k8sIngress.TypeMeta.APIVersion = "networking.k8s.io/v1"
			k8sIngress.TypeMeta.Kind = "Ingress"
			if service.Name != nil && route.Name != nil {
				k8sIngress.ObjectMeta.Name = *service.Name + "-" + *route.Name
			}
			ingressClassName := "kong"
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
				k8sIngress.ObjectMeta.Annotations["konghq.com/https-redirect-status-code"] = strconv.Itoa(*route.HTTPSRedirectStatusCode)
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
					}}
				for _, path := range route.Paths {
					// if path starts with ~ then add / to the beginning of the path
					// see: https://docs.konghq.com/kubernetes-ingress-controller/latest/guides/upgrade-kong-3x/#update-ingress-regular-expression-paths-for-kong-3x-compatibility
					if strings.HasPrefix(*path, "~") {
						*path = "/" + *path
					}
					if service.Port != nil {
						ingressRule.IngressRuleValue.HTTP.Paths = append(ingressRule.IngressRuleValue.HTTP.Paths, k8snetv1.HTTPIngressPath{
							Path:     *path,
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
						ingressRule.IngressRuleValue.HTTP.Paths = append(ingressRule.IngressRuleValue.HTTP.Paths, k8snetv1.HTTPIngressPath{
							Path:     *path,
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
						}}
					for _, path := range route.Paths {
						// if path starts with ~ then add / to the beginning of the path
						// see: https://docs.konghq.com/kubernetes-ingress-controller/latest/guides/upgrade-kong-3x/#update-ingress-regular-expression-paths-for-kong-3x-compatibility
						if strings.HasPrefix(*path, "~") {
							*path = "/" + *path
						}
						if service.Port != nil {
							ingressRule.IngressRuleValue.HTTP.Paths = append(ingressRule.IngressRuleValue.HTTP.Paths, k8snetv1.HTTPIngressPath{
								Path:     *path,
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
							ingressRule.IngressRuleValue.HTTP.Paths = append(ingressRule.IngressRuleValue.HTTP.Paths, k8snetv1.HTTPIngressPath{
								Path:     *path,
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

			// transform the plugin config from map[string]interface{} to apiextensionsv1.JSON
			// create a plugins annotation in the routeIngresses to link them to this plugin.
			// separate plugins with commas
			err := addPluginsToRoute(service, route, k8sIngress, kicContent)
			if err != nil {
				return err
			}
			kicContent.Ingresses = append(kicContent.Ingresses, k8sIngress)
		}
	}
	return nil
}
