// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: MPL-2.0

package profiles

import (
	"strings"
	"testing"

	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
	"github.com/stretchr/testify/require"
)

func TestList(t *testing.T) {
	t.Parallel()
	r := require.New(t)

	io := iostreams.Test()
	l := profile.TestLoader(t)
	output := format.New(io)

	opts := &ListOpts{
		IO:       io,
		Output:   output,
		Profiles: l,
	}

	// Create a few profiles
	p1, err := l.NewProfile("alpha")
	r.NoError(err)
	p1.OrganizationID = "alpha-org-id"
	p1.ProjectID = "alpha-project-id"
	r.NoError(p1.Write())

	p2, err := l.NewProfile("beta")
	r.NoError(err)
	p2.OrganizationID = "beta-org-id"
	p2.ProjectID = "beta-project-id"
	r.NoError(p2.Write())

	p3, err := l.NewProfile("zed")
	r.NoError(err)
	p3.OrganizationID = "zed-org-id"
	p3.ProjectID = "zed-project-id"
	r.NoError(p3.Write())

	// Set beta as active
	active, err := l.GetActiveProfile()
	r.NoError(err)
	active.Name = "beta"
	r.NoError(active.Write())

	// Call list
	r.NoError(listRun(opts))

	// Check we got the output we expected
	expected := [][]string{
		{"Name", "Active", "Organization ID", "Project ID"},
		{p1.Name, "false", p1.OrganizationID, p1.ProjectID},
		{p2.Name, "true", p2.OrganizationID, p2.ProjectID},
		{p3.Name, "false", p3.OrganizationID, p3.ProjectID},
	}

	lines := strings.Split(io.Output.String(), "\n")
	r.Len(lines, 5)
	r.Empty(lines[4])
	for i, expectedFields := range expected {
		for _, field := range expectedFields {
			r.Contains(lines[i], field)
		}
	}

}
