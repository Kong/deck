package cmd

import (
	"fmt"
	"io"
	"os"
	"slices"

	"github.com/kong/go-database-reconciler/pkg/file"
	"github.com/spf13/cobra"
)

// aiGatewayDocsURL documents the supported way to manage AI Gateway
// configuration on Konnect.
const aiGatewayDocsURL = "https://developer.konghq.com/ai-gateway/kongctl"

func newAiSubCmd() *cobra.Command {
	aiSubCmd := &cobra.Command{
		Use:   "ai [sub-command]...",
		Short: "Subcommand to host decK AI Gateway operations",
		Long:  `Subcommand to host decK AI Gateway operations.`,
	}

	return aiSubCmd
}

// printAIWarnings writes conversion warnings, one per line, to w. When w is nil
// it defaults to os.Stderr. It is a no-op when there are no warnings.
func printAIWarnings(w io.Writer, warnings []string) {
	if len(warnings) == 0 {
		return
	}
	if w == nil {
		w = os.Stderr
	}
	for _, warning := range warnings {
		fmt.Fprintf(w, "Warning: %v\n", warning)
	}
}

// errAIManagedEntitiesOnKonnect is returned when a user attempts to sync/apply
// AI Gateway entities (tagged managed_by:deck-ai) to Konnect using decK. Those
// entities must be managed with kongctl instead.
func errAIManagedEntitiesOnKonnect() error {
	return fmt.Errorf(
		"AI Gateway entities (tagged %q) cannot be managed on Konnect using decK.\n"+
			"Use kongctl to manage AI Gateway configuration on Konnect instead: %s",
		managedByAIDeckTag, aiGatewayDocsURL)
}

// contentHasmanagedByAIDeckTag reports whether the configuration is marked as managed
// by the AI Gateway tooling, either via a select tag in _info or via the tags
// of any individual entity.
//
// AI Gateway configuration only ever contains services, routes, plugins,
// consumers, consumer groups, vaults and AI models (plus the plugins/routes
// nested under them), so those collections are checked explicitly instead of
// reflecting over the whole state file.
func contentHasmanagedByAIDeckTag(content *file.Content) bool {
	if content == nil {
		return false
	}
	if content.Info != nil && slices.Contains(content.Info.SelectorTags, managedByAIDeckTag) {
		return true
	}

	for i := range content.Services {
		s := &content.Services[i]
		if hasAIDeckTag(s.Tags) || fpluginsHaveTag(s.Plugins) {
			return true
		}
		for _, r := range s.Routes {
			if r != nil && (hasAIDeckTag(r.Tags) || fpluginsHaveTag(r.Plugins)) {
				return true
			}
		}
	}

	for i := range content.Routes {
		r := &content.Routes[i]
		if hasAIDeckTag(r.Tags) || fpluginsHaveTag(r.Plugins) {
			return true
		}
	}

	for i := range content.Plugins {
		if hasAIDeckTag(content.Plugins[i].Tags) {
			return true
		}
	}

	for i := range content.Consumers {
		c := &content.Consumers[i]
		if hasAIDeckTag(c.Tags) || fpluginsHaveTag(c.Plugins) {
			return true
		}
	}

	for i := range content.ConsumerGroups {
		cg := &content.ConsumerGroups[i]
		if hasAIDeckTag(cg.Tags) {
			return true
		}
		for _, c := range cg.Consumers {
			if c != nil && hasAIDeckTag(c.Tags) {
				return true
			}
		}
		for _, p := range cg.Plugins {
			if p != nil && hasAIDeckTag(p.Tags) {
				return true
			}
		}
	}

	for i := range content.Vaults {
		if hasAIDeckTag(content.Vaults[i].Tags) {
			return true
		}
	}

	for i := range content.AIModels {
		if hasAIDeckTag(content.AIModels[i].Tags) {
			return true
		}
	}

	return false
}

// fpluginsHaveTag reports whether any plugin in plugins carries managedByAIDeckTag.
func fpluginsHaveTag(plugins []*file.FPlugin) bool {
	for _, p := range plugins {
		if p != nil && hasAIDeckTag(p.Tags) {
			return true
		}
	}
	return false
}

// hasAIDeckTag reports whether a []*string Tags field contains managedByAIDeckTag.
func hasAIDeckTag(tags []*string) bool {
	for _, tag := range tags {
		if tag != nil && *tag == managedByAIDeckTag {
			return true
		}
	}
	return false
}
