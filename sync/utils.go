package sync

import "strings"

func isPlaceHolder(id *string) bool {
	return id != nil && strings.HasPrefix(*id, "placeholder")
}
