// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: MPL-2.0

package profile

import (
	"strings"
	"testing"

	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
	"github.com/stretchr/testify/require"
)

func TestGet(t *testing.T) {
	t.Parallel()
	r := require.New(t)

	io := iostreams.Test()
	p := profile.TestProfile(t)
	p.OrganizationID = "123"
	p.ProjectID = "456"
	p.Core = &profile.Core{
		NoColor:      new(bool),
		OutputFormat: new(string),
	}
	*p.Core.NoColor = true
	*p.Core.OutputFormat = "json"

	expect := map[string]string{
		"organization_id":    "123",
		"project_id":         "456",
		"core/no_color":      "true",
		"core/output_format": "json",
	}

	for k, v := range expect {
		opts := &GetOpts{
			IO:       io,
			Profile:  p,
			Property: k,
		}
		r.NoError(getRun(opts))
		r.Equal(strings.TrimSpace(io.Output.String()), v)
		io.Output.Reset()
	}

	// Get an unset property
	opts := &GetOpts{
		IO:       io,
		Profile:  p,
		Property: "core/verbosity",
	}
	r.ErrorContains(getRun(opts), "property \"core/verbosity\" is not set")
}
