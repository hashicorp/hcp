// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apps

import (
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
)

func NewCmdApps(ctx *cmd.Context) *cmd.Command {
	cmd := &cmd.Command{
		Name:      "apps",
		ShortHelp: "Manage Vault Secrets apps.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp vault-secrets apps" }} command group lets you
		manage Vault Secrets applications.
		`),
		PersistentPreRun: func(c *cmd.Command, args []string) error {
			return cmd.RequireOrgAndProject(ctx)
		},
	}

	cmd.AddChild(NewCmdCreate(ctx, nil))
	cmd.AddChild(NewCmdDelete(ctx, nil))
	cmd.AddChild(NewCmdRead(ctx, nil))
	cmd.AddChild(NewCmdList(ctx, nil))
	cmd.AddChild(NewCmdUpdate(ctx, nil))
	return cmd
}
