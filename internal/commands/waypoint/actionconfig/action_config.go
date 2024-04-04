// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package actionconfig

import (
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
)

func NewCmdActionConfig(ctx *cmd.Context) *cmd.Command {
	cmd := &cmd.Command{
		Name:      "action-config",
		ShortHelp: "Manage action configuration options for HCP Waypoint.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp waypoint action-config" }} command
		group manages all action configuration options for HCP Waypoint. An action
		configuration is a set of options that define how an action is executed. This
		includes the action request type, and the action name. The action
		configuration is used to launch action runs depending on the Request type.
		`),
	}

	cmd.AddChild(NewCmdCreate(ctx))
	cmd.AddChild(NewCmdRead(ctx))
	cmd.AddChild(NewCmdUpdate(ctx))
	cmd.AddChild(NewCmdDelete(ctx))
	cmd.AddChild(NewCmdList(ctx))

	return cmd
}
