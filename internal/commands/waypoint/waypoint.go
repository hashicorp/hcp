// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package waypoint

import (
	addon "github.com/hashicorp/hcp/internal/commands/waypoint/add-ons"
	"github.com/hashicorp/hcp/internal/commands/waypoint/applications"
	"github.com/hashicorp/hcp/internal/commands/waypoint/templates"
	"github.com/hashicorp/hcp/internal/commands/waypoint/tfcconfig"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
)

func NewCmdWaypoint(ctx *cmd.Context) *cmd.Command {
	cmd := &cmd.Command{
		Name:      "waypoint",
		ShortHelp: "Manage HCP Waypoint.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp waypoint" }} command group lets you
		manage HCP Waypoint resources through the CLI. These commands let you to interact
		with their HCP Waypoint instance to manage their application deployment process.
		`),
	}

	cmd.AddChild(tfcconfig.NewCmdTFCConfig(ctx))
	// TODO: Enable later
	// cmd.AddChild(actions.NewCmdActionConfig(ctx))
	// cmd.AddChild(agent.NewCmdAgent(ctx))
	cmd.AddChild(templates.NewCmdTemplate(ctx))
	cmd.AddChild(addon.NewCmdAddOn(ctx))
	cmd.AddChild(applications.NewCmdApplications(ctx))

	return cmd
}
