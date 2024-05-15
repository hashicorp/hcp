// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package secrets

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-06-13/client/secret_service"
	models "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-06-13/models"
	mock_secret_service "github.com/hashicorp/hcp/internal/pkg/api/mocks/github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-06-13/client/secret_service"

	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewCmdCreate(t *testing.T) {
	t.Parallel()

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
		Expect  *CreateOpts
	}{
		{
			Name:    "Failed: No secret name arg specified",
			Profile: testProfile,
			Args:    []string{},
			Error:   "ERROR: accepts 1 arg(s), received 0",
		},
		{
			Name:    "Good: Secret name arg specified",
			Profile: testProfile,
			Args:    []string{"test"},
			Expect: &CreateOpts{
				AppName:    testProfile(t).VaultSecrets.AppName,
				SecretName: "test",
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

			var gotOpts *CreateOpts
			createCmd := NewCmdCreate(ctx, func(o *CreateOpts) error {
				gotOpts = o
				gotOpts.AppName = "test-app"
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
			r.NotNil(gotOpts)
			r.Equal(c.Expect.AppName, gotOpts.AppName)
			r.Equal(c.Expect.SecretName, gotOpts.SecretName)
		})
	}
}

func TestCreateRun(t *testing.T) {
	t.Parallel()

	testProfile := func(t *testing.T) *profile.Profile {
		tp := profile.TestProfile(t).SetOrgID("123").SetProjectID("456")
		tp.VaultSecrets = &profile.VaultSecretsConf{
			AppName: "test-app",
		}
		return tp
	}
	testSecretValue := "my super secret value"

	cases := []struct {
		Name             string
		RespErr          bool
		ReadViaStdin     bool
		EmptySecretValue bool
		ErrMsg           string
		MockCalled       bool
		AugmentOpts      func(*CreateOpts)
	}{
		{
			Name:   "Failed: Read via stdin as hypen not supplied for --data-file flag",
			ErrMsg: "data file path is required",
		},
		{
			Name:             "Failed: Read empty secret value via stdin",
			EmptySecretValue: true,
			ReadViaStdin:     true,
			RespErr:          true,
			AugmentOpts:      func(o *CreateOpts) { o.SecretFilePath = "-" },
			ErrMsg:           "secret value cannot be empty",
		},
		{
			Name:         "Success: Create secret via stdin",
			ReadViaStdin: true,
			AugmentOpts:  func(o *CreateOpts) { o.SecretFilePath = "-" },
			MockCalled:   true,
		},
		{
			Name:        "Failed: Max secret versions reached",
			RespErr:     true,
			ErrMsg:      "[POST /secrets/2023-11-28/organizations/{organization_id}/projects/{project_id}/apps/{app_name}/secret/kv][429] CreateAppKVSecret default  &{Code:8 Details:[] Message:maximum number of secret versions reached}",
			AugmentOpts: func(o *CreateOpts) { o.SecretValuePlaintext = testSecretValue },
			MockCalled:  true,
		},
		{
			Name:        "Success: Created secret",
			RespErr:     false,
			AugmentOpts: func(o *CreateOpts) { o.SecretValuePlaintext = testSecretValue },
			MockCalled:  true,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			r := require.New(t)

			io := iostreams.Test()
			if c.ReadViaStdin {
				io.InputTTY = true
				io.ErrorTTY = true

				if !c.EmptySecretValue {
					_, err := io.Input.WriteString(testSecretValue)
					r.NoError(err)
				}
			}
			vs := mock_secret_service.NewMockClientService(t)
			opts := &CreateOpts{
				Ctx:        context.Background(),
				IO:         io,
				Profile:    testProfile(t),
				Output:     format.New(io),
				Client:     vs,
				AppName:    testProfile(t).VaultSecrets.AppName,
				SecretName: "test_secret",
			}

			if c.AugmentOpts != nil {
				c.AugmentOpts(opts)
			}

			dt := strfmt.NewDateTime()
			if c.MockCalled {
				if c.RespErr {
					vs.EXPECT().CreateAppKVSecret(mock.Anything, mock.Anything).Return(nil, errors.New(c.ErrMsg)).Once()
				} else {
					vs.EXPECT().CreateAppKVSecret(&secret_service.CreateAppKVSecretParams{
						LocationOrganizationID: testProfile(t).OrganizationID,
						LocationProjectID:      testProfile(t).ProjectID,
						AppName:                testProfile(t).VaultSecrets.AppName,
						Body: secret_service.CreateAppKVSecretBody{
							Name:  opts.SecretName,
							Value: testSecretValue,
						},
						Context: opts.Ctx,
					}, mock.Anything).Return(&secret_service.CreateAppKVSecretOK{
						Payload: &models.Secrets20230613CreateAppKVSecretResponse{
							Secret: &models.Secrets20230613Secret{
								Name:      opts.SecretName,
								CreatedAt: dt,
								Version: &models.Secrets20230613SecretVersion{
									Version:   "2",
									CreatedAt: dt,
									Type:      "kv",
								},
								LatestVersion: "2",
							},
						},
					}, nil).Once()
				}
			}

			// Run the command
			err := createRun(opts)
			if c.ErrMsg != "" {
				r.Contains(err.Error(), c.ErrMsg)
				return
			}

			r.NoError(err)
			r.Contains(io.Error.String(), fmt.Sprintf("✓ Successfully created secret with name %q\n", opts.SecretName))
		})
	}
}
