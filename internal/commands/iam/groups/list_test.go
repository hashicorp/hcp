package groups

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
	}{
		{
			Name:    "No Org",
			Profile: profile.TestProfile,
			Args:    []string{},
			Error:   "Organization ID  must be configured",
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
			Args: []string{},
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

			cmd := NewCmdList(ctx, func(o *ListOpts) error {
				return nil
			})
			cmd.SetIO(io)

			code := cmd.Run(c.Args)
			if c.Error != "" {
				r.NotZero(code)
				r.Contains(io.Error.String(), c.Error)
				return
			}

			r.Zero(code, io.Error.String())
		})
	}
}

func TestListRun(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Name    string
		Resp    [][]*models.HashicorpCloudIamGroup
		RespErr bool
		Error   string
	}{
		{
			Name:    "Server error",
			RespErr: true,
			Error:   "failed to list groups: [GET /iam/2019-12-10/iam/{parent_resource_name}/groups][403]",
		},
		{
			Name: "Good no pagination",
			Resp: [][]*models.HashicorpCloudIamGroup{
				{
					{
						Description: "test group be testing",
						DisplayName: "Test Group",
						MemberCount: 123,
						ResourceID:  "iam.group:12141709",
					},
				},
			},
		},
		{
			Name: "Good pagination",
			Resp: [][]*models.HashicorpCloudIamGroup{
				{
					{
						Description: "test group be testing",
						DisplayName: "Test Group",
						MemberCount: 123,
						ResourceID:  "iam.group:12141709",
					},
					{
						Description: "description 2",
						DisplayName: "Another Group",
						MemberCount: 456,
						ResourceID:  "iam.group:56789324",
					},
				},
				{
					{
						Description: "hello",
						DisplayName: "TEST-group.with variety",
						MemberCount: 222,
						ResourceID:  "iam.group:999931225412",
					},
					{
						Description: "big group",
						DisplayName: "BIG_GROUP",
						MemberCount: 890,
						ResourceID:  "iam.group:7896413140124",
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
				Ctx:     context.Background(),
				Profile: profile.TestProfile(t).SetOrgID("123"),
				Output:  format.New(io),
				Client:  iam,
			}

			e := iam.EXPECT()
			for i := 0; i < len(c.Resp) || c.RespErr; i++ {
				i := i
				call := e.GroupsServiceListGroups(mock.MatchedBy(func(req *groups_service.GroupsServiceListGroupsParams) bool {
					// Expect an org
					if req.ParentResourceName != "organization/123" {
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
					call.Return(nil, groups_service.NewGroupsServiceListGroupsDefault(http.StatusForbidden))
					break
				} else {
					ok := groups_service.NewGroupsServiceListGroupsOK()
					ok.Payload = &models.HashicorpCloudIamListGroupsResponse{
						Groups: c.Resp[i],
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
					r.Contains(io.Output.String(), p.DisplayName)
					r.Contains(io.Output.String(), p.ResourceID)
					r.Contains(io.Output.String(), fmt.Sprintf("%d", p.MemberCount))
					r.Contains(io.Output.String(), p.Description)
				}
			}
		})
	}
}
