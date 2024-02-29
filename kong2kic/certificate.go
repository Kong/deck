package kong2kic

import (
	"crypto/sha256"
	"fmt"

	"github.com/kong/go-database-reconciler/pkg/file"
	k8scorev1 "k8s.io/api/core/v1"
)

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
		secret.Type = k8scorev1.SecretTypeTLS
		secret.ObjectMeta.Name = calculateSlug(secretName)
		secret.ObjectMeta.Annotations = map[string]string{IngressClass: ClassName}
		secret.StringData = make(map[string]string)
		secret.StringData["tls.crt"] = *cert.Cert
		secret.StringData["tls.key"] = *cert.Key
		// what to do with SNIs?

		file.Secrets = append(file.Secrets, secret)
	}
}
