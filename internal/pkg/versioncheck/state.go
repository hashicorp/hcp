// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: MPL-2.0

package versioncheck

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/hashicorp/hcp/internal/pkg/api/releasesapi/models"
)

// versionCheckState captures when the last check for a new version of the CLI
// occurred.
type versionCheckState struct {
	CheckedAt time.Time `json:"checked_at"`

	// latestRelease is the latest version of the CLI.
	latestRelease *models.ProductReleaseResponseV1

	// path is the path to write the state to.
	path string
}

// write writes the version state check to disk.
func (v *versionCheckState) write() error {

	f, err := os.Create(v.path)
	if err != nil {
		return fmt.Errorf("failed to create version check state file: %w", err)
	}

	m := json.NewEncoder(f)
	m.SetIndent("", "  ")
	if err := m.Encode(v); err != nil {
		return fmt.Errorf("failed to store version check state: %w", err)
	}

	return nil
}

// readVersionCheckState reads the version state from disk.
func readVersionCheckState(path string) (*versionCheckState, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	var s versionCheckState
	d := json.NewDecoder(f)
	if err := d.Decode(&s); err != nil {
		return nil, err
	}

	return &s, nil
}
