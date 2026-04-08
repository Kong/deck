package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/blang/semver/v4"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

const (
	updateCheckURL       = "https://api.github.com/repos/kong/deck/releases/latest"
	updateReleasePageURL = "https://github.com/kong/deck/releases"
	updateCheckTimeout   = 2 * time.Second
)

type githubRelease struct {
	TagName string `json:"tag_name"`
}

var (
	updateNoticeOnce sync.Once
	updateHTTPClient = &http.Client{Timeout: updateCheckTimeout}
)

func maybePrintUpdateNotice(cmd *cobra.Command) {
	if cmd == nil || suppressUpdateCheckEnabled(cmd) {
		return
	}

	updateNoticeOnce.Do(func() {
		notice, err := buildUpdateNotice(VERSION)
		if err != nil || notice == "" {
			return
		}
		fmt.Fprintln(cmd.ErrOrStderr(), notice)
	})
}

func suppressUpdateCheckEnabled(cmd *cobra.Command) bool {
	if cmd != nil {
		if flag := cmd.Flags().Lookup("suppress-update-check"); flag != nil && flag.Changed {
			value, err := cmd.Flags().GetBool("suppress-update-check")
			if err == nil {
				return value
			}
		}
	}

	if value, ok := os.LookupEnv("DECK_SUPPRESS_UPDATE_CHECK"); ok {
		enabled, err := strconv.ParseBool(value)
		if err == nil && enabled {
			return true
		}
	}
	return false
}

func buildUpdateNotice(localVersion string) (string, error) {
	currentVersion, err := parseReleaseVersion(localVersion)
	if err != nil {
		return "", nil
	}

	latestVersion, err := fetchLatestReleaseVersion()
	if err != nil {
		return "", nil
	}

	if !latestVersion.GT(currentVersion) {
		return "", nil
	}

	return formatUpdateNotice(currentVersion, latestVersion), nil
}

func parseReleaseVersion(version string) (semver.Version, error) {
	return semver.ParseTolerant(version)
}

func fetchLatestReleaseVersion() (semver.Version, error) {
	req, err := http.NewRequest(http.MethodGet, updateCheckURL, nil)
	if err != nil {
		return semver.Version{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", fmt.Sprintf("decK/%s", VERSION))

	resp, err := updateHTTPClient.Do(req)
	if err != nil {
		return semver.Version{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return semver.Version{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return semver.Version{}, err
	}

	return parseReleaseVersion(release.TagName)
}

func releaseNotesURL(version semver.Version) string {
	return fmt.Sprintf("%s/tag/%s", updateReleasePageURL, displayVersion(version))
}

func displayVersion(version semver.Version) string {
	return "v" + version.String()
}

func formatUpdateNotice(currentVersion, latestVersion semver.Version) string {
	header := color.New(color.FgYellow, color.Bold).Sprint("== Update available")

	return fmt.Sprintf(
		"%s %s -> %s\nDownload & Release Notes: %s\n",
		header,
		displayVersion(currentVersion),
		displayVersion(latestVersion),
		releaseNotesURL(latestVersion),
	)
}
