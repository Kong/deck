package cmd

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/blang/semver/v4"
	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func resetUpdateCheckState() {
	updateNoticeOnce = sync.Once{}
	updateHTTPClient = &http.Client{Timeout: updateCheckTimeout}
	rootConfig.Address = ""
}

func setVersionForTest(t *testing.T) {
	t.Helper()
	previousVersion := VERSION
	VERSION = "v1.2.3"
	t.Cleanup(func() {
		VERSION = previousVersion
	})
}

func disableColorForTest(t *testing.T) {
	t.Helper()
	previousNoColor := color.NoColor
	color.NoColor = true
	t.Cleanup(func() {
		color.NoColor = previousNoColor
	})
}

func TestBuildUpdateNotice(t *testing.T) {
	t.Cleanup(resetUpdateCheckState)
	setVersionForTest(t)
	disableColorForTest(t)

	updateHTTPClient = &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			require.Equal(t, updateCheckURL, req.URL.String())
			require.Equal(t, "application/vnd.github+json", req.Header.Get("Accept"))
			require.Equal(t, "decK/"+VERSION, req.Header.Get("User-Agent"))

			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"tag_name":"v1.2.4"}`)),
				Header:     make(http.Header),
			}, nil
		}),
	}

	notice := buildUpdateNotice("v1.2.3")
	assert.Contains(t, notice, "== Update available v1.2.3 -> v1.2.4")
	assert.Contains(t, notice, "Download & Release Notes: "+releaseNotesURL(mustParseVersion(t, "v1.2.4")))
	assert.True(t, strings.HasSuffix(notice, "\n"))
}

func TestBuildUpdateNoticeSkipsInvalidOrNonNewerVersions(t *testing.T) {
	tests := []struct {
		name         string
		localVersion string
		responseBody string
		statusCode   int
		wantNotice   string
	}{
		{
			name:         "local version is not a release",
			localVersion: "dev",
			responseBody: `{"tag_name":"v1.2.4"}`,
			statusCode:   http.StatusOK,
		},
		{
			name:         "same version",
			localVersion: "v1.2.4",
			responseBody: `{"tag_name":"v1.2.4"}`,
			statusCode:   http.StatusOK,
		},
		{
			name:         "newer local version",
			localVersion: "v1.2.5",
			responseBody: `{"tag_name":"v1.2.4"}`,
			statusCode:   http.StatusOK,
		},
		{
			name:         "remote tag missing",
			localVersion: "v1.2.3",
			responseBody: `{}`,
			statusCode:   http.StatusOK,
		},
		{
			name:         "remote tag invalid",
			localVersion: "v1.2.3",
			responseBody: `{"tag_name":"latest"}`,
			statusCode:   http.StatusOK,
		},
		{
			name:         "non-200 response",
			localVersion: "v1.2.3",
			responseBody: `{"tag_name":"v1.2.4"}`,
			statusCode:   http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Cleanup(resetUpdateCheckState)
			setVersionForTest(t)
			updateHTTPClient = &http.Client{
				Transport: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: tt.statusCode,
						Body:       io.NopCloser(strings.NewReader(tt.responseBody)),
						Header:     make(http.Header),
					}, nil
				}),
			}

			notice := buildUpdateNotice(tt.localVersion)
			assert.Equal(t, tt.wantNotice, notice)
		})
	}
}

func TestMaybePrintUpdateNoticePrintsOnceToStderr(t *testing.T) {
	t.Cleanup(resetUpdateCheckState)
	setVersionForTest(t)
	disableColorForTest(t)

	updateHTTPClient = &http.Client{
		Transport: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"tag_name":"v1.2.4"}`)),
				Header:     make(http.Header),
			}, nil
		}),
	}

	var stderr bytes.Buffer
	cmd := NewRootCmd()
	cmd.SetErr(&stderr)

	maybePrintUpdateNotice(cmd)
	maybePrintUpdateNotice(cmd)

	output := stderr.String()
	assert.Contains(t, output, "== Update available")
	assert.Equal(t, 1, strings.Count(output, "== Update available"))
	assert.Contains(t, output, "/tag/v1.2.4")
	assert.Contains(t, output, "\n\n")
}

func TestRootCommandSuppressesUpdateCheckWithFlag(t *testing.T) {
	t.Cleanup(resetUpdateCheckState)
	setVersionForTest(t)
	disableColorForTest(t)

	updateHTTPClient = &http.Client{
		Transport: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"tag_name":"v999.0.0"}`)),
				Header:     make(http.Header),
			}, nil
		}),
	}

	cmd := NewRootCmd()
	var stderr bytes.Buffer
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"version", "--suppress-update-check"})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.NotContains(t, stderr.String(), "== Update available")
}

func TestRootCommandSuppressesUpdateCheckWithEnvVar(t *testing.T) {
	t.Cleanup(resetUpdateCheckState)
	t.Setenv("DECK_SUPPRESS_UPDATE_CHECK", "true")
	setVersionForTest(t)
	disableColorForTest(t)

	updateHTTPClient = &http.Client{
		Transport: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"tag_name":"v999.0.0"}`)),
				Header:     make(http.Header),
			}, nil
		}),
	}

	cmd := NewRootCmd()
	var stderr bytes.Buffer
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"version"})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.NotContains(t, stderr.String(), "== Update available")
}

func TestRootCommandFlagFalseOverridesSuppressEnvVar(t *testing.T) {
	t.Cleanup(resetUpdateCheckState)
	t.Setenv("DECK_SUPPRESS_UPDATE_CHECK", "true")
	setVersionForTest(t)
	disableColorForTest(t)

	updateHTTPClient = &http.Client{
		Transport: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"tag_name":"v999.0.0"}`)),
				Header:     make(http.Header),
			}, nil
		}),
	}

	cmd := NewRootCmd()
	var stderr bytes.Buffer
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"version", "--suppress-update-check=false"})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, stderr.String(), "== Update available")
}

func TestRootCommandSuppressesUpdateCheckWithJSONOutput(t *testing.T) {
	t.Cleanup(resetUpdateCheckState)
	setVersionForTest(t)
	disableColorForTest(t)

	updateHTTPClient = &http.Client{
		Transport: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"tag_name":"v999.0.0"}`)),
				Header:     make(http.Header),
			}, nil
		}),
	}

	cmd := NewRootCmd()
	var stderr bytes.Buffer
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"gateway", "diff", "--json-output"})

	err := cmd.Execute()
	require.Error(t, err)
	assert.NotContains(t, stderr.String(), "== Update available")
}

func TestSuppressUpdateCheckEnabledFromEnvWithoutViperInitialization(t *testing.T) {
	t.Cleanup(resetUpdateCheckState)
	t.Setenv("DECK_SUPPRESS_UPDATE_CHECK", "true")

	cmd := NewRootCmd()

	assert.True(t, suppressUpdateCheckEnabled(cmd))
}

func TestRootHelpShowsUpdateNoticeOnStderr(t *testing.T) {
	t.Cleanup(resetUpdateCheckState)
	setVersionForTest(t)
	disableColorForTest(t)

	updateHTTPClient = &http.Client{
		Transport: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"tag_name":"v999.0.0"}`)),
				Header:     make(http.Header),
			}, nil
		}),
	}

	cmd := NewRootCmd()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, stdout.String(), "The deck tool helps you manage Kong clusters")
	assert.Contains(t, stderr.String(), "== Update available")
	assert.Contains(t, stderr.String(), "/tag/v999.0.0")
	assert.True(t, strings.HasSuffix(stderr.String(), "\n\n"))
}

func mustParseVersion(t *testing.T, version string) semver.Version {
	t.Helper()
	parsed, err := parseReleaseVersion(version)
	require.NoError(t, err)
	return parsed
}
