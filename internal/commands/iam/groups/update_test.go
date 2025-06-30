// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package groups

import (
	"context"
	"net/http"
	"testing"

	"github.com/go-openapi/runtime/client"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/client/groups_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/models"
	mock_groups_service "github.com/hashicorp/hcp/internal/pkg/api/mocks/github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/client/groups_service"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewCmdUpdate(t *testing.T) {
	t.Parallel()

	bar, baz := "bar", "baz"
	cases := []struct {
		Name    string
		Args    []string
		Profile func(t *testing.T) *profile.Profile
		Error   string
		Expect  *UpdateOpts
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
			Args: []string{"foo", "--display-name", "bar", "--description", "baz"},
			Expect: &UpdateOpts{
				Name:        "foo",
				DisplayName: &bar,
				Description: &baz,
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

			var gotOpts *UpdateOpts
			createCmd := NewCmdUpdate(ctx, func(o *UpdateOpts) error {
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
			r.Equal(c.Expect.DisplayName, gotOpts.DisplayName)
			r.EqualValues(c.Expect.Description, gotOpts.Description)
		})
	}
}

func TestCreateUpdateNoFields(t *testing.T) {
	t.Parallel()
	r := require.New(t)

	io := iostreams.Test()
	iam := mock_groups_service.NewMockClientService(t)
	opts := &UpdateOpts{
		Ctx:     context.Background(),
		IO:      io,
		Profile: profile.TestProfile(t).SetOrgID("123"),
		Client:  iam,
		Name:    "test",
	}

	err := updateRun(opts)
	r.ErrorContains(err, "either display name or description must be specified")

}

func TestCreateUpdate(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Name         string
		RespErr      bool
		GivenName    string
		ExpectedName string
		DisplayName  string
		Description  string
		Error        string
	}{
		{
			Name:         "Server error",
			GivenName:    "test-group",
			ExpectedName: "iam/organization/123/group/test-group",
			Description:  "This is a test",
			RespErr:      true,
			Error:        "failed to update group: [PATCH /iam/2019-12-10/{resource_name}][403]",
		},
		{
			Name:         "Good suffix",
			GivenName:    "test-group",
			ExpectedName: "iam/organization/123/group/test-group",
			DisplayName:  "new display name",
			Description:  "This is a test",
		},
		{
			Name:         "Good resource name",
			GivenName:    "iam/organization/456/group/test-group",
			ExpectedName: "iam/organization/456/group/test-group",
			DisplayName:  "new display name",
			Description:  "This is a test",
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			r := require.New(t)

			io := iostreams.Test()
			iam := mock_groups_service.NewMockClientService(t)
			opts := &UpdateOpts{
				Ctx:     context.Background(),
				IO:      io,
				Profile: profile.TestProfile(t).SetOrgID("123"),
				Client:  iam,
				Name:    c.GivenName,
			}
			if c.DisplayName != "" {
				opts.DisplayName = &c.DisplayName
			}
			if c.Description != "" {
				opts.Description = &c.Description
			}

			// Expect a request to get the user.
			call := iam.EXPECT().GroupsServiceUpdateGroup2(mock.MatchedBy(func(req *groups_service.GroupsServiceUpdateGroup2Params) bool {
				if req.ResourceName != c.ExpectedName {
					return false
				}

				if c.DisplayName != "" && req.Group.DisplayName != c.DisplayName {
					return false
				}

				if c.Description != "" && req.Group.Description != c.Description {
					return false
				}

				return true
			}), nil).Once()

			if c.RespErr {
				call.Return(nil, groups_service.NewGroupsServiceUpdateGroup2Default(http.StatusForbidden))
			} else {
				ok := groups_service.NewGroupsServiceUpdateGroup2OK()
				ok.Payload = &models.HashicorpCloudIamUpdateGroupResponse{
					OperationID: "test",
				}

				call.Return(ok, nil)
			}

			// Run the command
			err := updateRun(opts)
			if c.Error != "" {
				r.ErrorContains(err, c.Error)
				return
			}

			r.NoError(err)
			r.Contains(io.Error.String(), c.ExpectedName)
		})
	}
}
