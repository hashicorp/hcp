// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package agent

import (
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
)

func NewCmdAgent(ctx *cmd.Context) *cmd.Command {
	cmd := &cmd.Command{
		Name:      "agent",
		ShortHelp: "Run and manage a Waypoint Agent.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp waypoint agent" }} command group lets you 
		run and manage a local Waypoint agent.

		Agents are used in conjunction with HCP Waypoint Actions to allow actions to run on your
		own systems when initiated from HCP Waypoint.
		`),
	}

	cmd.AddChild(NewCmdRun(ctx))
	cmd.AddChild(NewCmdQueue(ctx))
	cmd.AddChild(NewCmdGroup(ctx))
	return cmd
}
