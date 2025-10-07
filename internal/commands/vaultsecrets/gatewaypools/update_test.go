// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: MPL-2.0

package gatewaypools

import (
	"context"
	"testing"

	"github.com/go-openapi/runtime/client"
	"github.com/stretchr/testify/require"

	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
)

func TestNewCmdUpdate(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Name    string
		Args    []string
		Profile func(t *testing.T) *profile.Profile
		Error   string
		Expect  *UpdateOpts
	}{
		{
			Name: "Good",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123").SetProjectID("abc")
			},
			Args: []string{"gp1", "--description", "gateway tunnel."},
			Expect: &UpdateOpts{
				GatewayPoolName: "gp1",
				Description:     "gateway tunnel.",
			},
		},
		{
			Name: "No description",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123").SetProjectID("abc")
			},
			Args:  []string{"gp1"},
			Error: "no fields specified for update",
		},
		{
			Name: "No gwp name",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123").SetProjectID("abc")
			},
			Args:  []string{"--description", "gateway tunnel."},
			Error: "accepts 1 arg(s), received 0",
		},
		{
			Name: "Too many args",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123").SetProjectID("abc")
			},
			Args:  []string{"gp1", "additional-arg", "--description", "gateway tunnel."},
			Error: "ERROR: accepts 1 arg(s), received 2",
		},
	}
	for _, c := range cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()

			r := require.New(t)
			io := iostreams.Test()

			ctx := &cmd.Context{
				IO:          io,
				Profile:     c.Profile(t),
				ShutdownCtx: context.Background(),
				HCP:         &client.Runtime{},
				Output:      format.New(io),
			}

			var updateOpts *UpdateOpts
			updateCmd := NewCmdUpdate(ctx, func(o *UpdateOpts) error {
				updateOpts = o
				return nil
			})
			updateCmd.SetIO(io)

			code := updateCmd.Run(c.Args)
			if c.Error != "" {
				r.NotZero(code)
				r.Contains(io.Error.String(), c.Error)
				return
			}

			r.Zero(code, io.Error.String())
			r.NotNil(updateOpts)
			r.Equal(c.Expect.GatewayPoolName, updateOpts.GatewayPoolName)
			r.Equal(c.Expect.Description, updateOpts.Description)
		})
	}
}
