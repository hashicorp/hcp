// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package integrations

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-openapi/runtime/client"
	"github.com/stretchr/testify/require"

	preview_secret_service "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/preview/2023-11-28/client/secret_service"
	preview_models "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/preview/2023-11-28/models"
	mock_preview_secret_service "github.com/hashicorp/hcp/internal/pkg/api/mocks/github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/preview/2023-11-28/client/secret_service"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
)

func TestNewCmdCreate(t *testing.T) {
	t.Parallel()

	testProfile := func(t *testing.T) *profile.Profile {
		tp := profile.TestProfile(t).SetOrgID("123").SetProjectID("456")
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
			Name:    "Good",
			Profile: testProfile,
			Args:    []string{"sample-integration", "--config-file", "path/to/file"},
			Expect: &CreateOpts{
				IntegrationName: "sample-integration",
				ConfigFilePath:  "path/to/file",
			},
		},
		{
			Name:    "Failed: No secret name arg specified",
			Profile: testProfile,
			Args:    []string{"--config-file", "path/to/file"},
			Error:   "ERROR: accepts 1 arg(s), received 0",
		},
		{
			Name:    "Failed: No config file flag specified",
			Profile: testProfile,
			Args:    []string{"sample-integration"},
			Error:   "ERROR: missing required flag: --config-file=CONFIG_FILE",
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
			r.Equal(c.Expect.IntegrationName, gotOpts.IntegrationName)
			r.Equal(c.Expect.ConfigFilePath, gotOpts.ConfigFilePath)
		})
	}
}

func TestCreateRun(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Name            string
		IntegrationName string
		Input           []byte
		Error           string
	}{
		{
			Name:            "Good",
			IntegrationName: "sample-integration",
			Input: []byte(`version: 1.0.0
type: "aws"
details:
  audience: abc
  role_arn: def`),
		},
		{
			Name:            "Missing a single required field",
			IntegrationName: "sample-integration",
			Input: []byte(`version: 1.0.0
type: "mongodb-atlas"
details:
  public_key: abc`),
			Error: "missing required field(s) in the config file: [private_key]",
		},
		{
			Name:            "Missing multiple required fields",
			IntegrationName: "sample-integration",
			Input: []byte(`version: 1.0.0
type: "twilio"
details:
  api_key_sid: ghi`),
			Error: "missing required field(s) in the config file: [account_sid api_key_secret]",
		},
	}

	for _, c := range cases {

		c := c
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			r := require.New(t)

			tempDir := t.TempDir()
			f, err := os.Create(filepath.Join(tempDir, "config.yaml"))
			r.NoError(err)
			_, err = f.Write(c.Input)
			r.NoError(err)

			io := iostreams.Test()
			vs := mock_preview_secret_service.NewMockClientService(t)

			opts := &CreateOpts{
				Ctx:             context.Background(),
				Profile:         profile.TestProfile(t).SetOrgID("123").SetProjectID("abc"),
				IO:              io,
				PreviewClient:   vs,
				Output:          format.New(io),
				IntegrationName: c.IntegrationName,
				ConfigFilePath:  f.Name(),
			}

			if c.Error == "" {
				vs.EXPECT().CreateAwsIntegration(&preview_secret_service.CreateAwsIntegrationParams{
					Context:        opts.Ctx,
					OrganizationID: "123",
					ProjectID:      "abc",
					Body: &preview_models.SecretServiceCreateAwsIntegrationBody{
						Name: opts.IntegrationName,
						FederatedWorkloadIdentity: &preview_models.Secrets20231128AwsFederatedWorkloadIdentityRequest{
							Audience: "abc",
							RoleArn:  "def",
						},
					},
				}, nil).Return(&preview_secret_service.CreateAwsIntegrationOK{
					Payload: &preview_models.Secrets20231128CreateAwsIntegrationResponse{
						Integration: &preview_models.Secrets20231128AwsIntegration{
							Name: opts.IntegrationName,
							FederatedWorkloadIdentity: &preview_models.Secrets20231128AwsFederatedWorkloadIdentityResponse{
								Audience: "abc",
								RoleArn:  "def",
							},
						},
					},
				}, nil).Once()
			}

			// Run the command
			err = createRun(opts)
			if c.Error != "" {
				r.ErrorContains(err, c.Error)
				return
			}

			r.NoError(err)
			r.Contains(io.Error.String(), fmt.Sprintf("✓ Successfully created integration with name %q\n", opts.IntegrationName))
		})
	}
}