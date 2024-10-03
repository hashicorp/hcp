// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package integrations

import (
	"context"
	"errors"
	"fmt"
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
	"github.com/stretchr/testify/mock"
)

func TestNewCmdRead(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Name    string
		Args    []string
		Profile func(t *testing.T) *profile.Profile
		Error   string
		Expect  *ReadOpts
	}{
		{
			Name: "Good",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123").SetProjectID("abc")
			},
			Args: []string{"sample-integration", "--type", "twilio"},
			Expect: &ReadOpts{
				IntegrationName: "sample-integration",
				Type:            "twilio",
			},
		},
	}
	for _, c := range cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()

			r := require.New(t)
			io := iostreams.Test()

			ctx := &cmd.Context{
				IO:          io,
				Profile:     c.Profile(t),
				ShutdownCtx: context.Background(),
				HCP:         &client.Runtime{},
				Output:      format.New(io),
			}

			var readOpts *ReadOpts
			readCmd := NewCmdRead(ctx, func(o *ReadOpts) error {
				readOpts = o
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
			r.NotNil(readOpts)
			r.Equal(c.Expect.IntegrationName, readOpts.IntegrationName)
			r.Equal(c.Expect.Type, readOpts.Type)

		})
	}
}

func TestReadRun(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Name            string
		ErrMsg          string
		IntegrationName string
		Type            IntegrationType
	}{
		{
			Name:   "Failed: Integration not found",
			ErrMsg: "[GET /secrets/2023-11-28/organizations/{organization_id}/projects/{project_id}/integrations/twilio/config/{integration_name}][404] GetTwilioIntegration",
			Type:   Twilio,
		},
		{
			Name:            "Success: Read integration",
			IntegrationName: "sample-integration",
			Type:            Twilio,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			r := require.New(t)

			io := iostreams.Test()
			vs := mock_preview_secret_service.NewMockClientService(t)

			opts := &ReadOpts{
				Ctx:             context.Background(),
				Profile:         profile.TestProfile(t).SetOrgID("123").SetProjectID("abc"),
				IO:              io,
				PreviewClient:   vs,
				Output:          format.New(io),
				IntegrationName: c.IntegrationName,
				Type:            c.Type,
			}

			if c.ErrMsg != "" {
				vs.EXPECT().GetTwilioIntegration(mock.Anything, mock.Anything).Return(nil, errors.New(c.ErrMsg)).Once()
			} else {
				vs.EXPECT().GetTwilioIntegration(&preview_secret_service.GetTwilioIntegrationParams{
					OrganizationID: "123",
					ProjectID:      "abc",
					Name:           opts.IntegrationName,
					Context:        opts.Ctx,
				}, nil).Return(&preview_secret_service.GetTwilioIntegrationOK{
					Payload: &preview_models.Secrets20231128GetTwilioIntegrationResponse{
						Integration: &preview_models.Secrets20231128TwilioIntegration{
							Name: opts.IntegrationName,
							StaticCredentialDetails: &preview_models.Secrets20231128TwilioStaticCredentialsResponse{
								AccountSid: "account_sid",
								APIKeySid:  "api_key_sid",
							},
						},
					},
				}, nil).Once()
			}

			// Run the command
			err := readRun(opts)
			if c.ErrMsg != "" {
				r.Contains(err.Error(), c.ErrMsg)
				return
			}

			r.NoError(err)
			r.Contains(io.Output.String(), fmt.Sprintf("Integration Name     Account SID   API Key SID\n%s   account_sid   api_key_sid\n", opts.IntegrationName))
		})
	}
}
