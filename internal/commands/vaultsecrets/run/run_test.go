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
		MockCalled      bool
	}{
		{
			Name:       "Failed: Secret not found",
			RespErr:    true,
			Secrets:    nil,
			ErrMsg:     "[GET /secrets/2023-11-28/organizations/{organization_id}/projects/{project_id}/apps/{app_name}/secrets:open][403]",
			MockCalled: true,
		},
		{
			Name:       "Success",
			RespErr:    false,
			MockCalled: true,
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
			MockCalled:      true,
			IOErrorContains: "\"STATIC_COLLISION\" was assigned more than once",
			Secrets: []*preview_models.Secrets20231128OpenSecret{
				{
					Name:          "static_collision",
					LatestVersion: 1,
					CreatedAt:     strfmt.DateTime(time.Now()),
					StaticVersion: &preview_models.Secrets20231128OpenSecretStaticVersion{},
				},
				{
					Name:          "static",
					LatestVersion: 1,
					CreatedAt:     strfmt.DateTime(time.Now()),
					RotatingVersion: &preview_models.Secrets20231128OpenSecretRotatingVersion{
						Values: map[string]string{"collision": ""},
					},
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
							Secrets: c.Secrets,
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

			// Check for log messages
			r.Contains(io.Error.String(), c.IOErrorContains)
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

func Test_processCollisions(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name              string
		expectedCollision bool
		fmtNames          []string
	}{
		{
			name:              "success",
			expectedCollision: false,
			fmtNames:          []string{"secret_1", "secret_2", "secret_3"},
		},
		{
			name:              "one collision",
			expectedCollision: true,
			fmtNames:          []string{"secret_1", "secret_1", "secret_2"},
		},
		{
			name:              "multiple single key collision",
			expectedCollision: true,
			fmtNames:          []string{"secret_1", "secret_1", "secret_1", "secret_2"},
		},
		{
			name:              "two key collision",
			expectedCollision: true,
			fmtNames:          []string{"secret_1", "secret_1", "secret_2", "secret_2"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			collisions := make(map[string]bool, 0)
			for _, fmtName := range tt.fmtNames {
				processCollisions(collisions, fmtName)
			}
			// drop false records
			for fmtName, collided := range collisions {
				if !collided {
					delete(collisions, fmtName)
				}
			}
			if len(collisions) > 0 != tt.expectedCollision {
				t.Fail()
			}
		})
	}
}
