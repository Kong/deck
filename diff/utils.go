package diff

import "strings"

const nodeKey = "node"

func isPlaceHolder(id *string) bool {
	return id != nil && strings.HasPrefix(*id, "placeholder")
}
