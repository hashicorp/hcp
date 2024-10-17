// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package applications

import (
	"github.com/hashicorp/hcp/internal/commands/waypoint/opts"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
)

type ApplicationOpts struct {
	opts.WaypointOpts

	Name               string
	TemplateName       string
	ActionConfigNames  []string
	ReadmeMarkdownFile string

	Variables     map[string]string
	VariablesFile string

	testFunc func(c *cmd.Command, args []string) error
}

func NewCmdApplications(ctx *cmd.Context) *cmd.Command {
	opts := &ApplicationOpts{
		WaypointOpts: opts.New(ctx),
	}

	cmd := &cmd.Command{
		Name:      "applications",
		ShortHelp: "Manage HCP Waypoint applications.",
		LongHelp: heredoc.New(ctx.IO).Must(`
The {{ template "mdCodeOrBold" "hcp waypoint applications" }} command group lets you manage
HCP Waypoint applications.
		`),
	}

	cmd.AddChild(NewCmdApplicationsCreate(ctx, opts))
	cmd.AddChild(NewCmdApplicationsDestroy(ctx, opts))
	cmd.AddChild(NewCmdApplicationsList(ctx, opts))
	cmd.AddChild(NewCmdApplicationsRead(ctx, opts))
	cmd.AddChild(NewCmdApplicationsUpdate(ctx, opts))

	return cmd
}
