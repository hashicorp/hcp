// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: MPL-2.0

package projects

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/project_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/models"
	cloud "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	mock_project_service "github.com/hashicorp/hcp/internal/pkg/api/mocks/github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/project_service"
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
			Error:   "Organization ID must be configured",
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

			listCmd := NewCmdList(ctx, func(o *ListOpts) error {
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
		})
	}
}

func TestListRun(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Name    string
		Resp    [][]*models.HashicorpCloudResourcemanagerProject
		RespErr bool
		Error   string
	}{
		{
			Name:    "Server error",
			RespErr: true,
			Error:   "failed to list projects: [GET /resource-manager/2019-12-10/projects][403]",
		},
		{
			Name: "Good no pagination",
			Resp: [][]*models.HashicorpCloudResourcemanagerProject{
				{
					{
						CreatedAt:   strfmt.DateTime(time.Now()),
						Description: "Good test",
						ID:          "456",
						Name:        "Good",
					},
				},
			},
		},
		{
			Name: "Good pagination",
			Resp: [][]*models.HashicorpCloudResourcemanagerProject{
				{
					{
						CreatedAt:   strfmt.DateTime(time.Now()),
						Description: "Good test",
						ID:          "456",
						Name:        "Good",
					},
					{
						CreatedAt:   strfmt.DateTime(time.Now()),
						Description: "Good test",
						ID:          "789",
						Name:        "Good 2",
					},
				},
				{
					{
						CreatedAt:   strfmt.DateTime(time.Now()),
						Description: "Good test",
						ID:          "012",
						Name:        "Page 2",
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
			project := mock_project_service.NewMockClientService(t)
			opts := &ListOpts{
				Ctx:     context.Background(),
				Profile: profile.TestProfile(t).SetOrgID("123"),
				IO:      io,
				Output:  format.New(io),
				Client:  project,
			}

			e := project.EXPECT()
			for i := 0; i < len(c.Resp) || c.RespErr; i++ {
				i := i
				call := e.ProjectServiceList(mock.MatchedBy(func(req *project_service.ProjectServiceListParams) bool {
					// Expect an org
					if req.ScopeID == nil || req.ScopeType == nil || *req.ScopeID != "123" ||
						*req.ScopeType != string(models.HashicorpCloudResourcemanagerResourceIDResourceTypeORGANIZATION) {
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
					call.Return(nil, project_service.NewProjectServiceListDefault(http.StatusForbidden))
					break
				} else {
					ok := project_service.NewProjectServiceListOK()
					ok.Payload = &models.HashicorpCloudResourcemanagerProjectListResponse{
						Projects: c.Resp[i],
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
					r.Contains(io.Output.String(), p.Name)
					r.Contains(io.Output.String(), p.Description)
				}
			}
		})
	}
}
