// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: MPL-2.0

package resourcename

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPop_FailingCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		resourceName  string
		numParts      int
		expectedError string
	}{
		{
			name:          "Negative numParts",
			resourceName:  "namespace/type-one/name1",
			numParts:      -1,
			expectedError: "numParts must be greater than zero",
		},
		{
			name:          "zero numParts",
			resourceName:  "namespace/type-one/name1",
			numParts:      0,
			expectedError: "numParts must be greater than zero",
		},
		{
			name:          "Pop an organization",
			resourceName:  "organization/aaaa1111-2222-3333-4444-555566667777",
			numParts:      1,
			expectedError: errorResourceNameMustBeValid,
		},
		{
			name:          "Pop a project",
			resourceName:  "project/aaaa1111-2222-3333-4444-555566667777",
			numParts:      1,
			expectedError: errorResourceNameMustBeValid,
		},
		{
			name:          "Pop equal to number of parts",
			resourceName:  "namespace/type-one/name1/type-two/name2",
			numParts:      2,
			expectedError: "numParts can not be equal to or greater than the parts the resource name contains",
		},
		{
			name:          "Pop more than the number of parts",
			resourceName:  "namespace/type-one/name1/type-two/name2",
			numParts:      3,
			expectedError: "numParts can not be equal to or greater than the parts the resource name contains",
		},
	}
	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			_, err := Pop(test.resourceName, test.numParts)

			require.EqualError(t, err, test.expectedError)
		})
	}
}

func TestPop_Good(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                 string
		resourceName         string
		numParts             int
		expectedResourceName string
	}{
		{
			name:                 "Pop one off nested",
			resourceName:         "namespace/type-one/name1/type-two/name2",
			numParts:             1,
			expectedResourceName: "namespace/type-one/name1",
		},
		{
			name:                 "Pop two off nested",
			resourceName:         "namespace/type-one/name1/type-two/name2/type-three/name3",
			numParts:             2,
			expectedResourceName: "namespace/type-one/name1",
		},
		{
			name:                 "Pop til project",
			resourceName:         "namespace/project/project-id/type-two/name2/type-three/name3",
			numParts:             2,
			expectedResourceName: "project/project-id",
		},
		{
			name:                 "Pop til organization",
			resourceName:         "namespace/organization/project-id/type-two/name2/type-three/name3",
			numParts:             2,
			expectedResourceName: "organization/project-id",
		},
		{
			name:                 "Ensure skipping of name validation",
			resourceName:         "namespace/type-one/⛔️/type-two/name2",
			numParts:             1,
			expectedResourceName: "namespace/type-one/⛔️",
		},
	}
	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			rn, err := Pop(test.resourceName, test.numParts)
			require.NoError(t, err)
			require.Equal(t, test.expectedResourceName, rn)
		})
	}
}

func TestPush_FailingCase(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		resourceName  string
		part          Part
		expectedError string
	}{
		{
			name:          "Invalid original resource name",
			resourceName:  "namespace",
			part:          Part{Type: "type-one", Name: "name1"},
			expectedError: "resource name must consist of a namespace, and a list of type and name parts",
		},
		{
			name:          "Invalid part given no options",
			resourceName:  "namespace/type-one/name1",
			part:          Part{Type: "type-two", Name: "*name2"},
			expectedError: "failed to generate a new Resource Name: a Resource Name's name parts must consist only of alphanumeric characters, dashes, underscores or dots",
		},
		{
			name:          "Bad part",
			resourceName:  "namespace/type-one/name1",
			part:          Part{Type: "type-two", Name: ""},
			expectedError: "failed to generate a new Resource Name: a Resource Name's name parts must consist only of alphanumeric characters, dashes, underscores or dots",
		},
	}
	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			_, err := Push(test.resourceName, test.part)

			require.EqualError(t, err, test.expectedError)
		})
	}
}

func TestPush_Good(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                 string
		resourceName         string
		part                 Part
		expectedResourceName string
	}{
		{
			name:                 "Push valid",
			resourceName:         "namespace/type-one/name1",
			part:                 Part{Type: "type-two", Name: "name2"},
			expectedResourceName: "namespace/type-one/name1/type-two/name2",
		},
		{
			name:                 "Push nested",
			resourceName:         "namespace/type-one/name1/type-two/name2",
			part:                 Part{Type: "type-three", Name: "name3"},
			expectedResourceName: "namespace/type-one/name1/type-two/name2/type-three/name3",
		},
		{
			name:                 "Push mixed case",
			resourceName:         "namespace/type-one/name1",
			part:                 Part{Type: "type-two", Name: "Name2"},
			expectedResourceName: "namespace/type-one/name1/type-two/Name2",
		},
	}
	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			rn, err := Push(test.resourceName, test.part)
			require.NoError(t, err)
			require.Equal(t, test.expectedResourceName, rn)
		})
	}
}
