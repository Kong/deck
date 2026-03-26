package cmd

import (
	"fmt"
	"log"
	"strings"

	"github.com/kong/go-apiops/deckformat"
	"github.com/kong/go-apiops/filebasics"
	"github.com/kong/go-apiops/logbasics"
	"github.com/kong/go-apiops/openapi2mcp"
	"github.com/spf13/cobra"
)

var (
	cmdO2MinputFilename       string
	cmdO2MoutputFilename      string
	cmdO2MdocName             string
	cmdO2MoutputFormat        string
	cmdO2MentityTags          []string
	cmdO2MskipID              bool
	cmdO2Mmode                string
	cmdO2MpathPrefix          string
	cmdO2MincludeDirectRoute  bool
	cmdO2MignoreSecurityError bool
)

// Executes the CLI command "openapi2mcp"
func executeOpenapi2MCP(cmd *cobra.Command, _ []string) error {
	verbosity, _ := cmd.Flags().GetInt("verbose")
	logbasics.Initialize(log.LstdFlags, verbosity)
	_ = sendAnalytics("file-openapi2mcp", "", modeLocal)

	if len(cmdO2MentityTags) == 0 {
		cmdO2MentityTags = nil
	}

	cmdO2MoutputFormat = strings.ToUpper(cmdO2MoutputFormat)

	options := openapi2mcp.O2MOptions{
		Tags:                 cmdO2MentityTags,
		DocName:              cmdO2MdocName,
		SkipID:               cmdO2MskipID,
		Mode:                 cmdO2Mmode,
		PathPrefix:           cmdO2MpathPrefix,
		IncludeDirectRoute:   cmdO2MincludeDirectRoute,
		IgnoreSecurityErrors: cmdO2MignoreSecurityError,
	}

	trackInfo := deckformat.HistoryNewEntry("openapi2mcp")
	trackInfo["input"] = cmdO2MinputFilename
	trackInfo["output"] = cmdO2MoutputFilename
	trackInfo["uuid-base"] = cmdO2MdocName

	// do the work: read/convert/write
	content, err := filebasics.ReadFile(cmdO2MinputFilename)
	if err != nil {
		return err
	}
	result, err := openapi2mcp.Convert(content, options)
	if err != nil {
		return fmt.Errorf("failed converting OpenAPI spec '%s' to MCP; %w", cmdO2MinputFilename, err)
	}
	deckformat.HistoryAppend(result, trackInfo)
	return filebasics.WriteSerializedFile(cmdO2MoutputFilename, result, filebasics.OutputFormat(cmdO2MoutputFormat))
}

//
//
// Define the CLI data for the openapi2mcp command
//
//

func newOpenapi2MCPCmd() *cobra.Command {
	openapi2mcpCmd := &cobra.Command{
		Use:   "openapi2mcp",
		Short: "Convert OpenAPI files to Kong's decK format with MCP (Model Context Protocol) configuration",
		Long: `Convert OpenAPI files to Kong's decK format with ai-mcp-proxy plugin configuration.

This command generates a Kong service with an MCP route that includes the ai-mcp-proxy
plugin configured with tools derived from the OpenAPI specification operations.

Each OpenAPI operation is mapped to an MCP tool definition:
  - operationId -> tool name (kebab-case normalized)
  - summary/description -> tool description
  - parameters -> tool parameters array
  - requestBody -> tool request_body

Security/ACL generation:
  When an oauth2 security scheme includes the x-kong-mcp-acl extension, ACL entries
  are automatically generated for each tool based on the operation's security scopes.
  The plugin config will include acl_attribute_type, access_token_claim_field, and
  per-tool acl.allow arrays. Use x-kong-mcp-default-acl at the document level to
  set a default ACL for the plugin. Use --ignore-security-errors to skip unsupported
  security configurations instead of failing.

Supported x-kong extensions:
  - x-kong-name: Custom entity naming
  - x-kong-tags: Tags for all entities
  - x-kong-service-defaults: Service entity defaults
  - x-kong-route-defaults: Route entity defaults
  - x-kong-upstream-defaults: Upstream entity defaults
  - x-kong-plugin-*: Additional plugins

MCP-specific extensions:
  - x-kong-mcp-tool-name: Override generated tool name
  - x-kong-mcp-tool-description: Override tool description
  - x-kong-mcp-exclude: Exclude operation from tool generation (boolean)
  - x-kong-mcp-proxy: Override ai-mcp-proxy plugin config at document level

Security extensions:
  - x-kong-mcp-acl: ACL config on oauth2 security scheme (acl_attribute_type, access_token_claim_field)
  - x-kong-mcp-default-acl: Default ACL array at document level (scope, allow)`,
		RunE: executeOpenapi2MCP,
		Args: cobra.NoArgs,
	}

	openapi2mcpCmd.Flags().StringVarP(&cmdO2MinputFilename, "spec", "s", "-",
		"OpenAPI spec file to process. Use - to read from stdin.")
	openapi2mcpCmd.Flags().StringVarP(&cmdO2MoutputFilename, "output-file", "o", "-",
		"Output file to write. Use - to write to stdout.")
	openapi2mcpCmd.Flags().StringVarP(&cmdO2MoutputFormat, "format", "", "yaml", "output format: yaml or json")
	openapi2mcpCmd.Flags().StringVarP(&cmdO2MdocName, "uuid-base", "", "",
		"The unique base-string for uuid-v5 generation of entity IDs. If omitted,\n"+
			"uses the root-level \"x-kong-name\" directive, or falls back to 'info.title'.")
	openapi2mcpCmd.Flags().StringSliceVar(&cmdO2MentityTags, "select-tag", nil,
		"Select tags to apply to all entities. If omitted, uses the \"x-kong-tags\"\n"+
			"directive from the file.")
	openapi2mcpCmd.Flags().BoolVar(&cmdO2MskipID, "no-id", false,
		"Do not generate UUIDs for entities.")
	openapi2mcpCmd.Flags().StringVarP(&cmdO2Mmode, "mode", "m", openapi2mcp.ModeConversionListener,
		"ai-mcp-proxy mode: 'conversion' (client mode) or 'conversion-listener' (server mode).")
	openapi2mcpCmd.Flags().StringVarP(&cmdO2MpathPrefix, "path-prefix", "p", "",
		"Custom path prefix for the MCP route (default: /{service-name}-mcp).")
	openapi2mcpCmd.Flags().BoolVar(&cmdO2MincludeDirectRoute, "include-direct-route", false,
		"Also generate non-MCP routes for direct API access.")
	openapi2mcpCmd.Flags().BoolVar(&cmdO2MignoreSecurityError, "ignore-security-errors", false,
		"Ignore errors for unsupported security schemes or missing x-kong-mcp-acl extensions.")

	return openapi2mcpCmd
}
