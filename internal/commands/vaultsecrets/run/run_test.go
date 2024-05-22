// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package run

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	preview_secret_service "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/preview/2023-11-28/client/secret_service"
	preview_models "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/preview/2023-11-28/models"
	mock_preview_secret_service "github.com/hashicorp/hcp/internal/pkg/api/mocks/github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/preview/2023-11-28/client/secret_service"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
)

func TestNewCmdRun(t *testing.T) {
	t.Parallel()

	testProfile := func(t *testing.T) *profile.Profile {
		tp := profile.TestProfile(t).SetOrgID("123").SetProjectID("abc")
		tp.VaultSecrets = &profile.VaultSecretsConf{
			AppName: "test-app",
		}
		return tp
	}

	cases := []struct {
		Name    string
		Args    []string
		Profile func(t *testing.T) *profile.Profile
		Error   string
		Expect  *RunOpts
	}{
		{
			Name:    "Missing command flag",
			Profile: testProfile,
			Args:    []string{},
			Error:   "missing required flag: --command=COMMAND",
		},
		{
			Name:    "Too many args",
			Profile: testProfile,
			Args:    []string{"foo", "--command", "env"},
			Error:   "no arguments allowed, but received 1",
		},
		{
			Name:    "Good",
			Profile: testProfile,
			Args:    []string{"--command", "env"},
			Expect: &RunOpts{
				App:     testProfile(t).VaultSecrets.AppName,
				Command: "env",
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

			var gotOpts *RunOpts
			runCmd := NewCmdRun(ctx, func(o *RunOpts) error {
				gotOpts = o
				gotOpts.App = c.Profile(t).VaultSecrets.AppName
				return nil
			})
			runCmd.SetIO(io)

			code := runCmd.Run(c.Args)
			if c.Error != "" {
				r.NotZero(code)
				r.Contains(io.Error.String(), c.Error)
				return
			}

			r.Zero(code, io.Error.String())
			r.NotNil(gotOpts)
			r.Equal(c.Expect.App, gotOpts.App)
			r.Equal(c.Expect.Command, gotOpts.Command)
		})
	}
}

func TestRunRun(t *testing.T) {
	t.Parallel()

	testProfile := func(t *testing.T) *profile.Profile {
		tp := profile.TestProfile(t).SetOrgID("123").SetProjectID("abc")
		tp.VaultSecrets = &profile.VaultSecretsConf{
			AppName: "test-app",
		}
		return tp
	}

	cases := []struct {
		Name       string
		RespErr    bool
		ErrMsg     string
		MockCalled bool
	}{
		{
			Name:       "Failed: Secret not found",
			RespErr:    true,
			ErrMsg:     "[GET /secrets/2023-11-28/organizations/{organization_id}/projects/{project_id}/apps/{app_name}/secrets:open][403]",
			MockCalled: true,
		},
		{
			Name:       "Success",
			RespErr:    false,
			MockCalled: true,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			r := require.New(t)

			io := iostreams.Test()
			io.ErrorTTY = true
			vs := mock_preview_secret_service.NewMockClientService(t)
			opts := &RunOpts{
				Ctx:           context.Background(),
				IO:            io,
				Profile:       testProfile(t),
				Output:        format.New(io),
				PreviewClient: vs,
				App:           testProfile(t).VaultSecrets.AppName,
				Command:       "echo \"Testing\"",
			}

			if c.MockCalled {
				if c.RespErr {
					vs.EXPECT().OpenAppSecrets(mock.Anything, mock.Anything).Return(nil, errors.New(c.ErrMsg)).Once()
				} else {
					vs.EXPECT().OpenAppSecrets(&preview_secret_service.OpenAppSecretsParams{
						OrganizationID: testProfile(t).OrganizationID,
						ProjectID:      testProfile(t).ProjectID,
						AppName:        testProfile(t).VaultSecrets.AppName,
						Context:        opts.Ctx,
					}, nil).Return(&preview_secret_service.OpenAppSecretsOK{
						Payload: &preview_models.Secrets20231128OpenAppSecretsResponse{
							Secrets: []*preview_models.Secrets20231128OpenSecret{
								{
									Name:          "secret_1",
									LatestVersion: 2,
									CreatedAt:     strfmt.DateTime(time.Now()),
								},
								{
									Name:          "secret_2",
									LatestVersion: 2,
									CreatedAt:     strfmt.DateTime(time.Now()),
								},
							},
						},
					}, nil).Once()
				}
			}

			// Run the command
			err := runRun(opts)
			if c.ErrMsg != "" {
				r.Contains(err.Error(), c.ErrMsg)
				return
			}

			r.NoError(err)
		})
	}
}