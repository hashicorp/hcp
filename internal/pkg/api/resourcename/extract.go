// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package resourcename

import (
	"fmt"
	"strings"
)

// ExtractOrganizationID extracts the organization ID set in the resource name
// or returns an error if the resource name is either invalid or doesn't contain
// an organization as its first part.
func ExtractOrganizationID(resourceName string) (string, error) {
	// Check if the resource name is just the organization resource name
	if parts := strings.SplitN(resourceName, "/", 3); len(parts) == 2 && parts[0] == OrganizationTypePart {
		return parts[1], nil
	}

	_, parts, err := parse(resourceName, withSkipNamePartValidation())
	if err != nil {
		return "", fmt.Errorf("invalid resource name: %w", err)
	}

	if parts[0].Type != OrganizationTypePart {
		return "", fmt.Errorf("resource name doesn't specify an organization ID")
	}

	return parts[0].Name, nil
}

// MustExtractOrganizationID extracts the organization ID from the resource name
// or panics if the passed resource name is invalid or doesn't contain an
// organization ID in its first part. This is intended to be a helper during
// tests only.
func MustExtractOrganizationID(resourceName string) string {
	org, err := ExtractOrganizationID(resourceName)
	if err != nil {
		panic(err)
	}

	return org
}

// ExtractProjectID extracts the project ID set in the resource name or returns
// an error if the resource name is either invalid or doesn't contain a project
// as its first part.
func ExtractProjectID(resourceName string) (string, error) {
	// Check if the resource name is just the project resource name
	if parts := strings.SplitN(resourceName, "/", 3); len(parts) == 2 && parts[0] == ProjectTypePart {
		return parts[1], nil
	}

	_, parts, err := parse(resourceName, withSkipNamePartValidation())
	if err != nil {
		return "", fmt.Errorf("invalid resource name: %w", err)
	}

	if parts[0].Type != ProjectTypePart {
		return "", fmt.Errorf("resource name doesn't specify a project ID")
	}

	return parts[0].Name, nil
}

// MustExtractProjectID extracts the project ID from the resource name or panics
// if the passed resource name is invalid or doesn't contain a project ID in its
// first part. This is intended to be a helper during tests only.
func MustExtractProjectID(resourceName string) string {
	proj, err := ExtractProjectID(resourceName)
	if err != nil {
		panic(err)
	}

	return proj
}

// ExtractOrganizationOrProjectID extracts either a project ID or organization
// ID from the resource name. It returns an error if the resource name is
// invalid or doesn't contain either a project or organization ID in its first
// type part.
func ExtractOrganizationOrProjectID(resourceName string) (organizationID, projectID string, err error) {
	// Check if the resource name is just the organization resource name
	if parts := strings.SplitN(resourceName, "/", 3); len(parts) == 2 {
		switch parts[0] {
		case OrganizationTypePart:
			organizationID = parts[1]
			return
		case ProjectTypePart:
			projectID = parts[1]
			return
		}
	}

	_, parts, err := parse(resourceName, withSkipNamePartValidation())
	if err != nil {
		return "", "", fmt.Errorf("invalid resource name: %w", err)
	}

	switch parts[0].Type {
	case OrganizationTypePart:
		organizationID = parts[0].Name
	case ProjectTypePart:
		projectID = parts[0].Name
	default:
		return "", "", fmt.Errorf("resource name doesn't specify an organization or project ID")
	}

	return
}

// MustExtractOrganizationOrProjectID extracts either a project ID or
// organization ID from the resource name. It panics if the resource name is
// invalid or doesn't contain either a project or organization ID in its first
// type part. This is intended to be a helper during tests only.
func MustExtractOrganizationOrProjectID(resourceName string) (organizationID, projectID string) {
	org, proj, err := ExtractOrganizationOrProjectID(resourceName)
	if err != nil {
		panic(err)
	}

	return org, proj
}

// ExtractGeo extracts the geography set in the resource name or returns
// an error if the resource name is either invalid or doesn't contain a geography.
func ExtractGeo(resourceName string) (string, error) {
	_, parts, err := parse(resourceName, withSkipNamePartValidation())
	if err != nil {
		return "", fmt.Errorf("invalid resource name: %w", err)
	}

	if parts[1].Type != GeoTypePart {
		return "", fmt.Errorf("resource name doesn't specify a geography")
	}

	return parts[1].Name, nil
}

// MustExtractGeo extracts the geography from the resource name or panics
// if the passed resource name is invalid or doesn't contain a geography.
// This is intended to be a helper during tests only.
func MustExtractGeo(resourceName string) string {
	geo, err := ExtractGeo(resourceName)
	if err != nil {
		panic(err)
	}

	return geo
}
