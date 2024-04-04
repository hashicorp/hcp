// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package version

import (
	_ "embed"
	"fmt"
	"runtime"
	"strings"
)

var (
	// GitCommit The git commit that was compiled. These will be filled in by the
	// compiler.
	GitCommit string

	// Version is the base version of the product repo
	//
	//go:embed VERSION
	fullVersion string

	Version, VersionPrerelease, _ = strings.Cut(FullVersion(), "-")
	VersionMetadata               = ""
)

// GetHumanVersion composes the parts of the version in a way that's suitable
// for displaying to humans.
func GetHumanVersion() string {
	version := Version
	release := VersionPrerelease
	metadata := VersionMetadata

	if release != "" {
		version += fmt.Sprintf("-%s", release)
	}

	if metadata != "" {
		version += fmt.Sprintf("+%s", metadata)
	}

	// Strip off any single quotes added by the git information.
	version = strings.ReplaceAll(version, "'", "")

	// Add the git commit if it's available.
	if GitCommit != "" {
		version += fmt.Sprintf(" (%s)", GitCommit)
	}

	// Add runtime information.
	version += fmt.Sprintf(" %s %s", runtime.Version(), runtime.GOARCH)

	// Add the command name prefix.
	version = fmt.Sprintf("hcp v%s", version)

	return version
}

// FullVersion returns the full version string including any prerelease tags.
func FullVersion() string {
	return strings.TrimSpace(fullVersion)
}

// GetSourceChannel returns the source channel for the CLI, including the current version.
func GetSourceChannel() string {
	return fmt.Sprintf("hcp-cli/%s", FullVersion())
}
