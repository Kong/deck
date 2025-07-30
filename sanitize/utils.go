package sanitize

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

func (s *Sanitizer) hashValue(value string) string {
	hasher := sha256.New()
	hasher.Write([]byte(s.salt + value))
	return hex.EncodeToString(hasher.Sum(nil))
}

func (s *Sanitizer) sanitiseValue(value string) string {
	hashedPath := s.hashValue(value)
	if !strings.Contains(value, "/") {
		return hashedPath
	}

	var redactedPath string

	// this means it is a path field
	if strings.HasPrefix(value, "/") || strings.HasPrefix(value, "~/") {
		redactedPath = "/redacted/path/"
	} else {
		// if it is not a path, but a string with a /, it could be a content-type
		redactedPath = "redacted/"
	}

	return redactedPath + hashedPath
}
