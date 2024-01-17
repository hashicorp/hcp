package resourcename

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation"
)

var (
	// IsNamePart returns ozzo validator that verifies if given string is valid resource name part.
	// Name parts are case-sensitive and may contain lower-case alphanumeric
	// characters as well as dashes (-), dots (.) and underscores (_)
	//
	// For example:
	//   - e1e04e11-d590-41cf-b818-1535bc4b4889
	//   - my_cluster
	IsNamePart = validation.By(func(value interface{}) error {
		s, ok := value.(string)
		if !ok {
			return fmt.Errorf("invalid resource name part type: %T", value)
		}

		return validateNamePart(s)
	})

	// IsResourceName is a validation rule that evaluates whether the provided Resource Name satisfies the
	// requirements. Resource Name must have the following format:
	//
	//	<namespace>/<type a>/<name a>/<type b>/<name b>/…/<type z>/<name z>.
	//
	// Where
	//   - namespace - is expected to be based on the service name and may contain
	//     lower-case alphabetic characters as well as dashes (-)
	//   - type part - may contain lower-case alphabetic characters as well as
	//     dashes (-)
	//   - name part - is case-sensitive and may contain mixed-case alphanumeric
	//     characters as well as dashes (-), dots (.) and underscores (_)
	//
	// For example:
	//   - vault/project/e1e04e11-d590-41cf-b818-1535bc4b4889/cluster/my-cluster
	//   - vagrant/organization/ubuntu/box/lunar64/version/v20230130.0.0
	//
	// See [RFC-344] for more information.
	//
	// [RFC-344]: https://docs.google.com/document/d/1VY5pkYqKQ9-uQgQUIEBqgfVCKYTI99zchyMcwySgpdU/edit
	IsResourceName = validation.By(func(value interface{}) error {
		s, ok := value.(string)
		if !ok {
			return fmt.Errorf("invalid resource name type: %T", value)
		}

		// Don't validate empty values
		if s == "" {
			return nil
		}

		return Validate(s)
	})
)

// Validate checks that the provided Resource Name satisfies the
// requirements. Resource Name must have the following format:
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
// Projects and organizations adopt shorter resource names that exclude the
// <namespace> part.
//
// For example:
//   - organization/863063b6-c485-4cf4-8df3-6991a2ffd4b4
//   - project/e1e04e11-d590-41cf-b818-1535bc4b4889
//   - project/gTmD2KMMRt9DWhJz
//
// See [RFC-344] for more information.
//
// [RFC-344]: https://docs.google.com/document/d/1VY5pkYqKQ9-uQgQUIEBqgfVCKYTI99zchyMcwySgpdU/edit
func Validate(resourceName string) error {
	// Check if the Resource Name belongs to a resource-manager resource.
	if parts := strings.Split(resourceName, "/"); consistsOfTwoPartsAndStartsWithOrganizationOrProject(parts) {
		return validateNamePart(parts[1])
	}

	// Try to parse the Resource Name consisting of 3 parts (namespace/type/name).
	_, _, err := Parse(resourceName)

	return err
}

// ValidatePattern validates whether the provided Resource Name is a valid
// resource name and that it matches the passed resource name pattern. See
// Validate for details on what consistitutes a valid resource name.
//
// The passed pattern should be the format of resource name that is expected,
// with the values replaced with a '*'. The namespace, and part types are
// checked against the resource name being validated but the values are ignored.
// The pattern can exclude the namespace for organization and project resource
// names.
//
// See the following examples where validation passes:
//
//	ValidatePattern("organization/e1e04e11-d590-41cf-b818-1535bc4b4889", "organization/*")
//
//	ValidatePattern("project/e1e04e11-d590-41cf-b818-1535bc4b4889", "project/*")
//
//	ValidatePattern("iam/project/e1e04e11-d590-41cf-b818-1535bc4b4889/service-principal/prod-sp", "iam/project/*/service-principal/*")
//
// See the following examples where validation fails:
//
//	ValidatePattern("project/e1e04e11-d590-41cf-b818-1535bc4b4889", "iam/project/*/service-principal/*")
//
//	ValidatePattern("iam/project/e1e04e11-d590-41cf-b818-1535bc4b4889/service-principal/prod-sp", "iam/organization/*/group/*")
//
//	ValidatePattern("network/project/e1e04e11-d590-41cf-b818-1535bc4b4889/hvn/aws", "iam/organization/*/group/*")
func ValidatePattern(resourceName, resourceNamePattern string) error {
	patternError := fmt.Errorf("expected a resource name matching the pattern %q; got %q", resourceNamePattern, resourceName)

	// Check if the pattern is matching against a resource-manager resource.
	if patternParts := strings.Split(resourceNamePattern, "/"); consistsOfTwoPartsAndStartsWithOrganizationOrProject(patternParts) {
		// Validate that the name part is valid on the pattern.
		if err := validateNamePart(patternParts[1], withWildcardValues()); err != nil {
			return err
		}

		// Check that the passed resource name match the pattern
		nameParts := strings.Split(resourceName, "/")
		if len(nameParts) != 2 || patternParts[0] != nameParts[0] {
			return patternError
		}

		return nil
	}

	// Parse the resource name pattern
	patternNs, patternParts, err := parse(resourceNamePattern, withWildcardValues())
	if err != nil {
		return fmt.Errorf("failed to parse resource name pattern: %w", err)
	}

	rnNs, rnParts, err := Parse(resourceName)
	if err != nil {
		return err
	}

	// Check the namespaces match
	if patternNs != rnNs {
		return patternError
	}

	// Check the number of parts match
	if len(patternParts) != len(rnParts) {
		return patternError
	}

	// Validate that each name part is of the same type
	for i, patternPart := range patternParts {
		if patternPart.Type != rnParts[i].Type {
			return patternError
		}
	}

	return nil
}

// HasResourceNamePattern is a validation rule that evaluates whether the
// provided Resource Name is a valid resource name and that it matches the
// passed resource name pattern. See Validate for details on what
// consistitutes a valid resource name.
//
// The passed pattern should be the format of resource name that is expected,
// with the values replaced with a '*'. The namespace, and part types are
// checked against the resource name being validated but the values are ignored.
// The pattern can exclude the namespace for organization and project resource
// names.
//
// See the following examples where validation passes:
//
//	HasResourceNamePattern("organization/*").
//	  Validate("organization/e1e04e11-d590-41cf-b818-1535bc4b4889")
//
//	HasResourceNamePattern("project/*").
//	  Validate("project/e1e04e11-d590-41cf-b818-1535bc4b4889")
//
//	HasResourceNamePattern("iam/project/*/service-principal/*").
//	  Validate("iam/project/e1e04e11-d590-41cf-b818-1535bc4b4889/service-principal/prod-sp")
//
// See the following examples where validation fails:
//
//	HasResourceNamePattern("iam/project/*/service-principal/*").
//	  Validate("project/e1e04e11-d590-41cf-b818-1535bc4b4889")
//
//	HasResourceNamePattern("iam/organization/*/group/*").
//	  Validate("iam/project/e1e04e11-d590-41cf-b818-1535bc4b4889/service-principal/prod-sp")
//
//	HasResourceNamePattern("iam/organization/*/group/*").
//	  Validate("network/project/e1e04e11-d590-41cf-b818-1535bc4b4889/hvn/aws")
func HasResourceNamePattern(pattern string) *hasResourceNamePattern {
	return &hasResourceNamePattern{
		pattern: pattern,
	}
}

// hasResourceNamePattern implements ozzo validator interface.
type hasResourceNamePattern struct {
	pattern string
}

// Validate implements ozzo validator interface.
func (v *hasResourceNamePattern) Validate(value interface{}) error {
	s, ok := value.(string)
	if !ok {
		return fmt.Errorf("invalid resource name type: %T", value)
	}

	return ValidatePattern(s, v.pattern)
}

func consistsOfTwoPartsAndStartsWithOrganizationOrProject(parts []string) bool {
	return len(parts) == 2 &&
		(parts[0] == OrganizationTypePart || parts[0] == ProjectTypePart)
}

const lowercaseAlphabeticAndDashes = "^[a-z][a-z-]*$"

// validateNamespace checks that the provided namespace satisfies
// requirements. Namespaces are expected to be based on the service name and may
// contain lower-case alphabetic characters as well as dashes (-). Namespace
// values such as "organization" or "project" are reserved for specific use cases
// of resource names. It is not allowed to have a namespace that matches either
// of these reserved terms.
//
// For example:
//   - vault
//   - vagrant
//
// See [RFC-344] for more information.
//
// [RFC-344]: https://docs.google.com/document/d/1VY5pkYqKQ9-uQgQUIEBqgfVCKYTI99zchyMcwySgpdU/edit
func validateNamespace(namespace string) error {
	if namespace == OrganizationTypePart || namespace == ProjectTypePart {
		return errors.New(`a Resource Name cannot have a namespace that matches either "organization" or "project"`)
	}

	if matched, _ := regexp.MatchString(lowercaseAlphabeticAndDashes, namespace); !matched {
		return errors.New("a Resource Name's namespace must consist only of lowercase alphabetic characters and dashes")
	}
	return nil
}

// validateTypePart checks that the provided type satisfies requirements.
// Type parts may contain lower-case alphabetic characters as well as dashes (-)
//
// For example:
//   - organization
//   - project
//   - cluster
//
// See [RFC-344] for more information.
//
// [RFC-344]: https://docs.google.com/document/d/1VY5pkYqKQ9-uQgQUIEBqgfVCKYTI99zchyMcwySgpdU/edit
func validateTypePart(resourceType string) error {
	if matched, _ := regexp.MatchString(lowercaseAlphabeticAndDashes, resourceType); !matched {
		return errors.New("a Resource Name's type parts must consist only of lowercase alphabetic characters and dashes")
	}
	return nil
}

const (
	mixedCaseAlphanumericAndDashes  = "^[A-Za-z0-9][A-Za-z0-9_.-]*$"
	patternAsteriskOnly             = "^*$"
	errorTextNameMustBeAlphaNumeric = "a Resource Name's name parts must consist only of alphanumeric characters, dashes, underscores or dots"
	errorTextNameMustBeAnAsterick   = "the Resource Name pattern must specify name parts with an asterick ('*')"
)

// validateNamePart checks that the provided ID satisfies requirements. See IsNamePart doc for more info.
func validateNamePart(resourceType string, options ...parserOption) error {
	validatorConfig := processOptions(options...)

	// Skip validation
	if validatorConfig.skipNamePartValidation {
		return nil
	}

	pattern := mixedCaseAlphanumericAndDashes
	errorText := errorTextNameMustBeAlphaNumeric
	if validatorConfig.allowWildcardNamePart {
		pattern = patternAsteriskOnly
		errorText = errorTextNameMustBeAnAsterick
	}

	if matched, _ := regexp.MatchString(pattern, resourceType); !matched {
		return errors.New(errorText)
	}

	return nil
}
