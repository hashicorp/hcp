// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package keys

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/go-openapi/runtime/client"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/client/service_principals_service"
	mock_service_principals_service "github.com/hashicorp/hcp/internal/pkg/api/mocks/github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/client/service_principals_service"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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
			Name:    "No Org",
			Profile: profile.TestProfile,
			Args:    []string{},
			Error:   "Organization ID must be configured before running the command.",
		},
		{
			Name: "Too many args",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123")
			},
			Args:  []string{"foo", "bar"},
			Error: "accepts 1 arg(s), received 2",
		},
		{
			Name: "Good",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123")
			},
			Args: []string{"foo"},
			Expect: &DeleteOpts{
				Name: "foo",
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
			r.Equal(c.Expect.Name, gotOpts.Name)
		})
	}
}

func TestDeleteRun(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Name         string
		Profile      func(t *testing.T) *profile.Profile
		RespErr      bool
		ResourceName string
		Error        string
	}{
		{
			Name: "Server error",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123")
			},
			ResourceName: "iam/organization/123/service-principal/test-sp/key/DSAKJD",
			RespErr:      true,
			Error:        "failed to delete service principal key: [DELETE /2019-12-10/{resource_name_2}][403]",
		},
		{
			Name: "Good",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123").SetProjectID("456")
			},
			ResourceName: "iam/organization/123/service-principal/test-sp/key/DSAKJD",
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			r := require.New(t)

			io := iostreams.Test()
			spService := mock_service_principals_service.NewMockClientService(t)
			opts := &DeleteOpts{
				Ctx:     context.Background(),
				Profile: c.Profile(t),
				IO:      io,
				Client:  spService,
				Name:    c.ResourceName,
			}

			// Expect a request to get the user.
			call := spService.EXPECT().ServicePrincipalsServiceDeleteServicePrincipalKey(mock.MatchedBy(func(req *service_principals_service.ServicePrincipalsServiceDeleteServicePrincipalKeyParams) bool {
				return req.ResourceName2 == c.ResourceName
			}), nil).Once()

			if c.RespErr {
				call.Return(nil, service_principals_service.NewServicePrincipalsServiceDeleteServicePrincipalKeyDefault(http.StatusForbidden))
			} else {
				ok := service_principals_service.NewServicePrincipalsServiceDeleteServicePrincipalKeyOK()
				call.Return(ok, nil)
			}

			// Run the command
			err := deleteRun(opts)
			if c.Error != "" {
				r.ErrorContains(err, c.Error)
				return
			}

			r.NoError(err)
			r.Contains(io.Error.String(), fmt.Sprintf("Service principal key %q deleted", c.ResourceName))
		})
	}
}

func TestDeleteRun_RejectPrompt(t *testing.T) {
	t.Parallel()
	r := require.New(t)

	io := iostreams.Test()
	io.ErrorTTY = true
	io.InputTTY = true

	spService := mock_service_principals_service.NewMockClientService(t)
	opts := &DeleteOpts{
		Ctx:     context.Background(),
		Profile: profile.TestProfile(t).SetOrgID("123"),
		IO:      io,
		Client:  spService,
		Name:    "iam/organization/123/service-principal/test-sp/key/DSAKJD",
	}

	// Reject the deletion
	_, err := io.Input.WriteRune('n')
	r.NoError(err)

	// Run the command
	err = deleteRun(opts)
	r.NoError(err)

	// Expect to be warned
	r.Contains(io.Error.String(), "The service principal key will be deleted.")

	// We did not mock a call to delete the project, so if successful, we
	// exited.
}
