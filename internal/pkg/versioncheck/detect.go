// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: MPL-2.0

package versioncheck

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hcp/internal/pkg/api/releasesapi/client"
	"github.com/hashicorp/hcp/internal/pkg/api/releasesapi/client/operations"
	"github.com/hashicorp/hcp/internal/pkg/api/releasesapi/models"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	hcpversion "github.com/hashicorp/hcp/version"
	"github.com/mitchellh/go-homedir"
)

const (
	// hcpProduct is the product name for HCP CLI in releases.
	hcpProduct = "hcp"
)

// Checker is a version checker for the HCP CLI.
type Checker struct {
	io             iostreams.IOStreams
	client         operations.ClientService
	currentVersion *version.Version
	checkState     *versionCheckState
	statePath      string

	// skipCICheck allows for skipping the CI variables. This is needed when
	// testing, since we test in CI environments.
	skipCICheck bool

	sync.Mutex
}

// New returns a new version checker for the HCP CLI. It will used the passed
// iostreams to display any new version information and will store the
// stateFilePath.
func New(io iostreams.IOStreams, stateFilePath string) (*Checker, error) {
	return newChecker(io, stateFilePath,
		hcpversion.FullVersion(), operations.New(client.Default.Transport, nil), false)
}

func newChecker(io iostreams.IOStreams, stateFilePath string,
	currentVersion string, client operations.ClientService, skipCICheck bool) (*Checker, error) {
	// Parse the current version.
	current, err := version.NewVersion(currentVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to parse current version (%q): %w", hcpversion.FullVersion(), err)
	}

	path, err := homedir.Expand(stateFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed expanding HCP config directory path %q: %w", stateFilePath, err)
	}

	// Ensure the folder exists.
	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); err != nil {
		// If the directory doesn't exist, create it.
		if errors.Is(err, fs.ErrNotExist) {
			if err := os.MkdirAll(dir, 0766); err != nil {
				return nil, fmt.Errorf("failed to created HCP config directory %q: %w", dir, err)
			}
		} else {
			return nil, fmt.Errorf("failed to check if HCP config directory exists: %w", err)
		}
	}

	return &Checker{
		io:             io,
		client:         client,
		currentVersion: current,
		statePath:      path,
		skipCICheck:    skipCICheck,
	}, nil
}

// Check checks for a new version of the HCP CLI.
func (c *Checker) Check(ctx context.Context) error {
	if c == nil {
		return nil
	}

	if !c.shouldCheckNewVersion() {
		return nil
	}

	// Get the most recent "hcp" CLI releases.
	resp, err := c.client.ListReleasesV1(&operations.ListReleasesV1Params{
		Product: hcpProduct,
		Context: ctx,
	})
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return nil
		}

		return fmt.Errorf("failed to list recent releases: %w", err)
	}

	// Find the latest release.
	var latestRelease *models.ProductReleaseResponseV1
	for _, release := range resp.Payload {
		if release.IsPrerelease {
			continue
		}

		latestRelease = release
		break
	}
	if latestRelease == nil {
		return fmt.Errorf("failed to determine latest release for %q", hcpProduct)
	}

	// Check if the latest release is newer than the current version.
	latest, err := version.NewVersion(*latestRelease.Version)
	if err != nil {
		return fmt.Errorf("failed to parse latest version (%q): %w", *latestRelease.Version, err)
	}

	if c.currentVersion.GreaterThanOrEqual(latest) {
		return nil
	}

	c.setCheckState(&versionCheckState{
		CheckedAt:     time.Now(),
		latestRelease: latestRelease,
		path:          c.statePath,
	})

	return nil
}

// Display displays the new version information to the user.
func (c *Checker) Display() {
	if c == nil {
		return
	}

	vs := c.getCheckState()
	if vs == nil {
		return
	}

	cs := c.io.ColorScheme()
	fmt.Fprintf(c.io.Err(), "\n%s %s: %s -> %s\n\n",
		cs.String("INFO:").Color(cs.Yellow()),
		"A new version of the HCP CLI is available",
		c.currentVersion,
		cs.String(*vs.latestRelease.Version).Bold(),
	)

	if isUnderHomebrew() {
		fmt.Fprintf(c.io.Err(), "To upgrade, run: %s\n", "brew upgrade hcp")
	}
	fmt.Fprintf(c.io.Err(), "Release Notes: %s\n", vs.latestRelease.URLChangelog)

	// Save the fact that we've shown the update message
	_ = vs.write()

}

// setCheckState sets the current check state. It first locks the checker to
// ensure there is no race between setting and getting the check state.
func (c *Checker) setCheckState(s *versionCheckState) {
	c.Lock()
	defer c.Unlock()
	c.checkState = s
}

// getCheckState gets the current check state. It first locks the checker to
// ensure there is no race between setting and getting the check state.
func (c *Checker) getCheckState() *versionCheckState {
	c.Lock()
	defer c.Unlock()
	return c.checkState
}

// shouldCheckNewVersion returns whether to check for a new version.
func (c *Checker) shouldCheckNewVersion() bool {
	// Don't check for new versions in CI environments.
	if !c.skipCICheck && isCI() {
		return false
	}

	// See if we've checked recently.
	s, _ := readVersionCheckState(c.statePath)
	if s != nil && time.Since(s.CheckedAt).Hours() < 24 {
		return false
	}

	return true
}

// isCI returns true if the current environment is a CI environment.
func isCI() bool {
	return os.Getenv("CI") != "" || // GitHub Actions, Travis CI, CircleCI, Cirrus CI, GitLab CI, AppVeyor, CodeShip, dsari
		os.Getenv("BUILD_NUMBER") != "" || // Jenkins, TeamCity
		os.Getenv("RUN_ID") != "" // TaskCluster, dsari
}

// isUnderHomebrew checks whether the hcp binary is under the Homebrew prefix
func isUnderHomebrew() bool {
	brewExe, err := exec.LookPath("brew")
	if err != nil {
		return false
	}

	brewPrefixBytes, err := exec.Command(brewExe, "--prefix").Output()
	if err != nil {
		return false
	}

	exe, err := os.Executable()
	if err != nil {
		exe = "hcp"
	}

	brewPrefix := strings.TrimSpace(string(brewPrefixBytes))
	brewBinPrefix := filepath.Join(brewPrefix, "bin") + string(filepath.Separator)
	return strings.HasPrefix(exe, brewBinPrefix)
}
