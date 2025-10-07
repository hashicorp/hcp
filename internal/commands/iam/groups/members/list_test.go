// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: MPL-2.0

package members

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/go-openapi/runtime/client"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/client/groups_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/models"
	cloud "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	mock_groups_service "github.com/hashicorp/hcp/internal/pkg/api/mocks/github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/client/groups_service"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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
			Name: "Good",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123")
			},
			Args: []string{"-g=foo"},
			Expect: &ListOpts{
				GroupName: "foo",
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
			r.Equal(c.Expect.GroupName, gotOpts.GroupName)
		})
	}
}

func TestListRun(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Name         string
		GroupName    string
		ResourceName string
		Resp         [][]*models.HashicorpCloudIamGroupMember
		RespErr      bool
		Error        string
	}{
		{
			Name:         "Server error",
			GroupName:    "test-group",
			ResourceName: "iam/organization/123/group/test-group",
			RespErr:      true,
			Error:        "failed to list group members: [GET /iam/2019-12-10/{resource_name}/members][403]",
		},
		{
			Name:         "Good suffix no pagination",
			GroupName:    "test-group",
			ResourceName: "iam/organization/123/group/test-group",
			Resp: [][]*models.HashicorpCloudIamGroupMember{
				{
					{
						Email: "alex@hashicorp.com",
						ID:    "123",
						Name:  "Alex Dadgar",
					},
				},
			},
		},
		{
			Name:         "Good full rn no pagination",
			GroupName:    "iam/organization/456/group/test-group",
			ResourceName: "iam/organization/456/group/test-group",
			Resp: [][]*models.HashicorpCloudIamGroupMember{
				{
					{
						Email: "alex@hashicorp.com",
						ID:    "123",
						Name:  "Alex Dadgar",
					},
				},
			},
		},
		{
			Name:         "Good pagination",
			GroupName:    "test-group",
			ResourceName: "iam/organization/123/group/test-group",
			Resp: [][]*models.HashicorpCloudIamGroupMember{
				{
					{
						Email: "alex@hashicorp.com",
						ID:    "123",
						Name:  "Alex Dadgar",
					},
					{
						Email: "bob@hashicorp.com",
						ID:    "456",
						Name:  "Bob Bobert",
					},
				},
				{
					{
						Email: "charlie@hashicorp.com",
						ID:    "789",
						Name:  "Charlie Charleston",
					},
					{
						Email: "david@hashicorp.com",
						ID:    "012",
						Name:  "David Bowie",
					},
				},
			},
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			r := require.New(t)

			io := iostreams.Test()
			iam := mock_groups_service.NewMockClientService(t)
			opts := &ListOpts{
				Ctx:       context.Background(),
				Profile:   profile.TestProfile(t).SetOrgID("123"),
				Output:    format.New(io),
				Client:    iam,
				GroupName: c.GroupName,
			}

			e := iam.EXPECT()
			for i := 0; i < len(c.Resp) || c.RespErr; i++ {
				i := i
				call := e.GroupsServiceListGroupMembers(mock.MatchedBy(func(req *groups_service.GroupsServiceListGroupMembersParams) bool {
					// Expect an org
					if req.ResourceName != c.ResourceName {
						return false
					}

					// No initial pagination
					if i == 0 && req.PaginationNextPageToken != nil {
						return false
					} else if i >= 1 && *req.PaginationNextPageToken != fmt.Sprintf("next-page-%d", i) {
						// Expect a page token
						return false
					}

					return true
				}), nil)

				if c.RespErr {
					call.Return(nil, groups_service.NewGroupsServiceListGroupMembersDefault(http.StatusForbidden))
					break
				} else {
					ok := groups_service.NewGroupsServiceListGroupMembersOK()
					ok.Payload = &models.HashicorpCloudIamListGroupMembersResponse{
						Members: c.Resp[i],
					}

					if i < len(c.Resp)-1 {
						ok.Payload.Pagination = &cloud.HashicorpCloudCommonPaginationResponse{
							NextPageToken: fmt.Sprintf("next-page-%d", i+1),
						}
					}

					call.Return(ok, nil)
				}
			}

			// Run the command
			err := listRun(opts)
			if c.Error != "" {
				r.ErrorContains(err, c.Error)
				return
			}

			r.NoError(err)

			// Check we outputted the project
			for _, page := range c.Resp {
				for _, p := range page {
					r.Contains(io.Output.String(), p.ID)
					r.Contains(io.Output.String(), p.Email)
					r.Contains(io.Output.String(), p.Name)
				}
			}
		})
	}
}
