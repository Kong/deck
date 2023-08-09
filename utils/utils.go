package utils

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"

	"github.com/blang/semver/v4"
	"github.com/kong/deck/cprint"
	"github.com/kong/go-kong/kong"
)

var (
	kongVersionRegex = regexp.MustCompile(`^\d+\.\d+`)
	pathRegexPattern = regexp.MustCompile(`[^a-zA-Z0-9._~/%-]`)

	Kong140Version = semver.MustParse("1.4.0")
	Kong300Version = semver.MustParse("3.0.0")
	Kong340Version = semver.MustParse("3.4.0")
)

var ErrorConsumerGroupUpgrade = errors.New(
	"a rate-limiting-advanced plugin with config.consumer_groups\n" +
		"and/or config.enforce_consumer_groups was found. Please use Consumer Groups scoped\n" +
		"Plugins when running against Kong Enterprise 3.4.0 and above.\n\n" +
		"Check https://docs.konghq.com/gateway/latest/kong-enterprise/consumer-groups/ for more information",
)

var UpgradeMessage = "Please upgrade your configuration to account for 3.0\n" +
	"breaking changes using the following command:\n\n" +
	"deck convert --from kong-gateway-2.x --to kong-gateway-3.x\n\n" +
	"This command performs the following changes:\n" +
	"  - upgrade the `_format_version` value to `3.0`\n" +
	"  - add the `~` prefix to all routes' paths containing a regex-pattern\n\n" +
	"These changes may not be correct or exhaustive enough.\n" +
	"It is strongly recommended to perform a manual audit\n" +
	"of the updated configuration file before applying\n" +
	"the configuration in production. Incorrect changes will result in\n" +
	"unintended traffic routing by Kong Gateway.\n\n" +

	"For more information about this and related changes,\n" +
	"please visit: https://docs.konghq.com/deck/latest/3.0-upgrade\n\n"

// IsPathRegexLike checks if a path string contains a regex pattern.
func IsPathRegexLike(path string) bool {
	return pathRegexPattern.MatchString(path)
}

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

// These GetFooReference functions return stripped copies (ID and Name only) of Kong resource
// structs. We use these within KongRawState structs to indicate entity relationships.
// While state files indicate relationships by nesting (A collection of services is
// [{name: "foo", id: "1234", connect_timeout: 600000, routes: [{name: "fooRoute"}]}]),
// KongRawState is flattened, with all entities listed independently at the top level.
// To preserve the relationships, these flattened entities include references (the route from
// earlier becomes {name: "fooRoute", service: {name: "foo", id: "1234"}}).

// GetConsumerReference returns a username+ID only copy of the input consumer,
// for use in references from other objects
func GetConsumerReference(c kong.Consumer) *kong.Consumer {
	consumer := &kong.Consumer{ID: kong.String(*c.ID)}
	if c.Username != nil {
		consumer.Username = kong.String(*c.Username)
	}
	return consumer
}

// GetConsumerGroupReference returns a name+ID only copy of the input consumer-group,
// for use in references from other objects
func GetConsumerGroupReference(c kong.ConsumerGroup) *kong.ConsumerGroup {
	consumerGroup := &kong.ConsumerGroup{ID: kong.String(*c.ID)}
	if c.Name != nil {
		consumerGroup.Name = kong.String(*c.Name)
	}
	return consumerGroup
}

// GetServiceReference returns a name+ID only copy of the input service,
// for use in references from other objects
func GetServiceReference(s kong.Service) *kong.Service {
	service := &kong.Service{ID: kong.String(*s.ID)}
	if s.Name != nil {
		service.Name = kong.String(*s.Name)
	}
	return service
}

// GetRouteReference returns a name+ID only copy of the input route,
// for use in references from other objects
func GetRouteReference(r kong.Route) *kong.Route {
	route := &kong.Route{ID: kong.String(*r.ID)}
	if r.Name != nil {
		route.Name = kong.String(*r.Name)
	}
	return route
}

// ParseKongVersion takes a version string from the Gateway and
// turns it into a semver-compliant version to be used for
// comparison across the code.
func ParseKongVersion(version string) (semver.Version, error) {
	v, err := CleanKongVersion(version)
	if err != nil {
		return semver.Version{}, err
	}
	return semver.ParseTolerant(v)
}

// ConfigFilesInDir traverses the directory rooted at dir and
// returns all the files with a case-insensitive extension of `yml` or `yaml`.
func ConfigFilesInDir(dir string) ([]string, error) {
	var res []string
	err := filepath.Walk(
		dir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			switch strings.ToLower(filepath.Ext(path)) {
			case ".yaml", ".yml", ".json":
				res = append(res, path)
			}
			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("reading state directory: %w", err)
	}
	return res, nil
}

// HasPathsWithRegex300AndAbove checks routes' paths format and returns true
// if these math a regex-pattern without a '~' prefix.
func HasPathsWithRegex300AndAbove(route kong.Route) bool {
	for _, p := range route.Paths {
		if strings.HasPrefix(*p, "~/") || !IsPathRegexLike(*p) {
			continue
		}
		return true
	}
	return false
}

// PrintRouteRegexWarning prints out a warning about 3.x routes' path usage.
func PrintRouteRegexWarning(unsupportedRoutes []string) {
	unsupportedRoutesLen := len(unsupportedRoutes)
	// do not consider more than 10 sample routes to print out.
	if unsupportedRoutesLen > 10 {
		unsupportedRoutes = unsupportedRoutes[:10]
	}
	cprint.UpdatePrintf(
		"%d unsupported routes' paths format with Kong version 3.0\n"+
			"or above were detected. Some of these routes are (not an exhaustive list):\n\n"+
			"%s\n\n"+UpgradeMessage,
		unsupportedRoutesLen, strings.Join(unsupportedRoutes, "\n"),
	)
}
