// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package versions

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
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

func TestNewCmdList(t *testing.T) {
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
		Expect  *ListOpts
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
			Expect: &ListOpts{
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

			var gotOpts *ListOpts
			versionsCmd := NewCmdList(ctx, func(o *ListOpts) error {
				gotOpts = o
				gotOpts.AppName = c.Profile(t).VaultSecrets.AppName
				gotOpts.SecretName = testSecretName
				return nil
			})
			versionsCmd.SetIO(io)

			code := versionsCmd.Run(c.Args)
			if c.Error != "" {
				r.NotZero(code)
				r.Contains(io.Error.String(), c.Error)
				return
			}

			r.Zero(code, io.Error.String())
			r.NotNil(gotOpts)
			r.Equal(c.Expect.AppName, gotOpts.AppName)
		})
	}
}

func TestListRun(t *testing.T) {
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
		RespErr bool
		ErrMsg  string
	}{
		{
			Name:    "Failed: Secret not found",
			RespErr: true,
			ErrMsg:  "[GET] /secrets/2023-11-28/organizations/{organization_id}/projects/{project_id}/apps/{app_name}/secrets/{secret_name}/versions][404] ListAppSecretVersions default  &{Code:5 Details:[] Message:secret not found}",
		},
		{
			Name: "Success: Paginated List call",
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
			opts := &ListOpts{
				Ctx:        context.Background(),
				IO:         io,
				Profile:    testProfile(t),
				Output:     format.New(io),
				Client:     vs,
				AppName:    testProfile(t).VaultSecrets.AppName,
				SecretName: testSecretName,
			}

			if c.RespErr {
				vs.EXPECT().ListAppSecretVersions(mock.Anything, mock.Anything).Return(nil, errors.New(c.ErrMsg)).Once()
			} else {
				paginationNextPageToken := "next_page_token"
				vs.EXPECT().ListAppSecretVersions(&secret_service.ListAppSecretVersionsParams{
					OrganizationID: testProfile(t).OrganizationID,
					ProjectID:      testProfile(t).ProjectID,
					AppName:        testProfile(t).VaultSecrets.AppName,
					SecretName:     testSecretName,
					Context:        opts.Ctx,
				}, mock.Anything).Return(&secret_service.ListAppSecretVersionsOK{
					Payload: &models.Secrets20231128ListAppSecretVersionsResponse{
						StaticVersions: getMockSecretList(0, 10),
						Pagination: &models.CommonPaginationResponse{
							NextPageToken: paginationNextPageToken,
						},
					},
				}, nil).Once()

				vs.EXPECT().ListAppSecretVersions(&secret_service.ListAppSecretVersionsParams{
					OrganizationID:          testProfile(t).OrganizationID,
					ProjectID:               testProfile(t).ProjectID,
					AppName:                 testProfile(t).VaultSecrets.AppName,
					SecretName:              testSecretName,
					Context:                 opts.Ctx,
					PaginationNextPageToken: &paginationNextPageToken,
				}, mock.Anything).Return(&secret_service.ListAppSecretVersionsOK{
					Payload: &models.Secrets20231128ListAppSecretVersionsResponse{
						StaticVersions: getMockSecretList(10, 5),
					},
				}, nil).Once()
			}

			// Run the command
			err := versionsRun(opts)
			if c.ErrMsg != "" {
				r.Contains(err.Error(), c.ErrMsg)
				return
			}

			r.NoError(err)
			r.Contains(io.Output.String(), "Version   Created At")
		})
	}
}

func getMockSecretList(start, limit int) *models.Secrets20231128SecretStaticVersionList {
	var secrets []*models.Secrets20231128SecretStaticVersion
	for i := start; i < (start + limit); i++ {
		secrets = append(secrets, &models.Secrets20231128SecretStaticVersion{
			Version:   int64(rand.Intn(5)),
			CreatedAt: strfmt.DateTime(time.Now()),
			CreatedBy: &models.Secrets20231128Principal{
				Email: fmt.Sprintf("test-user-%d@example.com", i),
				Name:  fmt.Sprintf("test-user-%d", i),
				Type:  "kv",
			},
		})
	}
	return &models.Secrets20231128SecretStaticVersionList{
		Versions: secrets,
	}
}
