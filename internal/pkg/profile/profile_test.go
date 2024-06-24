// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package profile

import (
	"strings"
	"testing"

	"github.com/posener/complete"
	"github.com/stretchr/testify/require"
)

func TestPropertyNames(t *testing.T) {
	t.Parallel()
	r := require.New(t)
	properties := PropertyNames()
	r.NotEmpty(properties)
	r.Contains(properties, "name")
	r.Contains(properties, "organization_id")
	r.Contains(properties, "project_id")
	r.Contains(properties, "core/no_color")
	r.Contains(properties, "core/quiet")
	r.Contains(properties, "core/verbosity")
	r.Contains(properties, "vault-secrets/app")
}

func TestProfile_Validate(t *testing.T) {
	t.Parallel()

	badOutputFormat := "random-format"

	cases := []struct {
		Name    string
		Profile *Profile
		Error   string
	}{
		{
			Name:    "empty",
			Profile: &Profile{},
			Error:   "profile name may only include",
		},
		{
			Name: "name too long",
			Profile: &Profile{
				Name: strings.Repeat("test", 100),
			},
			Error: "profile name may only include",
		},
		{
			Name: "invalid core",
			Profile: &Profile{
				Name:           "test",
				OrganizationID: "123",
				Core: &Core{
					OutputFormat: &badOutputFormat,
				},
			},
			Error: "invalid output_format",
		},
		{
			Name: "valid",
			Profile: &Profile{
				Name:           "test",
				OrganizationID: "123",
			},
			Error: "",
		},
	}

	for _, c := range cases {
		// Capture the test case
		c := c
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			r := require.New(t)

			err := c.Profile.Validate()
			if c.Error == "" {
				r.NoError(err)
			} else {
				r.ErrorContains(err, c.Error)
			}
		})
	}
}

func TestProfile_Predict(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Name     string
		Args     complete.Args
		Expected []string
	}{
		{
			Name: "empty",
			Args: complete.Args{
				All: []string{""},
			},
			Expected: []string{"organization_id", "project_id", "core/", "vault-secrets/"},
		},
		{
			Name: "specific field",
			Args: complete.Args{
				All: []string{"org"},
			},
			Expected: []string{"organization_id", "project_id", "core/", "vault-secrets/"},
		},
		{
			Name: "core",
			Args: complete.Args{
				All: []string{"core/"},
			},
			Expected: []string{"core/no_color", "core/output_format", "core/quiet", "core/verbosity"},
		},
		{
			Name: "vault-secrets",
			Args: complete.Args{
				All: []string{"vault-secrets/"},
			},
			Expected: []string{"vault-secrets/app"},
		},
	}

	for _, c := range cases {
		// Capture the test case
		c := c
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			r := require.New(t)

			// Create a profile
			p := &Profile{}

			// Predict
			out := p.Predict(c.Args)
			r.Equal(c.Expected, out)
		})
	}
}

func TestCore_Validate(t *testing.T) {
	t.Parallel()

	badValue := "random-test-value"

	cases := []struct {
		Name    string
		Profile *Core
		Error   string
	}{
		{
			Name:    "empty",
			Profile: &Core{},
			Error:   "",
		},
		{
			Name: "bad output_format",
			Profile: &Core{
				OutputFormat: &badValue,
			},
			Error: "invalid output_format",
		},
		{
			Name: "bad verbosity",
			Profile: &Core{
				Verbosity: &badValue,
			},
			Error: "invalid verbosity",
		},
	}

	for _, c := range cases {
		// Capture the test case
		c := c
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			require := require.New(t)

			err := c.Profile.Validate()
			if c.Error == "" {
				require.NoError(err)
			} else {
				require.ErrorContains(err, c.Error)
			}
		})
	}
}

func TestCore_Predict(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Name     string
		Args     complete.Args
		Expected []string
	}{
		{
			Name: "just core",
			Args: complete.Args{
				All: []string{"core/"},
			},
			Expected: []string{"core/no_color", "core/output_format", "core/quiet", "core/verbosity"},
		},
		{
			Name: "no_color",
			Args: complete.Args{
				All: []string{"core/no_color", ""},
			},
			Expected: []string{"true", "false"},
		},
		{
			Name: "output_format",
			Args: complete.Args{
				All: []string{"core/output_format", ""},
			},
			Expected: []string{"pretty", "table", "json"},
		},
		{
			Name: "verbosity",
			Args: complete.Args{
				All: []string{"core/verbosity", ""},
			},
			Expected: []string{"trace", "debug", "info", "warn", "error"},
		},
	}

	for _, c := range cases {
		// Capture the test case
		c := c
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			r := require.New(t)

			// Create a core
			p := &Core{}

			// Predict
			out := p.Predict(c.Args)
			r.Equal(c.Expected, out)
		})
	}
}

func TestCore_Getters(t *testing.T) {
	t.Parallel()
	r := require.New(t)

	// Start with a nil core struct
	var c *Core
	r.Empty(c.GetOutputFormat())
	r.Empty(c.GetVerbosity())

	// Instantiate an empty core
	c = &Core{}
	r.Empty(c.GetOutputFormat())
	r.Empty(c.GetVerbosity())

	// Instantiate a non-empty core
	v := "test"
	c = &Core{
		OutputFormat: &v,
		Verbosity:    &v,
	}
	r.Equal(v, c.GetOutputFormat())
	r.Equal(v, c.GetVerbosity())
}
