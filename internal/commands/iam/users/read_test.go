// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package users

import (
	"context"
	"net/http"
	"testing"

	"github.com/go-openapi/runtime/client"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/client/iam_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/models"
	mock_iam_service "github.com/hashicorp/hcp/internal/pkg/api/mocks/github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/client/iam_service"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewCmdRead(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Name     string
		Args     []string
		Profile  func(t *testing.T) *profile.Profile
		Error    string
		ExpectID string
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
				return profile.TestProfile(t).SetOrgID("123").SetProjectID("456")
			},
			Args:  []string{"foo", "bar"},
			Error: "accepts 1 arg(s), received 2",
		},
		{
			Name: "Good",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123").SetProjectID("456")
			},
			Args:     []string{"foo"},
			ExpectID: "foo",
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

			var gotOpts *ReadOpts
			readCmd := NewCmdRead(ctx, func(o *ReadOpts) error {
				gotOpts = o
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
			r.NotNil(gotOpts)
			r.Equal(c.ExpectID, gotOpts.ID)
		})
	}
}

func TestReadRun(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Name    string
		RespErr bool
		Error   string
	}{
		{
			Name:    "Server error",
			RespErr: true,
			Error:   "failed to read user principal: [GET /iam/2019-12-10/organizations/{organization_id}/user-principals/{user_principal_id}][403]",
		},
		{
			Name: "Good",
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			r := require.New(t)

			io := iostreams.Test()
			iam := mock_iam_service.NewMockClientService(t)
			opts := &ReadOpts{
				Ctx:     context.Background(),
				Profile: profile.TestProfile(t).SetOrgID("123"),
				Output:  format.New(io),
				Client:  iam,
				ID:      "456",
			}

			// Expect a request to get the user.
			call := iam.EXPECT().IamServiceGetUserPrincipalByIDInOrganization(mock.MatchedBy(func(req *iam_service.IamServiceGetUserPrincipalByIDInOrganizationParams) bool {
				return req.OrganizationID == "123" && req.UserPrincipalID == "456"
			}), nil).Once()

			if c.RespErr {
				call.Return(nil, iam_service.NewIamServiceGetUserPrincipalByIDInOrganizationDefault(http.StatusForbidden))
			} else {
				ok := iam_service.NewIamServiceGetUserPrincipalByIDInOrganizationOK()
				ok.Payload = &models.HashicorpCloudIamUserPrincipalResponse{
					UserPrincipal: &models.HashicorpCloudIamUserPrincipal{
						ID:       "456",
						FullName: "Test User",
						Email:    "test@hcp.com",
					},
				}

				call.Return(ok, nil)
			}

			// Run the command
			err := readRun(opts)
			if c.Error != "" {
				r.ErrorContains(err, c.Error)
				return
			}

			r.NoError(err)
			r.Contains(io.Output.String(), "Test User")
			r.Contains(io.Output.String(), "test@hcp.com")
			r.Contains(io.Output.String(), "456")
		})
	}
}
