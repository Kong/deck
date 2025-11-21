package kong2kic

import (
	"crypto/sha256"
	"fmt"
	"log"
	"strings"

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
		secret.APIVersion = "v1"
		secret.Kind = SecretKind
		secret.Type = k8scorev1.SecretTypeOpaque
		secret.Name = calculateSlug(secretName)
		secret.Annotations = map[string]string{IngressClass: ClassName}
		secret.StringData = make(map[string]string)
		if caCert.Cert != nil {
			secret.StringData["ca.crt"] = *caCert.Cert
		} else {
			log.Println("CA Certificate is empty. This is not recommended." +
				"Please, provide a certificate for the CA before generating Kong Ingress Controller manifests.")
			continue
		}
		if caCert.CertDigest != nil {
			secret.StringData[SecretCADigest] = *caCert.CertDigest
		}

		// add konghq.com/tags annotation if cacert.Tags is not nil
		if caCert.Tags != nil {
			var tags []string
			for _, tag := range caCert.Tags {
				if tag != nil {
					tags = append(tags, *tag)
				}
			}
			secret.Annotations[KongHQTags] = strings.Join(tags, ",")
		}

		file.Secrets = append(file.Secrets, secret)
	}
}
