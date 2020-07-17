package cmd

import (
	"fmt"
	"github.com/hbagdi/deck/file"
	"github.com/hbagdi/deck/utils"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	pingCmdKongStateFile []string
)

// pingCmd represents the ping command
var pingCmd = &cobra.Command{
	Use:   "ping",
	Short: "Verify connectivity with Kong",
	Long: `Ping command can be used to verify if decK
can connect to Kong's Admin API or not.`,
	Args: validateNoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := utils.GetKongClient(config)
		if err != nil {
			return errors.Wrap(err, "creating kong client")
		}

		version, err := kongVersion(config)
		if err != nil {

			targetContent, err := file.GetContentFromFiles(pingCmdKongStateFile)
			if err != nil {
				return errors.Wrap(err, "error reading Kong State File")
			}
			// prepare to read the current state from Kong
			config.Workspace = targetContent.Workspace

			version, err := kongVersion(config)
			if err != nil {
				return errors.Wrap(err, "reading Kong version")
			}
			fmt.Println("Successfully connected to Kong!")
			fmt.Println("Kong version: ", version)
			return nil
		}

		fmt.Println("Successfully connected to Kong!")
		fmt.Println("Kong version: ", version)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(pingCmd)
	rootCmd.Flags().StringSliceVarP(&pingCmdKongStateFile,
		"state", "s", []string{"kong.yaml"}, "file(s) containing Kong's configuration.\n"+
			"This flag can be specified multiple times for multiple files.\n"+
			"Use '-' to read from stdin.")
}
