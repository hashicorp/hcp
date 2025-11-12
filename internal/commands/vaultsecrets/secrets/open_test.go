// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: MPL-2.0

package secrets

import (
	"context"
	"errors"
	"os"
	"path"
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

func TestNewCmdOpen(t *testing.T) {
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
		Expect  *OpenOpts
	}{
		{
			Name:    "No many args",
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
			Expect: &OpenOpts{
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

			var gotOpts *OpenOpts
			readCmd := NewCmdOpen(ctx, func(o *OpenOpts) error {
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

func TestOpenRun(t *testing.T) {
	t.Parallel()

	testProfile := func(t *testing.T) *profile.Profile {
		tp := profile.TestProfile(t).SetOrgID("123").SetProjectID("456")
		tp.VaultSecrets = &profile.VaultSecretsConf{
			AppName: "test-app",
		}
		return tp
	}
	testSecretName := "test_secret"
	testSecretValue := "my super secret value"

	cases := []struct {
		Name        string
		RespErr     bool
		ErrMsg      string
		MockCalled  bool
		WriteToFile bool
	}{
		{
			Name:       "Failed: Secret not found",
			RespErr:    true,
			ErrMsg:     "[GET] /secrets/2023-06-13/organizations/{location.organization_id}/projects/{location.project_id}/apps/{app_name}/open/{secret_name}][404]OpenAppSecret default  &{Code:5 Details:[] Message:secret not found}",
			MockCalled: true,
		},
		{
			Name:       "Success: Opened plaintext secret",
			RespErr:    false,
			MockCalled: true,
		},
		{
			Name:        "Success: Wrote plaintext secret to the file",
			RespErr:     false,
			MockCalled:  true,
			WriteToFile: true,
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
			opts := &OpenOpts{
				Ctx:        context.Background(),
				IO:         io,
				Profile:    testProfile(t),
				Output:     format.New(io),
				Client:     vs,
				AppName:    testProfile(t).VaultSecrets.AppName,
				SecretName: testSecretName,
			}

			if c.WriteToFile {
				tempDir, err := os.MkdirTemp("", "test")
				if err != nil {
					t.Fatalf("Failed to create temp dir: %v", err)
				}
				opts.OutputFilePath = path.Join(tempDir, "secret_value")
			}
			defer func() {
				os.RemoveAll(path.Dir(opts.OutputFilePath))
			}()

			if c.MockCalled {
				if c.RespErr {
					vs.EXPECT().OpenAppSecret(mock.Anything, mock.Anything).Return(nil, errors.New(c.ErrMsg)).Once()
				} else {
					vs.EXPECT().OpenAppSecret(&secret_service.OpenAppSecretParams{
						OrganizationID: testProfile(t).OrganizationID,
						ProjectID:      testProfile(t).ProjectID,
						AppName:        testProfile(t).VaultSecrets.AppName,
						SecretName:     opts.SecretName,
						Context:        opts.Ctx,
					}, mock.Anything).Return(&secret_service.OpenAppSecretOK{
						Payload: &models.Secrets20231128OpenAppSecretResponse{
							Secret: &models.Secrets20231128OpenSecret{
								Name:          opts.SecretName,
								LatestVersion: 3,
								StaticVersion: &models.Secrets20231128OpenSecretStaticVersion{
									Version: 3,
									Value:   "my super secret value",
								},
								CreatedAt: strfmt.DateTime(time.Now()),
								Type:      secretTypeKV,
							},
						},
					}, nil).Once()

				}
			}

			// Run the command
			err := openRun(opts)
			if c.ErrMsg != "" {
				r.Contains(err.Error(), c.ErrMsg)
				return
			}
			r.NoError(err)
			if !c.WriteToFile {
				r.Contains(io.Output.String(), "Value:          my super secret value")
			} else {
				readContent, err := os.ReadFile(opts.OutputFilePath)
				if err != nil {
					t.Fatalf("Failed to read test file: %v", err)
				}
				if string(readContent) != testSecretValue {
					t.Errorf("Expected file content %q, but got %q", testSecretValue, string(readContent))
				}
			}
		})
	}
}
