// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: MPL-2.0

package agent

import (
	"fmt"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2024-11-22/client/waypoint_service"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/flagvalue"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/pkg/errors"
)

func NewCmdGroupRead(ctx *cmd.Context, opts *GroupOpts) *cmd.Command {
	cmd := &cmd.Command{
		Name:      "read",
		ShortHelp: "Read details about a Waypoint Agent group.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp waypoint agent group read" }} command reads details about a Waypoint Agent group.
		`),
		Examples: []cmd.Example{
			{
				Preamble: "Read a group by name:",
				Command:  "$ hcp waypoint agent group read -n='prod:us-west-2'",
			},
		},
		Flags: cmd.Flags{
			Local: []*cmd.Flag{
				{
					Name:         "name",
					Shorthand:    "n",
					DisplayValue: "NAME",
					Description:  "Name of the group to read.",
					Value:        flagvalue.Simple("", &opts.Name),
					Required:     true,
				},
			},
		},
		PersistentPreRun: func(c *cmd.Command, args []string) error {
			return cmd.RequireOrgAndProject(ctx)
		},
		RunF: func(c *cmd.Command, args []string) error {
			if opts.testFunc != nil {
				return opts.testFunc(c, args)
			}
			return agentGroupRead(opts)
		},
	}
	return cmd
}

func agentGroupRead(opts *GroupOpts) error {
	resp, err := opts.WS2024Client.WaypointServiceGetAgentGroup(&waypoint_service.WaypointServiceGetAgentGroupParams{
		Name:                            opts.Name,
		NamespaceLocationOrganizationID: opts.Profile.OrganizationID,
		NamespaceLocationProjectID:      opts.Profile.ProjectID,
		Context:                         opts.Ctx,
	}, nil)
	if err != nil {
		return errors.Wrapf(err, "failed to get agent group %q", opts.Name)
	}

	if resp.GetPayload() == nil || resp.GetPayload().Group == nil {
		return fmt.Errorf("no group found with name %q", opts.Name)
	}
	group := resp.GetPayload().Group

	fields := []format.Field{
		format.NewField("Name", "{{ .Name }}"),
		format.NewField("Description", "{{ .Description }}"),
	}
	d := format.NewDisplayer(group, format.Pretty, fields)
	return opts.Output.Display(d)
}
