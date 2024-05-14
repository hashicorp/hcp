// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package secrets

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/go-openapi/runtime/client"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-06-13/client/secret_service"
	mock_secret_service "github.com/hashicorp/hcp/internal/pkg/api/mocks/github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-06-13/client/secret_service"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewCmdDelete(t *testing.T) {
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
		Expect  *DeleteOpts
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
			Expect: &DeleteOpts{
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

			var gotOpts *DeleteOpts
			createCmd := NewCmdDelete(ctx, func(o *DeleteOpts) error {
				gotOpts = o
				gotOpts.AppName = c.Profile(t).VaultSecrets.AppName
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

func TestDeleteRun(t *testing.T) {
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
		RespErr bool
		ErrMsg  string
	}{
		{
			Name:    "Failed: Secret not there",
			RespErr: true,
			ErrMsg:  "failed to delete secret with name: \"test_secret\"",
		},
		{
			Name:    "Success: Created secret",
			RespErr: false,
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
			opts := &DeleteOpts{
				Ctx:        context.Background(),
				IO:         io,
				Profile:    testProfile(t),
				Output:     format.New(io),
				Client:     vs,
				AppName:    testProfile(t).VaultSecrets.AppName,
				SecretName: "test_secret",
			}

			if c.RespErr {
				vs.EXPECT().DeleteAppSecret(mock.Anything, mock.Anything).Return(nil, errors.New(c.ErrMsg)).Once()
			} else {
				vs.EXPECT().DeleteAppSecret(&secret_service.DeleteAppSecretParams{
					LocationOrganizationID: testProfile(t).OrganizationID,
					LocationProjectID:      testProfile(t).ProjectID,
					AppName:                testProfile(t).VaultSecrets.AppName,
					SecretName:             opts.SecretName,
					Context:                opts.Ctx,
				}, mock.Anything).Return(&secret_service.DeleteAppSecretOK{}, nil).Once()
			}

			// Run the command
			err := deleteRun(opts)
			if c.ErrMsg != "" {
				r.Contains(err.Error(), c.ErrMsg)
				return
			}

			r.NoError(err)
			r.Equal(io.Error.String(), fmt.Sprintf("✓ Successfully deleted secret with name: %q\n", opts.SecretName))
		})
	}
}
