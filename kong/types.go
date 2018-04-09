package kong

import "strings"

// String returns pointer to s.
func String(s string) *string {
	return &s
}

func isEmptyString(s *string) bool {
	return s == nil || strings.TrimSpace(*s) == ""
}
