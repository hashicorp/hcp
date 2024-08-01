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
	mock_preview_secret_service "github.com/hashicorp/hcp/internal/pkg/api/mocks/github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/preview/2023-11-28/client/secret_service"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
	"github.com/stretchr/testify/mock"
)

func TestNewCmdDelete(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Name    string
		Args    []string
		Profile func(t *testing.T) *profile.Profile
		Error   string
		Expect  *DeleteOpts
	}{
		{
			Name: "Good",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123").SetProjectID("abc")
			},
			Args: []string{"sample-integration", "--type", "twilio"},
			Expect: &DeleteOpts{
				IntegrationName: "sample-integration",
				Type:            "twilio",
			},
		},
		{
			Name: "Missing type flag",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123").SetProjectID("abc")
			},
			Args:  []string{"sample-integration"},
			Error: "ERROR: missing required flag: --type=TYPE",
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

			var deleteOpts *DeleteOpts
			readCmd := NewCmdDelete(ctx, func(o *DeleteOpts) error {
				deleteOpts = o
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
			r.NotNil(deleteOpts)
			r.Equal(c.Expect.IntegrationName, deleteOpts.IntegrationName)
			r.Equal(c.Expect.Type, deleteOpts.Type)
		})
	}
}

func TestDeleteRun(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Name            string
		ErrMsg          string
		IntegrationName string
		Type            IntegrationType
	}{
		{
			Name:   "Failed: Integration not found",
			ErrMsg: "[DELETE /secrets/2023-11-28/organizations/{organization_id}/projects/{project_id}/integrations/twilio/config/{integration_name}][404] DeleteTwilioIntegration",
			Type:   Twilio,
		},
		{
			Name:            "Success: Delete integration",
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

			opts := &DeleteOpts{
				Ctx:             context.Background(),
				Profile:         profile.TestProfile(t).SetOrgID("123").SetProjectID("abc"),
				IO:              io,
				PreviewClient:   vs,
				Output:          format.New(io),
				IntegrationName: c.IntegrationName,
				Type:            c.Type,
			}

			if c.ErrMsg != "" {
				vs.EXPECT().DeleteTwilioIntegration(mock.Anything, mock.Anything).Return(nil, errors.New(c.ErrMsg)).Once()
			} else {
				vs.EXPECT().DeleteTwilioIntegration(&preview_secret_service.DeleteTwilioIntegrationParams{
					OrganizationID:  "123",
					ProjectID:       "abc",
					IntegrationName: opts.IntegrationName,
					Name:            &opts.IntegrationName,
					Context:         opts.Ctx,
				}, nil).Return(&preview_secret_service.DeleteTwilioIntegrationOK{}, nil).Once()
			}

			// Run the command
			err := deleteRun(opts)
			if c.ErrMsg != "" {
				r.Contains(err.Error(), c.ErrMsg)
				return
			}

			r.NoError(err)
			r.Contains(io.Error.String(), fmt.Sprintf("✓ Successfully deleted integration with name \"%s\"\n", c.IntegrationName))
		})
	}
}
