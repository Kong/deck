// Copyright © 2018 Harry Bagdi <harrybagdi@gmail.com>

package cmd

import (
	"errors"
	"fmt"

	"github.com/kong/deck/utils"
	"github.com/spf13/cobra"
)

var count int

// uuidCmd represents the uuid command
var uuidCmd = &cobra.Command{
	Use:   "uuid",
	Short: "Generate a pseudorandom UUID",
	RunE: func(cmd *cobra.Command, args []string) error {
		if count < 1 {
			return errors.New("count should be at least 1")
		}
		for i := 0; i < count; i++ {
			fmt.Println(utils.UUID())
		}
		return nil
	},
}

func init() {
	// rootCmd.AddCommand(uuidCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// uuidCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	uuidCmd.Flags().IntVarP(&count, "count", "c", 1,
		"number of UUIDs to generate")
}
