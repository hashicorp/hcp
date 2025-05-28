// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package agent

import (
	"fmt"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2024-11-22/client/waypoint_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2024-11-22/models"

	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/flagvalue"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
)

func NewCmdGroupCreate(ctx *cmd.Context, opts *GroupOpts) *cmd.Command {
	cmd := &cmd.Command{
		Name:      "create",
		ShortHelp: "Create a new HCP Waypoint Agent group.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp waypoint agent group create" }} command creates a new Agent group.
		`),
		Flags: cmd.Flags{
			Local: []*cmd.Flag{
				{
					Name:         "name",
					Shorthand:    "n",
					DisplayValue: "NAME",
					Description:  "Name for the new group.",
					Value:        flagvalue.Simple("", &opts.Name),
					Required:     true,
				},
				{
					Name:         "description",
					Shorthand:    "d",
					DisplayValue: "DESCRIPTION",
					Description:  "Description for the group.",
					Value:        flagvalue.Simple("", &opts.Description),
				},
			},
		},
		Examples: []cmd.Example{
			{
				Preamble: "Create a new group:",
				Command:  "$ hcp waypoint agent group create -n='prod:us-west-2' -d='us west production access'",
			},
		},
		PersistentPreRun: func(c *cmd.Command, args []string) error {
			return cmd.RequireOrgAndProject(ctx)
		},
		RunF: func(c *cmd.Command, args []string) error {
			if opts.testFunc != nil {
				return opts.testFunc(c, args)
			}
			return agentGroupCreate(c.Logger(), opts)
		},
	}

	return cmd
}

func agentGroupCreate(log hclog.Logger, opts *GroupOpts) error {
	ctx := opts.Ctx

	grp := &models.HashicorpCloudWaypointAgentGroup{
		Description: opts.Description,
		Name:        opts.Name,
	}
	_, err := opts.WS2024Client.WaypointServiceCreateAgentGroup(&waypoint_service.WaypointServiceCreateAgentGroupParams{
		NamespaceLocationOrganizationID: opts.Profile.OrganizationID,
		NamespaceLocationProjectID:      opts.Profile.ProjectID,
		Body: &models.HashicorpCloudWaypointV20241122WaypointServiceCreateAgentGroupBody{
			Group: grp,
		},
		Context: ctx,
	}, nil)

	if err != nil {
		return fmt.Errorf("error creating group: %w", err)
	}

	fmt.Fprintf(opts.IO.Err(), "%s Group %q created\n",
		opts.IO.ColorScheme().SuccessIcon(),
		opts.Name,
	)
	return nil
}
