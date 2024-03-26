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

// GetHCPConfig retrieves the HCP configuration, applying any passed options. If
// a previous login occurred, GetHCPConfig will attempt to configure the HCP
// configuration based on the principal that was authenticated.
func GetHCPConfig(options ...hcpconf.HCPConfigOption) (hcpconf.HCPConfig, error) {
	return GetHCPConfigFromDir(CredentialsDir, options...)
}

// GetHCPConfigFromDir is like GetHCPConfig but can search non-default
// directories.
func GetHCPConfigFromDir(credFileDir string, options ...hcpconf.HCPConfigOption) (hcpconf.HCPConfig, error) {
	opts := []hcpconf.HCPConfigOption{
		hcpconf.WithoutLogging(),
		hcpconf.FromEnv(),
	}

	// Get the path the credential file
	credFilePath, err := GetHCPCredFilePath(credFileDir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve hcp's credential path: %w", err)
	}

	// Create a credential file
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

	// Build the path to the credential file
	credFilePath := filepath.Join(dir, CredFileName)
	return credFilePath, nil
}

// IsAuthenticated returns if there is a valid token available.
func IsAuthenticated() (bool, error) {
	// Create the HCP Config
	hcpCfg, err := GetHCPConfig(hcpconf.WithoutBrowserLogin())
	if err != nil {
		return false, fmt.Errorf("failed to instantiate HCP config to check authentication status: %w", err)
	}

	if tkn, err := hcpCfg.Token(); err != nil {
		return false, nil
	} else if tkn.Expiry.Before(time.Now()) {
		return false, nil
	}

	return true, nil
}
