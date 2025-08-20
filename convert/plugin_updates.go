package convert

import (
	"fmt"
	"strings"

	"github.com/kong/go-database-reconciler/pkg/cprint"
	"github.com/kong/go-database-reconciler/pkg/file"
	"github.com/kong/go-kong/kong"
)

const (
	rlaNamespaceDefaultLength = 32
)

func generateAutoFields(content *file.Content) error {
	for _, plugin := range content.Plugins {
		if *plugin.Name == rateLimitingAdvancedPluginName {
			plugin := plugin
			if err := autoGenerateNamespaceForRLAPlugin(&plugin); err != nil {
				return err
			}
		}
	}

	for _, service := range content.Services {
		for _, plugin := range service.Plugins {
			if *plugin.Name == rateLimitingAdvancedPluginName {
				if err := autoGenerateNamespaceForRLAPlugin(plugin); err != nil {
					return err
				}
			}
		}
	}

	for _, route := range content.Routes {
		for _, plugin := range route.Plugins {
			if *plugin.Name == rateLimitingAdvancedPluginName {
				if err := autoGenerateNamespaceForRLAPlugin(plugin); err != nil {
					return err
				}
			}
		}
	}

	for _, consumer := range content.Consumers {
		for _, plugin := range consumer.Plugins {
			if *plugin.Name == rateLimitingAdvancedPluginName {
				if err := autoGenerateNamespaceForRLAPlugin(plugin); err != nil {
					return err
				}
			}
		}
	}

	for _, consumerGroup := range content.ConsumerGroups {
		for _, plugin := range consumerGroup.Plugins {
			if *plugin.Name == rateLimitingAdvancedPluginName {
				if err := autoGenerateNamespaceForRLAPluginConsumerGroups(plugin); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func autoGenerateNamespaceForRLAPlugin(plugin *file.FPlugin) error {
	if plugin.Config != nil {
		ns, ok := plugin.Config["namespace"]
		if !ok || ns == nil {
			// namespace is not set, generate one.
			randomNamespace, err := randomString(rlaNamespaceDefaultLength)
			if err != nil {
				return fmt.Errorf("error generating random namespace: %w", err)
			}
			plugin.Config["namespace"] = randomNamespace
		}
	}
	return nil
}

func autoGenerateNamespaceForRLAPluginConsumerGroups(plugin *kong.ConsumerGroupPlugin) error {
	if plugin.Config != nil {
		ns, ok := plugin.Config["namespace"]
		if !ok || ns == nil {
			// namespace is not set, generate one.
			randomNamespace, err := randomString(rlaNamespaceDefaultLength)
			if err != nil {
				return fmt.Errorf("error generating random namespace: %w", err)
			}
			plugin.Config["namespace"] = randomNamespace
		}
	}
	return nil
}

func updatePlugins(content *file.Content) {
	for idx := range content.Plugins {
		plugin := &content.Plugins[idx]
		updateLegacyPluginConfig(plugin)
	}

	for _, service := range content.Services {
		for _, plugin := range service.Plugins {
			updateLegacyPluginConfig(plugin)
		}

		for _, route := range service.Routes {
			for _, plugin := range route.Plugins {
				updateLegacyPluginConfig(plugin)
			}
		}
	}

	for _, route := range content.Routes {
		for _, plugin := range route.Plugins {
			updateLegacyPluginConfig(plugin)
		}
	}

	for _, consumer := range content.Consumers {
		for _, plugin := range consumer.Plugins {
			updateLegacyPluginConfig(plugin)
		}
	}

	for _, consumerGroup := range content.ConsumerGroups {
		for _, plugin := range consumerGroup.Plugins {
			updateLegacyPluginConfig(&file.FPlugin{
				Plugin: kong.Plugin{
					ID:     plugin.ID,
					Name:   plugin.Name,
					Config: plugin.Config,
				},
			})
		}
	}
}

func updateLegacyPluginConfig(plugin *file.FPlugin) {
	if plugin != nil && plugin.Config != nil {
		config := plugin.Config.DeepCopy()

		var pluginName string
		if plugin.Name != nil {
			pluginName = *plugin.Name
		}

		config = updateLegacyFieldToNewField(config, "blacklist", "deny", pluginName)

		config = updateLegacyFieldToNewField(config, "whitelist", "allow", pluginName)

		if pluginName != "" {
			if pluginName == awsLambdaPluginName {
				config = removeDeprecatedFields3x(config, "proxy_scheme", pluginName)
			}
			if pluginName == prefunctionPluginName || pluginName == postfunctionPluginName {
				config = updateLegacyFieldToNewField(config, "functions", "access", pluginName)
			}
		}

		plugin.Config = config
	}
}

func updateLegacyFieldToNewField(pluginConfig kong.Configuration,
	oldField, newField, pluginName string,
) kong.Configuration {
	oldKeys := strings.Split(oldField, ".")
	newKeys := strings.Split(newField, ".")

	// Traverse to the old field's parent map
	current := pluginConfig
	for _, key := range oldKeys[:len(oldKeys)-1] {
		if nested, ok := current[key].(map[string]interface{}); ok {
			current = nested
		} else {
			return pluginConfig
		}
	}

	// Get the value of the old field
	oldKey := oldKeys[len(oldKeys)-1]
	value, ok := current[oldKey]
	if !ok {
		return pluginConfig
	}

	// Remove the old field
	delete(current, oldKey)

	// Traverse to the new field's parent map
	current = pluginConfig
	for _, key := range newKeys[:len(newKeys)-1] {
		if nested, ok := current[key].(map[string]interface{}); ok {
			current = nested
		} else {
			// Create nested map if it doesn't exist
			newMap := make(map[string]interface{})
			current[key] = newMap
			current = newMap
		}
	}

	// Set the value to the new field
	newKey := newKeys[len(newKeys)-1]
	current[newKey] = value

	cprint.UpdatePrintf("Automatically converted legacy configuration field \"%s\""+
		" to the new field \"%s\" in plugin %s\n",
		oldField, newField, pluginName)

	return pluginConfig
}

func removeDeprecatedFields3x(pluginConfig kong.Configuration, fieldName, pluginName string) kong.Configuration {
	if _, ok := pluginConfig[fieldName]; ok {
		delete(pluginConfig, fieldName)
		cprint.UpdatePrintf("Automatically removed deprecated config field \"%s\" from plugin %s\n", fieldName, pluginName)
	}
	return pluginConfig
}

func convertScalarToList(pluginConfig kong.Configuration,
	fieldName string, pluginName string,
) kong.Configuration {
	keys := strings.Split(fieldName, ".")
	current := pluginConfig
	for i, key := range keys {
		if i == len(keys)-1 {
			// Last key, perform the conversion
			if _, ok := current[key]; ok {
				switch v := current[key].(type) {
				case string:
					current[key] = []string{v}
				case int:
					current[key] = []int{v}
				case float64:
					current[key] = []float64{v}
				default:
					cprint.DeletePrintf("ERROR: Unexpected type for field \"%s\" in plugin %s: %T\n", fieldName, pluginName, v)
				}
			}
		} else {
			// Step into the nested map
			if nested, ok := current[key].(map[string]interface{}); ok {
				current = nested
				cprint.UpdatePrintf("Automatically converted configuration field \"%s\" from a single value "+
					"to a list in plugin %s\n", fieldName, pluginName)
			} else {
				cprint.UpdatePrintf("Field \"%s\" in plugin %s is not a nested object as expected\n", fieldName, pluginName)
			}
		}
	}

	return pluginConfig
}
