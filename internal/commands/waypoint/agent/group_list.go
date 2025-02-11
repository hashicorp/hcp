// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package agent

import (
	"fmt"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2024-11-22/client/waypoint_service"

	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
)

func NewCmdGroupList(ctx *cmd.Context, opts *GroupOpts) *cmd.Command {
	cmd := &cmd.Command{
		Name:      "list",
		ShortHelp: "List HCP Waypoint Agent groups.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp waypoint agent group list" }} command lists groups registered.
		`),
		Examples: []cmd.Example{
			{
				Preamble: "List all groups:",
				Command:  "$ hcp waypoint agent group list",
			},
		},
		PersistentPreRun: func(c *cmd.Command, args []string) error {
			return cmd.RequireOrgAndProject(ctx)
		},
		RunF: func(c *cmd.Command, args []string) error {
			return agentGroupList(c.Logger(), opts)
		},
	}

	return cmd
}

func agentGroupList(log hclog.Logger, opts *GroupOpts) error {
	ctx := opts.Ctx

	list, err := opts.WS2024Client.WaypointServiceListAgentGroups(&waypoint_service.WaypointServiceListAgentGroupsParams{
		NamespaceLocationOrganizationID: opts.Profile.OrganizationID,
		NamespaceLocationProjectID:      opts.Profile.ProjectID,
		Context:                         ctx,
	}, nil)

	if err != nil {
		return fmt.Errorf("error listing groups: %w", err)
	}

	return opts.Output.Show(list.Payload.Groups, format.Table, "Name", "Description")
}
