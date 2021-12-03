package utils

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var kongVersionRegex = regexp.MustCompile(`^\d+\.\d+`)

// Empty checks if a string referenced by s or s itself is empty.
func Empty(s *string) bool {
	return s == nil || *s == ""
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

// FilenameToName (partially) reverses NameToFilename, replacing all URL-encoded path separator characters
// with the path separator character. It does not re-add a leading separator, because there is no way to know
// if that separator was included originally, and only some names (document paths) typically include one.
func FilenameToName(filename string) string {
	return strings.ReplaceAll(filename, url.PathEscape(string(os.PathSeparator)), string(os.PathSeparator))
}

func BasicAuthFormat(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
