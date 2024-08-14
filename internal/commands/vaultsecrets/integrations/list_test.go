// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package integrations

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/go-openapi/runtime/client"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	preview_secret_service "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/preview/2023-11-28/client/secret_service"
	preview_models "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/preview/2023-11-28/models"
	mock_preview_secret_service "github.com/hashicorp/hcp/internal/pkg/api/mocks/github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/preview/2023-11-28/client/secret_service"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
)

func TestNewCmdList(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Name    string
		Args    []string
		Profile func(t *testing.T) *profile.Profile
		Error   string
		Expect  *ListOpts
	}{
		{
			Name: "Good",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123").SetProjectID("abc")
			},
			Args: []string{"--type=twilio"},
			Expect: &ListOpts{
				Type: "twilio",
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
		})
	}
}

func TestListRun(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Name    string
		RespErr bool
		ErrMsg  string
	}{
		{
			Name: "Success: List integrations",
		},
		{
			Name:    "Failed: Unable to list integrations",
			RespErr: true,
			ErrMsg:  "[GET /secrets/2023-11-28/organizations/{organization_id}/projects/{project_id}/integrations][404] ListIntegrations",
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			r := require.New(t)

			io := iostreams.Test()
			io.ErrorTTY = true
			vs := mock_preview_secret_service.NewMockClientService(t)

			opts := &ListOpts{
				Ctx:           context.Background(),
				IO:            io,
				Profile:       profile.TestProfile(t).SetOrgID("123").SetProjectID("abc"),
				Output:        format.New(io),
				PreviewClient: vs,
				Type:          "twilio",
			}

			if c.RespErr {
				vs.EXPECT().ListTwilioIntegrations(mock.Anything, mock.Anything).Return(nil, errors.New(c.ErrMsg)).Once()
			} else {
				paginationNextPageToken := "token"
				vs.EXPECT().ListTwilioIntegrations(&preview_secret_service.ListTwilioIntegrationsParams{
					OrganizationID: "123",
					ProjectID:      "abc",
					Context:        opts.Ctx,
				}, mock.Anything).Return(&preview_secret_service.ListTwilioIntegrationsOK{
					Payload: &preview_models.Secrets20231128ListTwilioIntegrationsResponse{
						Integrations: getIntegrations(0, 10),
						Pagination: &preview_models.CommonPaginationResponse{
							NextPageToken: paginationNextPageToken,
						},
					},
				}, nil).Once()

				vs.EXPECT().ListTwilioIntegrations(&preview_secret_service.ListTwilioIntegrationsParams{
					OrganizationID:          "123",
					ProjectID:               "abc",
					Context:                 opts.Ctx,
					PaginationNextPageToken: &paginationNextPageToken,
				}, mock.Anything).Return(&preview_secret_service.ListTwilioIntegrationsOK{
					Payload: &preview_models.Secrets20231128ListTwilioIntegrationsResponse{
						Integrations: getIntegrations(10, 5),
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

func getIntegrations(start, limit int) []*preview_models.Secrets20231128TwilioIntegration {
	var secrets []*preview_models.Secrets20231128TwilioIntegration
	for i := start; i < (start + limit); i++ {
		secrets = append(secrets, &preview_models.Secrets20231128TwilioIntegration{
			Name: fmt.Sprint("test_app_", i),
		})
	}
	return secrets
}
