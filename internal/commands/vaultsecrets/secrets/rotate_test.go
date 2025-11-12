// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: MPL-2.0

package secrets

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/go-openapi/runtime/client"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-11-28/client/secret_service"
	mock_secret_service "github.com/hashicorp/hcp/internal/pkg/api/mocks/github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-11-28/client/secret_service"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
)

func TestNewCmdRotate(t *testing.T) {
	t.Parallel()

	testSecretName := "test_secret"
	testProfile := func(t *testing.T) *profile.Profile {
		tp := profile.TestProfile(t).SetOrgID("123").SetProjectID("456")
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
		Expect  *RotateOpts
	}{
		{
			Name:    "No args",
			Profile: testProfile,
			Args:    []string{},
			Error:   "accepts 1 arg(s), received 0",
		},
		{
			Name:    "Too many args",
			Profile: testProfile,
			Args:    []string{"foo", "bar"},
			Error:   "accepts 1 arg(s), received 2",
		},
		{
			Name:    "Good",
			Profile: testProfile,
			Args:    []string{"foo"},
			Expect: &RotateOpts{
				AppName:    testProfile(t).VaultSecrets.AppName,
				SecretName: testSecretName,
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

			var gotOpts *RotateOpts
			rotateCmd := NewCmdRotate(ctx, func(o *RotateOpts) error {
				gotOpts = o
				gotOpts.AppName = c.Profile(t).VaultSecrets.AppName
				gotOpts.SecretName = testSecretName
				return nil
			})
			rotateCmd.SetIO(io)

			code := rotateCmd.Run(c.Args)
			if c.Error != "" {
				r.NotZero(code)
				r.Contains(io.Error.String(), c.Error)
				return
			}

			r.Zero(code, io.Error.String())
			r.NotNil(gotOpts)
			r.Equal(c.Expect.AppName, gotOpts.AppName)
			r.Equal(c.Expect.SecretName, gotOpts.SecretName)
		})
	}
}

func TestRotateRun(t *testing.T) {
	t.Parallel()

	testProfile := func(t *testing.T) *profile.Profile {
		tp := profile.TestProfile(t).SetOrgID("123").SetProjectID("456")
		tp.VaultSecrets = &profile.VaultSecretsConf{
			AppName: "test-app",
		}
		return tp
	}
	testSecretName := "test_secret"

	cases := []struct {
		Name       string
		RespErr    bool
		ErrMsg     string
		MockCalled bool
	}{
		{
			Name:       "Failed: Secret not found",
			RespErr:    true,
			ErrMsg:     "[POST] /secrets/2023-11-28/organizations/{organization_id}/projects/{project_id}/apps/{app_name}/secrets/{secret_name}:rotate][404] RotateSecret}",
			MockCalled: true,
		},
		{
			Name:       "Success: Rotate secret",
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
			vs := mock_secret_service.NewMockClientService(t)
			opts := &RotateOpts{
				Ctx:        context.Background(),
				IO:         io,
				Profile:    testProfile(t),
				Output:     format.New(io),
				Client:     vs,
				AppName:    testProfile(t).VaultSecrets.AppName,
				SecretName: testSecretName,
			}

			if c.MockCalled {
				if c.RespErr {
					vs.EXPECT().RotateSecret(mock.Anything, mock.Anything).Return(nil, errors.New(c.ErrMsg)).Once()
				} else {
					vs.EXPECT().RotateSecret(&secret_service.RotateSecretParams{
						OrganizationID: testProfile(t).OrganizationID,
						ProjectID:      testProfile(t).ProjectID,
						AppName:        testProfile(t).VaultSecrets.AppName,
						Name:           opts.SecretName,
						Context:        opts.Ctx,
					}, mock.Anything).Return(&secret_service.RotateSecretOK{}, nil).Once()
				}
			}

			// Run the command
			err := rotateRun(opts)
			if c.ErrMsg != "" {
				r.Contains(err.Error(), c.ErrMsg)
				return
			}

			r.NoError(err)
			r.Equal(io.Error.String(), fmt.Sprintf("✓ Successfully scheduled rotation of secret with name %q\n", opts.SecretName))
		})
	}
}
