// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package members

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/go-openapi/runtime/client"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/client/groups_service"
	mock_groups_service "github.com/hashicorp/hcp/internal/pkg/api/mocks/github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/client/groups_service"
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
			Error: "no arguments allowed, but received 2",
		},
		{
			Name: "No Members",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123")
			},
			Args:  []string{"-g=foo"},
			Error: "at least one member must be specified",
		},
		{
			Name: "Good",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123")
			},
			Args: []string{"-g", "foo", "--member", "123", "--member=456"},
			Expect: &DeleteOpts{
				GroupName: "foo",
				Members:   []string{"123", "456"},
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
			r.Equal(c.Expect.GroupName, gotOpts.GroupName)
			r.EqualValues(c.Expect.Members, gotOpts.Members)
		})
	}
}

func TestDeleteRun(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Name         string
		RespErr      bool
		GroupName    string
		ExpectedName string
		Members      []string
		Error        string
	}{
		{
			Name:         "Server error",
			GroupName:    "test-group",
			ExpectedName: "iam/organization/123/group/test-group",
			RespErr:      true,
			Error:        "failed to update group membership: [PUT /iam/2019-12-10/{resource_name}/members][403]",
		},
		{
			Name:         "Good suffix",
			GroupName:    "test-group",
			ExpectedName: "iam/organization/123/group/test-group",
			Members:      []string{"1", "2", "3"},
		},
		{
			Name:         "Good full",
			GroupName:    "iam/organization/456/group/test-group",
			ExpectedName: "iam/organization/456/group/test-group",
			Members:      []string{"1", "2", "3"},
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			r := require.New(t)

			io := iostreams.Test()
			iam := mock_groups_service.NewMockClientService(t)
			opts := &DeleteOpts{
				Ctx:       context.Background(),
				Profile:   profile.TestProfile(t).SetOrgID("123"),
				IO:        io,
				Client:    iam,
				GroupName: c.GroupName,
				Members:   c.Members,
			}

			// Expect a request
			call := iam.EXPECT().GroupsServiceUpdateGroupMembers(mock.MatchedBy(func(req *groups_service.GroupsServiceUpdateGroupMembersParams) bool {
				return req.ResourceName == c.ExpectedName && reflect.DeepEqual(req.Body.MemberPrincipalIdsToRemove, c.Members)
			}), nil).Once()

			if c.RespErr {
				call.Return(nil, groups_service.NewGroupsServiceUpdateGroupMembersDefault(http.StatusForbidden))
			} else {
				ok := groups_service.NewGroupsServiceUpdateGroupMembersOK()
				call.Return(ok, nil)
			}

			// Run the command
			err := deleteRun(opts)
			if c.Error != "" {
				r.ErrorContains(err, c.Error)
				return
			}

			r.NoError(err)
			r.Contains(io.Error.String(), fmt.Sprintf("Membership to group %q updated", c.ExpectedName))
		})
	}
}
