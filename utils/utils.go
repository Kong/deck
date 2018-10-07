package utils

func Empty(s *string) bool {
	return s == nil || *s == ""
}
