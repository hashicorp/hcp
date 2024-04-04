// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package resourcename

import (
	"testing"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/stretchr/testify/require"
)

func TestParseAndValidateHappyPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                string
		resourceName        string
		resourceNamePattern string
		expectedNamespace   string
		expectedParts       []Part
	}{
		{
			name:                "One-level resource name",
			resourceName:        "namespace/type/name",
			resourceNamePattern: "namespace/type/*",
			expectedNamespace:   "namespace",
			expectedParts:       []Part{{"type", "name"}},
		},
		{
			name:                "Two-level resource name",
			resourceName:        "namespace/type-one/name1/type-two/name2",
			resourceNamePattern: "namespace/type-one/*/type-two/*",
			expectedNamespace:   "namespace",
			expectedParts:       []Part{{"type-one", "name1"}, {"type-two", "name2"}},
		},
		{
			name:                "Legacy name part with mixed case characters",
			resourceName:        "namespace/type-one/nAme1/type-two/Name2",
			resourceNamePattern: "namespace/type-one/*/type-two/*",
			expectedNamespace:   "namespace",
			expectedParts:       []Part{{"type-one", "nAme1"}, {"type-two", "Name2"}},
		},
	}
	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := Validate(test.resourceName)

			require.NoError(t, err)
		})

		t.Run(test.name+"_Ozzo", func(t *testing.T) {
			t.Parallel()

			err := validation.Validate(test.resourceName, IsResourceName)

			require.NoError(t, err)
		})

		t.Run(test.name+"_Ozzo_HasPattern", func(t *testing.T) {
			t.Parallel()

			err := validation.Validate(test.resourceName,
				HasResourceNamePattern(test.resourceNamePattern))

			require.NoError(t, err)
		})

		t.Run(test.name+"_Parse", func(t *testing.T) {
			t.Parallel()

			namespace, parts, err := Parse(test.resourceName)

			require.NoError(t, err)
			require.Equal(t, test.expectedNamespace, namespace)
			require.Equal(t, test.expectedParts, parts)
		})
	}
}

func TestValidateForOrganizationAndProjectHappyPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                string
		resourceName        string
		resourceNamePattern string
	}{
		{
			name:                "Organization specific Resource Name",
			resourceName:        "organization/test",
			resourceNamePattern: "organization/*",
		},
		{
			name:                "Organization specific Resource Name with a UUID as a name part",
			resourceName:        "organization/aaaa1111-2222-3333-4444-555566667777",
			resourceNamePattern: "organization/*",
		},
		{
			name:                "Project specific Resource Name",
			resourceName:        "project/test",
			resourceNamePattern: "project/*",
		},
		{
			name:                "Organization specific Resource Name with a UUID as a name part",
			resourceName:        "project/aaaa1111-2222-3333-4444-555566667777",
			resourceNamePattern: "project/*",
		},
		{
			name:                "Organization specific Legacy Resource Name",
			resourceName:        "organization/TEST",
			resourceNamePattern: "organization/*",
		},
		{
			name:                "Organization specific Legacy Resource Name with a UUID as a name part",
			resourceName:        "organization/AAAA1111-2222-3333-4444-555566667777",
			resourceNamePattern: "organization/*",
		},
		{
			name:                "Project specific Legacy Resource Name",
			resourceName:        "project/TEST",
			resourceNamePattern: "project/*",
		},
		{
			name:                "Organization specific Legacy Resource Name with a UUID as a name part",
			resourceName:        "project/AAAA1111-2222-3333-4444-555566667777",
			resourceNamePattern: "project/*",
		},
	}
	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := Validate(test.resourceName)

			require.NoError(t, err)
		})

		t.Run(test.name+"_Ozzo", func(t *testing.T) {
			t.Parallel()

			err := validation.Validate(test.resourceName, IsResourceName)

			require.NoError(t, err)
		})

		t.Run(test.name+"_Ozzo_HasPattern", func(t *testing.T) {
			t.Parallel()

			err := validation.Validate(test.resourceName,
				HasResourceNamePattern(test.resourceNamePattern))

			require.NoError(t, err)
		})
	}
}

func TestParseAndValidateFailingCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                string
		resourceName        string
		resourceNamePattern string
		expectedError       string
	}{
		{
			name:                "Malformed namespace",
			resourceName:        "⛔️/type-one/name1/type-two/name2",
			resourceNamePattern: "namespace/type-one/*/type-two/*",
			expectedError:       "a Resource Name's namespace must consist only of lowercase alphabetic characters and dashes",
		},
		{
			name:                "Malformed resource type part",
			resourceName:        "namespace/⛔️/name1/type-two/name2",
			resourceNamePattern: "namespace/⛔️/*/type-two/*",
			expectedError:       "a Resource Name's type parts must consist only of lowercase alphabetic characters and dashes",
		},
		{
			name:                "Malformed resource name part",
			resourceName:        "namespace/type-one/⛔️/type-two/name2",
			resourceNamePattern: "", // Skip because it Is valid if values are ignored.
			expectedError:       "a Resource Name's name parts must consist only of alphanumeric characters, dashes, underscores or dots",
		},
		{
			name:                "Legacy name part with mixed case characters",
			resourceName:        "namespace/type-one/⛔️/type-two/Name2",
			resourceNamePattern: "namespace/type-one/*/type-two/*",
			expectedError:       "a Resource Name's name parts must consist only of alphanumeric characters, dashes, underscores or dots",
		},
		{
			name:                "Only namespace is present",
			resourceName:        "namespace",
			resourceNamePattern: "namespace",
			expectedError:       errorResourceNameMustBeValid,
		},
		{
			name:                "Name part is missing",
			resourceName:        "namespace/type-one",
			resourceNamePattern: "namespace/type-one",
			expectedError:       errorResourceNameMustBeValid,
		},
		{
			name:                "Name part is missing in a two-level hierarchy",
			resourceName:        "namespace/type-one/name1/type-two",
			resourceNamePattern: "namespace/type-one/*/type-two",
			expectedError:       errorResourceNameMustBeValid,
		},
		{
			name:                "Empty namespace",
			resourceName:        "/type-one/name1/type-two/name2",
			resourceNamePattern: "/type-one/*/type-two/*",
			expectedError:       "a Resource Name's namespace must consist only of lowercase alphabetic characters and dashes",
		},
		{
			name:                "Empty type part",
			resourceName:        "namespace//name1/type-two/name2",
			resourceNamePattern: "namespace//*/type-two/*",
			expectedError:       "a Resource Name's type parts must consist only of lowercase alphabetic characters and dashes",
		},
		{
			name:                "Empty name part",
			resourceName:        "namespace/type-one//type-two/name2",
			resourceNamePattern: "namespace/type-one//type-two/*",
			expectedError:       "a Resource Name's name parts must consist only of alphanumeric characters, dashes, underscores or dots",
		},
	}
	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := Validate(test.resourceName)

			require.EqualError(t, err, test.expectedError)
		})

		t.Run(test.name+"_Ozzo", func(t *testing.T) {
			t.Parallel()

			err := validation.Validate(test.resourceName, IsResourceName)

			require.EqualError(t, err, test.expectedError)
		})

		t.Run(test.name+"_Ozzo_HasPattern", func(t *testing.T) {
			t.Parallel()

			if test.resourceNamePattern == "" {
				t.Skip()
			}

			err := validation.Validate(test.resourceName, HasResourceNamePattern(test.resourceNamePattern))

			// Use contains since the error may be prefixed because
			// the pattern is invalid
			require.Contains(t, err.Error(), test.expectedError)
		})

		t.Run(test.name+"_Parse", func(t *testing.T) {
			t.Parallel()

			_, _, err := Parse(test.resourceName)

			require.EqualError(t, err, test.expectedError)
		})
	}
}

func TestValidateForOrganizationAndProjectFailingCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                string
		resourceName        string
		resourceNamePattern string
		expectedError       string
	}{
		{
			name:                "Organization specific Resource Name with a malformed resource name part",
			resourceName:        "organization/⛔️",
			resourceNamePattern: "", // Skip because it Is valid if values are ignored.
			expectedError:       "a Resource Name's name parts must consist only of alphanumeric characters, dashes, underscores or dots",
		},
		{
			name:                "Project specific Resource Name with a malformed resource name part",
			resourceName:        "project/⛔️",
			resourceNamePattern: "", // Skip because it Is valid if values are ignored.
			expectedError:       "a Resource Name's name parts must consist only of alphanumeric characters, dashes, underscores or dots",
		},
		{
			name:                "Namespace is an organization string",
			resourceName:        "organization/type/name",
			resourceNamePattern: "organization/type/*",
			expectedError:       `a Resource Name cannot have a namespace that matches either "organization" or "project"`,
		},
		{
			name:                "Namespace is a project string",
			resourceName:        "project/type/name",
			resourceNamePattern: "project/type/*",
			expectedError:       `a Resource Name cannot have a namespace that matches either "organization" or "project"`,
		},
	}
	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := Validate(test.resourceName)

			require.EqualError(t, err, test.expectedError)
		})

		t.Run(test.name+"_Ozzo", func(t *testing.T) {
			t.Parallel()

			err := validation.Validate(test.resourceName, IsResourceName)

			require.EqualError(t, err, test.expectedError)
		})

		t.Run(test.name+"_Ozzo_HasPattern", func(t *testing.T) {
			t.Parallel()

			if test.resourceNamePattern == "" {
				t.Skip()
			}

			err := validation.Validate(test.resourceName,
				HasResourceNamePattern(test.resourceNamePattern))

			// Use contains since the error may be prefixed because
			// the pattern is invalid
			require.Contains(t, err.Error(), test.expectedError)
		})
	}
}

func TestHasResourceNamePattern_Failing(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                string
		resourceName        string
		resourceNamePattern string
		expectedError       string
	}{
		{
			name:                "Want Organization; Got Project",
			resourceName:        "project/aaaa1111-2222-3333-4444-555566667777",
			resourceNamePattern: "organization/*",
			expectedError:       "expected a resource name matching the pattern \"organization/*\"; got \"project/aaaa1111-2222-3333-4444-555566667777\"",
		},
		{
			name:                "Want Organization; Got Other",
			resourceName:        "namespace/type-one/name1/type-two/name2",
			resourceNamePattern: "organization/*",
			expectedError:       "expected a resource name matching the pattern \"organization/*\"; got \"namespace/type-one/name1/type-two/name2\"",
		},
		{
			name:                "Want Project; Got Organization",
			resourceName:        "organization/aaaa1111-2222-3333-4444-555566667777",
			resourceNamePattern: "project/*",
			expectedError:       "expected a resource name matching the pattern \"project/*\"; got \"organization/aaaa1111-2222-3333-4444-555566667777\"",
		},
		{
			name:                "Want Project; Got Other",
			resourceName:        "namespace/type-one/name1/type-two/name2",
			resourceNamePattern: "project/*",
			expectedError:       "expected a resource name matching the pattern \"project/*\"; got \"namespace/type-one/name1/type-two/name2\"",
		},
		{
			name:                "Mismatched at first level",
			resourceName:        "namespace/type-wrong/name1/type-two/name2",
			resourceNamePattern: "namespace/type-one/*/type-two/*",
			expectedError:       "expected a resource name matching the pattern \"namespace/type-one/*/type-two/*\"; got \"namespace/type-wrong/name1/type-two/name2\"",
		},
		{
			name:                "Mismatched at second level",
			resourceName:        "namespace/type-one/name1/type-wrong/name2",
			resourceNamePattern: "namespace/type-one/*/type-two/*",
			expectedError:       "expected a resource name matching the pattern \"namespace/type-one/*/type-two/*\"; got \"namespace/type-one/name1/type-wrong/name2\"",
		},
		{
			name:                "Matching except namespace",
			resourceName:        "namespace-wrong/type-one/name1/type-two/name2",
			resourceNamePattern: "namespace/type-one/*/type-two/*",
			expectedError:       "expected a resource name matching the pattern \"namespace/type-one/*/type-two/*\"; got \"namespace-wrong/type-one/name1/type-two/name2\"",
		},
	}
	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := ValidatePattern(test.resourceName, test.resourceNamePattern)

			require.EqualError(t, err, test.expectedError)
		})

		t.Run(test.name+"_Ozzo", func(t *testing.T) {
			t.Parallel()

			err := validation.Validate(test.resourceName, HasResourceNamePattern(test.resourceNamePattern))

			require.EqualError(t, err, test.expectedError)
		})
	}
}

func TestOzzoValidateForEmptyResourceName(t *testing.T) {
	t.Parallel()

	err := validation.Validate("", IsResourceName)

	require.NoError(t, err)
}
