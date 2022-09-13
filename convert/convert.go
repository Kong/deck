package convert

import (
	"fmt"
	"strings"

	"github.com/kong/deck/cprint"
	"github.com/kong/deck/file"
	"github.com/kong/deck/utils"
	"github.com/kong/go-kong/kong"
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
	default:
		return "", fmt.Errorf("invalid format: '%v'", key)
	}
}

func Convert(inputFilename, outputFilename string, from, to Format) error {
	var (
		outputContent *file.Content
		err           error
	)

	inputContent, err := file.GetContentFromFiles([]string{inputFilename})
	if err != nil {
		return err
	}

	switch {
	case from == FormatKongGateway && to == FormatKonnect:
		outputContent, err = convertKongGatewayToKonnect(inputContent)
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
	for _, service := range outputContent.Services {
		for _, route := range service.Routes {
			convertedRoutes = append(convertedRoutes, migrateRoutesPathFieldPre300(route, filename))
		}
		service.Routes = convertedRoutes
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

func migrateRoutesPathFieldPre300(route *file.FRoute, filename string) *file.FRoute {
	changedPaths := []string{}
	for _, path := range route.Paths {
		if !strings.HasPrefix(*path, "~/") && utils.IsPathRegexLike(*path) {
			changedPaths = append(changedPaths, *path)
			*path = "~" + *path
		}
	}
	if len(changedPaths) > 0 {
		changedPathsLen := len(changedPaths)
		// do not consider more than 3 sample routes to print out.
		if changedPathsLen > 3 {
			changedPaths = changedPaths[:3]
		}
		cprint.UpdatePrintf(
			"From the '%s' config file,\n"+
				"%d routes paths matching an unsupported regex pattern usage\n"+
				"with Kong version 3.0 or above were detected\n"+
				"(e.g. %s).\n\n"+
				"Kong gateway versions 3.0 and above require that regular expressions\n"+
				"start with a '~' character to distinguish from simple prefix match.\n"+
				"In order to make these paths compatible with 3.x, a '~' prefix has been added.\n\n",
			filename, changedPathsLen, strings.Join(changedPaths, ", "))
	}
	return route
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
