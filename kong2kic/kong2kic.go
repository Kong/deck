package kong2kic

import (
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/gosimple/slug"
	"github.com/kong/go-database-reconciler/pkg/file"
)

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
