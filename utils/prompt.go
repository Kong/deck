package utils

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Confirm prompts a user for a confirmation with message
// and returns true with no error if input is "yes" or "y" (case-insensitive),
// otherwise false.
func Confirm(message string) (bool, error) {
	fmt.Print(message)
	validOptions := []string{"yes", "y"}
	var input string
	_, err := fmt.Scanln(&input)
	if err != nil {
		return false, err
	}
	input = strings.ToLower(input)
	for _, validOption := range validOptions {
		if input == validOption {
			return true, nil
		}
	}
	return false, nil
}

// ConfirmFileOverwrite is a helper function to determine whether or not the program should
// truncate and overwrite a file given its name and extension. If the file doesn't already exist
// in the filesystem, then this will return true, otherwise it will prompt the user for confirmation.
func ConfirmFileOverwrite(filename string, ext string, assumeYes bool) (bool, error) {
	if assumeYes {
		return true, nil
	}

	filename = AddExtToFilename(filename, ext)
	_, err := os.Stat(filename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return true, nil
		}
		return false, err
	}

	// file exists, prompt user
	return Confirm("File '" + filename + "' already exists. Do you want to overwrite it? ")
}
