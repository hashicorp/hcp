// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: MPL-2.0

// Package geography provides helpers for working with HCP geographies using the HCP SDK.
package geography

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/hcp-sdk-go/config/files"
	"github.com/hashicorp/hcp-sdk-go/config/geography"
	"github.com/mitchellh/go-homedir"
)

const (
	// US is the united states geography
	US = string(geography.US)

	// EU is the europe geography
	EU = string(geography.EU)
)

// GetSupportedGeographies returns a slice of all supported geography strings.
// This leverages the HCP SDK's Geographies slice to ensure consistency.
func GetSupportedGeographies() []string {
	var geographies []string
	for _, geo := range geography.Geographies {
		geographies = append(geographies, string(geo))
	}
	return geographies
}

// IsValidGeography checks if the provided geography string is valid.
// This leverages the HCP SDK's ValidateGeo function.
func IsValidGeography(geo string) bool {
	return geography.ValidateGeo(geography.Geo(geo))
}

// GetDefaultGeography returns the default geography as a string.
// This leverages the HCP SDK's Default constant.
func GetDefaultGeography() string {
	return string(geography.Default)
}

// CredentialCache represents the structure of the SDK credential cache file
type CredentialCache struct {
	Login CredentialCacheLogin `json:"login"`
}

// CredentialCacheLogin represents the login section of the SDK credential cache
type CredentialCacheLogin struct {
	AccessToken       string `json:"access_token"`
	RefreshToken      string `json:"refresh_token"`
	AccessTokenExpiry string `json:"access_token_expiry"`
	Geography         string `json:"geography"`
}

// GetCachedGeography reads the geography from the SDK's credential cache file.
// This returns the geography that the user actually authenticated against during browser login.
func GetCachedGeography(configDir string) (string, error) {
	// Get the path to the credential cache file
	homeDir, err := homedir.Expand(configDir)
	if err != nil {
		return "", fmt.Errorf("failed to resolve home directory: %w", err)
	}

	// Check if the cache file exists
	// The SDK stores the credential cache in ~/.config/hcp/creds-cache.json
	cacheFilePath := filepath.Join(homeDir, files.TokenCacheFileName)
	if _, err := os.Stat(cacheFilePath); os.IsNotExist(err) {
		return "", nil // No cache file means no authenticated geography
	}

	// Read and parse the cache file
	cacheData, err := os.ReadFile(cacheFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to read credential cache file: %w", err)
	}

	var cache CredentialCache
	if err := json.Unmarshal(cacheData, &cache); err != nil {
		return "", fmt.Errorf("failed to parse credential cache file: %w", err)
	}

	return cache.Login.Geography, nil
}
