// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package auth

import (
	"path/filepath"
	"testing"

	hcpconf "github.com/hashicorp/hcp-sdk-go/config"
	"github.com/stretchr/testify/require"
)

func TestGetHCPCredFilePath(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Name        string
		CredFileDir string
		Expected    string
	}{
		{
			Name:        "Default credential directory",
			CredFileDir: "~/.config/hcp/credentials/",
			Expected:    "cred_file.json", // Just check filename since homedir expansion varies
		},
		{
			Name:        "Custom directory",
			CredFileDir: "/tmp/test-creds/",
			Expected:    "cred_file.json",
		},
		{
			Name:        "Directory without trailing slash",
			CredFileDir: "/tmp/test-creds",
			Expected:    "cred_file.json",
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			r := require.New(t)

			path, err := GetHCPCredFilePath(c.CredFileDir)
			r.NoError(err)
			r.Contains(path, c.Expected, "Expected path to contain %s, got %s", c.Expected, path)
		})
	}
}

func TestGetHCPConfigFromDir_Integration(t *testing.T) {
	t.Parallel()

	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Test with non-existent credential file - should still create config
	cfg, err := GetHCPConfigFromDir(tempDir)
	require.NoError(t, err)
	require.NotNil(t, cfg)
}

func TestGetHCPConfig_DefaultDirectory(t *testing.T) {
	t.Parallel()

	// Test default GetHCPConfig function
	// This will use the default CredentialsDir and may not have actual credentials
	// but should not error during config creation
	cfg, err := GetHCPConfig(hcpconf.WithoutBrowserLogin())
	require.NoError(t, err)
	require.NotNil(t, cfg)
}

func TestIsAuthenticated_Integration(t *testing.T) {
	t.Parallel()

	// Test IsAuthenticated - this will likely return false in test environment
	// but should not error
	isAuth, err := IsAuthenticated()
	require.NoError(t, err)
	// Don't assert the value since it depends on actual credentials being present
	_ = isAuth
}

func TestConstants(t *testing.T) {
	t.Parallel()

	// Test that constants are defined correctly
	require.Equal(t, "cred_file.json", CredFileName)
	require.Equal(t, "~/.config/hcp/credentials/", CredentialsDir)
}

// TestEdgeCases tests various edge cases and error conditions
func TestEdgeCases(t *testing.T) {
	t.Parallel()

	// Test with empty directory
	path, err := GetHCPCredFilePath("")
	require.NoError(t, err)
	require.NotEmpty(t, path)
	filename := filepath.Base(path)
	require.Equal(t, CredFileName, filename)
}
