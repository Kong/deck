/*
Package stats provides functionality for collecting and analyzing statistics
from Kong configuration files. It calculates entity counts, proportions, and
plugin usage statistics.

The main entry point is the PrintContentStatistics function which generates
reports in various formats including text tables, CSV, HTML, Markdown, JSON,
and YAML.
*/
package stats

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/kong/go-database-reconciler/pkg/file"
	"gopkg.in/yaml.v2"
)





// print struct ContentStatistics using three tables
// entity counts and percentages
// for each plugin name, plugin counts
func PrintContentStatistics(content *file.Content, format string) ([]byte, error) {

	stats := calculateContentStatistics(content)

	// entity counts and percentages
	entityCountsBuf, entityCountsTable := generateEntityCountReport(stats)

	// for each plugin name, plugin counts
	pluginsBuf, pluginsTable := generatePluginCountReport(stats)

	switch OutputFormat(format) {
	case TextFormat:
		entityCountsTable.Render()
		pluginsTable.Render()
	case CSVFormat:
		entityCountsTable.RenderCSV()
		pluginsTable.RenderCSV()
	case HTMLFormat:
		entityCountsTable.RenderHTML()
		pluginsTable.RenderHTML()
	case MarkdownFormat:
		entityCountsTable.RenderMarkdown()
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
	buf.Write(pluginsBuf.Bytes())

	return buf.Bytes(), nil
}