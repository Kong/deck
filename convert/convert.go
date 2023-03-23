package convert

import (
	"context"
	"fmt"
	"strings"

	"github.com/blang/semver/v4"
	"github.com/kong/deck/cprint"
	"github.com/kong/deck/dump"
	"github.com/kong/deck/file"
	"github.com/kong/deck/state"
	"github.com/kong/deck/utils"
	"github.com/kong/go-kong/kong"
)

type Format string

const (
	// FormatDistributed represents the Deck configuration format.
	FormatDistributed Format = "distributed"
	// FormatKongGateway represents the Kong gateway format.
	FormatKongGateway Format = "kong-gateway"
	// FormatKonnect represents the Konnect format.
	FormatKonnect Format = "konnect"
	// FormatKongGateway2x represents the Kong gateway 2.x format.
	FormatKongGateway2x Format = "kong-gateway-2.x"
	// FormatKongGateway3x represents the Kong gateway 3.x format.
	FormatKongGateway3x Format = "kong-gateway-3.x"
)

// AllFormats contains all available formats.
var AllFormats = []Format{FormatKongGateway, FormatKonnect}

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
	case FormatDistributed:
		return FormatDistributed, nil
	default:
		return "", fmt.Errorf("invalid format: '%v'", key)
	}
}

func Convert(
	inputFilenames []string,
	outputFilename string,
	outputFormat file.Format,
	from Format,
	to Format,
	mockEnvVars bool,
) error {
	var outputContent *file.Content

	inputContent, err := file.GetContentFromFiles(inputFilenames, mockEnvVars)
	if err != nil {
		return err
	}

	switch {
	case from == FormatKongGateway && to == FormatKonnect:
		if len(inputFilenames) > 1 {
			return fmt.Errorf("only one input file can be provided when converting from Kong to Konnect format")
		}
		outputContent, err = convertKongGatewayToKonnect(inputContent)
		if err != nil {
			return err
		}

	case from == FormatKongGateway2x && to == FormatKongGateway3x:
		if len(inputFilenames) > 1 {
			return fmt.Errorf("only one input file can be provided when converting from Kong 2.x to Kong 3.x format")
		}
		outputContent, err = convertKongGateway2xTo3x(inputContent, inputFilenames[0])
		if err != nil {
			return err
		}

	case from == FormatDistributed && to == FormatKongGateway,
		from == FormatDistributed && to == FormatKongGateway2x,
		from == FormatDistributed && to == FormatKongGateway3x:
		outputContent, err = convertDistributedToKong(inputContent, outputFilename, outputFormat, to)
		if err != nil {
			return err
		}

	default:
		return fmt.Errorf("cannot convert from '%s' to '%s' format", from, to)
	}

	err = file.WriteContentToFile(outputContent, outputFilename, outputFormat)
	return err
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

func convertKongGatewayToKonnect(input *file.Content) (*file.Content, error) {
	if input == nil {
		return nil, fmt.Errorf("input content is nil")
	}
	outputContent := input.DeepCopy()

	for _, service := range outputContent.Services {
		servicePackage, err := kongServiceToKonnectServicePackage(service)
		if err != nil {
			return nil, err
		}
		outputContent.ServicePackages = append(outputContent.ServicePackages, servicePackage)
	}
	// Remove Kong Services from the file because all of them have been converted
	// into Service packages
	outputContent.Services = nil

	// all other entities are left as is

	return outputContent, nil
}

func kongServiceToKonnectServicePackage(service file.FService) (file.FServicePackage, error) {
	if service.Name == nil {
		return file.FServicePackage{}, fmt.Errorf("kong service with id '%s' doesn't have a name,"+
			"all services must be named to convert them from %s to %s format",
			*service.ID, FormatKongGateway, FormatKonnect)
	}

	serviceName := *service.Name
	// Kong service MUST contain an ID and no name in Konnect representation

	// convert Kong Service to a Service Package
	return file.FServicePackage{
		Name:        &serviceName,
		Description: kong.String("placeholder description for " + serviceName + " service package"),
		Versions: []file.FServiceVersion{
			{
				Version: kong.String("v1"),
				Implementation: &file.Implementation{
					Type: utils.ImplementationTypeKongGateway,
					Kong: &file.Kong{
						Service: removeServiceName(&service),
					},
				},
			},
		},
	}, nil
}

func removeServiceName(service *file.FService) *file.FService {
	serviceCopy := service.DeepCopy()
	serviceCopy.Name = nil
	serviceCopy.ID = kong.String(utils.UUID())
	return serviceCopy
}

// convertDistributedToKong is used to convert one or many distributed format
// files to create one Kong Gateway declarative config. It also leverages some
// deck features like the defaults/centralized plugin configurations.
func convertDistributedToKong(
	targetContent *file.Content,
	outputFilename string,
	format file.Format,
	kongFormat Format,
) (*file.Content, error) {
	var version semver.Version

	switch kongFormat { //nolint:exhaustive
	case FormatKongGateway,
		FormatKongGateway3x:
		version = semver.Version{Major: 3, Minor: 0}
	case FormatKongGateway2x:
		version = semver.Version{Major: 2, Minor: 8}
	}

	s, _ := state.NewKongState()
	rawState, err := file.Get(context.Background(), targetContent, file.RenderConfig{
		CurrentState: s,
		KongVersion:  version,
	}, dump.Config{}, nil)
	if err != nil {
		return nil, err
	}
	targetState, err := state.Get(rawState)
	if err != nil {
		return nil, err
	}

	// file.KongStateToContent calls file.WriteContentToFile
	return file.KongStateToContent(targetState, file.WriteConfig{
		Filename:    outputFilename,
		FileFormat:  format,
		KongVersion: version.String(),
	})
}
