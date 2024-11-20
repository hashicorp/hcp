// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package run

import (
	"context"
	"errors"
	"os/exec"
	"testing"
	"time"

	"github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	preview_secret_service "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-11-28/client/secret_service"
	preview_models "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-11-28/models"

	mock_preview_secret_service "github.com/hashicorp/hcp/internal/pkg/api/mocks/github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-11-28/client/secret_service"
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
			Name:    "Good",
			Profile: testProfile,
			Args:    []string{"--app=test-app", "--", "env", "-i"},
			Expect: &RunOpts{
				AppName: testProfile(t).VaultSecrets.AppName,
				Command: []string{"env", "-i"},
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
				gotOpts.AppName = c.Profile(t).VaultSecrets.AppName
				return nil
			})
			runCmd.SetIO(io)

			_ = runCmd.Run(c.Args)

			r.NotNil(gotOpts)
			r.Equal(c.Expect.AppName, gotOpts.AppName)
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
		Name            string
		Secrets         []*preview_models.Secrets20231128OpenSecret
		RespErr         bool
		ErrMsg          string
		IOErrorContains string
		PaginatedResp   bool
	}{
		{
			Name:    "Failed: Secret not found",
			RespErr: true,
			Secrets: nil,
			ErrMsg:  "[GET /secrets/2023-11-28/organizations/{organization_id}/projects/{project_id}/apps/{app_name}/secrets:open][403]",
		},
		{
			Name:    "Success",
			RespErr: false,
			Secrets: []*preview_models.Secrets20231128OpenSecret{
				{
					Name:          "static",
					StaticVersion: &preview_models.Secrets20231128OpenSecretStaticVersion{},
				},
				{
					Name: "rotating",
					RotatingVersion: &preview_models.Secrets20231128OpenSecretRotatingVersion{
						Values: map[string]string{"sub_key": "value"},
					},
				},
				{
					Name: "dynamic",
					DynamicInstance: &preview_models.Secrets20231128OpenSecretDynamicInstance{
						Values: map[string]string{"sub_key": "value"},
					},
				},
			},
		},
		{
			Name:            "Collide",
			RespErr:         false,
			ErrMsg:          "multiple secrets map to the same environment variable",
			IOErrorContains: "ERROR: \"static_collision\" [static], \"static\" [rotating] map to the same environment variable \"STATIC_COLLISION\"",
			Secrets: []*preview_models.Secrets20231128OpenSecret{
				{
					Name:          "static_collision",
					Type:          "static",
					LatestVersion: 1,
					CreatedAt:     strfmt.DateTime(time.Now()),
					StaticVersion: &preview_models.Secrets20231128OpenSecretStaticVersion{},
				},
				{
					Name:          "static",
					Type:          "rotating",
					LatestVersion: 1,
					CreatedAt:     strfmt.DateTime(time.Now()),
					RotatingVersion: &preview_models.Secrets20231128OpenSecretRotatingVersion{
						Values: map[string]string{"collision": ""},
					},
				},
			},
		},
		{
			Name:          "Paginated",
			PaginatedResp: true,
			RespErr:       false,
			Secrets: []*preview_models.Secrets20231128OpenSecret{
				{
					Name:          "static_1",
					StaticVersion: &preview_models.Secrets20231128OpenSecretStaticVersion{},
				},
				{
					Name:          "static_2",
					StaticVersion: &preview_models.Secrets20231128OpenSecretStaticVersion{},
				},
				{
					Name:          "static_3",
					StaticVersion: &preview_models.Secrets20231128OpenSecretStaticVersion{},
				},
				{
					Name:          "static_4",
					StaticVersion: &preview_models.Secrets20231128OpenSecretStaticVersion{},
				},
			},
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
				AppName:       testProfile(t).VaultSecrets.AppName,
				Command:       []string{"echo \"Testing\""},
			}

			if c.RespErr {
				vs.EXPECT().OpenAppSecrets(mock.Anything, mock.Anything).Return(nil, errors.New(c.ErrMsg)).Once()
			} else if c.PaginatedResp {
				paginationNextPageToken := "next_page_token"

				// expect first request to be missing the page token
				// provide half the secrets and a NextPageToken
				vs.EXPECT().OpenAppSecrets(&preview_secret_service.OpenAppSecretsParams{
					OrganizationID: testProfile(t).OrganizationID,
					ProjectID:      testProfile(t).ProjectID,
					AppName:        testProfile(t).VaultSecrets.AppName,
					Context:        opts.Ctx,
				}, mock.Anything).Return(&preview_secret_service.OpenAppSecretsOK{
					Payload: &preview_models.Secrets20231128OpenAppSecretsResponse{
						Secrets: c.Secrets[:len(c.Secrets)/2],
						Pagination: &preview_models.CommonPaginationResponse{
							NextPageToken: paginationNextPageToken,
						},
					},
				}, nil).Once()

				// expect second request to have a page token
				// provide later half of the secrets
				vs.EXPECT().OpenAppSecrets(&preview_secret_service.OpenAppSecretsParams{
					OrganizationID:          testProfile(t).OrganizationID,
					ProjectID:               testProfile(t).ProjectID,
					AppName:                 testProfile(t).VaultSecrets.AppName,
					Context:                 opts.Ctx,
					PaginationNextPageToken: &paginationNextPageToken,
				}, mock.Anything).Return(&preview_secret_service.OpenAppSecretsOK{
					Payload: &preview_models.Secrets20231128OpenAppSecretsResponse{
						Secrets: c.Secrets[len(c.Secrets)/2:],
					},
				}, nil).Once()
			} else {
				vs.EXPECT().OpenAppSecrets(&preview_secret_service.OpenAppSecretsParams{
					OrganizationID: testProfile(t).OrganizationID,
					ProjectID:      testProfile(t).ProjectID,
					AppName:        testProfile(t).VaultSecrets.AppName,
					Context:        opts.Ctx,
				}, nil).Return(&preview_secret_service.OpenAppSecretsOK{
					Payload: &preview_models.Secrets20231128OpenAppSecretsResponse{
						Secrets: c.Secrets,
					},
				}, nil).Once()
			}

			// Run the command
			err := runRun(opts)
			if c.ErrMsg != "" {
				// Check for additional error messages
				r.Contains(io.Error.String(), c.IOErrorContains)

				r.Contains(err.Error(), c.ErrMsg)
				return
			}

			r.NoError(err)
		})
	}
}

func TestSetupChildProcess(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Name        string
		EnvVars     []string
		Command     []string
		ExpectedCmd *exec.Cmd
	}{
		{
			Name:    "Single string command yields correct args",
			EnvVars: []string{"test=123"},
			Command: []string{"echo \"Testing\""},
			ExpectedCmd: &exec.Cmd{
				Args: []string{"echo", "\"Testing\""},
				Env:  []string{"test=123"},
			},
		},
		{
			Name:    "Arbitrary args and multiple env vars",
			EnvVars: []string{"test=123", "test2=abc"},
			Command: []string{"go", "run", "main.go", "--flag=value"},
			ExpectedCmd: &exec.Cmd{
				Args: []string{"go", "run", "main.go", "--flag=value"},
				Env:  []string{"test=123", "test2=abc"},
			},
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			r := require.New(t)

			cmd := setupChildProcess(context.Background(), c.Command, c.EnvVars)

			r.Equal(cmd.Args, c.ExpectedCmd.Args)
			r.Equal(cmd.Env, c.ExpectedCmd.Env)
			r.Contains(cmd.Path, c.ExpectedCmd.Args[0])
		})
	}
}
