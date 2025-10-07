// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: MPL-2.0

package secrets

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-11-28/client/secret_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-11-28/models"
	mock_secret_service "github.com/hashicorp/hcp/internal/pkg/api/mocks/github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-11-28/client/secret_service"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewCmdRead(t *testing.T) {
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
		Expect  *ReadOpts
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
			Expect: &ReadOpts{
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

			var gotOpts *ReadOpts
			readCmd := NewCmdRead(ctx, func(o *ReadOpts) error {
				gotOpts = o
				gotOpts.AppName = c.Profile(t).VaultSecrets.AppName
				gotOpts.SecretName = testSecretName
				return nil
			})
			readCmd.SetIO(io)

			code := readCmd.Run(c.Args)
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

func TestReadRun(t *testing.T) {
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
			ErrMsg:     "[GET] /secrets/2023-06-13/organizations/{location.organization_id}/projects/{location.project_id}/apps/{app_name}/secrets/{secret_name}][404] GetAppSecret default  &{Code:5 Details:[] Message:secret not found}",
			MockCalled: true,
		},
		{
			Name:       "Success: Read secret's metadata",
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
			opts := &ReadOpts{
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
					vs.EXPECT().GetAppSecret(mock.Anything, mock.Anything).Return(nil, errors.New(c.ErrMsg)).Once()
				} else {
					vs.EXPECT().GetAppSecret(&secret_service.GetAppSecretParams{
						OrganizationID: testProfile(t).OrganizationID,
						ProjectID:      testProfile(t).ProjectID,
						AppName:        testProfile(t).VaultSecrets.AppName,
						SecretName:     opts.SecretName,
						Context:        opts.Ctx,
					}, mock.Anything).Return(&secret_service.GetAppSecretOK{
						Payload: &models.Secrets20231128GetAppSecretResponse{
							Secret: &models.Secrets20231128Secret{
								Name:          opts.SecretName,
								LatestVersion: 2,
								CreatedAt:     strfmt.DateTime(time.Now()),
							},
						},
					}, nil).Once()
				}
			}

			// Run the command
			err := readRun(opts)
			if c.ErrMsg != "" {
				r.Contains(err.Error(), c.ErrMsg)
				return
			}

			r.NoError(err)
			r.Contains(io.Output.String(), "Secret Name:")
		})
	}
}
