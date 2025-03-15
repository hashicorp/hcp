// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package agent

import (
	"fmt"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2024-11-22/client/waypoint_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2024-11-22/models"
	wpopts "github.com/hashicorp/hcp/internal/commands/waypoint/opts"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
)

func NewCmdGroupList(ctx *cmd.Context, opts *GroupOpts) *cmd.Command {
	opts.WaypointOpts = wpopts.New(ctx)

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
	var groups groupsDisplayer

	list, err := opts.WS2024Client.WaypointServiceListAgentGroups(&waypoint_service.WaypointServiceListAgentGroupsParams{
		Context:                         opts.Ctx,
		NamespaceLocationOrganizationID: opts.Profile.OrganizationID,
		NamespaceLocationProjectID:      opts.Profile.ProjectID,
	}, nil)

	if err != nil {
		return fmt.Errorf("error listing groups: %w", err)
	}

	groups = append(groups, list.Payload.Groups...)

	return opts.Output.Display(groups)
}

type groupsDisplayer []*models.HashicorpCloudWaypointAgentGroup

func (d groupsDisplayer) DefaultFormat() format.Format {
	return format.Table
}

func (d groupsDisplayer) Payload() any {
	return d
}

func (d groupsDisplayer) FieldTemplates() []format.Field {
	return []format.Field{
		{
			Name:        "Name",
			ValueFormat: "{{ .Name }}",
		},
		{
			Name:        "Description",
			ValueFormat: "{{ .Description }}",
		},
	}
}
