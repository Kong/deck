package stats

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"reflect"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/kong/go-database-reconciler/pkg/file"
	"gopkg.in/yaml.v2"
)

type OutputFormat string

const (
	TextFormat     OutputFormat = "txt"
	CSVFormat      OutputFormat = "csv"
	HTMLFormat     OutputFormat = "html"
	MarkdownFormat OutputFormat = "md"
	JsonFormat     OutputFormat = "json"
	YamlFormat     OutputFormat = "yaml"
)

type ContentStatistics struct {
	WorspaceOrCPName    string                   `json:"workspaceOrCPName"`
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
	//RegexRoutesCount    int                      `json:"regexRoutesCount"`
	PathRoutesCount int `json:"PathRoutesCount"`
	// RegexRoutesPct      float64                  `json:"RegexRoutesPct"`
	AvgPathsPerRoute    float64        `json:"AvgPathsPerRoute"`
	PluginsCountPerName map[string]int `json:"pluginsCountPerName"`
	OSSPluginsCount     int            `json:"ossPluginsCount"`
	EntPluginsCount     int            `json:"entPluginsCount"`
	CustomPluginsCount  int            `json:"customPluginsCount"`
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

// filter Kong entitites by tags
func filterByTags(slice interface{}, selectorTags []string) interface{} {
	if len(selectorTags) == 0 {
		return slice
	}
	s := reflect.ValueOf(slice)

	if s.Kind() != reflect.Slice {
		log.Fatal("filterByTags() given a non-slice type")
	}

	result := make([]interface{}, 0, s.Len())

	for i := 0; i < s.Len(); i++ {
		elem := s.Index(i).Interface()
		elemValue := reflect.ValueOf(elem)

		tagsField := elemValue.FieldByName("Tags")
		if !tagsField.IsValid() {
			log.Fatal("filterByTags() given a slice of structs without a Tags field")
		}

		tags := tagsField.Interface().([]*string)
		if containsAllTags(tags, selectorTags) {
			result = append(result, elem)
		}
	}

	return result
}

// check if all tags are present in the entity tags
func containsAllTags(tags []*string, selectorTags []string) bool {
	for _, selectorTag := range selectorTags {
		if !containsTag(tags, selectorTag) {
			return false
		}
	}
	return true
}

// check if a tag is present in the entity tags
func containsTag(tags []*string, tag string) bool {
	for _, t := range tags {
		if *t == tag {
			return true
		}
	}
	return false
}

// calculateContentStatistics orchestrates the collection of statistics from a Kong configuration
func calculateContentStatistics(content *file.Content) ContentStatistics {
	stats := initializeStats(content)

	// Process entities and collect data
	processEntities(content, &stats)

	// Calculate percentages and other derived metrics
	calculateDerivedMetrics(&stats)

	return stats
}

// initializeStats initializes the ContentStatistics struct with basic counts
func initializeStats(content *file.Content) ContentStatistics {
	workspaceOrControlPlaneName := content.Workspace
	if workspaceOrControlPlaneName == "" {
		workspaceOrControlPlaneName = content.Konnect.ControlPlaneName
		if workspaceOrControlPlaneName == "" {
			workspaceOrControlPlaneName = "unknown"
		}
	}

	return ContentStatistics{
		WorspaceOrCPName:    workspaceOrControlPlaneName,
		ServicesCount:       len(content.Services),
		ConsumersCount:      len(content.Consumers),
		ConsumerGroupsCount: len(content.ConsumerGroups),
		UpstreamsCount:      len(content.Upstreams),
		CertificatesCount:   len(content.Certificates),
		CACertificatesCount: len(content.CACertificates),
		PluginsCountPerName: make(map[string]int),
		TagStats:            make(map[string]TagStatistics),
	}
}

// processEntities processes all entities in the content and updates statistics
func processEntities(content *file.Content, stats *ContentStatistics) {
	uniqueTagsMap := make(map[string]struct{})

	// Process services and related entities
	processServices(content.Services, stats, uniqueTagsMap)

	// Process top-level entities
	processConsumers(content.Consumers, stats, uniqueTagsMap)
	processConsumerGroups(content.ConsumerGroups, stats, uniqueTagsMap)
	processPlugins(content.Plugins, stats, uniqueTagsMap)
	processRoutes(content.Routes, stats, uniqueTagsMap)
	processUpstreams(content.Upstreams, stats, uniqueTagsMap)
	processCertificates(content.Certificates, stats, uniqueTagsMap)
	processCACertificates(content.CACertificates, stats, uniqueTagsMap)

	// Convert uniqueTagsMap to slice
	stats.UniqueTags = mapKeysToSlice(uniqueTagsMap)
}

// processServices handles services and their nested entities (routes, plugins)
func processServices(services []file.FService, stats *ContentStatistics, uniqueTagsMap map[string]struct{}) {
	for _, service := range services {
		addTagsToMap(service.Tags, uniqueTagsMap)
		incrementTagCounts(service.Tags, "service", stats)

		// Process routes within service
		for _, route := range service.Routes {
			stats.RoutesCount++
			addTagsToMap(route.Tags, uniqueTagsMap)
			incrementTagCounts(route.Tags, "route", stats)

			// Process plugins within route
			for _, plugin := range route.Plugins {
				stats.PluginsCount++
				addTagsToMap(plugin.Tags, uniqueTagsMap)
				incrementTagCounts(plugin.Tags, "plugin", stats)
				stats.PluginsCountPerName[*plugin.Name]++
			}

			// Count paths in routes
			stats.PathRoutesCount += len(route.Paths)
		}

		// Process plugins within service
		for _, plugin := range service.Plugins {
			stats.PluginsCount++
			addTagsToMap(plugin.Tags, uniqueTagsMap)
			incrementTagCounts(plugin.Tags, "plugin", stats)
			stats.PluginsCountPerName[*plugin.Name]++
		}
	}
}

// processConsumers handles consumers and their plugins
func processConsumers(consumers []file.FConsumer, stats *ContentStatistics, uniqueTagsMap map[string]struct{}) {
	for _, consumer := range consumers {
		addTagsToMap(consumer.Tags, uniqueTagsMap)
		incrementTagCounts(consumer.Tags, "consumer", stats)

		// Process plugins within consumer
		for _, plugin := range consumer.Plugins {
			stats.PluginsCount++
			addTagsToMap(plugin.Tags, uniqueTagsMap)
			incrementTagCounts(plugin.Tags, "plugin", stats)
			stats.PluginsCountPerName[*plugin.Name]++
		}
	}
}

// Helper functions for other entity types
func processConsumerGroups(groups []file.FConsumerGroupObject, stats *ContentStatistics, uniqueTagsMap map[string]struct{}) {
	for _, group := range groups {
		addTagsToMap(group.Tags, uniqueTagsMap)
		incrementTagCounts(group.Tags, "consumerGroup", stats)
	}
}

func processPlugins(plugins []file.FPlugin, stats *ContentStatistics, uniqueTagsMap map[string]struct{}) {
	for _, plugin := range plugins {
		stats.PluginsCount++
		addTagsToMap(plugin.Tags, uniqueTagsMap)
		incrementTagCounts(plugin.Tags, "plugin", stats)
		stats.PluginsCountPerName[*plugin.Name]++
	}
}

func processRoutes(routes []file.FRoute, stats *ContentStatistics, uniqueTagsMap map[string]struct{}) {
	for _, route := range routes {
		stats.RoutesCount++
		addTagsToMap(route.Tags, uniqueTagsMap)
		incrementTagCounts(route.Tags, "route", stats)
	}
}

func processUpstreams(upstreams []file.FUpstream, stats *ContentStatistics, uniqueTagsMap map[string]struct{}) {
	for _, upstream := range upstreams {
		addTagsToMap(upstream.Tags, uniqueTagsMap)
		incrementTagCounts(upstream.Tags, "upstream", stats)
	}
}

func processCertificates(certificates []file.FCertificate, stats *ContentStatistics, uniqueTagsMap map[string]struct{}) {
	for _, certificate := range certificates {
		addTagsToMap(certificate.Tags, uniqueTagsMap)
		incrementTagCounts(certificate.Tags, "certificate", stats)
	}
}

func processCACertificates(caCertificates []file.FCACertificate, stats *ContentStatistics, uniqueTagsMap map[string]struct{}) {
	for _, caCertificate := range caCertificates {
		addTagsToMap(caCertificate.Tags, uniqueTagsMap)
		incrementTagCounts(caCertificate.Tags, "caCertificate", stats)
	}
}

// Helper functions for tag processing
func addTagsToMap(tags []*string, uniqueTagsMap map[string]struct{}) {
	for _, tag := range tags {
		if tag != nil {
			uniqueTagsMap[*tag] = struct{}{}
		}
	}
}

func incrementTagCounts(tags []*string, entityType string, stats *ContentStatistics) {
	for _, tag := range tags {
		if tag == nil {
			continue
		}

		tagStat := stats.TagStats[*tag]

		switch entityType {
		case "service":
			tagStat.ServicesCount++
		case "route":
			tagStat.RoutesCount++
		case "consumer":
			tagStat.ConsumersCount++
		case "consumerGroup":
			tagStat.ConsumerGroupsCount++
		case "plugin":
			tagStat.PluginsCount++
		case "upstream":
			tagStat.UpstreamsCount++
		case "certificate":
			tagStat.CertificatesCount++
		case "caCertificate":
			tagStat.CACertificatesCount++
		}

		stats.TagStats[*tag] = tagStat
	}
}

func mapKeysToSlice(m map[string]struct{}) []string {
	result := make([]string, 0, len(m))
	for key := range m {
		result = append(result, key)
	}
	return result
}

// calculateDerivedMetrics calculates percentages and other derived metrics
func calculateDerivedMetrics(stats *ContentStatistics) {
	stats.TotalEntities = stats.ServicesCount + stats.RoutesCount + stats.ConsumersCount +
		stats.ConsumerGroupsCount + stats.PluginsCount + stats.UpstreamsCount +
		stats.CertificatesCount + stats.CACertificatesCount

	if stats.TotalEntities > 0 {
		stats.ServicesPct = float64(stats.ServicesCount) / float64(stats.TotalEntities) * 100
		stats.RoutesPct = float64(stats.RoutesCount) / float64(stats.TotalEntities) * 100
		stats.ConsumersPct = float64(stats.ConsumersCount) / float64(stats.TotalEntities) * 100
		stats.ConsumerGroupsPct = float64(stats.ConsumerGroupsCount) / float64(stats.TotalEntities) * 100
		stats.PluginsPct = float64(stats.PluginsCount) / float64(stats.TotalEntities) * 100
		stats.UpstreamsPct = float64(stats.UpstreamsCount) / float64(stats.TotalEntities) * 100
		stats.CertificatesPct = float64(stats.CertificatesCount) / float64(stats.TotalEntities) * 100
		stats.CACertificatesPct = float64(stats.CACertificatesCount) / float64(stats.TotalEntities) * 100
	}

	if stats.RoutesCount > 0 {
		stats.AvgPathsPerRoute = float64(stats.PathRoutesCount) / float64(stats.RoutesCount)
	}
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
	entityCountsBuf, entityCountsTable := generateEntityCountReport(style, stats)

	//regexCountsTable := generateRegexCountReport(style, stats)

	// 2. for each unique tag, entity counts and percentages
	tagsBuf, tagsTable := generateTagBasedReport(style, stats)

	// 3. for each plugin name, plugin counts
	pluginsBuf, pluginsTable := generatePluginCountReport(style, stats)

	switch OutputFormat(format) {
	case TextFormat:
		entityCountsTable.Render()
		//regexCountsTable.Render()
		tagsTable.Render()
		pluginsTable.Render()
	case CSVFormat:
		entityCountsTable.RenderCSV()
		//regexCountsTable.RenderCSV()
		tagsTable.RenderCSV()
		pluginsTable.RenderCSV()
	case HTMLFormat:
		entityCountsTable.RenderHTML()
		//regexCountsTable.RenderHTML()
		tagsTable.RenderHTML()
		pluginsTable.RenderHTML()
	case MarkdownFormat:
		entityCountsTable.RenderMarkdown()
		//regexCountsTable.RenderMarkdown()
		tagsTable.RenderMarkdown()
		pluginsTable.RenderMarkdown()
	case JsonFormat:
		jsonData, err := json.MarshalIndent(stats, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal stats to JSON: %v", err)
		}
		return jsonData, nil
	case YamlFormat:
		yamlData, err := yaml.Marshal(stats)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal stats to YAML: %v", err)
		}
		return yamlData, nil
	default:
		return nil, fmt.Errorf("invalid format '%s'", format)
	}

	// append entityCountsBuf, tagsBuf, pluginsBuf
	buf := new(bytes.Buffer)
	buf.Write(entityCountsBuf.Bytes())
	// buf.Write(regexCountsBuf.Bytes())
	buf.Write(pluginsBuf.Bytes())
	if includeTags {
		buf.Write(tagsBuf.Bytes())
	}

	return buf.Bytes(), nil
}

func generatePluginCountReport(style string, stats ContentStatistics) (*bytes.Buffer, table.Writer) {
	var EnterprisePluginsMap = map[string]bool{
		"ai-azure-content-safety":        true,
		"ai-proxy-advanced":              true,
		"ai-rate-limiting-advanced":      true,
		"ai-semantic-cache":              true,
		"ai-semantic-prompt-guard":       true,
		"app-dynamics":                   true,
		"application-registration":       true,
		"canary":                         true,
		"confluent":                      true,
		"degraphql":                      true,
		"exit-transformer":               true,
		"forward-proxy":                  true,
		"graphql-proxy-cache-advanced":   true,
		"graphql-rate-limiting-advanced": true,
		"header-cert-auth":               true,
		"injection-protection":           true,
		"jwe-decrypt":                    true,
		"jq":                             true,
		"json-threat-protection":         true,
		"jwt-signer":                     true,
		"kafka-log":                      true,
		"kafka-upstream":                 true,
		"key-auth-enc":                   true,
		"ldap-auth-advanced":             true,
		"mocking":                        true,
		"mtls-auth":                      true,
		"oauth2-introspection":           true,
		"oas-validation":                 true,
		"opa":                            true,
		"openid-connect":                 true,
		"proxy-cache-advanced":           true,
		"rate-limiting-advanced":         true,
		"request-transformer-advanced":   true,
		"request-validator":              true,
		"response-transformer-advanced":  true,
		"route-by-header":                true,
		"route-transformer-advanced":     true,
		"saml":                           true,
		"service-protection":             true,
		"statsd-advanced":                true,
		"tls-handshake-modifier":         true,
		"tls-metadata-headers":           true,
		"upstream-oauth":                 true,
		"upstream-timeout":               true,
		"vault-auth":                     true,
		"websocket-size-limit":           true,
		"websocket-validator":            true,
		"xml-threat-protection":          true,
	}

	var OSSPluginsMap = map[string]bool{
		"ai-proxy":                true,
		"ai-prompt-decorator":     true,
		"ai-prompt-guard":         true,
		"ai-prompt-template":      true,
		"ai-request-transformer":  true,
		"ai-response-transformer": true,
		"basic-auth":              true,
		"hmac-auth":               true,
		"jwt":                     true,
		"key-auth":                true,
		"ldap-auth":               true,
		"oauth2":                  true,
		"session":                 true,
		"acme":                    true,
		"bot-detection":           true,
		"cors":                    true,
		"ip-restriction":          true,
		"acl":                     true,
		"proxy-cache":             true,
		"rate-limiting":           true,
		"redirect":                true,
		"request-size-limiting":   true,
		"request-termination":     true,
		"response-ratelimiting":   true,
		"standard-webhooks":       true,
		"aws-lambda":              true,
		"azure-functions":         true,
		"pre-function":            true,
		"post-function":           true,
		"datadog":                 true,
		"opentelemetry":           true,
		"prometheus":              true,
		"statsd":                  true,
		"zipkin":                  true,
		"correlation-id":          true,
		"request-transformer":     true,
		"response-transformer":    true,
		"grpc-web":                true,
		"grpc-gateway":            true,
		"file-log":                true,
		"http-log":                true,
		"loggly":                  true,
		"syslog":                  true,
		"tcp-log":                 true,
		"udp-log":                 true,
	}

	pluginsBuf := new(bytes.Buffer)
	pluginsTable := table.NewWriter()
	pluginsTable.SetStyle(getTableStyle(style))
	pluginsTable.SetOutputMirror(pluginsBuf)
	//pluginsTable.SetPageSize(10)
	pluginsTable.SortBy([]table.SortBy{
		{Name: "Count", Mode: table.DscNumeric},
	})
	pluginsTable.SetTitle("Plugin count per type")
	pluginsTable.AppendHeader(table.Row{"PluginName", "Count", "Ent", "OSS", "Custom"})
	for pluginName, count := range stats.PluginsCountPerName {
		// if the plugin name is in the EnterprisePlugins list, mark it as Enterprise
		enterprise := ""
		oss := ""
		custom := ""
		if OSSPluginsMap[pluginName] {
			oss = "*"
			stats.OSSPluginsCount++
		}
		if EnterprisePluginsMap[pluginName] {
			enterprise = "*"
			stats.EntPluginsCount++
		}
		if enterprise == "" && oss == "" {
			custom = "*"
			stats.CustomPluginsCount++
		}
		pluginsTable.AppendRow(table.Row{pluginName, count, enterprise, oss, custom})
	}
	pluginsTable.AppendFooter(table.Row{"Total Plugins", stats.PluginsCount, stats.EntPluginsCount, stats.OSSPluginsCount, stats.CustomPluginsCount})
	return pluginsBuf, pluginsTable
}

func generateTagBasedReport(style string, stats ContentStatistics) (*bytes.Buffer, table.Writer) {
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
	return tagsBuf, tagsTable
}

//func generateRegexCountReport(style string, stats ContentStatistics) table.Writer {
//regexCountsBuf := new(bytes.Buffer)
//regexCountsTable := table.NewWriter()
//regexCountsTable.SetStyle(getTableStyle(style))
//regexCountsTable.SetOutputMirror(regexCountsBuf)
//regexCountsTable.SetTitle("Paths with regex count and percentatge")
//regexCountsTable.AppendRow(table.Row{"Total paths in routes", stats.PathRoutesCount})
//regexCountsTable.AppendRow(table.Row{"Average paths per route", fmt.Sprintf("%.2f", stats.AvgPathsPerRoute)})
//regexCountsTable.AppendRow(table.Row{"Total paths with regular expressions in routes", stats.RegexRoutesCount})
//regexCountsTable.AppendFooter(table.Row{"Percentage of paths with regular expressions in routes", fmt.Sprintf("%.2f%%", stats.RegexRoutesPct)})
//return regexCountsTable
//}

func generateEntityCountReport(style string, stats ContentStatistics) (*bytes.Buffer, table.Writer) {
	entityCountsBuf := new(bytes.Buffer)
	entityCountsTable := table.NewWriter()
	entityCountsTable.SetStyle(getTableStyle(style))
	entityCountsTable.SetOutputMirror(entityCountsBuf)
	//entityCountsTable.SetPageSize(10)
	entityCountsTable.SortBy([]table.SortBy{
		{Name: "Count", Mode: table.DscNumeric},
	})
	entityCountsTable.SetTitle("Entity count for workspace/CP: " + stats.WorspaceOrCPName)
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
	return entityCountsBuf, entityCountsTable
}
