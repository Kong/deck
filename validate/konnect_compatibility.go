package validate

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/kong/go-database-reconciler/pkg/dump"
	"github.com/kong/go-database-reconciler/pkg/file"
	"github.com/kong/go-kong/kong"
)

var (
	errKonnect            = "[konnect] section not specified - ensure details are set via cli flags"
	errWorkspace          = "[workspaces] not supported by Konnect - use control planes instead"
	errNoVersion          = "[version] unable to determine decK file version"
	errBadVersion         = fmt.Sprintf("[version] decK file version must be '%.1f' or greater", supportedVersion)
	errPluginIncompatible = "[%s] plugin is not compatible with Konnect"
	errPluginNoCluster    = "[%s] plugin can't be used with cluster strategy"
)

var supportedVersion = 3.0

func checkPlugin(name *string, config kong.Configuration) error {
	switch *name {
	case "jwt-signer", "vault-auth", "oauth2":
		return fmt.Errorf(errPluginIncompatible, *name)
	case "application-registration":
		return fmt.Errorf("[%s] available in Konnect, but doesn't require this plugin", *name)
	case "key-auth-enc":
		return fmt.Errorf("[%s] keys are automatically encrypted in Konnect, use the key auth plugin instead", *name)
	case "openwhisk":
		return fmt.Errorf("[%s] plugin not bundled with Kong Gateway - installed as a LuaRocks package", *name)
	case "rate-limiting", "rate-limiting-advanced", "response-ratelimiting", "graphql-rate-limiting-advanced":
		if config["strategy"] == "cluster" {
			return fmt.Errorf(errPluginNoCluster, *name)
		}
	default:
	}
	return nil
}

func KonnectCompatibility(targetContent *file.Content, dumpConfig dump.Config) []error {
	var errs []error

	if targetContent.Konnect == nil && dumpConfig.KonnectControlPlane == "" {
		errs = append(errs, errors.New(errKonnect))
	}

	versionNumber, err := strconv.ParseFloat(targetContent.FormatVersion, 32)
	if err != nil {
		errs = append(errs, errors.New(errNoVersion))
	} else {
		if versionNumber < supportedVersion {
			errs = append(errs, errors.New(errBadVersion))
		}
	}

	for _, plugin := range targetContent.Plugins {
		if plugin.Enabled != nil && *plugin.Enabled && plugin.Config != nil {
			err := checkPlugin(plugin.Name, plugin.Config)
			if err != nil {
				errs = append(errs, err)
			}
		}
	}

	for _, consumer := range targetContent.Consumers {
		for _, plugin := range consumer.Plugins {
			if plugin.Enabled != nil && *plugin.Enabled && plugin.Config != nil {
				err := checkPlugin(plugin.Name, plugin.Config)
				if err != nil {
					errs = append(errs, err)
				}
			}
		}
	}

	for _, consumerGroup := range targetContent.ConsumerGroups {
		for _, plugin := range consumerGroup.Plugins {
			err := checkPlugin(plugin.Name, plugin.Config)
			if err != nil {
				errs = append(errs, err)
			}
		}
	}

	for _, service := range targetContent.Services {
		for _, plugin := range service.Plugins {
			if plugin.Enabled != nil && *plugin.Enabled && plugin.Config != nil {
				err := checkPlugin(plugin.Name, plugin.Config)
				if err != nil {
					errs = append(errs, err)
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
						errs = append(errs, err)
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
					errs = append(errs, err)
				}
			}
		}
	}

	return errs
}
