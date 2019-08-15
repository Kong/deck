// Copyright © 2018 Harry Bagdi <harrybagdi@gmail.com>

package cmd

import (
	"net/url"
	"path"

	"github.com/hbagdi/deck/dump"
	"github.com/hbagdi/deck/file"
	"github.com/hbagdi/deck/utils"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	dumpCmdKongStateFile string
	dumpWorkspace        string
)

// dumpCmd represents the dump command
var dumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "Export Kong configuration to a file",
	Long: `Dump command reads all the entities present in Kong
and writes them to a file on disk.

The file can then be read using the Sync o Diff command to again
configure Kong.`,
	RunE: func(cmd *cobra.Command, args []string) error {

		if dumpWorkspace != "" {
			u, err := url.Parse(config.Address)
			if err != nil {
				return err
			}
			u.Path = path.Join(u.Path, dumpWorkspace)
			config.Address = u.String()
		}

		client, err := utils.GetKongClient(config)
		if err != nil {
			return err
		}

		ks, err := dump.GetState(client, dumpConfig)
		if err != nil {
			return errors.Wrap(err, "reading configuration from Kong")
		}
		if err := file.KongStateToFile(ks, dumpConfig.SelectorTags,
			dumpWorkspace, dumpCmdKongStateFile); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(dumpCmd)
	dumpCmd.Flags().StringVarP(&dumpCmdKongStateFile, "output-file", "o",
		"kong.yaml", "write Kong configuration to FILE. "+
			"Use '-' to write to stdout.")
	dumpCmd.Flags().StringVarP(&dumpWorkspace, "workspace", "w",
		"", "dump configuration of a specific workspace"+
			"(Kong Enterprise only).")
	dumpCmd.Flags().BoolVar(&dumpConfig.SkipConsumers, "skip-consumers",
		false, "skip exporting consumers and any plugins associated "+
			"with consumers")
	dumpCmd.Flags().StringSliceVar(&dumpConfig.SelectorTags,
		"select-tag", []string{},
		"only entities matching tags specified via this flag are exported.\n"+
			"Multiple tags are ANDed together.")

}
