package kong2kic

import (
	"crypto/sha256"
	"fmt"
	"log"
	"strings"

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
		if cert.Cert != nil && cert.Key != nil {
			secret.StringData["tls.crt"] = *cert.Cert
			secret.StringData["tls.key"] = *cert.Key
		} else {
			log.Println("Certificate or Key is empty. This is not recommended." +
				"Please, provide a certificate and key before generating Kong Ingress Controller manifests.")
			continue
		}
		// what to do with SNIs?

		// add konghq.com/tags annotation if cert.Tags is not nil
		if cert.Tags != nil {
			var tags []string
			for _, tag := range cert.Tags {
				if tag != nil {
					tags = append(tags, *tag)
				}
			}
			secret.ObjectMeta.Annotations["konghq.com/tags"] = strings.Join(tags, ",")
		}

		file.Secrets = append(file.Secrets, secret)
	}
}
