// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apps

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
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

func TestNewCmdCreate(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Name    string
		Args    []string
		Profile func(t *testing.T) *profile.Profile
		Error   string
		Expect  *CreateOpts
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
			Expect: &CreateOpts{
				Name: "foo",
			},
		},
		{
			Name: "Description",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123")
			},
			Args: []string{"foo", "--description", "test group"},
			Expect: &CreateOpts{
				Name:        "foo",
				Description: "test group",
			},
		},
		{
			Name: "Members",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123")
			},
			Args: []string{"foo", "--member", "123", "--member=456"},
			Expect: &CreateOpts{
				Name:    "foo",
				Members: []string{"123", "456"},
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

			var gotOpts *CreateOpts
			createCmd := NewCmdCreate(ctx, func(o *CreateOpts) error {
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
			r.Equal(c.Expect.Description, gotOpts.Description)
			r.EqualValues(c.Expect.Members, gotOpts.Members)
		})
	}
}

func TestCreateRun(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Name        string
		RespErr     bool
		GroupName   string
		Description string
		Members     []string
		Error       string
	}{
		{
			Name:      "Server error",
			GroupName: "test-group",
			RespErr:   true,
			Error:     "failed to create group: [POST /iam/2019-12-10/iam/{parent_resource_name}/groups][403]",
		},
		{
			Name:      "Good basic",
			GroupName: "test-group",
		},
		{
			Name:        "Good full",
			GroupName:   "test-group",
			Description: "This is a test",
			Members:     []string{"1", "2", "3"},
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			r := require.New(t)

			io := iostreams.Test()
			iam := mock_groups_service.NewMockClientService(t)
			opts := &CreateOpts{
				Ctx:         context.Background(),
				Profile:     profile.TestProfile(t).SetOrgID("123"),
				Output:      format.New(io),
				Client:      iam,
				Name:        c.GroupName,
				Description: c.Description,
				Members:     c.Members,
			}

			// Expect a request to get the user.
			call := iam.EXPECT().GroupsServiceCreateGroup(mock.MatchedBy(func(req *groups_service.GroupsServiceCreateGroupParams) bool {
				return req.ParentResourceName == "organization/123" &&
					req.Body.Name == c.GroupName && req.Body.Description == c.Description &&
					reflect.DeepEqual(req.Body.MemberPrincipalIds, c.Members)
			}), nil).Once()

			rn := fmt.Sprintf("iam/organization/123/group/%s", c.GroupName)
			if c.RespErr {
				call.Return(nil, groups_service.NewGroupsServiceCreateGroupDefault(http.StatusForbidden))
			} else {
				ok := groups_service.NewGroupsServiceCreateGroupOK()
				ok.Payload = &models.HashicorpCloudIamCreateGroupResponse{
					Group: &models.HashicorpCloudIamGroup{
						Description:  c.Description,
						DisplayName:  c.GroupName,
						MemberCount:  int32(len(c.Members)),
						ResourceID:   "iam.group:123456",
						ResourceName: rn,
					},
				}

				call.Return(ok, nil)
			}

			// Run the command
			err := createRun(opts)
			if c.Error != "" {
				r.ErrorContains(err, c.Error)
				return
			}

			r.NoError(err)
			r.Contains(io.Output.String(), c.GroupName)
			r.Contains(io.Output.String(), rn)
			r.Contains(io.Output.String(), c.Description)
			r.Contains(io.Output.String(), fmt.Sprintf("%d", len(c.Members)))
		})
	}
}
