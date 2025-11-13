package kong2kic

import (
	"strconv"

	"github.com/kong/go-database-reconciler/pkg/file"
	configurationv1 "github.com/kong/kubernetes-configuration/api/configuration/v1"
	k8scorev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Helper function to create credential secrets
func createCredentialSecret(consumerUsername, credentialType string, dataFields map[string]*string) k8scorev1.Secret {
	secretName := credentialType + "-" + consumerUsername
	stringData := make(map[string]string)
	labels := map[string]string{}
	if targetKICVersionAPI == KICV3GATEWAY || targetKICVersionAPI == KICV3INGRESS {
		labels[KongHQCredential] = credentialType
	} else {
		stringData[KongCredType] = credentialType
	}

	// Add the data fields to stringData
	for key, value := range dataFields {
		if value != nil {
			stringData[key] = *value
		}
	}

	return k8scorev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       SecretKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        calculateSlug(secretName),
			Annotations: map[string]string{IngressClass: ClassName},
			Labels:      labels,
		},
		Type:       "Opaque",
		StringData: stringData,
	}
}

// Functions to populate different credential types
func populateKICKeyAuthSecrets(consumer *file.FConsumer, kongConsumer *configurationv1.KongConsumer, file *KICContent) {
	for _, keyAuth := range consumer.KeyAuths {
		dataFields := map[string]*string{
			"key": keyAuth.Key,
		}
		secret := createCredentialSecret(*consumer.Username, "key-auth", dataFields)
		kongConsumer.Credentials = append(kongConsumer.Credentials, secret.Name)
		file.Secrets = append(file.Secrets, secret)
	}
}

func populateKICHMACSecrets(consumer *file.FConsumer, kongConsumer *configurationv1.KongConsumer, file *KICContent) {
	for _, hmacAuth := range consumer.HMACAuths {
		dataFields := map[string]*string{
			"username": hmacAuth.Username,
			"secret":   hmacAuth.Secret,
		}
		secret := createCredentialSecret(*consumer.Username, "hmac-auth", dataFields)
		kongConsumer.Credentials = append(kongConsumer.Credentials, secret.Name)
		file.Secrets = append(file.Secrets, secret)
	}
}

func populateKICJWTAuthSecrets(consumer *file.FConsumer, kongConsumer *configurationv1.KongConsumer, file *KICContent) {
	for _, jwtAuth := range consumer.JWTAuths {
		dataFields := map[string]*string{
			"key":            jwtAuth.Key,
			"algorithm":      jwtAuth.Algorithm,
			"rsa_public_key": jwtAuth.RSAPublicKey,
			"secret":         jwtAuth.Secret,
		}
		secret := createCredentialSecret(*consumer.Username, "jwt", dataFields)
		kongConsumer.Credentials = append(kongConsumer.Credentials, secret.Name)
		file.Secrets = append(file.Secrets, secret)
	}
}

func populateKICBasicAuthSecrets(
	consumer *file.FConsumer, kongConsumer *configurationv1.KongConsumer, file *KICContent,
) {
	for _, basicAuth := range consumer.BasicAuths {
		dataFields := map[string]*string{
			"username": basicAuth.Username,
			"password": basicAuth.Password,
		}
		secret := createCredentialSecret(*consumer.Username, "basic-auth", dataFields)
		kongConsumer.Credentials = append(kongConsumer.Credentials, secret.Name)
		file.Secrets = append(file.Secrets, secret)
	}
}

func populateKICOAuth2CredSecrets(
	consumer *file.FConsumer, kongConsumer *configurationv1.KongConsumer, file *KICContent,
) {
	for _, oauth2Cred := range consumer.Oauth2Creds {
		dataFields := map[string]*string{
			"name":          oauth2Cred.Name,
			"client_id":     oauth2Cred.ClientID,
			"client_secret": oauth2Cred.ClientSecret,
			"client_type":   oauth2Cred.ClientType,
		}
		if oauth2Cred.HashSecret != nil {
			hashSecretStr := strconv.FormatBool(*oauth2Cred.HashSecret)
			dataFields["hash_secret"] = &hashSecretStr
		}
		secret := createCredentialSecret(*consumer.Username, "oauth2", dataFields)
		kongConsumer.Credentials = append(kongConsumer.Credentials, secret.Name)
		file.Secrets = append(file.Secrets, secret)
	}
}

func populateKICACLGroupSecrets(
	consumer *file.FConsumer, kongConsumer *configurationv1.KongConsumer, file *KICContent,
) {
	for _, aclGroup := range consumer.ACLGroups {
		dataFields := map[string]*string{
			"group": aclGroup.Group,
		}
		secret := createCredentialSecret(*consumer.Username, "acl", dataFields)
		kongConsumer.Credentials = append(kongConsumer.Credentials, secret.Name)
		file.Secrets = append(file.Secrets, secret)
	}
}

func populateKICMTLSAuthSecrets(
	consumer *file.FConsumer, kongConsumer *configurationv1.KongConsumer, file *KICContent,
) {
	for _, mtlsAuth := range consumer.MTLSAuths {
		dataFields := map[string]*string{
			"subject_name": mtlsAuth.SubjectName,
			"id":           mtlsAuth.ID,
		}
		if mtlsAuth.CACertificate != nil && mtlsAuth.CACertificate.Cert != nil {
			dataFields["ca_certificate"] = mtlsAuth.CACertificate.Cert
		}
		secret := createCredentialSecret(*consumer.Username, "mtls-auth", dataFields)
		kongConsumer.Credentials = append(kongConsumer.Credentials, secret.Name)
		file.Secrets = append(file.Secrets, secret)
	}
}
