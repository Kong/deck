package convert

import (
	"strings"

	"github.com/kong/go-database-reconciler/pkg/file"
	"github.com/kong/go-database-reconciler/pkg/utils"
)

func migrateRoutesPathFieldPre300(route *file.FRoute) (*file.FRoute, bool) {
	var hasChanged bool
	for _, path := range route.Paths {
		if !strings.HasPrefix(*path, "~/") && utils.IsPathRegexLike(*path) {
			*path = "~" + *path
			hasChanged = true
		}
	}
	return route, hasChanged
}
