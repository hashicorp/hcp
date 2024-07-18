// Copyright (c) HashiCorp, Inc.
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

func TestNewCmdDelete(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Name    string
		Args    []string
		Profile func(t *testing.T) *profile.Profile
		Error   string
		Expect  *DeleteOpts
	}{
		{
			Name: "Good",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123").SetProjectID("abc")
			},
			Args: []string{"gp1"},
			Expect: &DeleteOpts{
				GatewayPoolName: "gp1",
			},
		},
		{
			Name: "No args",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123").SetProjectID("abc")
			},
			Args:  []string{},
			Error: "accepts 1 arg(s), received 0",
		},
		{
			Name: "Too many args",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123").SetProjectID("abc")
			},
			Args:  []string{"gp1", "additional-arg"},
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

			var deleteOpts *DeleteOpts
			deleteCmd := NewCmdDelete(ctx, func(o *DeleteOpts) error {
				deleteOpts = o
				return nil
			})
			deleteCmd.SetIO(io)

			code := deleteCmd.Run(c.Args)
			if c.Error != "" {
				r.NotZero(code)
				r.Contains(io.Error.String(), c.Error)
				return
			}

			r.Zero(code, io.Error.String())
			r.NotNil(deleteOpts)
			r.Equal(c.Expect.GatewayPoolName, deleteOpts.GatewayPoolName)
		})
	}
}
