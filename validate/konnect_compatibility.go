package validate

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/kong/go-database-reconciler/pkg/file"
	"github.com/kong/go-kong/kong"
)

func checkPlugin(name *string, config kong.Configuration) error {
	switch *name {
	case "jwt-signer", "vault-auth", "oauth2":
		return fmt.Errorf("[%s] plugin is not compatible with Konnect", *name)
	case "application-registration":
		return fmt.Errorf("[%s] available in Konnect, but doesn't require this plugin", *name)
	case "key-auth-enc":
		return fmt.Errorf("[%s] keys are automatically encrypted in Konnect, use the key auth plugin instead", *name)
	case "openwhisk":
		return fmt.Errorf("[%s] plugin not bundled with Kong Gateway - installed as a LuaRocks package", *name)
	case "rate-limiting", "rate-limiting-advanced", "response-ratelimiting", "graphql-rate-limiting-advanced":
		if config["strategy"] == "cluster" {
			return fmt.Errorf("[%s] plugin can't be used with cluster strategy", *name)
		}
	default:
	}
	return nil
}

func KonnectCompatibility(targetContent *file.Content) []error {
	var errs []error

	if targetContent.Workspace != "" {
		errs = append(errs, errors.New("[workspaces] not supported by Konnect - use control planes instead"))
	}

	if targetContent.Konnect == nil {
		errs = append(errs, errors.New("[konnect] section not specified - ensure details are set via cli flags"))
	}

	versionNumber, err := strconv.ParseFloat(targetContent.FormatVersion, 32)
	if err != nil {
		errs = append(errs, errors.New("[version] unable to determine decK file version"))
	} else {
		if versionNumber < 3.0 {
			errs = append(errs, errors.New("[version] decK file version must be '3.0' or greater"))
		}
	}

	for _, plugin := range targetContent.Plugins {
		if plugin.Enabled != nil && *plugin.Enabled && plugin.Config != nil {
			err := checkPlugin(plugin.Name, plugin.Config)
			if err != nil {
				errs = append(errs, checkPlugin(plugin.Name, plugin.Config))
			}
		}
	}

	for _, consumer := range targetContent.Consumers {
		for _, plugin := range consumer.Plugins {
			if plugin.Enabled != nil && *plugin.Enabled && plugin.Config != nil {
				err := checkPlugin(plugin.Name, plugin.Config)
				if err != nil {
					errs = append(errs, checkPlugin(plugin.Name, plugin.Config))
				}
			}
		}
	}

	for _, consumerGroup := range targetContent.ConsumerGroups {
		for _, plugin := range consumerGroup.Plugins {
			err := checkPlugin(plugin.Name, plugin.Config)
			if err != nil {
				errs = append(errs, checkPlugin(plugin.Name, plugin.Config))
			}
		}
	}

	for _, service := range targetContent.Services {
		for _, plugin := range service.Plugins {
			if plugin.Enabled != nil && *plugin.Enabled && plugin.Config != nil {
				err := checkPlugin(plugin.Name, plugin.Config)
				if err != nil {
					errs = append(errs, checkPlugin(plugin.Name, plugin.Config))
				}
			}
		}
	}

	for _, service := range targetContent.Services {
		for _, route := range service.Routes {
			for _, plugin := range route.Plugins {
				if plugin.Enabled != nil && *plugin.Enabled && plugin.Config != nil {
					err := checkPlugin(plugin.Name, plugin.Config)
					if err != nil {
						errs = append(errs, checkPlugin(plugin.Name, plugin.Config))
					}
				}
			}
		}
	}

	for _, route := range targetContent.Routes {
		for _, plugin := range route.Plugins {
			if plugin.Enabled != nil && *plugin.Enabled && plugin.Config != nil {
				err := checkPlugin(plugin.Name, plugin.Config)
				if err != nil {
					errs = append(errs, checkPlugin(plugin.Name, plugin.Config))
				}
			}
		}
	}

	return errs
}
