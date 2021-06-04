package convert

import (
	"fmt"
	"strings"

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
	default:
		return fmt.Errorf("cannot convert from '%s' to '%s' format", from, to)
	}

	err = file.WriteContentToFile(outputContent, outputFilename, file.YAML)
	if err != nil {
		return err
	}
	return nil
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
