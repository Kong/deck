package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// VERSION is the current version of decK.
// This should be substituted by git tag during the build process.
var VERSION = "dev"

// COMMIT is the short hash of the source tree.
// This should be substituted by Git commit hash  during the build process.
var COMMIT = "unknown"

var suppressVersionCheck bool

// GitHubRelease represents the structure of GitHub API response for latest release
type GitHubRelease struct {
	TagName     string    `json:"tag_name"`
	PublishedAt time.Time `json:"published_at"`
}

// checkLatestVersion checks GitHub for the latest release version
func checkLatestVersion() (*GitHubRelease, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/repos/Kong/deck/releases/latest", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	return &release, nil
}

// compareVersions returns true if latestVersion is newer than currentVersion
func compareVersions(currentVersion, latestVersion string) bool {
	// Remove 'v' prefix if present
	current := strings.TrimPrefix(currentVersion, "v")
	latest := strings.TrimPrefix(latestVersion, "v")

	// Split versions into parts
	currentParts := strings.Split(current, ".")
	latestParts := strings.Split(latest, ".")

	// Compare each part
	for i := 0; i < len(currentParts) && i < len(latestParts); i++ {
		var currentNum, latestNum int
		fmt.Sscanf(currentParts[i], "%d", &currentNum)
		fmt.Sscanf(latestParts[i], "%d", &latestNum)

		if latestNum > currentNum {
			return true
		} else if latestNum < currentNum {
			return false
		}
	}

	// If all compared parts are equal, check if latest has more parts
	return len(latestParts) > len(currentParts)
}

// newVersionCmd represents the version command
func newVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the decK version",
		Long: `The version command prints the version of decK along with a Git short
commit hash of the source tree.`,
		Args: validateNoArgs,
		Run: func(cmd *cobra.Command, _ []string) {
			fmt.Fprintf(cmd.OutOrStdout(), "decK %s (%s)\n", VERSION, COMMIT)

			// Check for latest version unless suppressed
			if !suppressVersionCheck && VERSION != "dev" {
				if release, err := checkLatestVersion(); err == nil {
					if compareVersions(VERSION, release.TagName) {
						fmt.Fprintf(cmd.OutOrStdout(), "\nA new version is available! -> %s (%s)\n", release.TagName, release.PublishedAt.Format("2006-01-02"))
						fmt.Fprintf(cmd.OutOrStdout(), "Download -> https://github.com/Kong/deck/releases\n")
					}
				}
			}
		},
	}

	cmd.Flags().BoolVar(&suppressVersionCheck, "suppress-version-check", false, "Disable checking for latest version")

	return cmd
}
