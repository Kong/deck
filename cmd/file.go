/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

//
//
// Define the CLI data for the file sub-command
//
//

func newAddFileCmd() *cobra.Command {
	addFileCmd := &cobra.Command{
		Use:   "file [sub-command]...",
		Short: "Subcommand to host the decK file manipulation operations",
		Long:  `Subcommand to host the decK file manipulation operations.`,
	}

	return addFileCmd
}
