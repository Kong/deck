package stats

import (
	"bytes"
	"fmt"

	"github.com/jedib0t/go-pretty/table"
)

func generatePluginCountReport(stats ContentStatistics) (*bytes.Buffer, table.Writer) {
	pluginsBuf := new(bytes.Buffer)
	pluginsTable := table.NewWriter()
	pluginsTable.SetOutputMirror(pluginsBuf)
	pluginsTable.SortBy([]table.SortBy{
		{Name: "Count", Mode: table.DscNumeric},
	})
	pluginsTable.SetTitle("Plugin count per type")
	pluginsTable.AppendHeader(table.Row{"PluginName", "Count", "Ent", "OSS", "Custom/\nPartner"})

	for pluginName, pluginStat := range stats.PluginsCountPerName {
		enterprise := ""
		oss := ""
		custom := ""

		if pluginStat.Enterprise {
			enterprise = "*"
		}
		if pluginStat.OSS {
			oss = "*"
		}
		if pluginStat.CustomOrPartner {
			custom = "*"
		}

		pluginsTable.AppendRow(table.Row{pluginName, pluginStat.Count, enterprise, oss, custom})
	}

	pluginsTable.AppendFooter(table.Row{
		"Total Plugins",
		stats.Plugins,
		stats.EnterprisePlugins,
		stats.OSSPlugins,
		stats.CustomPlugins,
	})

	return pluginsBuf, pluginsTable
}

func generateEntityCountReport(stats ContentStatistics) (*bytes.Buffer, table.Writer) {
	entityCountsBuf := new(bytes.Buffer)
	entityCountsTable := table.NewWriter()
	entityCountsTable.SetOutputMirror(entityCountsBuf)
	//entityCountsTable.SetPageSize(10)
	entityCountsTable.SortBy([]table.SortBy{
		{Name: "Count", Mode: table.DscNumeric},
	})
	entityCountsTable.SetTitle("Workspace/Control Plane:\n" + stats.WorkspaceOrCPName)
	entityCountsTable.AppendHeader(table.Row{"Entity", "Count", "Percentage"})
	entityCountsTable.AppendRow(table.Row{"Services", stats.Services, fmt.Sprintf("%.2f%%", stats.ServicesPct)})
	entityCountsTable.AppendRow(table.Row{"Routes", stats.Routes, fmt.Sprintf("%.2f%%", stats.RoutesPct)})
	entityCountsTable.AppendRow(table.Row{"Consumers", stats.Consumers, fmt.Sprintf("%.2f%%", stats.ConsumersPct)})
	entityCountsTable.AppendRow(table.Row{"Consumer Groups", stats.ConsumerGroups, fmt.Sprintf("%.2f%%", stats.ConsumerGroupsPct)})
	entityCountsTable.AppendRow(table.Row{"Plugins", stats.Plugins, fmt.Sprintf("%.2f%%", stats.PluginsPct)})
	entityCountsTable.AppendRow(table.Row{"Upstreams", stats.Upstreams, fmt.Sprintf("%.2f%%", stats.UpstreamsPct)})
	entityCountsTable.AppendRow(table.Row{"Certificates", stats.Certificates, fmt.Sprintf("%.2f%%", stats.CertificatesPct)})
	entityCountsTable.AppendRow(table.Row{"CA Certificates", stats.CACertificates, fmt.Sprintf("%.2f%%", stats.CACertificatesPct)})
	entityCountsTable.AppendRow(table.Row{"Vaults", stats.Vaults, fmt.Sprintf("%.2f%%", stats.VaultsPct)})
	entityCountsTable.AppendFooter(table.Row{"Total Entities", stats.TotalEntities, ""})
	return entityCountsBuf, entityCountsTable
}
