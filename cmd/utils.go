package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/kong/deck/utils"
)

// confirm prompts a user for a confirmation with message
// and returns true with no error if input is "yes" or "y" (case-insensitive),
// otherwise false.
func confirm(message string) (bool, error) {
	fmt.Print(message)
	yes := []string{"yes", "y"}
	var input string
	_, err := fmt.Scanln(&input)
	if err != nil {
		return false, err
	}
	input = strings.ToLower(input)
	for _, valid := range yes {
		if input == valid {
			return true, nil
		}
	}
	return false, nil
}

func confirmFileOverwrite(filename string, ext string, assumeYes bool) (bool, error) {
	if assumeYes {
		return true, nil
	}

	filename = utils.AddExtToFilename(filename, ext)
	_, err := os.Stat(filename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return true, nil
		}
		return false, err
	}

	// file exists, prompt user
	return confirm("File '" + filename + "' already exists. Do you want to overwrite it? ")
}
