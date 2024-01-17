package projects

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/project_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/models"
	mock_project_service "github.com/hashicorp/hcp/internal/pkg/api/mocks/github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/project_service"
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
		Name    string
		Args    []string
		Profile func(t *testing.T) *profile.Profile
		Error   string
	}{
		{
			Name:    "No Org",
			Profile: profile.TestProfile,
			Args:    []string{},
			Error:   "Organization ID and Project ID must be configured before running the command.",
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
			Error:   "failed to read project: [GET /resource-manager/2019-12-10/projects/{id}][403]",
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
			project := mock_project_service.NewMockClientService(t)
			opts := &ReadOpts{
				Ctx:     context.Background(),
				Profile: profile.TestProfile(t).SetOrgID("123").SetProjectID("456"),
				IO:      io,
				Output:  format.New(io),
				Client:  project,
			}

			// Expect a request to create the project.
			call := project.EXPECT().ProjectServiceGet(mock.MatchedBy(func(req *project_service.ProjectServiceGetParams) bool {
				return req.ID == "456"
			}), nil).Once()

			if c.RespErr {
				call.Return(nil, project_service.NewProjectServiceGetDefault(http.StatusForbidden))
			} else {
				ok := project_service.NewProjectServiceGetOK()
				ok.Payload = &models.HashicorpCloudResourcemanagerProjectGetResponse{
					Project: &models.HashicorpCloudResourcemanagerProject{
						CreatedAt:   strfmt.DateTime(time.Now()),
						Description: "World",
						ID:          "456",
						Name:        "Hello",
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
			r.Contains(io.Output.String(), "Hello")
			r.Contains(io.Output.String(), "World")
			r.Contains(io.Output.String(), "456")
		})
	}
}
