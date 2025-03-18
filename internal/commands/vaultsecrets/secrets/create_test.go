// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package secrets

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/hcp/internal/commands/vaultsecrets/integrations"

	"github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
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
			Name:    "Good: Secret name arg specified",
			Profile: testProfile,
			Args:    []string{"test", "--data-file=DATA_FILE_PATH"},
			Expect: &CreateOpts{
				AppName:    testProfile(t).VaultSecrets.AppName,
				SecretName: "test",
			},
		},
		{
			Name:    "Good: Rotating secret",
			Profile: testProfile,
			Args:    []string{"test", "--secret-type=rotating", "--data-file=DATA_FILE_PATH"},
			Expect: &CreateOpts{
				AppName:    testProfile(t).VaultSecrets.AppName,
				SecretName: "test",
			},
		},
		{
			Name:    "Good: Dynamic secret",
			Profile: testProfile,
			Args:    []string{"test", "--secret-type=dynamic", "--data-file=DATA_FILE_PATH"},
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
		Provider         integrations.IntegrationType
		Input            []byte
	}{
		{
			Name:   "Failed: Read via stdin as hyphen not supplied for --data-file flag",
			ErrMsg: "data file path is required",
		},
		{
			Name:             "Failed: Read empty secret value via stdin",
			EmptySecretValue: true,
			ReadViaStdin:     true,
			RespErr:          true,
			AugmentOpts: func(o *CreateOpts) {
				o.SecretFilePath = "-"
				o.Type = secretTypeKV
			},
			ErrMsg: "secret value cannot be empty",
		},
		{
			Name:         "Success: Create secret via stdin",
			ReadViaStdin: true,
			AugmentOpts: func(o *CreateOpts) {
				o.SecretFilePath = "-"
				o.Type = secretTypeKV
			},
			MockCalled: true,
		},
		{
			Name:    "Failed: Max secret versions reached",
			RespErr: true,
			ErrMsg:  "[POST /secrets/2023-11-28/organizations/{organization_id}/projects/{project_id}/apps/{app_name}/secret/kv][429] CreateAppKVSecret default  &{Code:8 Details:[] Message:maximum number of secret versions reached}",
			AugmentOpts: func(o *CreateOpts) {
				o.SecretValuePlaintext = testSecretValue
				o.Type = secretTypeKV
			},
			MockCalled: true,
		},
		{
			Name:    "Success: Created secret",
			RespErr: false,
			AugmentOpts: func(o *CreateOpts) {
				o.SecretValuePlaintext = testSecretValue
				o.Type = secretTypeKV
			},
			MockCalled: true,
		},
		{
			Name:    "Success: Create a MongoDB rotating secret",
			RespErr: false,
			AugmentOpts: func(o *CreateOpts) {
				o.Type = secretTypeRotating
			},
			MockCalled: true,
			Provider:   integrations.MongoDBAtlas,
			Input: []byte(`type = "mongodb-atlas"
details = {
  integration_name = "mongo-db-integration"
  rotation_policy_name = "built-in:60-days-2-active"
  secret_details = {
    mongodb_group_id = "mbdgi"
    mongodb_roles = [{
      "role_name" = "rn1"
      "database_name" = "dn1"
      "collection_name" = "cn1"
    },
	{
	  "role_name" = "rn2"
	  "database_name" = "dn2"
	  "collection_name" = "cn2"
	}]
  }
}`),
		},
		{
			Name:    "Success: Create a Postgres rotating secret",
			RespErr: false,
			AugmentOpts: func(o *CreateOpts) {
				o.Type = secretTypeRotating
			},
			MockCalled: true,
			Provider:   integrations.Postgres,
			Input: []byte(`type = "postgres"
details = {
  integration_name = "postgres-integration"
  rotation_policy_name = "built-in:60-days-2-active"
  postgres_params = {
  	usernames = ["postgres_user_1", "postgres_user_2"]
  }
}`),
		},
		{
			Name:    "Success: Create an Aws dynamic secret",
			RespErr: false,
			AugmentOpts: func(o *CreateOpts) {
				o.Type = secretTypeDynamic
			},
			MockCalled: true,
			Input: []byte(`type = "aws"

details = {
  "integration_name" = "Aws-Int-12"
  "default_ttl" = "3600s"

  "assume_role" = {
    "role_arn" = "ra"
  }
}`),
		},
		{
			Name:    "Failed: Unsupported secret type",
			RespErr: true,
			AugmentOpts: func(o *CreateOpts) {
				o.Type = "random"
			},
			Input:  []byte{},
			ErrMsg: "\"random\" is an unsupported secret type; \"static\", \"rotating\", \"dynamic\" are available types",
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

			if (opts.Type == secretTypeRotating || opts.Type == secretTypeDynamic) && c.Input != nil {
				tempDir := t.TempDir()
				f, err := os.Create(filepath.Join(tempDir, "config.hcl"))
				r.NoError(err)
				_, err = f.Write(c.Input)
				r.NoError(err)
				opts.SecretFilePath = f.Name()
			}

			dt := strfmt.NewDateTime()
			if opts.Type == secretTypeKV {
				if c.MockCalled {
					if c.RespErr {
						vs.EXPECT().CreateAppKVSecret(mock.Anything, mock.Anything).Return(nil, errors.New(c.ErrMsg)).Once()
					} else {
						vs.EXPECT().CreateAppKVSecret(&secret_service.CreateAppKVSecretParams{
							OrganizationID: testProfile(t).OrganizationID,
							ProjectID:      testProfile(t).ProjectID,
							AppName:        testProfile(t).VaultSecrets.AppName,
							Body: &models.SecretServiceCreateAppKVSecretBody{
								Name:  opts.SecretName,
								Value: testSecretValue,
							},
							Context: opts.Ctx,
						}, mock.Anything).Return(&secret_service.CreateAppKVSecretOK{
							Payload: &models.Secrets20231128CreateAppKVSecretResponse{
								Secret: &models.Secrets20231128Secret{
									Name:      opts.SecretName,
									CreatedAt: dt,
									StaticVersion: &models.Secrets20231128SecretStaticVersion{
										Version:   2,
										CreatedAt: dt,
									},
									LatestVersion: 2,
								},
							},
						}, nil).Once()
					}
				}
			} else if opts.Type == secretTypeRotating {
				if c.MockCalled {
					switch c.Provider {
					case integrations.MongoDBAtlas:
						if c.RespErr {
							vs.EXPECT().CreateMongoDBAtlasRotatingSecret(mock.Anything, mock.Anything).Return(nil, errors.New(c.ErrMsg)).Once()
						} else {
							vs.EXPECT().CreateMongoDBAtlasRotatingSecret(&secret_service.CreateMongoDBAtlasRotatingSecretParams{
								OrganizationID: testProfile(t).OrganizationID,
								ProjectID:      testProfile(t).ProjectID,
								AppName:        testProfile(t).VaultSecrets.AppName,
								Body: &models.SecretServiceCreateMongoDBAtlasRotatingSecretBody{
									Name:               opts.SecretName,
									IntegrationName:    "mongo-db-integration",
									RotationPolicyName: "built-in:60-days-2-active",
									SecretDetails: &models.Secrets20231128MongoDBAtlasSecretDetails{
										MongodbGroupID: "mbdgi",
										MongodbRoles: []*models.Secrets20231128MongoDBRole{
											{
												RoleName:       "rn1",
												DatabaseName:   "dn1",
												CollectionName: "cn1",
											},
											{
												RoleName:       "rn2",
												DatabaseName:   "dn2",
												CollectionName: "cn2",
											},
										},
									},
								},
								Context: opts.Ctx,
							}, mock.Anything).Return(&secret_service.CreateMongoDBAtlasRotatingSecretOK{
								Payload: &models.Secrets20231128CreateMongoDBAtlasRotatingSecretResponse{
									Config: &models.Secrets20231128RotatingSecretConfig{
										AppName:            opts.AppName,
										CreatedAt:          dt,
										IntegrationName:    "mongo-db-integration",
										RotationPolicyName: "built-in:60-days-2-active",
										Name:               opts.SecretName,
									},
								},
							}, nil).Once()
						}
					case integrations.Postgres:
						if c.RespErr {
							vs.EXPECT().CreatePostgresRotatingSecret(mock.Anything, mock.Anything).Return(nil, errors.New(c.ErrMsg)).Once()
						} else {
							println(testProfile(t).ProjectID)
							vs.EXPECT().CreatePostgresRotatingSecret(&secret_service.CreatePostgresRotatingSecretParams{
								OrganizationID: testProfile(t).OrganizationID,
								ProjectID:      testProfile(t).ProjectID,
								AppName:        testProfile(t).VaultSecrets.AppName,
								Body: &models.SecretServiceCreatePostgresRotatingSecretBody{
									Name:               opts.SecretName,
									IntegrationName:    "postgres-integration",
									RotationPolicyName: "built-in:60-days-2-active",
									PostgresParams:     &models.Secrets20231128PostgresParams{Usernames: []string{"postgres_user_1", "postgres_user_2"}},
								},
								Context: opts.Ctx,
							}, mock.Anything).Return(&secret_service.CreatePostgresRotatingSecretOK{
								Payload: &models.Secrets20231128CreatePostgresRotatingSecretResponse{
									Config: &models.Secrets20231128PostgresRotatingSecretConfig{
										AppName:            opts.AppName,
										CreatedAt:          dt,
										IntegrationName:    "postgres-integration",
										RotationPolicyName: "built-in:60-days-2-active",
										Name:               opts.SecretName,
										PostgresParams:     &models.Secrets20231128PostgresParams{Usernames: []string{"postgres_user_1", "postgres_user_2"}},
									},
								},
							}, nil).Once()
						}
					}
				}
			} else if opts.Type == secretTypeDynamic {
				if c.MockCalled {
					if c.RespErr {
						vs.EXPECT().CreateAwsDynamicSecret(mock.Anything, mock.Anything).Return(nil, errors.New(c.ErrMsg)).Once()
					} else {
						vs.EXPECT().CreateAwsDynamicSecret(&secret_service.CreateAwsDynamicSecretParams{
							OrganizationID: testProfile(t).OrganizationID,
							ProjectID:      testProfile(t).ProjectID,
							AppName:        testProfile(t).VaultSecrets.AppName,
							Body: &models.SecretServiceCreateAwsDynamicSecretBody{
								IntegrationName: "Aws-Int-12",
								Name:            opts.SecretName,
								DefaultTTL:      "3600s",
								AssumeRole: &models.Secrets20231128AssumeRoleRequest{
									RoleArn: "ra",
								},
							},
							Context: opts.Ctx,
						}, mock.Anything).Return(&secret_service.CreateAwsDynamicSecretOK{
							Payload: &models.Secrets20231128CreateAwsDynamicSecretResponse{
								Secret: &models.Secrets20231128AwsDynamicSecret{
									AssumeRole: &models.Secrets20231128AssumeRoleResponse{
										RoleArn: "ra",
									},
									DefaultTTL:      "3600s",
									CreatedAt:       dt,
									IntegrationName: "Aws-Int-12",
									Name:            opts.SecretName,
								},
							},
						}, nil).Once()
					}
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
