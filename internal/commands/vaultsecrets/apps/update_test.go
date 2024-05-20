// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apps

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/go-openapi/runtime/client"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-06-13/client/secret_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-06-13/models"
	mock_secret_service "github.com/hashicorp/hcp/internal/pkg/api/mocks/github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-06-13/client/secret_service"
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
			Args: []string{"company-card", "--description", "Stores corporate card info."},
			Expect: &UpdateOpts{
				AppName:     "company-card",
				Description: "Stores corporate card info.",
			},
		},
		{
			Name: "No description",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123").SetProjectID("abc")
			},
			Args:  []string{"company-card"},
			Error: "ERROR: missing required flag: --description=DESCRIPTION",
		},
		{
			Name: "No app name",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123").SetProjectID("abc")
			},
			Args:  []string{"--description", "Stores corporate card info."},
			Error: "accepts 1 arg(s), received 0",
		},
		{
			Name: "Too many args",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123").SetProjectID("abc")
			},
			Args:  []string{"company-card", "additional-arg", "--description", "Stores corporate card info."},
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
			r.Equal(c.Expect.AppName, updateOpts.AppName)
			r.Equal(c.Expect.Description, updateOpts.Description)
		})
	}
}

func TestUpdateRun(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Name           string
		AppName        string
		AppDescription string
		Error          string
	}{
		{
			Name:           "Good",
			AppName:        "company-card",
			AppDescription: "Stores corporate card info.",
		},
		{
			Name:           "Missing app name",
			AppDescription: "Stores corporate card info.",
			Error:          "missing required app name",
		},
		{
			Name:    "Missing description",
			AppName: "company-card",
			Error:   "missing required description",
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			r := require.New(t)

			io := iostreams.Test()
			vs := mock_secret_service.NewMockClientService(t)

			opts := &UpdateOpts{
				Ctx:         context.Background(),
				Profile:     profile.TestProfile(t).SetOrgID("123").SetProjectID("abc"),
				IO:          io,
				Client:      vs,
				Output:      format.New(io),
				AppName:     c.AppName,
				Description: c.AppDescription,
			}

			if c.Error != "" {
				vs.EXPECT().UpdateApp(mock.Anything, mock.Anything).Return(nil, errors.New(c.Error)).Once()
			} else {
				vs.EXPECT().UpdateApp(&secret_service.UpdateAppParams{
					Context:                opts.Ctx,
					LocationOrganizationID: "123",
					LocationProjectID:      "abc",
					Name:                   opts.AppName,
					Body: secret_service.UpdateAppBody{
						Description: opts.Description,
					},
				}, nil).Return(&secret_service.UpdateAppOK{
					Payload: &models.Secrets20230613UpdateAppResponse{
						App: &models.Secrets20230613App{
							Name:        opts.AppName,
							Description: opts.Description,
						},
					},
				}, nil).Once()
			}

			// Run the command
			err := updateRun(opts)
			if c.Error != "" {
				r.ErrorContains(err, c.Error)
				return
			}

			r.NoError(err)
			r.Contains(io.Error.String(), fmt.Sprintf("✓ Successfully updated application with name %q\n", opts.AppName))
		})
	}
}
