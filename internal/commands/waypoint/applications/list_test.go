// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: MPL-2.0

package applications

import (
	"testing"

	"github.com/hashicorp/hcp/internal/pkg/profile"
)

func TestNewCmdListApplications(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Name    string
		Args    []string
		Profile func(t *testing.T) *profile.Profile
		Error   string
		Expect  *ApplicationOpts
	}{
		{
			Name:    "No Org",
			Profile: profile.TestProfile,
			Args:    []string{},
			Error:   "Organization ID must be configured",
		},
		// No flags are needed for the list command at this time, but if that changes,
		// we should add a test case here to test those flags
	}

	for _, c := range cases {
		c := c

		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
		})
	}
}
