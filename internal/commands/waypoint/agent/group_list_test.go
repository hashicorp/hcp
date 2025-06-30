// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package agent

import (
	"context"
	"net/http"
	"testing"

	"github.com/go-openapi/runtime/client"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2024-11-22/client/waypoint_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2024-11-22/models"
	"github.com/hashicorp/hcp/internal/commands/waypoint/opts"
	mock_waypoint_service "github.com/hashicorp/hcp/internal/pkg/api/mocks/github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2024-11-22/client/waypoint_service"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCmdGroupList(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Name    string
		Args    []string
		Profile func(t *testing.T) *profile.Profile
		Error   string
		Expect  *GroupOpts
	}{
		{
			Name:    "No org or project",
			Profile: profile.TestProfile,
			Args:    []string{},
			Error:   "Organization ID and Project ID must be configured before running the command.",
		},
		{
			Name: "Good",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123").SetProjectID("456")
			},
			Args: []string{},
			Expect: &GroupOpts{
				WaypointOpts: opts.WaypointOpts{
					Profile: profile.TestProfile(t).SetOrgID("123").SetProjectID("456"),
				},
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

			var gotOpts GroupOpts
			cmd := NewCmdGroupList(ctx, &gotOpts)
			cmd.SetIO(io)

			if c.Error == "" {
				// Mock the WaypointServiceListAgentGroups call
				ws := mock_waypoint_service.NewMockClientService(t)
				gotOpts.WS2024Client = ws

				call := ws.EXPECT().WaypointServiceListAgentGroups(mock.Anything, mock.Anything)

				ok := waypoint_service.NewWaypointServiceListAgentGroupsOK()
				ok.Payload = &models.HashicorpCloudWaypointV20241122ListAgentGroupsResponse{
					Groups: []*models.HashicorpCloudWaypointV20241122AgentGroup{
						{
							Name:        "test-group",
							Description: "Test Group",
						},
					},
				}
				call.Return(ok, nil)
			}

			code := cmd.Run(c.Args)
			if c.Error != "" {
				r.NotZero(code)
				r.Contains(io.Error.String(), c.Error)
				return
			}

			r.Zero(code, io.Error.String())
			r.NotNil(gotOpts)
			if c.Expect != nil {
				r.Equal(c.Expect.Profile.OrganizationID, gotOpts.Profile.OrganizationID)
				r.Equal(c.Expect.Profile.ProjectID, gotOpts.Profile.ProjectID)
			}
		})
	}
}

func TestGroupListRun(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Name    string
		Resp    []*models.HashicorpCloudWaypointV20241122AgentGroup
		RespErr bool
		Error   string
	}{
		{
			Name:    "Server error",
			RespErr: true,
			Error:   "error listing groups: [GET /waypoint/2024-11-22/organizations/{namespace.location.organization_id}/projects/{namespace.location.project_id}/agent/group][403]",
		},
		{
			Name: "Good empty",
			Resp: []*models.HashicorpCloudWaypointV20241122AgentGroup{},
		},
		{
			Name: "Good",
			Resp: []*models.HashicorpCloudWaypointV20241122AgentGroup{
				{
					Name:        "test-group",
					Description: "Test Group",
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
			ws := mock_waypoint_service.NewMockClientService(t)
			opts := &GroupOpts{
				WaypointOpts: opts.WaypointOpts{
					Ctx:          context.Background(),
					Profile:      profile.TestProfile(t).SetOrgID("123").SetProjectID("456"),
					Output:       format.New(io),
					WS2024Client: ws,
				},
			}

			call := ws.EXPECT().WaypointServiceListAgentGroups(mock.Anything, mock.Anything)

			if c.RespErr {
				call.Return(nil, waypoint_service.NewWaypointServiceListAgentGroupsDefault(http.StatusForbidden))
			} else {
				ok := waypoint_service.NewWaypointServiceListAgentGroupsOK()
				ok.Payload = &models.HashicorpCloudWaypointV20241122ListAgentGroupsResponse{
					Groups: c.Resp,
				}
				call.Return(ok, nil)
			}

			err := agentGroupList(hclog.NewNullLogger(), opts)
			if c.Error != "" {
				r.ErrorContains(err, c.Error)
				return
			}

			r.NoError(err)

			for _, group := range c.Resp {
				r.Contains(io.Output.String(), group.Name)
				r.Contains(io.Output.String(), group.Description)
			}
		})
	}
}
