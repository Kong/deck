package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	kongVersionRegex = regexp.MustCompile(`^\d+\.\d+`)
)

// Empty checks if a string referenced by s or s itself is empty.
func Empty(s *string) bool {
	return s == nil || *s == ""
}

// UUID will generate a random v14 unique identifier based upon random numbers
func UUID() string {
	version := byte(4)
	uuid := make([]byte, 16)
	_, err := rand.Read(uuid)
	if err != nil {
		panic("failed to read from random generator: " + err.Error())
	}

	// Set version
	uuid[6] = (uuid[6] & 0x0f) | (version << 4)

	// Set variant
	uuid[8] = (uuid[8] & 0xbf) | 0x80

	buf := make([]byte, 36)
	var dash byte = '-'
	hex.Encode(buf[0:8], uuid[0:4])
	buf[8] = dash
	hex.Encode(buf[9:13], uuid[4:6])
	buf[13] = dash
	hex.Encode(buf[14:18], uuid[6:8])
	buf[18] = dash
	hex.Encode(buf[19:23], uuid[8:10])
	buf[23] = dash
	hex.Encode(buf[24:], uuid[10:])

	return string(buf)
}

// CleanKongVersion takes a version of Kong and returns back a string in
// the form of `/major.minor` version. There are various dashes and dots
// and other descriptors in Kong version strings, which has often created
// confusion in code and incorrect parsing, and hence this function does
// not return the patch version (on which shouldn't rely on anyways).
func CleanKongVersion(version string) (string, error) {
	matches := kongVersionRegex.FindStringSubmatch(version)
	if len(matches) < 1 {
		return "", fmt.Errorf("unknown Kong version")
	}
	return matches[0], nil
}

func AddExtToFilename(filename, ext string) string {
	if filepath.Ext(filename) == "" {
		filename = filename + "." + ext
	}
	return filename
}

// NameToFilename clears path separators from strings. Some entity names in Kong and Konnect
// allow path directory separators. Some decK operations write files using entity names,
// which is not compatible with names that contain path separators. NameToFilename strips leading
// separator characters and replaces other instances of the separator with its URL-encoded representation.
func NameToFilename(name string) string {
	s := strings.TrimPrefix(name, string(os.PathSeparator))
	s = strings.ReplaceAll(s, string(os.PathSeparator), url.PathEscape(string(os.PathSeparator)))
	return s
}
