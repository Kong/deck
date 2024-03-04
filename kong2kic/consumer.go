package kong2kic

import (
	"encoding/json"
	"log"
	"strconv"
	"strings"

	"github.com/kong/go-database-reconciler/pkg/file"
	"github.com/kong/go-kong/kong"
	kicv1 "github.com/kong/kubernetes-ingress-controller/v3/pkg/apis/configuration/v1"
	k8scorev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

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

		// add konghq.com/tags annotation if consumer.Tags is not nil
		if consumer.Tags != nil {
			var tags []string
			for _, tag := range consumer.Tags {
				if tag != nil {
					tags = append(tags, *tag)
				}
			}
			kongConsumer.ObjectMeta.Annotations["konghq.com/tags"] = strings.Join(tags, ",")
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
				continue
			}

			// add konghq.com/tags annotation if plugin.Tags is not nil
			if plugin.Tags != nil {
				var tags []string
				for _, tag := range plugin.Tags {
					if tag != nil {
						tags = append(tags, *tag)
					}
				}
				kongPlugin.ObjectMeta.Annotations["konghq.com/tags"] = strings.Join(tags, ",")
			}

			// populate enabled, runon, ordering and protocols
			if plugin.Enabled != nil {
				kongPlugin.Disabled = !*plugin.Enabled
			}
			if plugin.RunOn != nil {
				kongPlugin.RunOn = *plugin.RunOn
			}
			if plugin.Ordering != nil {
				kongPlugin.Ordering = &kong.PluginOrdering{
					Before: plugin.Ordering.Before,
					After:  plugin.Ordering.After,
				}
			}
			if plugin.Protocols != nil {
				protocols := make([]string, len(plugin.Protocols))
				for i, protocol := range plugin.Protocols {
					if protocol != nil {
						protocols[i] = *protocol
					}
				}
				kongPlugin.Protocols = kicv1.StringsToKongProtocols(protocols)
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
