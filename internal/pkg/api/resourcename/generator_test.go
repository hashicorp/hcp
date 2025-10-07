// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: MPL-2.0

package resourcename

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerate_HappyPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                 string
		namespace            string
		typeIDs              []Part
		expectedResourceName string
	}{
		{
			name:                 "Simple resource",
			namespace:            "namespace",
			typeIDs:              []Part{{"type", "name"}},
			expectedResourceName: "namespace/type/name",
		},
		{
			name:                 "Simple resource mixed case",
			namespace:            "namespace",
			typeIDs:              []Part{{"type", "namePART"}},
			expectedResourceName: "namespace/type/namePART",
		},
		{
			name:                 "Nested resource",
			namespace:            "namespace",
			typeIDs:              []Part{{"type-one", "name1"}, {"type-two", "name2"}},
			expectedResourceName: "namespace/type-one/name1/type-two/name2",
		},
	}
	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			actualResourceName, err := Generate(test.namespace, test.typeIDs...)

			require.NoError(t, err)
			require.Equal(t, test.expectedResourceName, actualResourceName)
		})
	}
}

func TestGenerate_FailingCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		namespace     string
		typeIDs       []Part
		expectedError string
	}{
		{
			name:          "Namespace containing illegal characters",
			namespace:     "🔭",
			typeIDs:       []Part{{"type", "name"}},
			expectedError: "failed to generate a new Resource Name: a Resource Name's namespace must consist only of lowercase alphabetic characters and dashes",
		},
		{
			name:          "Type containing illegal characters",
			namespace:     "namespace",
			typeIDs:       []Part{{"😅", "name"}},
			expectedError: "failed to generate a new Resource Name: a Resource Name's type parts must consist only of lowercase alphabetic characters and dashes",
		},
		{
			name:          "ID containing illegal characters",
			namespace:     "namespace",
			typeIDs:       []Part{{"type", "🚀"}},
			expectedError: "failed to generate a new Resource Name: a Resource Name's name parts must consist only of alphanumeric characters, dashes, underscores or dots",
		},
	}
	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			actualResourceName, err := Generate(test.namespace, test.typeIDs...)

			require.EqualError(t, err, test.expectedError)
			require.Empty(t, actualResourceName)
		})
	}
}

func TestOrganizationPart(t *testing.T) {
	t.Parallel()

	givenOrganizationID := "test"

	resourceNamePart := OrganizationPart(givenOrganizationID)

	require.Equal(t, OrganizationTypePart, resourceNamePart.Type)
	require.Equal(t, givenOrganizationID, resourceNamePart.Name)
}

func TestProjectPart(t *testing.T) {
	t.Parallel()

	givenProjectID := "test"

	resourceNamePart := ProjectPart(givenProjectID)

	require.Equal(t, ProjectTypePart, resourceNamePart.Type)
	require.Equal(t, givenProjectID, resourceNamePart.Name)
}

func TestGeoPart(t *testing.T) {
	t.Parallel()

	givenGeo := "us"

	resourceNamePart := GeoPart(givenGeo)

	require.Equal(t, GeoTypePart, resourceNamePart.Type)
	require.Equal(t, givenGeo, resourceNamePart.Name)
}
