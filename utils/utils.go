package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

var (
	kongVersionRegex = regexp.MustCompile(`^\d+\.\d+`)
)

// Empty checks if a string referenced by s or s itself is empty.
func Empty(s *string) bool {
	return s == nil || *s == ""
}

// UUID will generate a random v14 unique identifier based upon random numbers
func UUID() string {
	version := byte(4)
	uuid := make([]byte, 16)
	_, err := rand.Read(uuid)
	if err != nil {
		panic("failed to read from random generator: " + err.Error())
	}

	// Set version
	uuid[6] = (uuid[6] & 0x0f) | (version << 4)

	// Set variant
	uuid[8] = (uuid[8] & 0xbf) | 0x80

	buf := make([]byte, 36)
	var dash byte = '-'
	hex.Encode(buf[0:8], uuid[0:4])
	buf[8] = dash
	hex.Encode(buf[9:13], uuid[4:6])
	buf[13] = dash
	hex.Encode(buf[14:18], uuid[6:8])
	buf[18] = dash
	hex.Encode(buf[19:23], uuid[8:10])
	buf[23] = dash
	hex.Encode(buf[24:], uuid[10:])

	return string(buf)
}

// CleanKongVersion takes a version of Kong and returns back a string in
// the form of `/major.minor` version. There are various dashes and dots
// and other descriptors in Kong version strings, which has often created
// confusion in code and incorrect parsing, and hence this function does
// not return the patch version (on which shouldn't rely on anyways).
func CleanKongVersion(version string) (string, error) {
	matches := kongVersionRegex.FindStringSubmatch(version)
	if len(matches) < 1 {
		return "", fmt.Errorf("unknown Kong version")
	}
	return matches[0], nil
}

func AddExtToFilename(filename, ext string) string {
	if filepath.Ext(filename) == "" {
		filename = filename + "." + ext
	}
	return filename
}

// confirm prompts a user for a confirmation with message
// and returns true with no error if input is "yes" or "y" (case-insensitive),
// otherwise false.
func Confirm(message string) (bool, error) {
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
