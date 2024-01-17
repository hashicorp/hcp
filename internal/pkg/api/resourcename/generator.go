package resourcename

import (
	"fmt"
	"strings"
)

// Part is a tuple representing a resource level in a Resource Name as described
// in [RFC-344].
// It contains a resource type and a resource name.
//
// For example, in a Resource Name "namespace/type1/name1/type2/name2", Parts are
//
//	type1:name1 and type2:name2
//
// [RFC-344]: https://docs.google.com/document/d/1VY5pkYqKQ9-uQgQUIEBqgfVCKYTI99zchyMcwySgpdU/edit
type Part struct {
	Type string
	Name string
}

// String returns the string representation of an individual Part.
func (p Part) String() string {
	return fmt.Sprintf("%s/%s", p.Type, p.Name)
}

// Generate generates a Resource Name from a namespace and a list of parts.
func Generate(namespace string, parts ...Part) (string, error) {
	err := validateNamespace(namespace)
	if err != nil {
		return "", fmt.Errorf("failed to generate a new Resource Name: %w", err)
	}

	partStrings := make([]string, 0, len(parts)+1)

	partStrings = append(partStrings, namespace)
	for _, part := range parts {
		err := validateTypePart(part.Type)
		if err != nil {
			return "", fmt.Errorf("failed to generate a new Resource Name: %w", err)
		}
		partStrings = append(partStrings, part.Type)

		err = validateNamePart(part.Name)
		if err != nil {
			return "", fmt.Errorf("failed to generate a new Resource Name: %w", err)
		}
		partStrings = append(partStrings, part.Name)
	}

	return strings.Join(partStrings, "/"), nil
}

// OrganizationPart is a helper method that returns a Part based on the provided
// organization ID.
//
// For example:
//
//	name, err := resourcename.Generate(
//	  "namespace",
//	  resources.OrganizationPart("organizationID"),
//	)
//
//	Will return "namespace/organization/organizationID"
func OrganizationPart(organizationID string) Part {
	return Part{
		Type: OrganizationTypePart,
		Name: organizationID,
	}
}

// ProjectPart is a helper method that returns a Part based on the provided
// project ID.
//
// For example:
//
//	name, err := resourcename.Generate(
//	  "namespace",
//	  resources.ProjectPart("projectID"),
//	)
//
//	Will return "namespace/project/projectID
func ProjectPart(projectID string) Part {
	return Part{
		Type: ProjectTypePart,
		Name: projectID,
	}
}

// GeoPart is a helper method that returns a Part based on the provided
// geography.
//
// For example:
//
//	name, err := resourcename.Generate(
//	  "namespace",
//	  resources.ProjectPart("projectID"),
//	  resources.GeoPart("us"),
//	)
//
//	Will return "namespace/geo/geography
func GeoPart(geo string) Part {
	return Part{
		Type: GeoTypePart,
		Name: geo,
	}
}
