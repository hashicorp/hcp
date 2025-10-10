// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package geography

import (
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
}

func TestIsValidGeography(t *testing.T) {
	t.Parallel()
	r := require.New(t)

	// Valid geographies
	r.True(IsValidGeography("us"))
	r.True(IsValidGeography("eu"))

	// Invalid geographies
	r.False(IsValidGeography("ap"))
	r.False(IsValidGeography("invalid"))
	r.False(IsValidGeography(""))
}

func TestGetDefaultGeography(t *testing.T) {
	t.Parallel()
	r := require.New(t)

	defaultGeo := GetDefaultGeography()
	r.Equal("us", defaultGeo) // SDK default should be US
}

func TestConstants(t *testing.T) {
	t.Parallel()
	r := require.New(t)

	r.Equal("us", US)
	r.Equal("eu", EU)
}
