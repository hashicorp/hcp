// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package projects

import (
	"github.com/hashicorp/hcp/internal/commands/projects/iam"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
)

var (
	// projectFields is the set of fields to display for a project.
	projectFields = []format.Field{
		format.NewField("Name", "{{ .Name }}"),
		format.NewField("ID", "{{ .ID }}"),
		format.NewField("Description", "{{ .Description }}"),
		format.NewField("Created At", "{{ .CreatedAt }}"),
	}
)

func NewCmdProjects(ctx *cmd.Context) *cmd.Command {
	cmd := &cmd.Command{
		Name:      "projects",
		ShortHelp: "Create and manage projects.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp projects" }} command group lets you create a new
		HCP Project, view existing projects, and manage access to a project.

		A principal can be granted access to a project by using {{ template "mdCodeOrBold" "hcp projects iam add-binding" }}.

		To view the IAM Policy for the project, run {{ template "mdCodeOrBold" "hcp projects iam read-policy" }}.

		To set a project as the default project for the active profile,
		run {{ template "mdCodeOrBold" "hcp profile set project_id PROJECT_ID" }}.
		`),
	}

	cmd.AddChild(NewCmdCreate(ctx, nil))
	cmd.AddChild(NewCmdRead(ctx, nil))
	cmd.AddChild(NewCmdList(ctx, nil))
	cmd.AddChild(NewCmdDelete(ctx, nil))
	cmd.AddChild(NewCmdUpdate(ctx, nil))
	cmd.AddChild(iam.NewCmdIAM(ctx))
	return cmd
}
