package stats

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/kong/go-database-reconciler/pkg/file"
)

type OutputFormat string

const (
	TextFormat     OutputFormat = "txt"
	CSVFormat      OutputFormat = "csv"
	HTMLFormat     OutputFormat = "html"
	MarkdownFormat OutputFormat = "md"
)

type ContentStatistics struct {
	ServicesCount       int                      `json:"servicesCount"`
	RoutesCount         int                      `json:"routesCount"`
	ConsumersCount      int                      `json:"consumersCount"`
	ConsumerGroupsCount int                      `json:"consumerGroupsCount"`
	PluginsCount        int                      `json:"pluginsCount"`
	UpstreamsCount      int                      `json:"upstreamsCount"`
	CertificatesCount   int                      `json:"certificatesCount"`
	CACertificatesCount int                      `json:"caCertificatesCount"`
	TotalEntities       int                      `json:"totalEntities"`
	ServicesPct         float64                  `json:"servicesPct"`
	RoutesPct           float64                  `json:"routesPct"`
	ConsumersPct        float64                  `json:"consumersPct"`
	ConsumerGroupsPct   float64                  `json:"consumerGroupsPct"`
	PluginsPct          float64                  `json:"pluginsPct"`
	UpstreamsPct        float64                  `json:"upstreamsPct"`
	CertificatesPct     float64                  `json:"certificatesPct"`
	CACertificatesPct   float64                  `json:"caCertificatesPct"`
	UniqueTags          []string                 `json:"uniqueTags"`
	TagStats            map[string]TagStatistics `json:"tagStats"`
	RegexRoutesCount    int                      `json:"regexRoutesCount"`
	PathRoutesCount     int                      `json:"PathRoutesCount"`
	RegexRoutesPct      float64                  `json:"RegexRoutesPct"`
	AvgPathsPerRoute    float64                  `json:"AvgPathsPerRoute"`
	PluginsCountPerName map[string]int           `json:"pluginsCountPerName"`
}

type TagStatistics struct {
	ServicesCount       int `json:"servicesCount"`
	RoutesCount         int `json:"routesCount"`
	ConsumersCount      int `json:"consumersCount"`
	ConsumerGroupsCount int `json:"consumerGroupsCount"`
	PluginsCount        int `json:"pluginsCount"`
	UpstreamsCount      int `json:"upstreamsCount"`
	CertificatesCount   int `json:"certificatesCount"`
	CACertificatesCount int `json:"caCertificatesCount"`
}

func filterContentByTags(content *file.Content, selectorTags []string) *file.Content {
	if len(selectorTags) == 0 {
		return content
	}
	v := reflect.ValueOf(content).Elem()

	fieldsToFilter := map[string]bool{
		"Services":       true,
		"Routes":         true,
		"Consumers":      true,
		"ConsumerGroups": true,
		"Plugins":        true,
		"Upstreams":      true,
		"Certificates":   true,
		"CACertificates": true,
	}

	consumerFieldsToFilter := map[string]bool{
		"Plugins":     true,
		"KeyAuths":    true,
		"HMACAuths":   true,
		"JWTAuths":    true,
		"BasicAuths":  true,
		"Oauth2Creds": true,
		"ACLGroups":   true,
		"MTLSAuths":   true,
	}

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldName := v.Type().Field(i).Name

		if fieldsToFilter[fieldName] && field.Kind() == reflect.Slice {
			filtered := filterByTags(field.Interface(), selectorTags)
			filteredSlice := reflect.MakeSlice(field.Type(), len(filtered.([]interface{})), len(filtered.([]interface{})))
			for i, val := range filtered.([]interface{}) {
				filteredSlice.Index(i).Set(reflect.ValueOf(val))
			}
			field.Set(filteredSlice)
		}

		if fieldName == "Consumers" && field.Kind() == reflect.Slice {
			for j := 0; j < field.Len(); j++ {
				consumer := field.Index(j)
				for k := 0; k < consumer.NumField(); k++ {
					consumerField := consumer.Field(k)
					consumerFieldName := consumer.Type().Field(k).Name

					if consumerFieldsToFilter[consumerFieldName] && consumerField.Kind() == reflect.Slice {
						filtered := filterByTags(consumerField.Interface(), selectorTags)
						filteredSlice := reflect.MakeSlice(consumerField.Type(), len(filtered.([]interface{})), len(filtered.([]interface{})))
						for i, val := range filtered.([]interface{}) {
							filteredSlice.Index(i).Set(reflect.ValueOf(val))
						}
						consumerField.Set(filteredSlice)
					}
				}
			}
		}
	}

	return content
}

func filterByTags(slice interface{}, selectorTags []string) interface{} {
	if len(selectorTags) == 0 {
		return slice
	}
	s := reflect.ValueOf(slice)

	if s.Kind() != reflect.Slice {
		panic("filterByTags() given a non-slice type")
	}

	result := make([]interface{}, 0, s.Len())

	for i := 0; i < s.Len(); i++ {
		elem := s.Index(i).Interface()
		elemValue := reflect.ValueOf(elem)

		tagsField := elemValue.FieldByName("Tags")
		if !tagsField.IsValid() {
			panic("filterByTags() given a slice of structs without a Tags field")
		}

		tags := tagsField.Interface().([]*string)
		if containsAllTags(tags, selectorTags) {
			result = append(result, elem)
		}
	}

	return result
}

func containsAllTags(tags []*string, selectorTags []string) bool {
	for _, selectorTag := range selectorTags {
		if !containsTag(tags, selectorTag) {
			return false
		}
	}
	return true
}

func containsTag(tags []*string, tag string) bool {
	for _, t := range tags {
		if *t == tag {
			return true
		}
	}
	return false
}

func calculateContentStatistics(content *file.Content) ContentStatistics {
	stats := ContentStatistics{
		ServicesCount:       len(content.Services),
		ConsumersCount:      len(content.Consumers),
		ConsumerGroupsCount: len(content.ConsumerGroups),
		UpstreamsCount:      len(content.Upstreams),
		CertificatesCount:   len(content.Certificates),
		CACertificatesCount: len(content.CACertificates),
	}

	// unique tags
	uniqueTagsMap := make(map[string]struct{})
	addTags := func(tags []*string) {
		for _, tag := range tags {
			uniqueTagsMap[*tag] = struct{}{}
		}
	}

	// entities per tag
	tagStats := make(map[string]TagStatistics)
	incrementCount := func(tags []*string, entity string) {
		for _, tag := range tags {
			stats := tagStats[*tag]
			switch entity {
			case "service":
				stats.ServicesCount++
			case "route":
				stats.RoutesCount++
			case "consumer":
				stats.ConsumersCount++
			case "consumerGroup":
				stats.ConsumerGroupsCount++
			case "plugin":
				stats.PluginsCount++
			case "upstream":
				stats.UpstreamsCount++
			case "certificate":
				stats.CertificatesCount++
			case "caCertificate":
				stats.CACertificatesCount++
			}
			tagStats[*tag] = stats
		}
	}

	// plugin count per name
	stats.PluginsCountPerName = make(map[string]int)

	for _, service := range content.Services {
		addTags(service.Tags)
		incrementCount(service.Tags, "service")
		for _, route := range service.Routes {
			stats.RoutesCount++
			addTags(route.Tags)
			incrementCount(route.Tags, "route")
			for _, plugin := range route.Plugins {
				stats.PluginsCount++
				addTags(plugin.Tags)
				incrementCount(plugin.Tags, "plugin")
				stats.PluginsCountPerName[*plugin.Name]++
			}
			for _, path := range route.Paths {
				stats.PathRoutesCount++
				if strings.HasPrefix(*path, "~") {
					stats.RegexRoutesCount++
				}
			}
		}
		for _, plugin := range service.Plugins {
			stats.PluginsCount++
			addTags(plugin.Tags)
			incrementCount(plugin.Tags, "plugin")
			stats.PluginsCountPerName[*plugin.Name]++
		}
		for _, upstream := range content.Upstreams {
			addTags(upstream.Tags)
			incrementCount(upstream.Tags, "upstream")
		}
	}
	for _, consumer := range content.Consumers {
		addTags(consumer.Tags)
		incrementCount(consumer.Tags, "consumer")
		for _, plugin := range consumer.Plugins {
			stats.PluginsCount++
			addTags(plugin.Tags)
			incrementCount(plugin.Tags, "plugin")
			stats.PluginsCountPerName[*plugin.Name]++
		}
	}
	for _, consumerGroup := range content.ConsumerGroups {
		addTags(consumerGroup.Tags)
		incrementCount(consumerGroup.Tags, "consumerGroup")
	}
	for _, plugin := range content.Plugins {
		stats.PluginsCount++
		addTags(plugin.Tags)
		incrementCount(plugin.Tags, "plugin")
		stats.PluginsCountPerName[*plugin.Name]++
	}
	for _, certificate := range content.Certificates {
		addTags(certificate.Tags)
		incrementCount(certificate.Tags, "certificate")
	}
	for _, caCertificate := range content.CACertificates {
		addTags(caCertificate.Tags)
		incrementCount(caCertificate.Tags, "caCertificate")
	}

	stats.UniqueTags = make([]string, 0, len(uniqueTagsMap))
	for tag := range uniqueTagsMap {
		stats.UniqueTags = append(stats.UniqueTags, tag)
	}

	stats.TagStats = tagStats

	stats.TotalEntities = stats.ServicesCount + stats.RoutesCount + stats.ConsumersCount +
		stats.ConsumerGroupsCount + stats.PluginsCount + stats.UpstreamsCount +
		stats.CertificatesCount + stats.CACertificatesCount

	stats.ServicesPct = float64(stats.ServicesCount) / float64(stats.TotalEntities) * 100
	stats.RoutesPct = float64(stats.RoutesCount) / float64(stats.TotalEntities) * 100
	stats.ConsumersPct = float64(stats.ConsumersCount) / float64(stats.TotalEntities) * 100
	stats.ConsumerGroupsPct = float64(stats.ConsumerGroupsCount) / float64(stats.TotalEntities) * 100
	stats.PluginsPct = float64(stats.PluginsCount) / float64(stats.TotalEntities) * 100
	stats.UpstreamsPct = float64(stats.UpstreamsCount) / float64(stats.TotalEntities) * 100
	stats.CertificatesPct = float64(stats.CertificatesCount) / float64(stats.TotalEntities) * 100
	stats.CACertificatesPct = float64(stats.CACertificatesCount) / float64(stats.TotalEntities) * 100
	stats.RegexRoutesPct = float64(stats.RegexRoutesCount) / float64(stats.PathRoutesCount) * 100
	stats.AvgPathsPerRoute = float64(stats.PathRoutesCount) / float64(stats.RoutesCount)

	return stats
}

func getTableStyle(styleName string) table.Style {
	switch styleName {
	case "StyleDefault":
		return table.StyleDefault
	case "StyleBold":
		return table.StyleBold
	case "StyleColoredBright":
		return table.StyleColoredBright
	case "StyleColoredDark":
		return table.StyleColoredDark
	case "StyleColoredBlackOnBlueWhite":
		return table.StyleColoredBlackOnBlueWhite
	case "StyleColoredBlackOnCyanWhite":
		return table.StyleColoredBlackOnCyanWhite
	case "StyleColoredBlackOnGreenWhite":
		return table.StyleColoredBlackOnGreenWhite
	case "StyleColoredBlackOnMagentaWhite":
		return table.StyleColoredBlackOnMagentaWhite
	case "StyleColoredBlackOnRedWhite":
		return table.StyleColoredBlackOnRedWhite
	case "StyleColoredBlackOnYellowWhite":
		return table.StyleColoredBlackOnYellowWhite
	case "StyleColoredBlueWhiteOnBlack":
		return table.StyleColoredBlueWhiteOnBlack
	case "StyleColoredCyanWhiteOnBlack":
		return table.StyleColoredCyanWhiteOnBlack
	case "StyleColoredGreenWhiteOnBlack":
		return table.StyleColoredGreenWhiteOnBlack
	case "StyleColoredMagentaWhiteOnBlack":
		return table.StyleColoredMagentaWhiteOnBlack
	case "StyleColoredRedWhiteOnBlack":
		return table.StyleColoredRedWhiteOnBlack
	case "StyleColoredYellowWhiteOnBlack":
		return table.StyleColoredYellowWhiteOnBlack
	case "StyleDouble":
		return table.StyleDouble
	case "StyleLight":
		return table.StyleLight
	case "StyleRounded":
		return table.StyleRounded
	default:
		return table.StyleDefault
	}
}

// print struct ContentStatistics using three tables
// 1. entity counts and percentages
// 2. for each unique tag, entity counts and percentages
// 3. for each plugin name, plugin counts
func PrintContentStatistics(content *file.Content, style string, format string, includeTags bool, selectedTags []string) ([]byte, error) {
	// filter content by tags if selectedTags is not empty
	if len(selectedTags) > 0 {
		content = filterContentByTags(content, selectedTags)
	}
	stats := calculateContentStatistics(content)

	// 1. entity counts and percentages
	entityCountsBuf := new(bytes.Buffer)
	entityCountsTable := table.NewWriter()
	entityCountsTable.SetStyle(getTableStyle(style))
	entityCountsTable.SetOutputMirror(entityCountsBuf)
	//entityCountsTable.SetPageSize(10)
	entityCountsTable.SortBy([]table.SortBy{
		{Name: "Count", Mode: table.DscNumeric},
	})
	entityCountsTable.SetTitle("Entity count and percentage")
	entityCountsTable.AppendHeader(table.Row{"Entity", "Count", "Percentage"})
	entityCountsTable.AppendRow(table.Row{"Services", stats.ServicesCount, fmt.Sprintf("%.2f%%", stats.ServicesPct)})
	entityCountsTable.AppendRow(table.Row{"Routes", stats.RoutesCount, fmt.Sprintf("%.2f%%", stats.RoutesPct)})
	entityCountsTable.AppendRow(table.Row{"Consumers", stats.ConsumersCount, fmt.Sprintf("%.2f%%", stats.ConsumersPct)})
	entityCountsTable.AppendRow(table.Row{"Consumer Groups", stats.ConsumerGroupsCount, fmt.Sprintf("%.2f%%", stats.ConsumerGroupsPct)})
	entityCountsTable.AppendRow(table.Row{"Plugins", stats.PluginsCount, fmt.Sprintf("%.2f%%", stats.PluginsPct)})
	entityCountsTable.AppendRow(table.Row{"Upstreams", stats.UpstreamsCount, fmt.Sprintf("%.2f%%", stats.UpstreamsPct)})
	entityCountsTable.AppendRow(table.Row{"Certificates", stats.CertificatesCount, fmt.Sprintf("%.2f%%", stats.CertificatesPct)})
	entityCountsTable.AppendRow(table.Row{"CA Certificates", stats.CACertificatesCount, fmt.Sprintf("%.2f%%", stats.CACertificatesPct)})
	entityCountsTable.AppendFooter(table.Row{"Total Entities", stats.TotalEntities, ""})

	regexCountsBuf := new(bytes.Buffer)
	regexCountsTable := table.NewWriter()
	regexCountsTable.SetStyle(getTableStyle(style))
	regexCountsTable.SetOutputMirror(regexCountsBuf)
	regexCountsTable.SetTitle("Paths with regex count and percentatge")
	regexCountsTable.AppendRow(table.Row{"Total paths in routes", stats.PathRoutesCount})
	regexCountsTable.AppendRow(table.Row{"Average paths per route", fmt.Sprintf("%.2f", stats.AvgPathsPerRoute)})
	regexCountsTable.AppendRow(table.Row{"Total paths with regular expressions in routes", stats.RegexRoutesCount})
	regexCountsTable.AppendFooter(table.Row{"Percentage of paths with regular expressions in routes", fmt.Sprintf("%.2f%%", stats.RegexRoutesPct)})

	// 2. for each unique tag, entity counts and percentages
	tagsBuf := new(bytes.Buffer)
	tagsTable := table.NewWriter()
	tagsTable.SetStyle(getTableStyle(style))
	tagsTable.SetOutputMirror(tagsBuf)
	tagsTable.SetPageSize(10)
	tagsTable.SortBy([]table.SortBy{
		{Name: "Services", Mode: table.DscNumeric},
		{Name: "Routes", Mode: table.DscNumeric},
	})
	tagsTable.SetTitle("Entity count per tag")
	tagsTable.AppendHeader(table.Row{"Tag", "Services", "Routes", "Consumers", "Consumer Groups", "Plugins", "Upstreams", "Certificates", "CA Certificates"})
	for _, tag := range stats.UniqueTags {
		tagsTable.AppendRow(table.Row{
			tag,
			stats.TagStats[tag].ServicesCount,
			stats.TagStats[tag].RoutesCount,
			stats.TagStats[tag].ConsumersCount,
			stats.TagStats[tag].ConsumerGroupsCount,
			stats.TagStats[tag].PluginsCount,
			stats.TagStats[tag].UpstreamsCount,
			stats.TagStats[tag].CertificatesCount,
			stats.TagStats[tag].CACertificatesCount,
		})
	}

	// 3. for each plugin name, plugin counts
	pluginsBuf := new(bytes.Buffer)
	pluginsTable := table.NewWriter()
	pluginsTable.SetStyle(getTableStyle(style))
	pluginsTable.SetOutputMirror(pluginsBuf)
	//pluginsTable.SetPageSize(10)
	pluginsTable.SortBy([]table.SortBy{
		{Name: "Count", Mode: table.DscNumeric},
	})
	pluginsTable.SetTitle("Plugin count per type")
	pluginsTable.AppendHeader(table.Row{"PluginName", "Count"})
	for pluginName, count := range stats.PluginsCountPerName {
		pluginsTable.AppendRow(table.Row{pluginName, count})
	}
	pluginsTable.AppendFooter(table.Row{"Total Plugins", stats.PluginsCount})

	switch OutputFormat(format) {
	case TextFormat:
		entityCountsTable.Render()
		regexCountsTable.Render()
		tagsTable.Render()
		pluginsTable.Render()
	case CSVFormat:
		entityCountsTable.RenderCSV()
		regexCountsTable.RenderCSV()
		tagsTable.RenderCSV()
		pluginsTable.RenderCSV()
	case HTMLFormat:
		entityCountsTable.RenderHTML()
		regexCountsTable.RenderHTML()
		tagsTable.RenderHTML()
		pluginsTable.RenderHTML()
	case MarkdownFormat:
		entityCountsTable.RenderMarkdown()
		regexCountsTable.RenderMarkdown()
		tagsTable.RenderMarkdown()
		pluginsTable.RenderMarkdown()
	default:
		return nil, fmt.Errorf("invalid format '%s'", format)
	}

	// append entityCountsBuf, tagsBuf, pluginsBuf
	buf := new(bytes.Buffer)
	buf.Write(entityCountsBuf.Bytes())
	buf.Write(regexCountsBuf.Bytes())
	buf.Write(pluginsBuf.Bytes())
	if includeTags {
		buf.Write(tagsBuf.Bytes())
	}

	return buf.Bytes(), nil
}
