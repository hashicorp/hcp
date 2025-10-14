// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package auth

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	hcpconf "github.com/hashicorp/hcp-sdk-go/config"
	"github.com/mitchellh/go-homedir"
)

const (
	// CredFileName is the name of the credential file written to the CLI's
	// config directory
	CredFileName = "cred_file.json"

	// CredentialsDir is the directory to store credentials from authentication.
	CredentialsDir = "~/.config/hcp/credentials/"
)

// GetHCPConfig retrieves the HCP configuration with geography support.
// Geography configuration is handled by the HCP SDK through the WithGeography option.
func GetHCPConfig(options ...hcpconf.HCPConfigOption) (hcpconf.HCPConfig, error) {
	return GetHCPConfigFromDir(CredentialsDir, options...)
}

// GetHCPConfigFromDir is like GetHCPConfig but can search non-default directories.
func GetHCPConfigFromDir(credFileDir string, options ...hcpconf.HCPConfigOption) (hcpconf.HCPConfig, error) {
	opts := []hcpconf.HCPConfigOption{
		hcpconf.WithoutLogging(),
		hcpconf.FromEnv(),
	}

	// Get the path to the credential file
	credFilePath, err := GetHCPCredFilePath(credFileDir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve hcp's credential path: %w", err)
	}

	// Use credential file if it exists
	if _, err := os.Stat(credFilePath); err == nil {
		opts = append(opts, hcpconf.WithCredentialFilePath(credFilePath))
	}

	opts = append(opts, options...)
	hcpCfg, err := hcpconf.NewHCPConfig(opts...)
	if err != nil {
		return nil, err
	}

	return hcpCfg, nil
}

// GetHCPCredFilePath returns the path to the cli's credential file.
func GetHCPCredFilePath(credFileDir string) (string, error) {
	// Expand the directory
	dir, err := homedir.Expand(credFileDir)
	if err != nil {
		return "", fmt.Errorf("failed to resolve hcp's credential directory path: %w", err)
	}

	credFilePath := filepath.Join(dir, CredFileName)
	return credFilePath, nil
}

// IsAuthenticated returns if there is a valid token available.
func IsAuthenticated() (bool, error) {
	hcpCfg, err := GetHCPConfig(hcpconf.WithoutBrowserLogin())
	if err != nil {
		return false, fmt.Errorf("failed to instantiate HCP config to check authentication status: %w", err)
	}

	return isTokenValid(hcpCfg)
}

// isTokenValid checks if a token from the given HCP config is valid and not expired.
func isTokenValid(hcpCfg hcpconf.HCPConfig) (bool, error) {
	tkn, err := hcpCfg.Token()
	if err != nil {
		return false, nil
	}

	if tkn.Expiry.Before(time.Now()) {
		return false, nil
	}

	return true, nil
}
