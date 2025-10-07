// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: MPL-2.0

package secrets

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
			Name:    "Good: List succeeded",
			Profile: testProfile,
			Expect: &ListOpts{
				AppName: testProfile(t).VaultSecrets.AppName,
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
			listCmd := NewCmdList(ctx, func(o *ListOpts) error {
				gotOpts = o
				gotOpts.AppName = c.Profile(t).VaultSecrets.AppName
				return nil
			})
			listCmd.SetIO(io)

			code := listCmd.Run(c.Args)
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
			Name:    "Failed: App not found",
			RespErr: true,
			ErrMsg:  "[GET /secrets/2023-11-28/organizations/{organization_id}/projects/{project_id}/apps/{app_name}/secrets][403] ListAppSecrets default  &{Code:7 Details:[] Message:}",
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
				Ctx:     context.Background(),
				IO:      io,
				Profile: testProfile(t),
				Output:  format.New(io),
				Client:  vs,
				AppName: testProfile(t).VaultSecrets.AppName,
			}

			if c.RespErr {
				vs.EXPECT().ListAppSecrets(mock.Anything, mock.Anything).Return(nil, errors.New(c.ErrMsg)).Once()
			} else {
				paginationNextPageToken := "next_page_token"
				vs.EXPECT().ListAppSecrets(&secret_service.ListAppSecretsParams{
					OrganizationID: testProfile(t).OrganizationID,
					ProjectID:      testProfile(t).ProjectID,
					AppName:        testProfile(t).VaultSecrets.AppName,
					Context:        opts.Ctx,
				}, mock.Anything).Return(&secret_service.ListAppSecretsOK{
					Payload: &models.Secrets20231128ListAppSecretsResponse{
						Secrets: getMockSecrets(0, 10),
						Pagination: &models.CommonPaginationResponse{
							NextPageToken: paginationNextPageToken,
						},
					},
				}, nil).Once()

				vs.EXPECT().ListAppSecrets(&secret_service.ListAppSecretsParams{
					OrganizationID:          testProfile(t).OrganizationID,
					ProjectID:               testProfile(t).ProjectID,
					AppName:                 testProfile(t).VaultSecrets.AppName,
					Context:                 opts.Ctx,
					PaginationNextPageToken: &paginationNextPageToken,
				}, mock.Anything).Return(&secret_service.ListAppSecretsOK{
					Payload: &models.Secrets20231128ListAppSecretsResponse{
						Secrets: getMockSecrets(10, 5),
					},
				}, nil).Once()
			}

			// Run the command
			err := listRun(opts)
			if c.ErrMsg != "" {
				r.Contains(err.Error(), c.ErrMsg)
				return
			}

			r.NoError(err)
			r.NotNil(io.Error.String())
		})
	}
}

func getMockSecrets(start, limit int) []*models.Secrets20231128Secret {
	var secrets []*models.Secrets20231128Secret
	for i := start; i < (start + limit); i++ {
		secrets = append(secrets, &models.Secrets20231128Secret{
			Name:          fmt.Sprint("test_secret_", i),
			LatestVersion: int64(rand.Intn(5)),
			CreatedAt:     strfmt.DateTime(time.Now()),
		})
	}
	return secrets
}
