// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package agent

import (
	"fmt"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2024-11-22/client/waypoint_service"

	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/flagvalue"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
)

func NewCmdGroupDelete(ctx *cmd.Context, opts *GroupOpts) *cmd.Command {
	cmd := &cmd.Command{
		Name:      "delete",
		ShortHelp: "Delete a HCP Waypoint Agent group.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp waypoint agent group delete" }} command deletes an Agent group.
		`),
		Flags: cmd.Flags{
			Local: []*cmd.Flag{
				{
					Name:         "name",
					Shorthand:    "n",
					DisplayValue: "NAME",
					Description:  "Name of the group to delete.",
					Value:        flagvalue.Simple("", &opts.Name),
					Required:     true,
				},
			},
		},
		Examples: []cmd.Example{
			{
				Preamble: "Delete a group:",
				Command:  "$ hcp waypoint agent group delete -n='prod:us-west-2'",
			},
		},
		PersistentPreRun: func(c *cmd.Command, args []string) error {
			return cmd.RequireOrgAndProject(ctx)
		},
		RunF: func(c *cmd.Command, args []string) error {
			return agentGroupDelete(c.Logger(), opts)
		},
	}

	return cmd
}

func agentGroupDelete(log hclog.Logger, opts *GroupOpts) error {
	ctx := opts.Ctx

	_, err := opts.WS2024Client.WaypointServiceDeleteAgentGroup(&waypoint_service.WaypointServiceDeleteAgentGroupParams{
		Name:                            opts.Name,
		NamespaceLocationOrganizationID: opts.Profile.OrganizationID,
		NamespaceLocationProjectID:      opts.Profile.ProjectID,
		Context:                         ctx,
	}, nil)

	if err != nil {
		return fmt.Errorf("error deleting group: %w", err)
	}

	fmt.Fprintf(opts.IO.Err(), "Group deleted\n")
	return nil
}
