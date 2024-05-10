// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package secrets

import (
	"context"
	"errors"
	"testing"

	"github.com/go-openapi/runtime/client"
	mock_preview_secret_service "github.com/hashicorp/hcp/internal/pkg/api/mocks/github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/preview/2023-11-28/client/secret_service"
	mock_secret_service "github.com/hashicorp/hcp/internal/pkg/api/mocks/github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-06-13/client/secret_service"

	// preview_secret_models "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/preview/2023-11-28/models"
	_ "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/preview/2023-11-28/models"
	_ "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-06-13/models"

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
			AppName: "test-app-name",
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
				gotOpts.AppName = "test-app-name"
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
			AppName: "test-app-name",
		}
		return tp
	}

	cases := []struct {
		Name         string
		RespErr      bool
		EnablePrompt bool
		ErrMsg       string
		MockCalled   bool
		AugmentOpts  func(*CreateOpts)
	}{
		{
			Name:    "Failed: Secret Value cannot be empty",
			RespErr: true,
			ErrMsg:  "secret value cannot be empty",
		},
		{
			Name:        "Failed: Max secret versions reached",
			RespErr:     true,
			ErrMsg:      "[POST /secrets/2023-11-28/organizations/{organization_id}/projects/{project_id}/apps/{app_name}/secret/kv][429] CreateAppKVSecret default  &{Code:8 Details:[] Message:maximum number of secret versions reached}",
			AugmentOpts: func(o *CreateOpts) { o.SecretValuePlaintext = "test_secret" },
			MockCalled:  true,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			r := require.New(t)

			io := iostreams.Test()
			vs := mock_preview_secret_service.NewMockClientService(t)

			opts := &CreateOpts{
				Ctx:           context.Background(),
				IO:            iostreams.Test(),
				Profile:       testProfile(t),
				Output:        format.New(io),
				PreviewClient: vs,
				Client:        mock_secret_service.NewMockClientService(t),
				AppName:       testProfile(t).VaultSecrets.AppName,
				SecretName:    "test_secret",
			}

			if c.AugmentOpts != nil {
				c.AugmentOpts(opts)
			}

			if c.MockCalled {
				vs.EXPECT().CreateAppKVSecret(mock.Anything, mock.Anything).Return(nil, errors.New(c.ErrMsg)).Once()
			}

			// Run the command
			err := createRun(opts)
			if c.ErrMsg != "" {
				r.Contains(err.Error(), c.ErrMsg)
				return
			}

			r.NoError(err)
			// r.Contains(io.Output.String(), c.GroupName)
			// r.Contains(io.Output.String(), rn)
			// r.Contains(io.Output.String(), c.Description)
			// r.Contains(io.Output.String(), fmt.Sprintf("%d", len(c.Members)))
		})
	}
}
