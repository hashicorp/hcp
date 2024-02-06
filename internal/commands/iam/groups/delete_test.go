package groups

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/go-openapi/runtime/client"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/client/groups_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/project_service"
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

			var gotOpts *DeleteOpts
			deleteCmd := NewCmdDelete(ctx, func(o *DeleteOpts) error {
				gotOpts = o
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
			r.NotNil(gotOpts)
			r.Equal(c.ExpectID, gotOpts.Name)
		})
	}
}

func TestDeleteRun(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Name         string
		RespErr      bool
		GivenName    string
		ExpectedName string
		Error        string
	}{
		{
			Name:         "Server error",
			GivenName:    "deleting-group",
			ExpectedName: "iam/organization/123/group/deleting-group",
			RespErr:      true,
			Error:        "failed to delete group: [DELETE /resource-manager/2019-12-10/projects/{id}][403]",
		},
		{
			Name:         "Good suffix",
			GivenName:    "deleting-group",
			ExpectedName: "iam/organization/123/group/deleting-group",
		},
		{
			Name:         "Good full",
			GivenName:    "iam/organization/456/group/deleting-group",
			ExpectedName: "iam/organization/456/group/deleting-group",
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			r := require.New(t)

			io := iostreams.Test()
			groups := mock_groups_service.NewMockClientService(t)
			opts := &DeleteOpts{
				Ctx:     context.Background(),
				Profile: profile.TestProfile(t).SetOrgID("123"),
				IO:      io,
				Client:  groups,
				Name:    c.GivenName,
			}

			// Expect a request to create the project.
			call := groups.EXPECT().GroupsServiceDeleteGroup(mock.MatchedBy(func(req *groups_service.GroupsServiceDeleteGroupParams) bool {
				return req.ResourceName == c.ExpectedName
			}), nil).Once()

			if c.RespErr {
				call.Return(nil, project_service.NewProjectServiceDeleteDefault(http.StatusForbidden))
			} else {
				ok := groups_service.NewGroupsServiceDeleteGroupOK()
				call.Return(ok, nil)
			}

			// Run the command
			err := deleteRun(opts)
			if c.Error != "" {
				r.ErrorContains(err, c.Error)
				return
			}

			r.NoError(err)
			r.Contains(io.Error.String(), fmt.Sprintf("Group %q deleted", c.ExpectedName))
		})
	}
}

func TestDeleteRun_RejectPrompt(t *testing.T) {
	t.Parallel()
	r := require.New(t)

	io := iostreams.Test()
	io.ErrorTTY = true
	io.InputTTY = true

	groups := mock_groups_service.NewMockClientService(t)
	opts := &DeleteOpts{
		Ctx:     context.Background(),
		Profile: profile.TestProfile(t).SetOrgID("123"),
		IO:      io,
		Client:  groups,
		Name:    "test",
	}

	// Reject the deletion
	_, err := io.Input.WriteRune('n')
	r.NoError(err)

	// Run the command
	err = deleteRun(opts)
	r.NoError(err)

	// Expect to be warned
	r.Contains(io.Error.String(), "The group will be deleted.")

	// We did not mock a call to delete the project, so if successful, we
	// exited.
}
