package cmd

import (
	"net/http"
	"strings"

	"github.com/hbagdi/deck/dump"
	"github.com/hbagdi/deck/file"
	"github.com/hbagdi/deck/state"
	"github.com/hbagdi/deck/utils"
	"github.com/hbagdi/go-kong/kong"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	dumpCmdKongStateFile string
	dumpCmdStateFormat   string
	dumpWorkspace        string
	dumpAllWorkspaces    bool
	dumpWithID           bool
)

func listWorkspaces(client *kong.Client, baseURL string) ([]string, error) {
	type Workspace struct {
		Name string
	}
	type Response struct {
		Data []Workspace
	}

	var response Response
	// TODO handle pagination
	req, err := http.NewRequest("GET", baseURL+"/workspaces?size=1000", nil)
	if err != nil {
		return nil, errors.Wrap(err, "building request for fetching workspaces")
	}
	_, err = client.Do(nil, req, &response)
	if err != nil {
		return nil, errors.Wrap(err, "fetching workspaces from Kong")
	}
	var res []string
	for _, workspace := range response.Data {
		res = append(res, workspace.Name)
	}

	return res, nil
}

// dumpCmd represents the dump command
var dumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "Export Kong configuration to a file",
	Long: `Dump command reads all the entities present in Kong
and writes them to a file on disk.

The file can then be read using the Sync o Diff command to again
configure Kong.`,
	Args: validateNoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {

		client, err := utils.GetKongClient(config)
		if err != nil {
			return err
		}

		format := file.Format(strings.ToUpper(dumpCmdStateFormat))

		// Kong Enterprise dump all workspace
		if dumpAllWorkspaces {
			if dumpWorkspace != "" {
				return errors.New("workspace cannot be specified with --all-workspace flag")
			}
			if dumpCmdKongStateFile != "kong" {
				return errors.New("output-file cannot be specified with --all-workspace flag")
			}
			workspaces, err := listWorkspaces(client, config.Address)
			if err != nil {
				return err
			}

			for _, workspace := range workspaces {
				config.Workspace = workspace
				client, err := utils.GetKongClient(config)
				if err != nil {
					return err
				}

				rawState, err := dump.Get(client, dumpConfig)
				if err != nil {
					return errors.Wrap(err, "reading configuration from Kong")
				}
				ks, err := state.Get(rawState)
				if err != nil {
					return errors.Wrap(err, "building state")
				}

				if err := file.KongStateToFile(ks, file.WriteConfig{
					SelectTags: dumpConfig.SelectorTags,
					Workspace:  workspace,
					Filename:   workspace,
					FileFormat: format,
					WithID:     dumpWithID,
				}); err != nil {
					return err
				}
			}
			return nil
		}

		// Kong OSS
		// or Kong Enterprise single workspace
		if dumpWorkspace != "" {
			config.Workspace = dumpWorkspace

			if err := checkWorkspace(config); err != nil {
				return err
			}

			client, err = utils.GetKongClient(config)
			if err != nil {
				return err
			}
		}

		rawState, err := dump.Get(client, dumpConfig)
		if err != nil {
			return errors.Wrap(err, "reading configuration from Kong")
		}
		ks, err := state.Get(rawState)
		if err != nil {
			return errors.Wrap(err, "building state")
		}
		if err := file.KongStateToFile(ks, file.WriteConfig{
			SelectTags: dumpConfig.SelectorTags,
			Workspace:  dumpWorkspace,
			Filename:   dumpCmdKongStateFile,
			FileFormat: format,
			WithID:     dumpWithID,
		}); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(dumpCmd)
	dumpCmd.Flags().StringVarP(&dumpCmdKongStateFile, "output-file", "o",
		"kong", "file to which to write Kong's configuration."+
			"Use '-' to write to stdout.")
	dumpCmd.Flags().StringVar(&dumpCmdStateFormat, "format",
		"yaml", "output file format: json or yaml")
	dumpCmd.Flags().BoolVar(&dumpWithID, "with-id",
		false, "write ID of all entities in the output")
	dumpCmd.Flags().StringVarP(&dumpWorkspace, "workspace", "w",
		"", "dump configuration of a specific workspace"+
			"(Kong Enterprise only).")
	dumpCmd.Flags().BoolVar(&dumpAllWorkspaces, "all-workspaces",
		false, "dump configuration of all workspaces (Kong Enterprise only).")
	dumpCmd.Flags().BoolVar(&dumpConfig.SkipConsumers, "skip-consumers",
		false, "skip exporting consumers and any plugins associated "+
			"with consumers")
	dumpCmd.Flags().StringSliceVar(&dumpConfig.SelectorTags,
		"select-tag", []string{},
		"only entities matching tags specified via this flag are exported.\n"+
			"Multiple tags are ANDed together.")

}
