package utils

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"

	"github.com/kong/go-kong/kong"
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

func CallGetAll(obj interface{}) (reflect.Value, error) {
	// call GetAll method on entity
	var result reflect.Value
	method := reflect.ValueOf(obj).MethodByName("GetAll")
	if !method.IsValid() {
		return result, fmt.Errorf("GetAll() method not found for type '%v'. "+
			"Please file a bug with Kong Inc", reflect.ValueOf(obj).Type())
	}
	entities := method.Call([]reflect.Value{})[0].Interface()
	result = reflect.ValueOf(entities)
	return result, nil
}

func alreadyInSlice(elem string, slice []string) bool {
	for _, s := range slice {
		if s == elem {
			return true
		}
	}
	return false
}

// RemoveDuplicates removes duplicated elements from a slice.
func RemoveDuplicates(slice *[]string) {
	newSlice := []string{}
	for _, s := range *slice {
		if alreadyInSlice(s, newSlice) {
			continue
		}
		newSlice = append(newSlice, s)
	}
	*slice = newSlice
}

func WorkspaceExists(ctx context.Context, client *kong.Client) (bool, error) {
	if client == nil {
		return false, nil
	}
	workspace := client.Workspace()
	if workspace == "" {
		return true, nil
	}
	return client.Workspaces.Exists(ctx, &workspace)
}
