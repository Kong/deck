package cmd

import (
	"fmt"
	"log"
	"strings"

	"github.com/kong/go-apiops/deckformat"
	"github.com/kong/go-apiops/filebasics"
	"github.com/kong/go-apiops/jsonbasics"
	"github.com/kong/go-apiops/logbasics"
	"github.com/kong/go-apiops/namespace"
	"github.com/kong/go-apiops/yamlbasics"
	"github.com/spf13/cobra"
)

var (
	cmdNamespaceInputFilename       string
	cmdNamespaceOutputFilename      string
	cmdNamespaceOutputFormat        string
	cmdNamespaceSelectors           []string
	cmdNamespacePathPrefix          string
	cmdNamespaceAllowEmptySelectors bool
	cmdNamespaceHosts               []string
	cmdClearHosts                   bool
)

// Executes the CLI command "namespace"
func executeNamespace(cmd *cobra.Command, _ []string) error {
	verbosity, _ := cmd.Flags().GetInt("verbose")
	logbasics.Initialize(log.LstdFlags, verbosity)
	_ = sendAnalytics("file-namespace", "", modeLocal)

	if cmdNamespacePathPrefix != "" {
		err := namespace.CheckNamespace(cmdNamespacePathPrefix)
		if err != nil {
			return fmt.Errorf("invalid path-prefix '%s': %w", cmdNamespacePathPrefix, err)
		}
	}

	cmdNamespaceOutputFormat = strings.ToUpper(cmdNamespaceOutputFormat)

	trackInfo := deckformat.HistoryNewEntry("namespace")
	trackInfo["input"] = cmdNamespaceInputFilename
	trackInfo["output"] = cmdNamespaceOutputFilename
	trackInfo["selectors"] = cmdNamespaceSelectors
	trackInfo["path-prefix"] = cmdNamespacePathPrefix
	trackInfo["hosts"] = cmdNamespaceHosts
	trackInfo["clear-hosts"] = cmdClearHosts

	// do the work: read/namespace/write
	data, err := filebasics.DeserializeFile(cmdNamespaceInputFilename)
	if err != nil {
		return fmt.Errorf("failed to read input file '%s'; %w", cmdNamespaceInputFilename, err)
	}
	deckformat.HistoryAppend(data, trackInfo)

	yamlNode := jsonbasics.ConvertToYamlNode(data)

	// var selectors yamlbasics.SelectorSet
	selectors, err := yamlbasics.NewSelectorSet(cmdNamespaceSelectors)
	if err != nil {
		return err
	}

	// apply the path-based namespace
	if cmdNamespacePathPrefix != "" {
		err = namespace.Apply(yamlNode, selectors, cmdNamespacePathPrefix, cmdNamespaceAllowEmptySelectors)
		if err != nil {
			if strings.Contains(err.Error(), "no routes matched the selectors") {
				// append CLI specific message
				err = fmt.Errorf("%w (use --allow-empty-selectors to suppress this error)", err)
			}
			return fmt.Errorf("failed to apply the path namespace: %w", err)
		}
	}

	// apply the host-based namespace
	if len(cmdNamespaceHosts) > 0 || cmdClearHosts {
		err = namespace.ApplyNamespaceHost(yamlNode, selectors, cmdNamespaceHosts,
			cmdClearHosts, cmdNamespaceAllowEmptySelectors)
		if err != nil {
			if strings.Contains(err.Error(), "no routes matched the selectors") {
				// append CLI specific message
				err = fmt.Errorf("%w (use --allow-empty-selectors to suppress this error)", err)
			}
			return fmt.Errorf("failed to apply the host namespace: %w", err)
		}
	}

	data = jsonbasics.ConvertToJSONobject(yamlNode)

	return filebasics.WriteSerializedFile(cmdNamespaceOutputFilename, data,
		filebasics.OutputFormat(cmdNamespaceOutputFormat))
}

//
//
// Define the CLI data for the namespace command
//
//

func newNamespaceCmd() *cobra.Command {
	namespaceCmd := &cobra.Command{
		Use:   "namespace [flags]",
		Short: "Apply a namespace to routes in a decK file by path or hostname",
		Long: `Apply a namespace to routes in a decK file by path or hostname.

There are 2 main ways to namespace api's:

1. use path prefixes, all on the same hostname;
   a. http://api.acme.com/service1/somepath
   b. http://api.acme.com/service2/somepath
2. use separate hostnames
   a. http://service1.api.acme.com/somepath
   b. http://service2.api.acme.com/somepath

For hostnames the --host and --clear-hosts flags are used. Just using --host appends
to the existing hosts, adding --clear-hosts will effectively replace the existing ones.
For path prefixes the --path-prefix flag is used. Combining them is possible.

Note on path-prefixing: To remain transparent to the backend services, the added path
prefix must be removed from the path before the request is routed to the service.
To remove the prefix the following approaches are used (in order):
- if the route has 'strip_path=true' then the added prefix will already be stripped
- if the related service has a 'path' property that matches the prefix, then the
  'service.path' property is updated to remove the prefix
- a "pre-function" plugin will be added to remove the prefix from the path

`,
		RunE: executeNamespace,
		Example: `# Apply namespace to a deckfile, path and host:
deck file namespace --path-prefix=/kong --host=konghq.com --state=deckfile.yaml

# Apply namespace to a deckfile, and write to a new file
# Example file 'kong.yaml':
routes:
- paths:
  - ~/tracks/system$
  strip_path: true
- paths:
  - ~/list$
  strip_path: false

# Apply namespace to the deckfile, and write to stdout:
cat kong.yaml | deck file namespace --path-prefix=/kong

# Output:
routes:
- paths:
  - ~/kong/tracks/system$
  strip_path: true
  hosts:
  - konghq.com
- paths:
  - ~/kong/list$
  strip_path: false
  hosts:
  - konghq.com
  plugins:
  - name: pre-function
    config:
      access:
      - "local ns='/kong' -- this strips the '/kong' namespace from the path\nlocal <more code here>"

`,
	}

	namespaceCmd.Flags().StringVarP(&cmdNamespaceInputFilename, "state", "s", "-",
		"decK file to process. Use - to read from stdin.")
	namespaceCmd.Flags().StringVarP(&cmdNamespaceOutputFilename, "output-file", "o", "-",
		"Output file to write. Use - to write to stdout.")
	namespaceCmd.Flags().StringVarP(&cmdNamespaceOutputFormat, "format", "", "yaml",
		"Output format: yaml or json.")
	namespaceCmd.Flags().StringArrayVarP(&cmdNamespaceSelectors, "selector", "", []string{},
		"json-pointer identifying element to patch. Repeat for multiple selectors. Defaults "+
			"to selecting all routes.")
	namespaceCmd.Flags().StringVarP(&cmdNamespacePathPrefix, "path-prefix", "p", "",
		"The path based namespace to apply.")
	namespaceCmd.Flags().BoolVarP(&cmdNamespaceAllowEmptySelectors, "allow-empty-selectors",
		"", false, "Do not error out if the selectors return empty")
	namespaceCmd.Flags().StringArrayVarP(&cmdNamespaceHosts, "host", "", []string{},
		"Hostname to add for host-based namespacing. Repeat for multiple hosts.")
	namespaceCmd.Flags().BoolVarP(&cmdClearHosts, "clear-hosts", "c", false,
		"Clear existing hosts.")

	return namespaceCmd
}
