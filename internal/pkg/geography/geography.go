// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

// Package geography provides helpers for working with HCP geographies using the HCP SDK.
package geography

import (
	"github.com/hashicorp/hcp-sdk-go/config/geography"
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
