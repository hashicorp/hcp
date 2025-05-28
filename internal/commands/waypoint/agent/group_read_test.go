// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package agent

import (
	"context"
	"testing"

	"github.com/go-openapi/runtime/client"
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

func TestCmdGroupRead(t *testing.T) {
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
			Args:    []string{"--name=foo"},
			Error:   "Organization ID and Project ID must be configured before running the command.",
		},
		{
			Name: "No name",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123").SetProjectID("456")
			},
			Args:  []string{},
			Error: "missing required flag: --name=NAME",
		},
		{
			Name: "Good",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123").SetProjectID("456")
			},
			Args: []string{"-n", "foo"},
			Expect: &GroupOpts{
				Name: "foo",
			},
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			r := require.New(t)

			io := iostreams.Test()
			ctx := &cmd.Context{
				IO:          io,
				Profile:     c.Profile(t),
				Output:      format.New(io),
				HCP:         &client.Runtime{},
				ShutdownCtx: context.Background(),
			}

			var gotOpts GroupOpts
			gotOpts.testFunc = func(c *cmd.Command, args []string) error {
				return nil
			}
			gotOpts.WS2024Client = mock_waypoint_service.NewMockClientService(t)
			cmd := NewCmdGroupRead(ctx, &gotOpts)
			cmd.SetIO(io)

			code := cmd.Run(c.Args)

			if c.Error != "" {
				r.NotZero(code)
				r.Contains(io.Error.String(), c.Error)
				return
			}

			r.Zero(code, io.Error.String())
			r.NotNil(gotOpts)
			if c.Expect != nil {
				r.Equal(c.Expect.Name, gotOpts.Name)
			}
		})
	}
}

func TestAgentGroupRead(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Name      string
		SetupMock func(ws *mock_waypoint_service.MockClientService)
		Opts      *GroupOpts
		Error     string
	}{
		{
			Name: "API error",
			SetupMock: func(ws *mock_waypoint_service.MockClientService) {
				ws.EXPECT().WaypointServiceGetAgentGroup(mock.Anything, mock.Anything).Return(nil, context.DeadlineExceeded)
			},
			Opts: &GroupOpts{
				Name: "foo",
				WaypointOpts: opts.WaypointOpts{
					Ctx:     context.Background(),
					Profile: profile.TestProfile(t).SetOrgID("123").SetProjectID("456"),
					IO:      iostreams.Test(),
					Output:  format.New(iostreams.Test()),
				},
			},
			Error: "failed to get agent group \"foo\"",
		},
		{
			Name: "Not found",
			SetupMock: func(ws *mock_waypoint_service.MockClientService) {
				ok := waypoint_service.NewWaypointServiceGetAgentGroupOK()
				ok.Payload = &models.HashicorpCloudWaypointGetAgentGroupResponse{Group: nil}
				ws.EXPECT().WaypointServiceGetAgentGroup(mock.Anything, mock.Anything).Return(ok, nil)
			},
			Opts: &GroupOpts{
				Name: "foo",
				WaypointOpts: opts.WaypointOpts{
					Ctx:     context.Background(),
					Profile: profile.TestProfile(t).SetOrgID("123").SetProjectID("456"),
					IO:      iostreams.Test(),
					Output:  format.New(iostreams.Test()),
				},
			},
			Error: "no group found with name \"foo\"",
		},
		{
			Name: "Success",
			SetupMock: func(ws *mock_waypoint_service.MockClientService) {
				ok := waypoint_service.NewWaypointServiceGetAgentGroupOK()
				ok.Payload = &models.HashicorpCloudWaypointGetAgentGroupResponse{
					Group: &models.HashicorpCloudWaypointAgentGroup{
						Name:        "foo",
						Description: "desc",
					},
				}
				ws.EXPECT().WaypointServiceGetAgentGroup(mock.Anything, mock.Anything).Return(ok, nil)
			},
			Opts: &GroupOpts{
				Name: "foo",
				WaypointOpts: opts.WaypointOpts{
					Ctx:     context.Background(),
					Profile: profile.TestProfile(t).SetOrgID("123").SetProjectID("456"),
					IO:      iostreams.Test(),
					Output:  format.New(iostreams.Test()),
				},
			},
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			r := require.New(t)

			ws := mock_waypoint_service.NewMockClientService(t)
			if c.SetupMock != nil {
				c.SetupMock(ws)
			}
			c.Opts.WS2024Client = ws

			err := agentGroupRead(c.Opts)
			if c.Error != "" {
				r.ErrorContains(err, c.Error)
				return
			}
			r.NoError(err)
		})
	}
}
