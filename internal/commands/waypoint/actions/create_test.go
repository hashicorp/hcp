// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package actions

import (
	"context"
	"errors"
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

func TestCmdCreate(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Name    string
		Args    []string
		Profile func(t *testing.T) *profile.Profile
		Setup   func(ws *mock_waypoint_service.MockClientService)
		Error   string
		Expect  *CreateOpts
	}{
		{
			Name:    "No org or project",
			Profile: profile.TestProfile,
			Args:    []string{"--name=foo"},
			Error:   "Organization ID and Project ID must be configured before running the command.",
		},
		{
			Name: "Custom action success",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123").SetProjectID("456")
			},
			Args: []string{"--name=foo", "--url=https://example.com", "--method=POST"},
			Expect: &CreateOpts{
				Name: "foo",
				Request: &models.HashicorpCloudWaypointActionConfigRequest{
					Custom: &models.HashicorpCloudWaypointActionConfigFlavorCustom{
						URL: "https://example.com",
					},
				},
				RequestCustomMethod: "POST",
			},
		},
		{
			Name: "Custom action with multiple headers",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123").SetProjectID("456")
			},
			Args: []string{
				"--name=foo",
				"--url=https://example.com",
				"--method=POST",
				"--header=X-First=abc",
				"--header=Second=123",
			},
			Expect: &CreateOpts{
				Name: "foo",
				Request: &models.HashicorpCloudWaypointActionConfigRequest{
					Custom: &models.HashicorpCloudWaypointActionConfigFlavorCustom{
						URL: "https://example.com",
						Headers: []*models.HashicorpCloudWaypointActionConfigFlavorCustomHeader{
							{Key: "X-First", Value: "abc"},
							{Key: "Second", Value: "123"},
						},
					},
				},
				RequestCustomMethod: "POST",
			},
		},
		{
			Name: "Agent action success",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123").SetProjectID("456")
			},
			Args: []string{"--name=foo", "--agent-group=bar", "--agent-operation=launch"},
			Expect: &CreateOpts{
				Name:           "foo",
				AgentGroup:     "bar",
				AgentOperation: "launch",
			},
		},
		{
			Name: "Mix of custom and agent options",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123").SetProjectID("456")
			},
			Args:  []string{"--name=foo", "--url=https://example.com", "--agent-group=bar"},
			Error: "cannot specify both custom action and agent action flags",
		},
		{
			Name: "API error",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123").SetProjectID("456")
			},
			Args:  []string{"--name=foo", "--url=https://example.com"},
			Error: "failed to create action \"foo\": api error",
			Setup: func(ws *mock_waypoint_service.MockClientService) {
				call := ws.EXPECT().WaypointServiceCreateActionConfig(mock.Anything, mock.Anything)
				call.Return(nil, errors.New("api error"))
			},
		},
		{
			Name: "Missing name",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123").SetProjectID("456")
			},
			Args:  []string{"--description=foo", "--url=https://example.com"},
			Error: "missing required flag: --name=NAME",
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

			var gotOpts CreateOpts
			cmd := NewCmdCreate(ctx, &gotOpts)
			cmd.SetIO(io)

			ws := mock_waypoint_service.NewMockClientService(t)
			gotOpts.WS2024Client = ws
			if c.Error == "" {
				call := ws.EXPECT().WaypointServiceCreateActionConfig(mock.Anything, mock.Anything)
				ok := waypoint_service.NewWaypointServiceCreateActionConfigOK()
				ok.Payload = &models.HashicorpCloudWaypointCreateActionConfigResponse{
					ActionConfig: &models.HashicorpCloudWaypointActionConfig{
						Name:        gotOpts.Name,
						Description: gotOpts.Description,
						Request:     gotOpts.Request,
					},
				}
				call.Return(ok, nil)
			}
			if c.Setup != nil {
				c.Setup(ws)
			}

			code := cmd.Run(c.Args)
			if c.Error != "" {
				r.NotZero(code)
				r.Contains(io.Error.String(), c.Error)
				return
			}

			r.Zero(code, io.Error.String())
			if c.Expect != nil {
				r.NotNil(gotOpts)
				r.Equal(c.Expect.Name, gotOpts.Name)
				if c.Expect.Description != "" {
					r.Equal(c.Expect.Description, gotOpts.Description)
				}
				if c.Expect.Request != nil {
					r.Equal(c.Expect.Request.Custom.URL, gotOpts.Request.Custom.URL)
					r.Equal(c.Expect.RequestCustomMethod, gotOpts.RequestCustomMethod)
					if c.Expect.Request.Custom.Headers != nil {
						r.ElementsMatch(c.Expect.Request.Custom.Headers, gotOpts.Request.Custom.Headers)
					}
				}
				if c.Expect.AgentGroup != "" {
					r.Equal(c.Expect.AgentGroup, gotOpts.AgentGroup)
				}
				if c.Expect.AgentOperation != "" {
					r.Equal(c.Expect.AgentOperation, gotOpts.AgentOperation)
				}
			}
		})
	}
}

func TestCreateAction(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Name      string
		Setup     func(*CreateOpts, *mock_waypoint_service.MockClientService)
		ExpectErr string
		ExpectOut string
	}{
		{
			Name: "Custom action success",
			Setup: func(opts *CreateOpts, ws *mock_waypoint_service.MockClientService) {
				opts.Name = "foo"
				opts.Description = "desc"
				opts.Request.Custom.URL = "https://example.com"
				opts.RequestCustomMethod = "POST"
				call := ws.EXPECT().WaypointServiceCreateActionConfig(mock.Anything, mock.Anything)
				ok := waypoint_service.NewWaypointServiceCreateActionConfigOK()
				ok.Payload = &models.HashicorpCloudWaypointCreateActionConfigResponse{
					ActionConfig: &models.HashicorpCloudWaypointActionConfig{
						Name:        opts.Name,
						Description: opts.Description,
						Request: &models.HashicorpCloudWaypointActionConfigRequest{
							Custom: &models.HashicorpCloudWaypointActionConfigFlavorCustom{
								URL: opts.Request.Custom.URL,
							},
						},
					},
				}
				call.Return(ok, nil)
			},
			ExpectErr: "",
			ExpectOut: "foo",
		},
		{
			Name: "Agent action success",
			Setup: func(opts *CreateOpts, ws *mock_waypoint_service.MockClientService) {
				opts.Name = "bar"
				opts.Description = "desc2"
				opts.AgentGroup = "group1"
				opts.AgentOperation = "op1"
				call := ws.EXPECT().WaypointServiceCreateActionConfig(mock.Anything, mock.Anything)
				ok := waypoint_service.NewWaypointServiceCreateActionConfigOK()
				ok.Payload = &models.HashicorpCloudWaypointCreateActionConfigResponse{
					ActionConfig: &models.HashicorpCloudWaypointActionConfig{
						Name:        opts.Name,
						Description: opts.Description,
						Request: &models.HashicorpCloudWaypointActionConfigRequest{
							Agent: &models.HashicorpCloudWaypointActionConfigFlavorAgent{
								Op: &models.HashicorpCloudWaypointAgentOperation{
									Group: opts.AgentGroup,
									ID:    opts.AgentOperation,
								},
							},
						},
					},
				}
				call.Return(ok, nil)
			},
			ExpectErr: "",
			ExpectOut: "bar",
		},
		{
			Name: "API error",
			Setup: func(opts *CreateOpts, ws *mock_waypoint_service.MockClientService) {
				opts.Name = "fail"
				opts.Request.Custom.URL = "https://fail.com"
				call := ws.EXPECT().WaypointServiceCreateActionConfig(mock.Anything, mock.Anything)
				call.Return(nil, errors.New("api error"))
			},
			ExpectErr: "failed to create action \"fail\": api error",
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			r := require.New(t)

			io := iostreams.Test()
			ws := mock_waypoint_service.NewMockClientService(t)
			opts := &CreateOpts{
				WaypointOpts: opts.WaypointOpts{
					Ctx:          context.Background(),
					Profile:      profile.TestProfile(t).SetOrgID("123").SetProjectID("456"),
					Output:       format.New(io),
					WS2024Client: ws,
				},
				Request: &models.HashicorpCloudWaypointActionConfigRequest{
					Custom: &models.HashicorpCloudWaypointActionConfigFlavorCustom{},
				},
			}
			c.Setup(opts, ws)

			err := createAction(nil, nil, opts)
			if c.ExpectErr != "" {
				r.Error(err)
				r.Contains(err.Error(), c.ExpectErr)
				return
			}
			r.NoError(err)
			if c.ExpectOut != "" {
				r.Contains(io.Output.String(), c.ExpectOut)
			}
		})
	}
}
