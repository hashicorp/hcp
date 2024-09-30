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

func TestNewCmdUpdate(t *testing.T) {
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
		Expect  *UpdateOpts
	}{
		{
			Name:    "Good",
			Profile: testProfile,
			Args:    []string{"sample-integration", "--config-file", "path/to/file"},
			Expect: &UpdateOpts{
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
			Name:    "Failed: No secret name arg specified",
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

			var gotOpts *UpdateOpts
			createCmd := NewCmdUpdate(ctx, func(o *UpdateOpts) error {
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

func TestUpdateRun(t *testing.T) {
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
			Input: []byte(`type = "aws"
details = {
  "federated_workload_identity" = {
    "audience" = "abc",
    "role_arn" = "def"
  }

  "capabilities" = [
	"ROTATION", "DYNAMIC"
  ]
}`),
		},
	}

	for _, c := range cases {

		c := c
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			r := require.New(t)

			tempDir := t.TempDir()
			f, err := os.Create(filepath.Join(tempDir, "config.hcl"))
			r.NoError(err)
			_, err = f.Write(c.Input)
			r.NoError(err)

			io := iostreams.Test()
			vs := mock_preview_secret_service.NewMockClientService(t)

			opts := &UpdateOpts{
				Ctx:             context.Background(),
				Profile:         profile.TestProfile(t).SetOrgID("123").SetProjectID("abc"),
				IO:              io,
				PreviewClient:   vs,
				Output:          format.New(io),
				IntegrationName: c.IntegrationName,
				ConfigFilePath:  f.Name(),
			}

			if c.Error == "" {
				vs.EXPECT().UpdateAwsIntegration(&preview_secret_service.UpdateAwsIntegrationParams{
					Context:        opts.Ctx,
					OrganizationID: "123",
					ProjectID:      "abc",
					Name:           opts.IntegrationName,
					Body: &preview_models.SecretServiceUpdateAwsIntegrationBody{
						FederatedWorkloadIdentity: &preview_models.Secrets20231128AwsFederatedWorkloadIdentityRequest{
							Audience: "abc",
							RoleArn:  "def",
						},
						Capabilities: []*preview_models.Secrets20231128Capability{
							preview_models.Secrets20231128CapabilityROTATION.Pointer(),
							preview_models.Secrets20231128CapabilityDYNAMIC.Pointer(),
						},
					},
				}, nil).Return(&preview_secret_service.UpdateAwsIntegrationOK{
					Payload: &preview_models.Secrets20231128UpdateAwsIntegrationResponse{
						Integration: &preview_models.Secrets20231128AwsIntegration{
							Name: opts.IntegrationName,
							FederatedWorkloadIdentity: &preview_models.Secrets20231128AwsFederatedWorkloadIdentityResponse{
								Audience: "abc",
								RoleArn:  "def",
							},
							Capabilities: []*preview_models.Secrets20231128Capability{
								preview_models.Secrets20231128CapabilityROTATION.Pointer(),
								preview_models.Secrets20231128CapabilityDYNAMIC.Pointer(),
							},
						},
					},
				}, nil).Once()
			}

			// Run the command
			err = updateRun(opts)
			if c.Error != "" {
				r.ErrorContains(err, c.Error)
				return
			}

			r.NoError(err)
			r.Contains(io.Error.String(), fmt.Sprintf("✓ Successfully updated integration with name %q\n", opts.IntegrationName))
		})
	}
}
