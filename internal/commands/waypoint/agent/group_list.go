package agent

import (
	"fmt"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2023-08-18/client/waypoint_service"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/pkg/errors"
)

func NewCmdGroupList(ctx *cmd.Context, opts *GroupOpts) *cmd.Command {
	cmd := &cmd.Command{
		Name:      "list",
		ShortHelp: "List HCP Waypoint Agent groups.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ Bold "hcp waypoint agent group list" }} command lists groups registered.
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
	resp, err := opts.WS.WaypointServiceGetNamespace(&waypoint_service.WaypointServiceGetNamespaceParams{
		LocationOrganizationID: opts.Profile.OrganizationID,
		LocationProjectID:      opts.Profile.ProjectID,
		Context:                opts.Ctx,
	}, nil)
	if err != nil {
		return errors.Wrapf(err, "Unable to access HCP project")
	}

	ns := resp.Payload.Namespace

	ctx := opts.Ctx

	list, err := opts.WS.WaypointServiceListAgentGroups(&waypoint_service.WaypointServiceListAgentGroupsParams{
		NamespaceID: ns.ID,
		Context:     ctx,
	}, nil)

	if err != nil {
		return fmt.Errorf("error listing groups: %w", err)
	}

	return opts.Output.Show(list.Payload.Groups, format.Table, "name", "description")
}
