// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package resourcename

import (
	"fmt"
	"strings"
)

// Pop pops the requested number of parts off of a resource name. See the
// following examples for behavior:
//
//	Pop("iam/project/e1e04e11-d590-41cf-b818-1535bc4b4889/service-principal/foo/key/bar", 1)
//	  returns "iam/project/e1e04e11-d590-41cf-b818-1535bc4b4889/service-principal/foo"
//
//	Pop("iam/project/e1e04e11-d590-41cf-b818-1535bc4b4889/service-principal/foo/key/bar", 2)
//	  returns "project/e1e04e11-d590-41cf-b818-1535bc4b4889"
func Pop(resourceName string, numParts int) (string, error) {
	// Validate numParts is positive
	if numParts <= 0 {
		return "", fmt.Errorf("numParts must be greater than zero")
	}

	ns, parts, err := parse(resourceName, withSkipNamePartValidation())
	if err != nil {
		return "", err
	}

	// Ensure we are not asked to pop more parts than is possible
	remainingParts := len(parts) - numParts
	if remainingParts <= 0 {
		return "", fmt.Errorf("numParts can not be equal to or greater than the parts the resource name contains")
	}

	// Check if the remaining part would just be an organization or project,
	// since then the namespace needs to be stripped
	if remainingParts == 1 {
		p := parts[0]
		switch p.Type {
		case OrganizationTypePart:
			return OrganizationPart(p.Name).String(), nil
		case ProjectTypePart:
			return ProjectPart(p.Name).String(), nil
		}

	}

	// Generate the new resource name
	partStrings := make([]string, 0, remainingParts+1)
	partStrings = append(partStrings, ns)
	for _, part := range parts[0:remainingParts] {
		partStrings = append(partStrings, part.Type, part.Name)
	}

	return strings.Join(partStrings, "/"), nil
}

// Push pushes the passed part onto an existing resource name. See the example:
//
//	 Push("iam/project/example/service-principal/foo", Part{Type: "key", Name: "bar"})
//		 returns "iam/project/example/service-principal/foo/key/bar"
func Push(resourceName string, part Part) (string, error) {
	// Parse the existing resource name
	ns, parts, err := Parse(resourceName)
	if err != nil {
		return "", err
	}

	// Push the new Part
	parts = append(parts, part)

	// Generate the new resource name
	return Generate(ns, parts...)
}
