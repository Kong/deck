package stats

import "github.com/kong/go-database-reconciler/pkg/file"

// calculateContentStatistics orchestrates the collection of statistics from a Kong configuration
func calculateContentStatistics(content *file.Content) ContentStatistics {
	stats := initializeStats(content)
	// Calculate percentages and other derived metrics
	calculateDerivedMetrics(&stats)

	return stats
}

// initializeStats initializes the ContentStatistics struct with basic counts
func initializeStats(content *file.Content) ContentStatistics {
	workspaceOrControlPlaneName := content.Workspace
	if workspaceOrControlPlaneName == "" && content.Konnect != nil {
		workspaceOrControlPlaneName = content.Konnect.ControlPlaneName
	}
	if workspaceOrControlPlaneName == "" {
		workspaceOrControlPlaneName = "default"
	}

	return ContentStatistics{
		WorkspaceOrCPName:   workspaceOrControlPlaneName,
		Services:            len(content.Services),
		Routes:              countRoutes(content),
		Plugins:             countPlugins(content),
		Consumers:           len(content.Consumers),
		ConsumerGroups:      len(content.ConsumerGroups),
		Upstreams:           len(content.Upstreams),
		Certificates:        len(content.Certificates),
		CACertificates:      len(content.CACertificates),
		Vaults:              len(content.Vaults),
		PluginsCountPerName: countPluginsGroupedByName(content),
	}
}

func countRoutes(content *file.Content) int {
	count := 0
	for _, service := range content.Services {
		count += len(service.Routes)
	}
	count += len(content.Routes)
	return count
}

func countPlugins(content *file.Content) int {
	count := 0
	// count plugins under services and routes
	for _, service := range content.Services {
		count += len(service.Plugins)
		for _, route := range service.Routes {
			count += len(route.Plugins)
		}
	}
	// count plugins under consumers and consumer groups
	for _, consumer := range content.Consumers {
		count += len(consumer.Plugins)
	}
	for _, group := range content.ConsumerGroups {
		count += len(group.Plugins)
	}
	// global plugins
	count += len(content.Plugins)
	return count
}

func countPluginsGroupedByName(content *file.Content) map[string]PluginStats {
	pluginsCountPerName := make(map[string]PluginStats)

	// Helper function to increment plugin count
	incrementPluginCount := func(pluginName string) {
		stats, exists := pluginsCountPerName[pluginName]
		if !exists {
			// First time seeing this plugin, determine its type
			stats = PluginStats{
				Count:      0,
				Enterprise: EnterprisePluginsMap[pluginName],
				OSS:        OSSPluginsMap[pluginName],
				CustomOrPartner:     !EnterprisePluginsMap[pluginName] && !OSSPluginsMap[pluginName],
			}
		}
		stats.Count++
		pluginsCountPerName[pluginName] = stats
	}

	// count plugins under services and routes
	for _, service := range content.Services {
		for _, plugin := range service.Plugins {
			incrementPluginCount(*plugin.Name)
		}
		for _, route := range service.Routes {
			for _, plugin := range route.Plugins {
				incrementPluginCount(*plugin.Name)
			}
		}
	}

	// count plugins under consumers and consumer groups
	for _, consumer := range content.Consumers {
		for _, plugin := range consumer.Plugins {
			incrementPluginCount(*plugin.Name)
		}
	}
	for _, group := range content.ConsumerGroups {
		for _, plugin := range group.Plugins {
			incrementPluginCount(*plugin.Name)
		}
	}

	// global plugins
	for _, plugin := range content.Plugins {
		incrementPluginCount(*plugin.Name)
	}
	return pluginsCountPerName
}

// calculateDerivedMetrics calculates percentages and other derived metrics
func calculateDerivedMetrics(stats *ContentStatistics) {
	stats.TotalEntities = stats.Services + stats.Routes + stats.Consumers +
		stats.ConsumerGroups + stats.Plugins + stats.Upstreams +
		stats.Certificates + stats.CACertificates

	if stats.TotalEntities > 0 {
		stats.ServicesPct = float64(stats.Services) / float64(stats.TotalEntities) * 100
		stats.RoutesPct = float64(stats.Routes) / float64(stats.TotalEntities) * 100
		stats.ConsumersPct = float64(stats.Consumers) / float64(stats.TotalEntities) * 100
		stats.ConsumerGroupsPct = float64(stats.ConsumerGroups) / float64(stats.TotalEntities) * 100
		stats.PluginsPct = float64(stats.Plugins) / float64(stats.TotalEntities) * 100
		stats.UpstreamsPct = float64(stats.Upstreams) / float64(stats.TotalEntities) * 100
		stats.CertificatesPct = float64(stats.Certificates) / float64(stats.TotalEntities) * 100
		stats.CACertificatesPct = float64(stats.CACertificates) / float64(stats.TotalEntities) * 100
		stats.VaultsPct = float64(stats.Vaults) / float64(stats.TotalEntities) * 100
	}
	// Calculate plugin type counts
	stats.OSSPlugins = 0
	stats.EnterprisePlugins = 0
	stats.CustomPlugins = 0
	for _, pluginStat := range stats.PluginsCountPerName {
		if pluginStat.OSS {
			stats.OSSPlugins++
		}
		if pluginStat.Enterprise {
			stats.EnterprisePlugins++
		}
		if pluginStat.CustomOrPartner {
			stats.CustomPlugins++
		}
	}
}
