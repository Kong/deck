package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// VERSION is the current version of decK.
// This should be substituted by git tag during the build process.
var VERSION = "dev"

// COMMIT is the short hash of the source tree.
// This should be substituted by Git commit hash  during the build process.
var COMMIT = "unknown"

// newVersionCmd represents the version command
func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the decK version",
		Long: `The version command prints the version of decK along with a Git short
commit hash of the source tree.`,
		Args: validateNoArgs,
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Printf("decK %s (%s) \n", VERSION, COMMIT)
		},
	}
}
