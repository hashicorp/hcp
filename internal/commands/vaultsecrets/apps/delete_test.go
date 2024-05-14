package apps

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/go-openapi/runtime/client"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-06-13/client/secret_service"
	mock_secret_service "github.com/hashicorp/hcp/internal/pkg/api/mocks/github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-06-13/client/secret_service"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
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
			Args: []string{"company-card"},
			Expect: &DeleteOpts{
				AppName: "company-card",
			},
		},
		{
			Name:    "No Org or Project ID",
			Profile: profile.TestProfile,
			Args:    []string{},
			Error:   "Organization ID and Project ID must be configured before running the command.",
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

			var deleteOpts *DeleteOpts
			deleteCmd := NewCmdDelete(ctx, func(o *DeleteOpts) error {
				deleteOpts = o
				return nil
			})
			deleteCmd.SetIO(io)

			code := deleteCmd.Run(c.Args)
			if c.Error != "" {
				r.NotZero(code)
				r.Contains(io.Error.String(), c.Error)
				return
			}

			r.Zero(code, io.Error.String())
			r.NotNil(deleteOpts)
			r.Equal(c.Expect.AppName, deleteOpts.AppName)
		})
	}
}

func TestDeleteRun(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Name    string
		ErrMsg  string
		AppName string
	}{
		{
			Name:   "Failed: App not found",
			ErrMsg: "[DELETE /secrets/2023-06-13/organizations/{location.organization_id}/projects/{location.project_id}/apps/{name}][403] DeleteApp",
		},
		{
			Name:    "Success: Delete app",
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

			opts := &DeleteOpts{
				Ctx:     context.Background(),
				Profile: profile.TestProfile(t).SetOrgID("123").SetProjectID("abc"),
				IO:      io,
				Client:  vs,
				Output:  format.New(io),
				AppName: c.AppName,
			}

			if c.ErrMsg != "" {
				vs.EXPECT().DeleteApp(mock.Anything, mock.Anything).Return(nil, errors.New(c.ErrMsg)).Once()
			} else {
				vs.EXPECT().DeleteApp(&secret_service.DeleteAppParams{
					LocationOrganizationID: "123",
					LocationProjectID:      "abc",
					Name:                   opts.AppName,
					Context:                opts.Ctx,
				}, nil).Return(&secret_service.DeleteAppOK{}, nil).Once()
			}

			// Run the command
			err := deleteRun(opts)
			if c.ErrMsg != "" {
				r.Contains(err.Error(), c.ErrMsg)
				return
			}

			r.NoError(err)
			r.Equal(io.Error.String(), fmt.Sprintf("✓ Successfully deleted application with name %q\n", opts.AppName))
		})
	}
}
