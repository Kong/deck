package file

import (
	"encoding/json"
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
	getContent() *KICContent
}

const (
	CUSTOM_RESOURCE = "CUSTOM_RESOURCE"
	ANNOTATIONS     = "ANNOTATIONS"
	GATEWAY         = "GATEWAY"
)

func getBuilder(builderType string) IBuilder {
	if builderType == CUSTOM_RESOURCE {
		return newCustomResourceBuilder()
	}

	if builderType == ANNOTATIONS {
		return newAnnotationsBuilder()
	}

	if builderType == GATEWAY {
		// TODO: implement gateway builder
	}
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
	return kicContent.marshalKICContentToJson()

}

func convertKongToKIC(content *Content, builderType string) *KICContent {
	builder := getBuilder(builderType)
	director := newDirector(builder)
	return director.buildManifests(content)
}

/////
// Functions valid for both custom resources and annotations based manifests
/////

func populateKICKongClusterPlugins(content *Content, file *KICContent) error {

	// Global Plugins map to KongClusterPlugins
	// iterate content.Plugins and copy them into kicv1.KongPlugin manifests
	// add the kicv1.KongPlugin to the KICContent.KongClusterPlugins slice
	for _, plugin := range content.Plugins {
		var kongPlugin kicv1.KongClusterPlugin
		kongPlugin.APIVersion = "configuration.konghq.com/v1"
		kongPlugin.Kind = "KongClusterPlugin"
		if plugin.InstanceName != nil {
			kongPlugin.ObjectMeta.Name = *plugin.InstanceName
		}
		kongPlugin.ObjectMeta.Annotations = map[string]string{"kubernetes.io/ingress.class": "kong"}
		if plugin.Name != nil {
			kongPlugin.PluginName = *plugin.Name
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
	for _, consumer := range content.Consumers {
		var kongConsumer kicv1.KongConsumer
		kongConsumer.APIVersion = "configuration.konghq.com/v1"
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
			kongPlugin.APIVersion = "configuration.konghq.com/v1"
			kongPlugin.Kind = "KongPlugin"
			if plugin.InstanceName != nil {
				kongPlugin.ObjectMeta.Name = *plugin.InstanceName
			}
			kongPlugin.ObjectMeta.Annotations = map[string]string{"kubernetes.io/ingress.class": "kong"}
			if plugin.Name != nil {
				kongPlugin.PluginName = *plugin.Name
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
				kongConsumer.ObjectMeta.Annotations["konghq.com/plugins"] = kongConsumer.ObjectMeta.Annotations["konghq.com/plugins"] + "," + kongPlugin.PluginName
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
		var secretName = "mtls-auth-" + *consumer.Username
		secret.TypeMeta.APIVersion = "v1"
		secret.TypeMeta.Kind = "Secret"
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
		var secretName = "acl-group-" + *consumer.Username
		secret.TypeMeta.APIVersion = "v1"
		secret.TypeMeta.Kind = "Secret"
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
		var secretName = "oauth2cred-" + *consumer.Username
		secret.TypeMeta.APIVersion = "v1"
		secret.TypeMeta.Kind = "Secret"
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
		var secretName = "basic-auth-" + *consumer.Username
		secret.TypeMeta.APIVersion = "v1"
		secret.TypeMeta.Kind = "Secret"
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
		var secretName = "jwt-auth-" + *consumer.Username
		secret.TypeMeta.APIVersion = "v1"
		secret.TypeMeta.Kind = "Secret"
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
		var secretName = "hmac-auth-" + *consumer.Username
		secret.TypeMeta.APIVersion = "v1"
		secret.TypeMeta.Kind = "Secret"
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
		var secretName = "key-auth-" + *consumer.Username
		secret.TypeMeta.APIVersion = "v1"
		secret.TypeMeta.Kind = "Secret"
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
		kongIngress.APIVersion = "configuration.konghq.com/v1"
		kongIngress.Kind = "KongIngress"
		kongIngress.ObjectMeta.Name = *service.Name
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
		kongPlugin.APIVersion = "configuration.konghq.com/v1"
		kongPlugin.Kind = "KongPlugin"
		if plugin.Name != nil {
			kongPlugin.ObjectMeta.Name = *plugin.Name
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
			k8sService.ObjectMeta.Annotations["konghq.com/plugins"] = k8sService.ObjectMeta.Annotations["konghq.com/plugins"] + "," + kongPlugin.PluginName
		}

		kicContent.KongPlugins = append(kicContent.KongPlugins, kongPlugin)
	}
	return nil
}

func addPluginsToRoute(route *FRoute, routeIngresses []k8snetv1.Ingress, kicContent *KICContent) error {
	for _, plugin := range route.Plugins {
		var kongPlugin kicv1.KongPlugin
		kongPlugin.APIVersion = "configuration.konghq.com/v1"
		kongPlugin.Kind = "KongPlugin"
		if plugin.Name != nil {
			kongPlugin.ObjectMeta.Name = *plugin.Name
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

		for _, k8sIngress := range routeIngresses {
			if k8sIngress.ObjectMeta.Annotations["konghq.com/plugins"] == "" {
				k8sIngress.ObjectMeta.Annotations["konghq.com/plugins"] = kongPlugin.PluginName
			} else {
				k8sIngress.ObjectMeta.Annotations["konghq.com/plugins"] = k8sIngress.ObjectMeta.Annotations["konghq.com/plugins"] + "," + kongPlugin.PluginName
			}
		}

		kicContent.KongPlugins = append(kicContent.KongPlugins, kongPlugin)
	}
	return nil
}

func fillIngressHostAndPortSection(host *string, service FService, k8sIngress *k8snetv1.Ingress, path *string, pathTypeImplSpecific k8snetv1.PathType) {
	if host != nil && service.Port != nil {
		k8sIngress.Spec.Rules = append(k8sIngress.Spec.Rules, k8snetv1.IngressRule{
			Host: *host,
			IngressRuleValue: k8snetv1.IngressRuleValue{
				HTTP: &k8snetv1.HTTPIngressRuleValue{
					Paths: []k8snetv1.HTTPIngressPath{
						{
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
						},
					},
				},
			},
		})
	} else if host == nil && service.Port != nil {
		k8sIngress.Spec.Rules = append(k8sIngress.Spec.Rules, k8snetv1.IngressRule{
			IngressRuleValue: k8snetv1.IngressRuleValue{
				HTTP: &k8snetv1.HTTPIngressRuleValue{
					Paths: []k8snetv1.HTTPIngressPath{
						{
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
						},
					},
				},
			},
		})
	} else if host != nil && service.Port == nil {
		k8sIngress.Spec.Rules = append(k8sIngress.Spec.Rules, k8snetv1.IngressRule{
			Host: *host,
			IngressRuleValue: k8snetv1.IngressRuleValue{
				HTTP: &k8snetv1.HTTPIngressRuleValue{
					Paths: []k8snetv1.HTTPIngressPath{
						{
							Path:     *path,
							PathType: &pathTypeImplSpecific,
							Backend: k8snetv1.IngressBackend{
								Service: &k8snetv1.IngressServiceBackend{
									Name: *service.Name,
								},
							},
						},
					},
				},
			},
		})
	} else {

		k8sIngress.Spec.Rules = append(k8sIngress.Spec.Rules, k8snetv1.IngressRule{
			IngressRuleValue: k8snetv1.IngressRuleValue{
				HTTP: &k8snetv1.HTTPIngressRuleValue{
					Paths: []k8snetv1.HTTPIngressPath{
						{
							Path:     *path,
							PathType: &pathTypeImplSpecific,
							Backend: k8snetv1.IngressBackend{
								Service: &k8snetv1.IngressServiceBackend{
									Name: *service.Name,
								},
							},
						},
					},
				},
			},
		})
	}
}

func populateKICConsumerGroups(content *Content, kicContent *KICContent) error {
	// iterate over the consumer groups and create a KongConsumerGroup for each one
	for _, consumerGroup := range content.ConsumerGroups {
		var kongConsumerGroup kicv1beta1.KongConsumerGroup
		kongConsumerGroup.APIVersion = "configuration.konghq.com/v1"
		kongConsumerGroup.Kind = "KongConsumerGroup"
		kongConsumerGroup.ObjectMeta.Name = *consumerGroup.Name
		kongConsumerGroup.ObjectMeta.Annotations = map[string]string{"kubernetes.io/ingress.class": "kong"}
		kongConsumerGroup.Name = *consumerGroup.Name

		// Iterate over the consumers in consumerGroup and
		// find the KongConsumer with the same username in kicContent.KongConsumers
		// and add it to the KongConsumerGroup
		for _, consumer := range consumerGroup.Consumers {
			for idx, _ := range kicContent.KongConsumers {
				if kicContent.KongConsumers[idx].Username == *consumer.Username {
					if kicContent.KongConsumers[idx].ConsumerGroups == nil {
						kicContent.KongConsumers[idx].ConsumerGroups = make([]string, 0)
					}
					kicContent.KongConsumers[idx].ConsumerGroups = append(kicContent.KongConsumers[idx].ConsumerGroups, *consumerGroup.Name)
				}
			}
		}

		// for each consumerGroup.plugin, create a KongPlugin and a plugin annotation in the kongConsumerGroup
		// to link the plugin
		for _, plugin := range consumerGroup.Plugins {
			var kongPlugin kicv1.KongPlugin
			kongPlugin.APIVersion = "configuration.konghq.com/v1"
			kongPlugin.Kind = "KongPlugin"
			kongPlugin.ObjectMeta.Annotations = map[string]string{"kubernetes.io/ingress.class": "kong"}
			if plugin.Name != nil {
				kongPlugin.PluginName = *plugin.Name
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
				kongConsumerGroup.ObjectMeta.Annotations["konghq.com/plugins"] = kongConsumerGroup.ObjectMeta.Annotations["konghq.com/plugins"] + "," + kongPlugin.PluginName
			}
		}

		kicContent.KongConsumerGroups = append(kicContent.KongConsumerGroups, kongConsumerGroup)
	}

	return nil

}

/////
// Functions for CUSTOM RESOURCES based manifests
/////

func populateKICServiceProxyAndUpstreamCustomResources(content *Content, service *FService, k8sservice *k8scorev1.Service, kicContent *KICContent) {

	// add Kong specific configuration to the k8s service via a KongIngress resource

	var kongIngress kicv1.KongIngress
	kongIngress.APIVersion = "configuration.konghq.com/v1"
	kongIngress.Kind = "KongIngress"
	kongIngress.ObjectMeta.Name = *service.Name
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
	for _, service := range content.Services {
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
		error := addPluginsToService(service, k8sService, kicContent)
		if error != nil {
			return error
		}
		kicContent.Services = append(kicContent.Services, k8sService)

	}
	return nil
}

func populateKICIngressesWithCustomResources(content *Content, kicContent *KICContent) error {

	// Transform routes into k8s Ingress and KongIngress resources
	// Assume each pair host/path will get its own ingress manifest
	for _, service := range content.Services {
		for _, route := range service.Routes {
			// save all ingresses we create for this route so we can then
			// assign them the plugins defined for the route
			var routeIngresses []k8snetv1.Ingress

			// if there are no hosts just use the paths
			if len(route.Hosts) == 0 {
				routeIngresses = KongRoutePathToIngressPathCustomResources(route, nil, routeIngresses, kicContent, service)
			} else {
				// iterate over the hosts and paths and create an ingress for each

				for _, host := range route.Hosts {
					//  create a KongIngress resource and copy route data into it
					// add annotation to the ingress to link it to the kongIngress
					routeIngresses = KongRoutePathToIngressPathCustomResources(route, host, routeIngresses, kicContent, service)

				}
			}
			// transform the plugin config from map[string]interface{} to apiextensionsv1.JSON
			// create a plugins annotation in the routeIngresses to link them to this plugin.
			// separate plugins with commas
			error := addPluginsToRoute(route, routeIngresses, kicContent)
			if error != nil {
				return error
			}
		}
	}
	return nil

}

func KongRoutePathToIngressPathCustomResources(route *FRoute, host *string, routeIngresses []k8snetv1.Ingress, kicContent *KICContent, service FService) []k8snetv1.Ingress {
	for _, path := range route.Paths {
		var k8sIngress k8snetv1.Ingress
		var pathTypeImplSpecific k8snetv1.PathType = k8snetv1.PathTypeImplementationSpecific
		k8sIngress.TypeMeta.APIVersion = "networking.k8s.io/v1"
		k8sIngress.TypeMeta.Kind = "Ingress"
		k8sIngress.ObjectMeta.Name = *route.Name
		ingressClassName := "kong"
		k8sIngress.Spec.IngressClassName = &ingressClassName

		// Host and/or Service.Port can be nil. There are 4 possible combinations.
		// host == nil && service.Port == nil
		fillIngressHostAndPortSection(host, service, &k8sIngress, path, pathTypeImplSpecific)

		// Create a KongIngress resource and copy Kong specific route data into it
		var kongIngress kicv1.KongIngress
		kongIngress.APIVersion = "configuration.konghq.com/v1"
		kongIngress.Kind = "KongIngress"
		kongIngress.ObjectMeta.Name = *route.Name
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

		routeIngresses = append(routeIngresses, k8sIngress)

		kicContent.Ingresses = append(kicContent.Ingresses, k8sIngress)
		kicContent.KongIngresses = append(kicContent.KongIngresses, kongIngress)
	}
	return routeIngresses
}

/////
// Functions for ANNOTATION based manifests
/////

func populateKICServicesWithAnnotations(content *Content, kicContent *KICContent) error {

	// Iterate Kong Services and create k8s Services,
	// then create KongIngress resources for Kong Service Upstream data.
	// Finally, create KongPlugin resources for each plugin
	// associated with the service.
	for _, service := range content.Services {
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
		error := addPluginsToService(service, k8sService, kicContent)
		if error != nil {
			return error
		}

		kicContent.Services = append(kicContent.Services, k8sService)

	}
	return nil
}

func populateKICIngressesWithAnnotations(content *Content, kicContent *KICContent) error {

	// Transform routes into k8s Ingress and KongIngress resources
	// Assume each pair host/path will get its own ingress manifest
	for _, service := range content.Services {
		for _, route := range service.Routes {
			// save all ingresses we create for this route so we can then
			// assign them the plugins defined for the route
			var routeIngresses []k8snetv1.Ingress

			// if there are no hosts just use the paths
			if len(route.Hosts) == 0 {
				routeIngresses = KongRoutePathToIngressPathAnnotations(route, nil, routeIngresses, kicContent, service)
			} else {
				// iterate over the hosts and paths and create an ingress for each

				for _, host := range route.Hosts {
					//  create a KongIngress resource and copy route data into it
					// add annotation to the ingress to link it to the kongIngress
					routeIngresses = KongRoutePathToIngressPathAnnotations(route, host, routeIngresses, kicContent, service)

				}
			}
			// transform the plugin config from map[string]interface{} to apiextensionsv1.JSON
			// create a plugins annotation in the routeIngresses to link them to this plugin.
			// separate plugins with commas
			error := addPluginsToRoute(route, routeIngresses, kicContent)
			if error != nil {
				return error
			}
		}
	}
	return nil
}

func KongRoutePathToIngressPathAnnotations(route *FRoute, host *string, routeIngresses []k8snetv1.Ingress, file *KICContent, service FService) []k8snetv1.Ingress {
	for _, path := range route.Paths {
		var k8sIngress k8snetv1.Ingress
		var pathTypeImplSpecific k8snetv1.PathType = k8snetv1.PathTypeImplementationSpecific
		k8sIngress.TypeMeta.APIVersion = "networking.k8s.io/v1"
		k8sIngress.TypeMeta.Kind = "Ingress"
		k8sIngress.ObjectMeta.Name = *route.Name
		ingressClassName := "kong"
		k8sIngress.Spec.IngressClassName = &ingressClassName
		k8sIngress.ObjectMeta.Annotations = make(map[string]string)

		// Host and/or Service.Port can be nil. There are 4 possible combinations.
		// host == nil && service.Port == nil
		fillIngressHostAndPortSection(host, service, &k8sIngress, path, pathTypeImplSpecific)

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

		routeIngresses = append(routeIngresses, k8sIngress)

		file.Ingresses = append(file.Ingresses, k8sIngress)
	}
	return routeIngresses
}
