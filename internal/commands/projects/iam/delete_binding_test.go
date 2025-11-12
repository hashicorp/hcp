// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-openapi/runtime/client"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/models"
	"github.com/hashicorp/hcp/internal/pkg/api/iampolicy"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewCmdDeleteBinding(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Name    string
		Args    []string
		Profile func(t *testing.T) *profile.Profile
		Error   string
		Expect  *AddBindingOpts
	}{
		{
			Name:    "No Org",
			Profile: profile.TestProfile,
			Args:    []string{"--member=123", "--role=admin"},
			Error:   "Organization ID and Project ID must be configured",
		},
		{
			Name: "No Project passed/profile",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123")
			},
			Args:  []string{"--member=123", "--role=admin"},
			Error: "Organization ID and Project ID must be configured",
		},
		{
			Name: "Too many args",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123").SetProjectID("456")
			},
			Args:  []string{"--member=123", "--role=admin", "foo", "bar"},
			Error: "no arguments allowed, but received 2",
		},
		{
			Name: "Good",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123").SetProjectID("456")
			},
			Args: []string{"--member=123", "--role=admin"},
			Expect: &AddBindingOpts{
				PrincipalID: "123",
				Role:        "admin",
			},
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			r := require.New(t)

			// Create a context.
			io := iostreams.Test()
			ctx := &cmd.Context{
				IO:          io,
				Profile:     c.Profile(t),
				Output:      format.New(io),
				HCP:         &client.Runtime{},
				ShutdownCtx: context.Background(),
			}

			var gotOpts *DeleteBindingOpts
			deleteCmd := NewCmdDeleteBinding(ctx, func(o *DeleteBindingOpts) error {
				gotOpts = o
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
			r.NotNil(gotOpts)
			r.Equal(c.Expect.PrincipalID, gotOpts.PrincipalID)
			r.Equal(c.Expect.Role, gotOpts.Role)
			r.NotNil(gotOpts.Setter)
		})
	}
}

func TestDeleteBindingRun(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Name    string
		RespErr error
		Error   string
	}{
		{
			Name:    "Server error",
			RespErr: fmt.Errorf("failed to add policy"),
			Error:   "failed to add policy",
		},
		{
			Name: "Good",
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			r := require.New(t)

			io := iostreams.Test()
			setter := iampolicy.NewMockSetter(t)
			opts := &DeleteBindingOpts{
				Ctx:         context.Background(),
				IO:          io,
				Setter:      setter,
				PrincipalID: "principal-123",
				Role:        "roles/test",
			}

			// Expect a request to add a binding.
			call := setter.EXPECT().DeleteBinding(mock.Anything, opts.PrincipalID, opts.Role).Once()

			if c.RespErr != nil {
				call.Return(nil, c.RespErr)
			} else {
				// TODO
				call.Return(&models.HashicorpCloudResourcemanagerPolicy{}, nil)
			}

			// Run the command
			err := deleteBindingRun(opts)
			if c.Error != "" {
				r.ErrorContains(err, c.Error)
				return
			}

			// Check we outputted the project
			r.NoError(err)
			r.Contains(io.Error.String(), `Principal "principal-123" binding to role "roles/test" deleted.`)
		})
	}
}
