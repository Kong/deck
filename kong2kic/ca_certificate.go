package kong2kic

import (
	"crypto/sha256"
	"fmt"

	"github.com/kong/go-database-reconciler/pkg/file"
	k8scorev1 "k8s.io/api/core/v1"
)

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
		secret.Type = k8scorev1.SecretTypeOpaque
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
