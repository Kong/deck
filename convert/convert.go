package convert

import (
	"fmt"
	"strings"

	"github.com/kong/deck/cprint"
	"github.com/kong/deck/file"
	"github.com/kong/deck/utils"
)

type Format string

const (
	// FormatKongGateway represents the Kong gateway format.
	FormatKongGateway Format = "kong-gateway"
	// FormatKonnect represents the Konnect format.
	FormatKonnect Format = "konnect"
	// FormatKongGateway2x represents the Kong gateway 2.x format.
	FormatKongGateway2x Format = "kong-gateway-2.x"
	// FormatKongGateway3x represents the Kong gateway 3.x format.
	FormatKongGateway3x Format = "kong-gateway-3.x"
)

func ParseFormat(key string) (Format, error) {
	format := Format(strings.ToLower(key))
	switch format {
	case FormatKongGateway:
		return FormatKongGateway, nil
	case FormatKonnect:
		return FormatKonnect, nil
	case FormatKongGateway2x:
		return FormatKongGateway2x, nil
	case FormatKongGateway3x:
		return FormatKongGateway3x, nil
	default:
		return "", fmt.Errorf("invalid format: '%v'", key)
	}
}

type Opts struct {
	InputFilename  string
	OutputFilename string
	FromFormat     Format
	ToFormat       Format

	// konnect-specific
	RuntimeGroupName string
}

func Convert(opts Opts) error {
	var (
		outputContent *file.Content
		err           error
	)

	inputFilename, outputFilename := opts.InputFilename, opts.OutputFilename
	from, to := opts.FromFormat, opts.ToFormat
	inputContent, err := file.GetContentFromFiles([]string{inputFilename})
	if err != nil {
		return err
	}

	switch {
	case from == FormatKongGateway && to == FormatKonnect:
		outputContent, err = convertKongGatewayToKonnect(inputContent, opts.RuntimeGroupName)
		if err != nil {
			return err
		}
	case from == FormatKongGateway2x && to == FormatKongGateway3x:
		outputContent, err = convertKongGateway2xTo3x(inputContent, inputFilename)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("cannot convert from '%s' to '%s' format", from, to)
	}

	err = file.WriteContentToFile(outputContent, outputFilename, file.YAML)
	if err != nil {
		return err
	}
	return nil
}

func convertKongGateway2xTo3x(input *file.Content, filename string) (*file.Content, error) {
	if input == nil {
		return nil, fmt.Errorf("input content is nil")
	}
	outputContent := input.DeepCopy()

	convertedRoutes := []*file.FRoute{}
	changedRoutes := []string{}
	for _, service := range outputContent.Services {
		for _, route := range service.Routes {
			route, hasChanged := migrateRoutesPathFieldPre300(route)
			convertedRoutes = append(convertedRoutes, route)
			if hasChanged {
				if route.ID != nil {
					changedRoutes = append(changedRoutes, *route.ID)
				} else {
					changedRoutes = append(changedRoutes, *route.Name)
				}
			}
		}
		service.Routes = convertedRoutes
	}

	if len(changedRoutes) > 0 {
		changedRoutesLen := len(changedRoutes)
		// do not consider more than 10 sample routes to print out.
		if changedRoutesLen > 10 {
			changedRoutes = changedRoutes[:10]
		}
		cprint.UpdatePrintf(
			"From the '%s' config file,\n"+
				"%d unsupported routes' paths format with Kong version 3.0\n"+
				"or above were detected. Some of these routes are (not an exhaustive list):\n\n"+
				"%s\n\n"+
				"Kong gateway versions 3.0 and above require that regular expressions\n"+
				"start with a '~' character to distinguish from simple prefix match.\n"+
				"In order to make these paths compatible with 3.x, a '~' prefix has been added.\n\n",
			filename, changedRoutesLen, strings.Join(changedRoutes, "\n"))
	}

	cprint.UpdatePrintf(
		"From the '%s' config file,\n"+
			"the _format_version field has been migrated from '%s' to '%s'.\n"+
			"These automatic changes may not be correct or exhaustive enough, please\n"+
			"perform a manual audit of the config file.\n\n"+
			"For related information, please visit:\n"+
			"https://docs.konghq.com/deck/latest/3.0-upgrade\n\n",
		filename, outputContent.FormatVersion, "3.0")
	outputContent.FormatVersion = "3.0"

	return outputContent, nil
}

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

// convertKongGatewayToKonnect convert a Kong configuration file format to a Konnect
// configuration file format. The two formats are almost identical, except for a few gotchas:
// - the `_format_version` option must be set to '3.0'
// - no `_workspace` must be present in the Konnect format (since no Workspaces exist there)
//
// This function makes sure to set the appropriate `_format_version` value and handle
// the removal of the `_workspace` configuration following this algorithm:
//  1. if present, remove the `_workspace` entry
//  2. if `--konnect-runtime-group-name foo` flag is used, then set `_konnect.runtime_group_name: foo`
//  3. if no `--konnect-runtime-group-name` flag is used, then set `_konnect.runtime_group_name`
//     to the value `_workspace` was precedently set to
//  4. if no `_workspace` was present and no `--konnect-runtime-group-name` flag is used,
//     then leave the whole `_konnect` entry unset
func convertKongGatewayToKonnect(
	input *file.Content, runtimeGroupName string,
) (*file.Content, error) {
	if input == nil {
		return nil, fmt.Errorf("input content is nil")
	}
	outputContent := input.DeepCopy()

	workspace := outputContent.Workspace
	outputContent.Workspace = ""
	// the `--konnect-runtime-group-name` flag takes precedence over the `_workspace` value
	if runtimeGroupName != "" {
		if outputContent.Konnect == nil {
			outputContent.Konnect = &file.Konnect{}
		}
		outputContent.Konnect.RuntimeGroupName = runtimeGroupName
	} else if workspace != "" {
		if outputContent.Konnect == nil {
			outputContent.Konnect = &file.Konnect{}
		}
		outputContent.Konnect.RuntimeGroupName = workspace
	}
	outputContent.FormatVersion = "3.0"
	return outputContent, nil
}
