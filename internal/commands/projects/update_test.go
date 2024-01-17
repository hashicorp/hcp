package projects

import (
	"context"
	"net/http"
	"testing"

	"github.com/go-openapi/runtime/client"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/project_service"
	mock_project_service "github.com/hashicorp/hcp/internal/pkg/api/mocks/github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/project_service"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewCmdUpdate(t *testing.T) {
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
			Error:   "Organization ID and Project ID must be configured",
		},
		{
			Name: "No Project ID",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123")
			},
			Args:  []string{},
			Error: "Organization ID and Project ID must be configured",
		},
		{
			Name: "Too many args",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123").SetProjectID("456")
			},
			Args:  []string{"foo", "bar"},
			Error: "no arguments allowed, but received 2",
		},
		{
			Name: "Good",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123").SetProjectID("456")
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
			updateCmd := NewCmdUpdate(ctx, func(o *UpdateOpts) error {
				gotOpts = o
				return nil
			})
			updateCmd.SetIO(io)

			code := updateCmd.Run(c.Args)
			if c.Error != "" {
				r.NotZero(code)
				r.Contains(io.Error.String(), c.Error)
				return
			}

			r.Zero(code, io.Error.String())
			r.NotNil(gotOpts)
		})
	}
}

func TestUpdateRun(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Name           string
		NewName        string
		NewDescription string
		RespErr        bool
		Error          string
	}{
		{
			Name:           "Server error",
			NewName:        "foo",
			NewDescription: "bar",
			RespErr:        true,
			Error:          "failed to update project name: [PUT /resource-manager/2019-12-10/projects/{id}/name][403]",
		},
		{
			Name:  "No flags",
			Error: "either name or description must be specified",
		},
		{
			Name:           "Good",
			NewName:        "foo",
			NewDescription: "bar",
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			r := require.New(t)

			io := iostreams.Test()
			project := mock_project_service.NewMockClientService(t)
			opts := &UpdateOpts{
				Ctx:     context.Background(),
				Profile: profile.TestProfile(t).SetOrgID("123").SetProjectID("456"),
				IO:      io,
				Client:  project,
			}
			if c.NewName != "" {
				opts.Name = &c.NewName
			}
			if c.NewDescription != "" {
				opts.Description = &c.NewDescription
			}

			// Expect a request to update the name
			if c.NewName != "" {
				call := project.EXPECT().ProjectServiceSetName(mock.MatchedBy(func(req *project_service.ProjectServiceSetNameParams) bool {
					return req.ID == "456" && req.Body.Name == *opts.Name
				}), nil).Once()

				if c.RespErr {
					call.Return(nil, project_service.NewProjectServiceSetNameDefault(http.StatusForbidden))
				} else {
					call.Return(project_service.NewProjectServiceSetNameOK(), nil)
				}
			}

			if c.NewDescription != "" && !c.RespErr {
				project.EXPECT().ProjectServiceSetDescription(mock.MatchedBy(func(req *project_service.ProjectServiceSetDescriptionParams) bool {
					return req.ID == "456" && req.Body.Description == *opts.Description
				}), nil).Once().Return(project_service.NewProjectServiceSetDescriptionOK(), nil)
			}

			// Run the command
			err := updateRun(opts)
			if c.Error != "" {
				r.ErrorContains(err, c.Error)
				return
			}

			r.NoError(err)
			r.Contains(io.Error.String(), "Project \"456\" updated")
		})
	}
}
