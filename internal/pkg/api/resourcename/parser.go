// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package resourcename

import (
	"errors"
	"strings"
)

const errorResourceNameMustBeValid = "resource name must consist of a namespace, and a list of type and name parts"

// Parse parses the provided Resource Name and returns individual
// components. If the Resource Name doesn't satisfy the requirements, the
// function returns a validation error. Resource Name must have the following
// format:
//
//	<namespace>/<type a>/<name a>/<type b>/<name b>/…/<type z>/<name z>.
//
// Where
//   - namespace - is expected to be based on the service name and may contain
//     lower-case alphabetic characters as well as dashes (-)
//   - type part - may contain lower-case alphabetic characters as well as
//     dashes (-)
//   - name part - is case-sensitive and may contain lower-case alphanumeric
//     characters as well as dashes (-), dots (.) and underscores (_)
//
// For example:
//   - vault/project/e1e04e11-d590-41cf-b818-1535bc4b4889/cluster/my-cluster
//   - vagrant/organization/ubuntu/box/lunar64/version/v20230130.0.0
//
// See [RFC-344] for more information.
//
// [RFC-344]: https://docs.google.com/document/d/1VY5pkYqKQ9-uQgQUIEBqgfVCKYTI99zchyMcwySgpdU/edit
func Parse(resourceName string) (namespace string, parts []Part, error error) {
	return parse(resourceName)
}

func parse(resourceName string, options ...parserOption) (namespace string, parts []Part, error error) {
	partStrings := strings.Split(resourceName, "/")

	// A Resource Name must consist of at least 3 parts (namespace/type/name) and
	// have an odd number of parts.

	if len(partStrings) < 3 || len(partStrings)%2 == 0 {
		return "", nil, errors.New(errorResourceNameMustBeValid)
	}

	namespace = partStrings[0]

	err := validateNamespace(namespace)
	if err != nil {
		return "", nil, err
	}

	typeAndNameParts := partStrings[1:]

	parts = make([]Part, 0, len(typeAndNameParts)/2)
	for i := 0; i < len(typeAndNameParts); i += 2 {
		err := validateTypePart(typeAndNameParts[i])
		if err != nil {
			return "", nil, err
		}

		err = validateNamePart(typeAndNameParts[i+1], options...)
		if err != nil {
			return "", nil, err
		}

		parts = append(parts, Part{typeAndNameParts[i], typeAndNameParts[i+1]})
	}

	return namespace, parts, nil
}

// parserConfig defines Resource Name parser settings.
type parserConfig struct {
	// allowWildcardNamePart allows resource names with a name part value of '*'
	// to be valid.
	allowWildcardNamePart bool

	// skipNamePartValidation skips validation of the name part.
	skipNamePartValidation bool
}

// parserOption provides a way to pass parser settings parameters as a function
// argument.
type parserOption func(*parserConfig)

func processOptions(options ...parserOption) parserConfig {
	validatorConfig := parserConfig{}

	for _, option := range options {
		option(&validatorConfig)
	}

	return validatorConfig
}

// withWildcardValues enables a resource name with name part values of '*' to be
// valid.
func withWildcardValues() parserOption {
	return func(config *parserConfig) {
		config.allowWildcardNamePart = true
	}
}

// withSkipNamePartValidation skips validation of name parts. This can be useful
// when parsing a resource name after validation has already occurred.
func withSkipNamePartValidation() parserOption {
	return func(config *parserConfig) {
		config.skipNamePartValidation = true
	}
}
