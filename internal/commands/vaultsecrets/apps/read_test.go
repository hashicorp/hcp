// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apps

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/go-openapi/runtime/client"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-11-28/client/secret_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-11-28/models"
	mock_secret_service "github.com/hashicorp/hcp/internal/pkg/api/mocks/github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-11-28/client/secret_service"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
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
			Args: []string{"company-card"},
			Expect: &ReadOpts{
				AppName: "company-card",
			},
		},
		{
			Name: "No args",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123").SetProjectID("abc")
			},
			Args:  []string{},
			Error: "accepts 1 arg(s), received 0",
		},
		{
			Name: "Too many args",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123").SetProjectID("abc")
			},
			Args:  []string{"company-card", "additional-arg"},
			Error: "ERROR: accepts 1 arg(s), received 2",
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
			r.Equal(c.Expect.AppName, readOpts.AppName)
		})
	}
}

func TestReadRun(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Name    string
		ErrMsg  string
		AppName string
	}{
		{
			Name:   "Failed: App not found",
			ErrMsg: "[GET /secrets/2023-06-13/organizations/{location.organization_id}/projects/{location.project_id}/apps/{name}][403] GetApp",
		},
		{
			Name:    "Success: Read app",
			AppName: "company-card",
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			r := require.New(t)

			io := iostreams.Test()
			vs := mock_secret_service.NewMockClientService(t)

			opts := &ReadOpts{
				Ctx:     context.Background(),
				Profile: profile.TestProfile(t).SetOrgID("123").SetProjectID("abc"),
				IO:      io,
				Client:  vs,
				Output:  format.New(io),
				AppName: c.AppName,
			}

			if c.ErrMsg != "" {
				vs.EXPECT().GetApp(mock.Anything, mock.Anything).Return(nil, errors.New(c.ErrMsg)).Once()
			} else {
				vs.EXPECT().GetApp(&secret_service.GetAppParams{
					OrganizationID: "123",
					ProjectID:      "abc",
					Name:           opts.AppName,
					Context:        opts.Ctx,
				}, nil).Return(&secret_service.GetAppOK{
					Payload: &models.Secrets20231128GetAppResponse{
						App: &models.Secrets20231128App{
							Name: opts.AppName,
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
			r.Contains(io.Output.String(), fmt.Sprintf("App Name       Description\n%s", opts.AppName))
		})
	}
}
