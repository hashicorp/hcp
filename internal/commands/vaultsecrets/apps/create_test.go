// Copyright IBM Corp. 2024, 2025
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

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-11-28/client/secret_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-11-28/models"
	mock_secret_service "github.com/hashicorp/hcp/internal/pkg/api/mocks/github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-11-28/client/secret_service"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
)

func TestNewCmdCreate(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Name    string
		Args    []string
		Profile func(t *testing.T) *profile.Profile
		Error   string
		Expect  *CreateOpts
	}{
		{
			Name: "Good",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123").SetProjectID("abc")
			},
			Args: []string{"company-card"},
			Expect: &CreateOpts{
				AppName: "company-card",
			},
		},
		{
			Name: "Good with description",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123").SetProjectID("abc")
			},
			Args: []string{"company-card", "--description", "Stores corporate card info."},
			Expect: &CreateOpts{
				AppName:     "company-card",
				Description: "Stores corporate card info.",
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
			Args:  []string{"company-card", "additional-arg"},
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

			var createOpts *CreateOpts
			createCmd := NewCmdCreate(ctx, func(o *CreateOpts) error {
				createOpts = o
				return nil
			})
			createCmd.SetIO(io)

			code := createCmd.Run(c.Args)
			if c.Error != "" {
				r.NotZero(code)
				r.Contains(io.Error.String(), c.Error)
				return
			}

			r.Zero(code, io.Error.String())
			r.NotNil(createOpts)
			r.Equal(c.Expect.AppName, createOpts.AppName)
			r.Equal(c.Expect.Description, createOpts.Description)
		})
	}
}

func TestCreateRun(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Name           string
		AppName        string
		AppDescription string
		Error          string
	}{
		{
			Name:  "Missing app name",
			Error: "failed to create application",
		},
		{
			Name:           "Good",
			AppName:        "company-card",
			AppDescription: "Stores corporate card info.",
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			r := require.New(t)

			io := iostreams.Test()
			vs := mock_secret_service.NewMockClientService(t)

			opts := &CreateOpts{
				Ctx:         context.Background(),
				Profile:     profile.TestProfile(t).SetOrgID("123").SetProjectID("abc"),
				IO:          io,
				Client:      vs,
				Output:      format.New(io),
				AppName:     c.AppName,
				Description: c.AppDescription,
			}

			if c.Error != "" {
				vs.EXPECT().CreateApp(mock.Anything, mock.Anything).Return(nil, errors.New("missing app name")).Once()
			} else {
				vs.EXPECT().CreateApp(&secret_service.CreateAppParams{
					Context:        opts.Ctx,
					OrganizationID: "123",
					ProjectID:      "abc",
					Body: &models.SecretServiceCreateAppBody{
						Name:        opts.AppName,
						Description: opts.Description,
					},
				}, nil).Return(&secret_service.CreateAppOK{
					Payload: &models.Secrets20231128CreateAppResponse{
						App: &models.Secrets20231128App{
							Name:        opts.AppName,
							Description: opts.Description,
						},
					},
				}, nil).Once()
			}

			// Run the command
			err := createRun(opts)
			if c.Error != "" {
				r.ErrorContains(err, c.Error)
				return
			}

			r.NoError(err)
			r.Contains(io.Error.String(), fmt.Sprintf("✓ Successfully created application with name %q\n", opts.AppName))
		})
	}
}
