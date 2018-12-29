// Copyright © 2018 Harry Bagdi <harrybagdi@gmail.com>

package cmd

import (
	"fmt"

	"github.com/kong/deck/utils"
	"github.com/pkg/errors"
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
	// Disabled since it is not useful at the moment.
	// rootCmd.AddCommand(uuidCmd)
	uuidCmd.Flags().IntVarP(&count, "count", "c", 1,
		"number of UUIDs to generate")
}
