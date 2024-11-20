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

	"github.com/go-openapi/strfmt"
	preview_secret_service "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-11-28/client/secret_service"
	preview_models "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-11-28/models"
	mock_secret_service "github.com/hashicorp/hcp/internal/pkg/api/mocks/github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-06-13/client/secret_service"
	mock_preview_secret_service "github.com/hashicorp/hcp/internal/pkg/api/mocks/github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-11-28/client/secret_service"
	"github.com/stretchr/testify/mock"

	"github.com/go-openapi/runtime/client"
	"github.com/stretchr/testify/require"

	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
)

func TestNewCmdUpdate(t *testing.T) {
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
		Expect  *UpdateOpts
	}{
		{
			Name:    "Failed: No secret name arg specified",
			Profile: testProfile,
			Args:    []string{},
			Error:   "ERROR: missing required flag: --data-file=DATA_FILE_PATH",
		},
		{
			Name:    "Good: Secret name arg specified",
			Profile: testProfile,
			Args:    []string{"test", "--data-file=DATA_FILE_PATH"},
			Expect: &UpdateOpts{
				AppName:    testProfile(t).VaultSecrets.AppName,
				SecretName: "test",
			},
		},
		{
			Name:    "Good: Rotating secret",
			Profile: testProfile,
			Args:    []string{"test", "--secret-type=rotating", "--data-file=DATA_FILE_PATH"},
			Expect: &UpdateOpts{
				AppName:    testProfile(t).VaultSecrets.AppName,
				SecretName: "test",
			},
		},
		{
			Name:    "Good: Dynamic secret",
			Profile: testProfile,
			Args:    []string{"test", "--secret-type=dynamic", "--data-file=DATA_FILE_PATH"},
			Expect: &UpdateOpts{
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

			var gotOpts *UpdateOpts
			updateCmd := NewCmdUpdate(ctx, func(o *UpdateOpts) error {
				gotOpts = o
				gotOpts.AppName = "test-app"
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
			r.NotNil(gotOpts)
			r.Equal(c.Expect.AppName, gotOpts.AppName)
			r.Equal(c.Expect.SecretName, gotOpts.SecretName)
		})
	}
}

func TestUpdateRun(t *testing.T) {
	t.Parallel()

	testProfile := func(t *testing.T) *profile.Profile {
		tp := profile.TestProfile(t).SetOrgID("123").SetProjectID("456")
		tp.VaultSecrets = &profile.VaultSecretsConf{
			AppName: "test-app",
		}
		return tp
	}

	cases := []struct {
		Name             string
		RespErr          bool
		ReadViaStdin     bool
		EmptySecretValue bool
		ErrMsg           string
		MockCalled       bool
		AugmentOpts      func(opts *UpdateOpts)
		Provider         integrations.IntegrationType
		Input            []byte
	}{
		{
			Name:    "Success: Update a MongoDB rotating secret",
			RespErr: false,
			AugmentOpts: func(o *UpdateOpts) {
				o.Type = secretTypeRotating
			},
			MockCalled: true,
			Provider:   integrations.MongoDBAtlas,
			Input: []byte(`type = "mongodb-atlas"
details = {
  rotate_on_update = true
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
			Name:    "Success: Update a Postgres rotating secret",
			RespErr: false,
			AugmentOpts: func(o *UpdateOpts) {
				o.Type = secretTypeRotating
			},
			MockCalled: true,
			Provider:   integrations.Postgres,
			Input: []byte(`type = "postgres"
details = {
  rotate_on_update = true
  integration_name = "postgres-integration"
  rotation_policy_name = "built-in:60-days-2-active"
  postgres_params = {
  	usernames = ["postgres_user_1", "postgres_user_2"]
  }
}`),
		},
		{
			Name:    "Success: Update an Aws dynamic secret",
			RespErr: false,
			AugmentOpts: func(o *UpdateOpts) {
				o.Type = secretTypeDynamic
			},
			MockCalled: true,
			Input: []byte(`type = "aws"

details = {
  "default_ttl" = "3600s"

  "assume_role" = {
    "role_arn" = "ra2"
  }
}`),
		},
		{
			Name:    "Failed: Unable to update integration name of a secret",
			RespErr: true,
			AugmentOpts: func(o *UpdateOpts) {
				o.Type = secretTypeDynamic
			},
			ErrMsg: "Unsupported argument; An argument named \"integration_name\" is not expected here",
			Input: []byte(`type = "aws"
integration_name = "Aws-Int-12"

details = {
  "default_ttl" = "3600s"

  "assume_role" = {
    "role_arn" = "ra2"
  }
}`),
		},
		{
			Name:    "Failed: Unsupported secret type",
			RespErr: true,
			AugmentOpts: func(o *UpdateOpts) {
				o.Type = "random"
			},
			Input:  []byte{},
			ErrMsg: "\"random\" is an unsupported secret type; \"rotating\" and \"dynamic\" are available types",
		},
		{
			Name:    "Failed: Unsupported static secret type",
			RespErr: true,
			AugmentOpts: func(o *UpdateOpts) {
				o.Type = secretTypeKV
			},
			Input:  []byte{},
			ErrMsg: "\"kv\" is an unsupported secret type; \"rotating\" and \"dynamic\" are available types",
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			r := require.New(t)

			io := iostreams.Test()

			vs := mock_secret_service.NewMockClientService(t)
			pvs := mock_preview_secret_service.NewMockClientService(t)

			opts := &UpdateOpts{
				Ctx:           context.Background(),
				IO:            io,
				Profile:       testProfile(t),
				Output:        format.New(io),
				Client:        vs,
				PreviewClient: pvs,
				AppName:       testProfile(t).VaultSecrets.AppName,
				SecretName:    "test_secret",
			}

			if c.AugmentOpts != nil {
				c.AugmentOpts(opts)
			}

			tempDir := t.TempDir()
			f, err := os.Create(filepath.Join(tempDir, "config.hcl"))
			r.NoError(err)
			_, err = f.Write(c.Input)
			r.NoError(err)
			opts.SecretFilePath = f.Name()

			dt := strfmt.NewDateTime()
			if opts.Type == secretTypeRotating {
				if c.MockCalled {
					switch c.Provider {
					case integrations.MongoDBAtlas:
						if c.RespErr {
							pvs.EXPECT().UpdateMongoDBAtlasRotatingSecret(mock.Anything, mock.Anything).Return(nil, errors.New(c.ErrMsg)).Once()
						} else {
							pvs.EXPECT().UpdateMongoDBAtlasRotatingSecret(&preview_secret_service.UpdateMongoDBAtlasRotatingSecretParams{
								OrganizationID: testProfile(t).OrganizationID,
								ProjectID:      testProfile(t).ProjectID,
								AppName:        testProfile(t).VaultSecrets.AppName,
								Name:           "test_secret",
								Body: &preview_models.SecretServiceUpdateMongoDBAtlasRotatingSecretBody{
									RotateOnUpdate:     true,
									RotationPolicyName: "built-in:60-days-2-active",
									SecretDetails: &preview_models.Secrets20231128MongoDBAtlasSecretDetails{
										MongodbGroupID: "mbdgi",
										MongodbRoles: []*preview_models.Secrets20231128MongoDBRole{
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
							}, mock.Anything).Return(&preview_secret_service.UpdateMongoDBAtlasRotatingSecretOK{
								Payload: &preview_models.Secrets20231128UpdateMongoDBAtlasRotatingSecretResponse{
									Config: &preview_models.Secrets20231128RotatingSecretConfig{
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
							pvs.EXPECT().UpdatePostgresRotatingSecret(mock.Anything, mock.Anything).Return(nil, errors.New(c.ErrMsg)).Once()
						} else {
							pvs.EXPECT().UpdatePostgresRotatingSecret(&preview_secret_service.UpdatePostgresRotatingSecretParams{
								OrganizationID: testProfile(t).OrganizationID,
								ProjectID:      testProfile(t).ProjectID,
								AppName:        testProfile(t).VaultSecrets.AppName,
								Name:           "test_secret",
								Body: &preview_models.SecretServiceUpdatePostgresRotatingSecretBody{
									RotateOnUpdate:     true,
									RotationPolicyName: "built-in:60-days-2-active",
									PostgresParams: &preview_models.Secrets20231128PostgresParams{
										Usernames: []string{"postgres_user_1", "postgres_user_2"},
									},
								},
								Context: opts.Ctx,
							}, mock.Anything).Return(&preview_secret_service.UpdatePostgresRotatingSecretOK{
								Payload: &preview_models.Secrets20231128UpdatePostgresRotatingSecretResponse{
									Config: &preview_models.Secrets20231128PostgresRotatingSecretConfig{
										AppName:            opts.AppName,
										CreatedAt:          dt,
										IntegrationName:    "postgres-integration",
										RotationPolicyName: "built-in:60-days-2-active",
										Name:               opts.SecretName,
										PostgresParams: &preview_models.Secrets20231128PostgresParams{
											Usernames: []string{"postgres_user_1", "postgres_user_2"},
										},
									},
								},
							}, nil).Once()
						}
					}

				}
			} else if opts.Type == secretTypeDynamic {
				if c.MockCalled {
					if c.RespErr {
						pvs.EXPECT().UpdateAwsDynamicSecret(mock.Anything, mock.Anything).Return(nil, errors.New(c.ErrMsg)).Once()
					} else {
						pvs.EXPECT().UpdateAwsDynamicSecret(&preview_secret_service.UpdateAwsDynamicSecretParams{
							OrganizationID: testProfile(t).OrganizationID,
							ProjectID:      testProfile(t).ProjectID,
							AppName:        testProfile(t).VaultSecrets.AppName,
							Name:           opts.SecretName,
							Body: &preview_models.SecretServiceUpdateAwsDynamicSecretBody{
								DefaultTTL: "3600s",
								AssumeRole: &preview_models.Secrets20231128AssumeRoleRequest{
									RoleArn: "ra2",
								},
							},
							Context: opts.Ctx,
						}, mock.Anything).Return(&preview_secret_service.UpdateAwsDynamicSecretOK{
							Payload: &preview_models.Secrets20231128UpdateAwsDynamicSecretResponse{
								Secret: &preview_models.Secrets20231128AwsDynamicSecret{
									AssumeRole: &preview_models.Secrets20231128AssumeRoleResponse{
										RoleArn: "ra2",
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
			err = updateRun(opts)
			if c.ErrMsg != "" {
				r.Contains(err.Error(), c.ErrMsg)
				return
			}

			r.NoError(err)
			r.Contains(io.Error.String(), fmt.Sprintf("✓ Successfully updated secret with name %q\n", opts.SecretName))
		})
	}

}
