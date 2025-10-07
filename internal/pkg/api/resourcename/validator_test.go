// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: MPL-2.0

package resourcename

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateNamespaceHappyPath(t *testing.T) {
	t.Parallel()

	err := validateNamespace("resource-manager")

	require.NoError(t, err)
}

func TestValidateNamespaceFailingCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		namespace     string
		expectedError string
	}{
		{
			name:          "Namespace consists of uppercase characters",
			namespace:     "nAMESPACE",
			expectedError: "a Resource Name's namespace must consist only of lowercase alphabetic characters and dashes",
		},
		{
			name:          "Namespace consists of non-latin characters",
			namespace:     "хашикорп",
			expectedError: "a Resource Name's namespace must consist only of lowercase alphabetic characters and dashes",
		},
		{
			name:          "Namespace consists of non-alphabetic characters",
			namespace:     "@",
			expectedError: "a Resource Name's namespace must consist only of lowercase alphabetic characters and dashes",
		},
		{
			name:          "Namespace contains only numbers",
			namespace:     "1",
			expectedError: "a Resource Name's namespace must consist only of lowercase alphabetic characters and dashes",
		},
		{
			name:          "Empty string",
			namespace:     "",
			expectedError: "a Resource Name's namespace must consist only of lowercase alphabetic characters and dashes",
		},
		{
			name:          "Namespace starting with a dash",
			namespace:     "-namespace",
			expectedError: "a Resource Name's namespace must consist only of lowercase alphabetic characters and dashes",
		},
		{
			name:          "Namespace is an organization string",
			namespace:     "organization",
			expectedError: `a Resource Name cannot have a namespace that matches either "organization" or "project"`,
		},
		{
			name:          "Namespace is a project string",
			namespace:     "project",
			expectedError: `a Resource Name cannot have a namespace that matches either "organization" or "project"`,
		},
	}
	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := validateNamespace(test.namespace)

			require.EqualError(t, err, test.expectedError)
		})
	}
}

func TestValidateTypePartHappyPath(t *testing.T) {
	t.Parallel()

	err := validateTypePart("resource-manager-test")

	require.NoError(t, err)
}

func TestValidateTypePartFailingCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		typePart string
	}{
		{
			name:     "Type consists of uppercase characters",
			typePart: "tYPE",
		},
		{
			name:     "Type consists of non-latin characters",
			typePart: "хашикорп",
		},
		{
			name:     "Type consists of non-alphabetic characters",
			typePart: "@",
		},
		{
			name:     "Type contains only numbers",
			typePart: "1",
		},
		{
			name:     "Empty string",
			typePart: "",
		},
		{
			name:     "Type starting with a dash",
			typePart: "-type",
		},
	}
	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := validateTypePart(test.typePart)

			require.EqualError(t, err, "a Resource Name's type parts must consist only of lowercase alphabetic characters and dashes")
		})
	}
}

func TestValidateNamePartHappyPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		namePart string
	}{
		{
			name:     "Name part consists of lowercase characters",
			namePart: "a1-test.hello_world",
		},
		{
			name:     "Name part consists of upercase characters",
			namePart: "HELLO-WORLD",
		},
		{
			name:     "Name part starts with a digit",
			namePart: "9f7c164a-e838-4d96-8329-0593b4b66900",
		},
		{
			name:     "Name part consists of mixed case characters",
			namePart: "A1-Test.Hello_World",
		},
		{
			name:     "Name part starts with a digit and mixed case characters",
			namePart: "9F7C164A-e838-4d96-8329-0593b4b66900",
		},
	}
	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := validateNamePart(test.namePart)

			require.NoError(t, err)
		})
	}
}

func TestValidateNamePartFailingCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		namePart      string
		expectedError string
	}{
		{
			name:          "Name part consists of non-latin characters",
			namePart:      "хашикорп",
			expectedError: "a Resource Name's name parts must consist only of alphanumeric characters, dashes, underscores or dots",
		},
		{
			name:          "Name part consists of non-alphabetic characters",
			namePart:      "@",
			expectedError: "a Resource Name's name parts must consist only of alphanumeric characters, dashes, underscores or dots",
		},
		{
			name:          "Empty string",
			namePart:      "",
			expectedError: "a Resource Name's name parts must consist only of alphanumeric characters, dashes, underscores or dots",
		},
		{
			name:          "Name part starting with a dot",
			namePart:      ".name",
			expectedError: "a Resource Name's name parts must consist only of alphanumeric characters, dashes, underscores or dots",
		},
		{
			name:          "Name part starting with a dash",
			namePart:      "-name",
			expectedError: "a Resource Name's name parts must consist only of alphanumeric characters, dashes, underscores or dots",
		},
		{
			name:          "Name part starting with an underscore",
			namePart:      "_name",
			expectedError: "a Resource Name's name parts must consist only of alphanumeric characters, dashes, underscores or dots",
		},
		{
			name:          "Name part consists of non-latin characters",
			namePart:      "хашикорп",
			expectedError: "a Resource Name's name parts must consist only of alphanumeric characters, dashes, underscores or dots",
		},
		{
			name:          "Name part consists of non-alphabetic characters",
			namePart:      "@",
			expectedError: "a Resource Name's name parts must consist only of alphanumeric characters, dashes, underscores or dots",
		},
		{
			name:          "Empty string",
			namePart:      "",
			expectedError: "a Resource Name's name parts must consist only of alphanumeric characters, dashes, underscores or dots",
		},
		{
			name:          "Name part starting with a dot",
			namePart:      ".name",
			expectedError: "a Resource Name's name parts must consist only of alphanumeric characters, dashes, underscores or dots",
		},
		{
			name:          "Name part starting with a dash",
			namePart:      "-name",
			expectedError: "a Resource Name's name parts must consist only of alphanumeric characters, dashes, underscores or dots",
		},
		{
			name:          "Name part starting with an underscore",
			namePart:      "_name",
			expectedError: "a Resource Name's name parts must consist only of alphanumeric characters, dashes, underscores or dots",
		},
	}
	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := validateNamePart(test.namePart)
			require.EqualError(t, err, test.expectedError)
		})
	}
}
