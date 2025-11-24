// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: MPL-2.0

package geography

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetSupportedGeographies(t *testing.T) {
	t.Parallel()
	r := require.New(t)

	geographies := GetSupportedGeographies()
	r.NotEmpty(geographies)
	r.Contains(geographies, "us")
	r.Contains(geographies, "eu")
	r.Len(geographies, 2) // Expecting exactly two geographies for now

	// Test that all returned geographies are valid
	for _, geo := range geographies {
		r.True(IsValidGeography(geo), "geography %q should be valid", geo)
	}
}

func TestIsValidGeography(t *testing.T) {
	t.Parallel()
	r := require.New(t)

	// Valid geographies
	r.True(IsValidGeography("us"))
	r.True(IsValidGeography("eu"))

	// Invalid geographies
	r.False(IsValidGeography("ap"))
	r.False(IsValidGeography("asia"))
	r.False(IsValidGeography("invalid"))
	r.False(IsValidGeography(""))
	r.False(IsValidGeography(" "))            // whitespace
	r.False(IsValidGeography("US"))           // case sensitivity
	r.False(IsValidGeography("EU"))           // case sensitivity
	r.False(IsValidGeography("us-east-1"))    // AWS-style region
	r.False(IsValidGeography("europe-west1")) // GCP-style region
}

func TestGetDefaultGeography(t *testing.T) {
	t.Parallel()
	r := require.New(t)

	defaultGeo := GetDefaultGeography()
	r.Equal("us", defaultGeo) // SDK default should be US
	r.True(IsValidGeography(defaultGeo), "default geography should be valid")
}

func TestConstants(t *testing.T) {
	t.Parallel()
	r := require.New(t)

	r.Equal("us", US)
	r.Equal("eu", EU)

	// Verify constants are valid geographies
	r.True(IsValidGeography(US))
	r.True(IsValidGeography(EU))

	// Verify constants match default if applicable
	r.Equal(US, GetDefaultGeography())
}

func TestCredentialCacheStructures(t *testing.T) {
	t.Parallel()
	r := require.New(t)

	// Test CredentialCache JSON marshaling/unmarshaling
	originalCache := CredentialCache{
		Login: CredentialCacheLogin{
			AccessToken:       "test-access-token",
			RefreshToken:      "test-refresh-token",
			AccessTokenExpiry: "2024-12-31T23:59:59Z",
			Geography:         "eu",
		},
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(originalCache)
	r.NoError(err)
	r.Contains(string(jsonData), "geography")
	r.Contains(string(jsonData), "eu")

	// Unmarshal back
	var unmarshaledCache CredentialCache
	err = json.Unmarshal(jsonData, &unmarshaledCache)
	r.NoError(err)
	r.Equal(originalCache.Login.Geography, unmarshaledCache.Login.Geography)
	r.Equal(originalCache.Login.AccessToken, unmarshaledCache.Login.AccessToken)
	r.Equal(originalCache.Login.RefreshToken, unmarshaledCache.Login.RefreshToken)
	r.Equal(originalCache.Login.AccessTokenExpiry, unmarshaledCache.Login.AccessTokenExpiry)
}

func TestGetCachedGeography(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		setupCache    func(t *testing.T, tempDir string)
		expectedGeo   string
		expectedError string
	}{
		{
			name: "valid_cache_with_us_geography",
			setupCache: func(t *testing.T, tempDir string) {
				createValidCacheFile(t, tempDir, "us")
			},
			expectedGeo: "us",
		},
		{
			name: "valid_cache_with_eu_geography",
			setupCache: func(t *testing.T, tempDir string) {
				createValidCacheFile(t, tempDir, "eu")
			},
			expectedGeo: "eu",
		},
		{
			name: "valid_cache_with_empty_geography",
			setupCache: func(t *testing.T, tempDir string) {
				createValidCacheFile(t, tempDir, "")
			},
			expectedGeo: "",
		},
		{
			name: "no_cache_file",
			setupCache: func(t *testing.T, tempDir string) {
				// Don't create any cache file
			},
			expectedGeo: "", // Should return empty string when file doesn't exist
		},
		{
			name: "invalid_json_cache_file",
			setupCache: func(t *testing.T, tempDir string) {
				createInvalidJSONCacheFile(t, tempDir)
			},
			expectedError: "failed to parse credential cache file",
		},
		{
			name: "empty_cache_file",
			setupCache: func(t *testing.T, tempDir string) {
				createEmptyCacheFile(t, tempDir)
			},
			expectedError: "failed to parse credential cache file",
		},
		{
			name: "cache_file_with_missing_login_section",
			setupCache: func(t *testing.T, tempDir string) {
				createCacheFileWithMissingLogin(t, tempDir)
			},
			expectedGeo: "", // Should return empty geography when login section is missing
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := require.New(t)

			// Create a temporary directory for this test
			tempDir := t.TempDir()

			// Setup the cache file (or not)
			tt.setupCache(t, tempDir)

			// Test GetCachedGeography
			actualGeo, err := GetCachedGeography(tempDir)

			if tt.expectedError != "" {
				r.Error(err)
				r.Contains(err.Error(), tt.expectedError)
			} else {
				r.NoError(err)
				r.Equal(tt.expectedGeo, actualGeo)
			}
		})
	}
}

func TestGetCachedGeography_InvalidDirectory(t *testing.T) {
	t.Parallel()
	r := require.New(t)

	// Test with directory that contains null bytes (invalid on most filesystems)
	geo, err := GetCachedGeography("/invalid\x00path")
	r.Error(err)
	// This should fail when trying to read the file, not resolve the directory
	r.Contains(err.Error(), "failed to read credential cache file")
	r.Equal("", geo)
}

func TestGetCachedGeography_ReadPermissionError(t *testing.T) {
	// Skip this test on Windows as file permissions work differently
	if os.Getenv("GOOS") == "windows" {
		t.Skip("Skipping permission test on Windows")
	}

	t.Parallel()
	r := require.New(t)

	// Create a temporary directory
	tempDir := t.TempDir()

	// Create a cache file
	createValidCacheFile(t, tempDir, "us")

	// Remove read permissions
	cacheFile := filepath.Join(tempDir, "creds-cache.json")
	err := os.Chmod(cacheFile, 0000)
	r.NoError(err)

	// Restore permissions after test to allow cleanup
	t.Cleanup(func() {
		r.NoError(os.Chmod(cacheFile, 0644))
	})

	// Test should fail with permission error
	geo, err := GetCachedGeography(tempDir)
	r.Error(err)
	r.Contains(err.Error(), "failed to read credential cache file")
	r.Equal("", geo)
}

// Helper functions for test setup

func createValidCacheFile(t *testing.T, tempDir, geography string) {
	cache := CredentialCache{
		Login: CredentialCacheLogin{
			AccessToken:       "test-token",
			RefreshToken:      "test-refresh",
			AccessTokenExpiry: "2024-12-31T23:59:59Z",
			Geography:         geography,
		},
	}

	data, err := json.Marshal(cache)
	require.NoError(t, err)

	cacheFile := filepath.Join(tempDir, "creds-cache.json")
	err = os.WriteFile(cacheFile, data, 0644)
	require.NoError(t, err)
}

func createInvalidJSONCacheFile(t *testing.T, tempDir string) {
	cacheFile := filepath.Join(tempDir, "creds-cache.json")
	invalidJSON := `{"login": {"access_token": "incomplete json"`
	err := os.WriteFile(cacheFile, []byte(invalidJSON), 0644)
	require.NoError(t, err)
}

func createEmptyCacheFile(t *testing.T, tempDir string) {
	cacheFile := filepath.Join(tempDir, "creds-cache.json")
	err := os.WriteFile(cacheFile, []byte(""), 0644)
	require.NoError(t, err)
}

func createCacheFileWithMissingLogin(t *testing.T, tempDir string) {
	// Create a valid JSON file but without the login section
	data := `{"other_field": "value"}`
	cacheFile := filepath.Join(tempDir, "creds-cache.json")
	err := os.WriteFile(cacheFile, []byte(data), 0644)
	require.NoError(t, err)
}

// Integration-style tests that simulate real-world usage scenarios

func TestGeographyWorkflow_ProfileInitScenario(t *testing.T) {
	t.Parallel()
	r := require.New(t)

	// Scenario: User has authenticated and cached geography exists, profile init should use it
	tempDir := t.TempDir()

	// Step 1: Simulate successful authentication with cached geography
	createValidCacheFile(t, tempDir, "eu")

	// Step 2: Simulate profile init trying to get cached geography
	cachedGeo, err := GetCachedGeography(tempDir)
	r.NoError(err)
	r.Equal("eu", cachedGeo)

	// Step 3: Verify the cached geography is valid
	r.True(IsValidGeography(cachedGeo))

	// Step 4: Profile should prefer cached over default
	defaultGeo := GetDefaultGeography()
	r.NotEqual(defaultGeo, cachedGeo) // In this test, they should be different
	r.Equal("us", defaultGeo)         // Verify default is still US
	r.Equal("eu", cachedGeo)          // But cached is EU
}

func TestGeographyWorkflow_AuthLoginSyncScenario(t *testing.T) {
	t.Parallel()
	r := require.New(t)

	// Scenario: User authenticates, changes region, needs to sync geography
	tempDir := t.TempDir()

	// Step 1: User initially authenticated to US
	createValidCacheFile(t, tempDir, "us")
	cachedGeo, err := GetCachedGeography(tempDir)
	r.NoError(err)
	r.Equal("us", cachedGeo)

	// Step 2: User switches authentication to EU (simulates new login)
	createValidCacheFile(t, tempDir, "eu") // Overwrites the cache
	newCachedGeo, err := GetCachedGeography(tempDir)
	r.NoError(err)
	r.Equal("eu", newCachedGeo)
	r.NotEqual(cachedGeo, newCachedGeo) // Geography changed

	// Step 3: Both old and new geographies should be valid
	r.True(IsValidGeography(cachedGeo))
	r.True(IsValidGeography(newCachedGeo))
}

func TestGeographyWorkflow_CleanSlateScenario(t *testing.T) {
	t.Parallel()
	r := require.New(t)

	// Scenario: New user with no cache file, should use defaults
	tempDir := t.TempDir()

	// Step 1: No cache file exists
	cachedGeo, err := GetCachedGeography(tempDir)
	r.NoError(err)
	r.Equal("", cachedGeo) // No cached geography

	// Step 2: Should fall back to default geography
	defaultGeo := GetDefaultGeography()
	r.Equal("us", defaultGeo)
	r.True(IsValidGeography(defaultGeo))

	// Step 3: All supported geographies should be valid options
	supportedGeos := GetSupportedGeographies()
	for _, geo := range supportedGeos {
		r.True(IsValidGeography(geo))
	}
	r.Contains(supportedGeos, defaultGeo)
}

func TestGeographyWorkflow_ValidationWorkflow(t *testing.T) {
	t.Parallel()
	r := require.New(t)

	// Scenario: Validate geography values throughout the workflow
	supportedGeos := GetSupportedGeographies()
	defaultGeo := GetDefaultGeography()

	// Step 1: Default should always be valid and supported
	r.True(IsValidGeography(defaultGeo))
	r.Contains(supportedGeos, defaultGeo)

	// Step 2: All constants should be valid and supported
	r.True(IsValidGeography(US))
	r.True(IsValidGeography(EU))
	r.Contains(supportedGeos, US)
	r.Contains(supportedGeos, EU)

	// Step 3: Test common invalid values that users might try
	invalidGeos := []string{
		"ap", "asia", "pacific", "americas", "europe", "asia-pacific",
		"us-east-1", "eu-west-1", "US", "EU", "", " ", "\t", "\n",
	}
	for _, invalidGeo := range invalidGeos {
		r.False(IsValidGeography(invalidGeo), "geography %q should be invalid", invalidGeo)
		r.NotContains(supportedGeos, invalidGeo)
	}
}

func TestGeographyWorkflow_CacheFileEvolution(t *testing.T) {
	t.Parallel()
	r := require.New(t)

	// Scenario: Test how the system handles cache file changes over time
	tempDir := t.TempDir()

	// Step 1: Start with no cache (new installation)
	geo, err := GetCachedGeography(tempDir)
	r.NoError(err)
	r.Equal("", geo)

	// Step 2: User authenticates to US
	createValidCacheFile(t, tempDir, "us")
	geo, err = GetCachedGeography(tempDir)
	r.NoError(err)
	r.Equal("us", geo)

	// Step 3: User switches to EU
	createValidCacheFile(t, tempDir, "eu")
	geo, err = GetCachedGeography(tempDir)
	r.NoError(err)
	r.Equal("eu", geo)

	// Step 4: Cache gets corrupted (simulating file system issues)
	createInvalidJSONCacheFile(t, tempDir)
	geo, err = GetCachedGeography(tempDir)
	r.Error(err)
	r.Contains(err.Error(), "failed to parse credential cache file")
	r.Equal("", geo)

	// Step 5: User re-authenticates, cache is restored
	createValidCacheFile(t, tempDir, "us")
	geo, err = GetCachedGeography(tempDir)
	r.NoError(err)
	r.Equal("us", geo)
}

func TestGeographyWorkflow_EdgeCases(t *testing.T) {
	t.Parallel()
	r := require.New(t)

	// Test various edge cases in a realistic workflow
	tempDir := t.TempDir()

	// Edge case 1: Cache file with empty geography string
	createValidCacheFile(t, tempDir, "")
	geo, err := GetCachedGeography(tempDir)
	r.NoError(err)
	r.Equal("", geo)

	// Edge case 2: Cache file with geography that becomes invalid (hypothetical future scenario)
	// Create a cache with a geography that's currently valid
	createValidCacheFile(t, tempDir, "us")
	geo, err = GetCachedGeography(tempDir)
	r.NoError(err)
	r.Equal("us", geo)
	r.True(IsValidGeography(geo))

	// Edge case 3: Verify the system can handle all supported geographies
	for _, supportedGeo := range GetSupportedGeographies() {
		createValidCacheFile(t, tempDir, supportedGeo)
		geo, err := GetCachedGeography(tempDir)
		r.NoError(err)
		r.Equal(supportedGeo, geo)
		r.True(IsValidGeography(geo))
	}
}
