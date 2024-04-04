// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package definitions

import (
	"github.com/hashicorp/hcp/internal/commands/waypoint/opts"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
)

type AddOnDefinitionOpts struct {
	opts.WaypointOpts

	Name                       string
	Summary                    string
	Description                string
	ReadmeMarkdownTemplateFile string
	Labels                     []string

	TerraformNoCodeModuleSource  string
	TerraformNoCodeModuleVersion string
	TerraformCloudProjectName    string
	TerraformCloudProjectID      string

	// testFunc is used for testing, so that the command can be tested without
	// using the real API.
	testFunc func(c *cmd.Command, args []string) error
}

func NewCmdAddOnDefinition(ctx *cmd.Context) *cmd.Command {
	opts := &AddOnDefinitionOpts{
		WaypointOpts: opts.New(ctx),
	}

	cmd := &cmd.Command{
		Name:      "definitions",
		ShortHelp: "Manage HCP Waypoint add-on definitions.",
		LongHelp: heredoc.New(ctx.IO).Must(`
The {{ template "mdCodeOrBold" "hcp waypoint add-ons definitions" }} command
group lets you manage HCP Waypoint add-on definitions.
`),
	}

	cmd.AddChild(NewCmdAddOnDefinitionCreate(ctx, opts))
	cmd.AddChild(NewCmdAddOnDefinitionDelete(ctx, opts))
	cmd.AddChild(NewCmdAddOnDefinitionList(ctx, opts))
	cmd.AddChild(NewCmdAddOnDefinitionRead(ctx, opts))
	cmd.AddChild(NewCmdAddOnDefinitionUpdate(ctx, opts))

	return cmd
}
