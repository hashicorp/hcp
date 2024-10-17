// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package templates

import (
	"github.com/hashicorp/hcp/internal/commands/waypoint/opts"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
)

type TemplateOpts struct {
	opts.WaypointOpts

	ID string

	// Name is the name of a new template, or the name of an existing template.
	// When used during update operations, it is the name of the template to be
	// updated.
	Name string

	// UpdatedName is used for updates, and is the new name for the template.
	UpdatedName                string
	Summary                    string
	Description                string
	ReadmeMarkdownTemplateFile string
	Labels                     []string
	Tags                       map[string]string

	TerraformNoCodeModuleSource string
	TerraformCloudProjectName   string
	TerraformCloudProjectID     string
	TerraformExecutionMode      string
	TerraformAgentPoolID        string

	VariableOptionsFile string

	// testFunc is used for testing, so that the command can be tested without
	// using the real API.
	testFunc func(c *cmd.Command, args []string) error
}

func NewCmdTemplate(ctx *cmd.Context) *cmd.Command {
	opts := &TemplateOpts{
		WaypointOpts: opts.New(ctx),
	}

	cmd := &cmd.Command{
		Name:      "templates",
		ShortHelp: "Manage HCP Waypoint templates.",
		LongHelp: heredoc.New(ctx.IO).Must(`
The {{ template "mdCodeOrBold" "hcp waypoint templates" }} command group lets you manage
HCP Waypoint templates. A template is a reusable configuration for creating
applications.
		`),
	}

	cmd.AddChild(NewCmdCreate(ctx, opts))
	cmd.AddChild(NewCmdDelete(ctx, opts))
	cmd.AddChild(NewCmdRead(ctx, opts))
	cmd.AddChild(NewCmdList(ctx, opts))
	cmd.AddChild(NewCmdUpdate(ctx, opts))

	return cmd
}
