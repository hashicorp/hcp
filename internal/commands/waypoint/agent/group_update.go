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

func NewCmdGroupUpdate(ctx *cmd.Context, opts *GroupOpts) *cmd.Command {
	cmd := &cmd.Command{
		Name:      "update",
		ShortHelp: "Update an HCP Waypoint Agent group.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp waypoint agent group update" }} command updates an existing Agent group.
		`),
		Flags: cmd.Flags{
			Local: []*cmd.Flag{
				{
					Name:         "name",
					Shorthand:    "n",
					DisplayValue: "NAME",
					Description:  "Name of the group to update.",
					Value:        flagvalue.Simple("", &opts.Name),
					Required:     true,
				},
				{
					Name:         "description",
					Shorthand:    "d",
					DisplayValue: "DESCRIPTION",
					Description:  "New description for the group.",
					Value:        flagvalue.Simple("", &opts.Description),
				},
			},
		},
		Examples: []cmd.Example{
			{
				Preamble: "Update a group's description:",
				Command:  "$ hcp waypoint agent group update -n='prod:us-west-2' -d='updated description'",
			},
		},
		PersistentPreRun: func(c *cmd.Command, args []string) error {
			return cmd.RequireOrgAndProject(ctx)
		},
		RunF: func(c *cmd.Command, args []string) error {
			if opts.testFunc != nil {
				return opts.testFunc(c, args)
			}
			return agentGroupUpdate(c.Logger(), opts)
		},
	}

	return cmd
}

func agentGroupUpdate(log hclog.Logger, opts *GroupOpts) error {
	ctx := opts.Ctx

	update := &models.HashicorpCloudWaypointV20241122WaypointServiceUpdateAgentGroupBody{
		Description: opts.Description,
	}

	_, err := opts.WS2024Client.WaypointServiceUpdateAgentGroup(&waypoint_service.WaypointServiceUpdateAgentGroupParams{
		Name:                            opts.Name,
		NamespaceLocationOrganizationID: opts.Profile.OrganizationID,
		NamespaceLocationProjectID:      opts.Profile.ProjectID,
		Body:                            update,
		Context:                         ctx,
	}, nil)

	if err != nil {
		return fmt.Errorf("error updating group: %w", err)
	}

	fmt.Fprintf(opts.IO.Err(), "%s Group %q updated\n",
		opts.IO.ColorScheme().SuccessIcon(),
		opts.Name,
	)
	return nil
}
