// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: MPL-2.0

package profile

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/hashicorp/hcp/internal/pkg/api/resourcename"
	"github.com/posener/complete"
	"golang.org/x/exp/maps"
)

const (
	// ActiveProfileFileName is the file name of the active profile stored in
	// the ConfigDir.
	ActiveProfileFileName = "active_profile.hcl"
)

var (
	// ErrNoProfileFilePresent is returned when the requested profile does not
	// exist.
	ErrNoProfileFilePresent = errors.New("profile configuration file doesn't exist")

	// ErrInvalidProfileName is returned if a profile is created with an invalid
	// profile name.
	ErrInvalidProfileName = errors.New("profile name may only include a-z, A-Z, 0-9, or '-', must start with a letter, and can be no longer than 64 characters")
)

// ActiveProfile stores the active profile.
type ActiveProfile struct {
	Name string `hcl:"name"`

	// dir is the directory the active profile should be written to.
	dir string
}

// Write writes the active profile to disk.
func (c *ActiveProfile) Write() error {
	path := filepath.Join(c.dir, ActiveProfileFileName)
	f := hclwrite.NewEmptyFile()
	gohcl.EncodeIntoBody(c, f.Body())
	return os.WriteFile(path, f.Bytes(), 0o666)
}

// Profile is a named set of configuration for the HCP CLI. It captures common
// configuration values such as the organization and project being interacted
// with, but also allows storing service specific configuration.
type Profile struct {
	// Name is the name of the profile
	Name string `hcl:"name"`

	// OrganizationID stores the organization_id to make requests against.
	OrganizationID string `hcl:"organization_id"`

	// ProjectID stores the project_id to make requests against.
	ProjectID string `hcl:"project_id"`

	// Core stores core CLI configuration values.
	Core *Core `hcl:"core,block" json:",omitempty"`

	/*
		To add a new component:

		1. Implement a struct with the configuration fields. Set the hcl name and json omitempty.
		2. Implement Predict method for autocompleting. This is used for hcp profile set/unset/read.
		3. Implement a Validate function if the config can be statically validated.
		4. Implement a isEmpty function and add to the profile.Clean() method.
		   This ensures the written configuration doesn't include empty stanzas.
		5. Document the properties in internal/commands/profile/property_docs.go
	*/

	// VaultSecrets stores vault-secrets CLI configuration values
	VaultSecrets *VaultSecretsConf `hcl:"vault-secrets,block" json:",omitempty"`

	// dir is the directory the profile should write to.
	dir string
}

// Predict predicts the HCL key names and basic settable values
func (p *Profile) Predict(args complete.Args) []string {
	sub := map[string]complete.Predictor{
		"core":          p.Core,
		"vault-secrets": p.VaultSecrets,
	}

	if len(args.All) >= 1 {
		c, ok := sub[strings.Split(args.All[0], "/")[0]]
		if ok {
			return c.Predict(args)
		}
	}

	// predicting the property
	if len(args.All) == 1 {
		return []string{"organization_id", "project_id", "core/", "vault-secrets/"}
	}

	return nil
}

// Validate validates that the set values are valid. It validates parameters
// that do not require any communication with HCP.
func (p *Profile) Validate() error {
	var err *multierror.Error

	const nameRegex = "^[A-Za-z][A-Za-z0-9-]{0,63}$"
	if matched, _ := regexp.MatchString(nameRegex, p.Name); !matched {
		err = multierror.Append(err, ErrInvalidProfileName)
	}

	err = multierror.Append(err, p.Core.Validate())

	err.ErrorFormat = func(errors []error) string {
		if len(errors) == 1 {
			return errors[0].Error()
		}

		numErrors := len(errors)
		var buf bytes.Buffer
		fmt.Fprintln(&buf)
		fmt.Fprintln(&buf)
		for i, e := range errors {
			fmt.Fprintf(&buf, "  * %s", e)
			if i != numErrors-1 {
				fmt.Fprintln(&buf)
			}
		}
		return buf.String()
	}

	return err.ErrorOrNil()
}

// Clean nils any empty component.
func (p *Profile) Clean() {
	if p.Core.isEmpty() {
		p.Core = nil
	}

	if p.VaultSecrets.isEmpty() {
		p.VaultSecrets = nil
	}
}

// Write writes the profile to disk.
func (p *Profile) Write() error {
	// Remove any empty components before writing
	p.Clean()

	path := fmt.Sprintf("%s/%s.hcl", p.dir, p.Name)
	f := hclwrite.NewEmptyFile()
	gohcl.EncodeIntoBody(p, f.Body())
	return os.WriteFile(path, f.Bytes(), 0o666)
}

// String returns an HCL formatted string representation of the profile.
func (p *Profile) String() string {
	f := hclwrite.NewEmptyFile()
	gohcl.EncodeIntoBody(p, f.Body())
	return strings.TrimSpace(string(f.Bytes()))
}

// PropertyNames returns the name of the properties in a profile. If the
// property is in a struct, such as Core, the property name will be
// <struct_name>/<property_name>, such as "core/no_color".
func PropertyNames() map[string]struct{} {
	keys := make(map[string]struct{})
	var p Profile
	doWalkStructElements("", reflect.TypeOf(p), keys)
	return keys
}

func doWalkStructElements(path string, t reflect.Type, keys map[string]struct{}) {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Get the tag
		name := field.Tag.Get("hcl")
		if name == "" {
			continue
		}

		name = strings.Split(name, ",")[0]
		if path != "" {
			name = fmt.Sprintf("%s/%s", path, name)
		}

		v := field.Type
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}

		if v.Kind() == reflect.Struct {
			doWalkStructElements(name, v, keys)
		} else {
			keys[name] = struct{}{}
		}

	}
}

// SetOrgID sets the OrganizationID
func (p *Profile) SetOrgID(id string) *Profile {
	if p == nil {
		return nil
	}

	p.OrganizationID = id
	return p
}

// SetProjectID sets the ProjectID
func (p *Profile) SetProjectID(id string) *Profile {
	if p == nil {
		return nil
	}

	p.ProjectID = id
	return p
}

func (p *Profile) GetOrgResourcePart() resourcename.Part {
	return resourcename.OrganizationPart(p.OrganizationID)
}

func (p *Profile) GetProjectResourcePart() resourcename.Part {
	return resourcename.ProjectPart(p.ProjectID)
}

// Core stores configuration settings that impact the CLIs behavior.
type Core struct {
	// NoColor disables color output
	NoColor *bool `hcl:"no_color,optional" json:",omitempty"`

	// Quiet disables prompting for user input and minimizes output.
	Quiet *bool `hcl:"quiet,optional" json:",omitempty"`

	// OutputFormat dictates the default output format if unspecified.
	OutputFormat *string `hcl:"output_format,optional" json:",omitempty"`

	// Verbosity is the default verbosity to log at
	Verbosity *string `hcl:"verbosity,optional" json:",omitempty"`
}

func (c *Core) Predict(args complete.Args) []string {
	properties := map[string][]string{
		"core/no_color":      {"true", "false"},
		"core/quiet":         {"true", "false"},
		"core/output_format": {"pretty", "table", "json"},
		"core/verbosity":     {"trace", "debug", "info", "warn", "error"},
	}

	// If the property has been specified, return possible values.
	if len(args.All) >= 1 {
		prediction, ok := properties[args.All[0]]
		if ok {
			return prediction
		}
	}

	// Predicting the property
	if len(args.All) == 1 {
		keys := maps.Keys(properties)
		slices.Sort(keys)
		return keys
	}

	return nil
}

func (c *Core) Validate() error {
	if c == nil {
		return nil
	}

	var err *multierror.Error
	allowedFormats := []string{"pretty", "json", "table"}
	if f := c.GetOutputFormat(); f != "" && !slices.Contains(allowedFormats, f) {
		err = multierror.Append(err, fmt.Errorf("invalid output_format %q. Must be one of: %q", f, allowedFormats))
	}

	allowedVerbosities := []string{"trace", "debug", "info", "warn", "error"}
	if f := c.GetVerbosity(); f != "" && !slices.Contains(allowedVerbosities, f) {
		err = multierror.Append(err, fmt.Errorf("invalid verbosity %q. Must be one of: %q", f, allowedVerbosities))
	}

	return err.ErrorOrNil()
}

func (c *Core) isEmpty() bool {
	if c == nil {
		return true
	}

	if c.NoColor != nil || c.OutputFormat != nil || c.Verbosity != nil ||
		c.Quiet != nil {
		return false
	}

	return true
}

// GetOutputFormat returns the set output format or an empty string if it has
// not been configured.
func (c *Core) GetOutputFormat() string {
	if c == nil {
		return ""
	}

	if c.OutputFormat == nil {
		return ""
	}

	return *c.OutputFormat
}

// GetVerbosity returns the set verbosity or an empty string if it has not been
// configured.
func (c *Core) GetVerbosity() string {
	if c == nil {
		return ""
	}

	if c.Verbosity == nil {
		return ""
	}

	return *c.Verbosity
}

// IsQueit returns whether the quiet property has been configured to be quiet.
func (c *Core) IsQuiet() bool {
	if c == nil {
		return false
	}

	if c.Quiet == nil {
		return false
	}

	return *c.Quiet
}
