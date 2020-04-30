package cmd

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// VERSION is the current version of decK.
// This should be substituted by git tag during the build process.
var VERSION = "dev"

// COMMIT is the short hash of the source tree.
// This should be substituted by Git commit hash  during the build process.
var COMMIT = "unknown"

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of decK",
	Long: `version prints the version of decK along with git short
commit hash of the source tree`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			return errors.New("version command cannot take any positional arguments. " +
				"Try using a flag instead.\n" +
				"For usage information: deck version -h")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("decK %s (%s) \n", VERSION, COMMIT)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
