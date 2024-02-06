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
			r.Equal(c.ExpectID, gotOpts.Name)
		})
	}
}

func TestReadRun(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Name                 string
		RespErr              bool
		GivenName            string
		ExpectedResourceName string
		Error                string
	}{
		{
			Name:                 "Server error",
			GivenName:            "test-group",
			ExpectedResourceName: "iam/organization/123/group/test-group",
			RespErr:              true,
			Error:                "failed to read group: [GET /iam/2019-12-10/{resource_name}][403]",
		},
		{
			Name:                 "Good suffix",
			GivenName:            "test-group",
			ExpectedResourceName: "iam/organization/123/group/test-group",
		},
		{
			Name:                 "Good full",
			GivenName:            "iam/organization/456/group/test-group",
			ExpectedResourceName: "iam/organization/456/group/test-group",
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			r := require.New(t)

			io := iostreams.Test()
			iam := mock_groups_service.NewMockClientService(t)
			opts := &ReadOpts{
				Ctx:     context.Background(),
				Profile: profile.TestProfile(t).SetOrgID("123"),
				Output:  format.New(io),
				Client:  iam,
				Name:    c.GivenName,
			}

			// Expect a request to get the user.
			call := iam.EXPECT().GroupsServiceGetGroup(mock.MatchedBy(func(req *groups_service.GroupsServiceGetGroupParams) bool {
				return req.ResourceName == c.ExpectedResourceName
			}), nil).Once()

			if c.RespErr {
				call.Return(nil, groups_service.NewGroupsServiceGetGroupDefault(http.StatusForbidden))
			} else {
				ok := groups_service.NewGroupsServiceGetGroupOK()
				ok.Payload = &models.HashicorpCloudIamGetGroupResponse{
					Group: &models.HashicorpCloudIamGroup{
						Description:  "This is a test",
						DisplayName:  "Test-GROUP",
						MemberCount:  11,
						ResourceID:   "iam.group:123456",
						ResourceName: c.ExpectedResourceName,
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
			r.Contains(io.Output.String(), "This is a test")
			r.Contains(io.Output.String(), c.ExpectedResourceName)
			r.Contains(io.Output.String(), "Test-GROUP")
			r.Contains(io.Output.String(), "11")
		})
	}
}
