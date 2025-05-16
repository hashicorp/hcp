// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package actions

import (
	"github.com/hashicorp/hcp/internal/commands/waypoint/opts"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
)

func NewCmdActionConfig(ctx *cmd.Context) *cmd.Command {
	cmd := &cmd.Command{
		Name:      "actions",
		ShortHelp: "Manage action configuration options for HCP Waypoint.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp waypoint actions" }} command group
		manages all action options for HCP Waypoint. An action is a set of
		options that define how an action is executed. This includes the action
		request type, and the action name. The action is used to launch action
		runs depending on the Request type.
		`),
	}

	cmd.AddChild(NewCmdCreate(ctx, &CreateOpts{WaypointOpts: opts.New(ctx)}))
	cmd.AddChild(NewCmdRead(ctx))
	cmd.AddChild(NewCmdUpdate(ctx))
	cmd.AddChild(NewCmdDelete(ctx))
	cmd.AddChild(NewCmdList(ctx))

	return cmd
}
