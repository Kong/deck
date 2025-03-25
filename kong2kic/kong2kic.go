package kong2kic

import (
	"github.com/kong/go-database-reconciler/pkg/file"
)

// ClassName is set by the CLI flag --class-name
var ClassName = "kong"

// targetKICVersionAPI is KIC v3.x Gateway API by default.
// Can be overridden by CLI flags.
var targetKICVersionAPI = KICV3GATEWAY

func MarshalKongToKIC(content *file.Content, builderType string, format string) ([]byte, error) {
	targetKICVersionAPI = builderType
	processTopLevelEntities(content)
	kicContent := convertKongToKIC(content, builderType)
	return kicContent.marshalKICContentToFormat(format)
}

func convertKongToKIC(content *file.Content, builderType string) *KICContent {
	builder := getBuilder(builderType)
	director := newDirector(builder)
	return director.buildManifests(content)
}
