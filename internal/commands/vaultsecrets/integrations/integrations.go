// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package integrations

import (
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
)

func NewCmdIntegrations(ctx *cmd.Context) *cmd.Command {
	cmd := &cmd.Command{
		Name:      "integrations",
		ShortHelp: "Manage Vault Secrets integrations.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp vault-secrets integrations" }} command group lets you
		manage Vault Secrets integrations.
		`),
		PersistentPreRun: func(c *cmd.Command, args []string) error {
			return cmd.RequireOrgAndProject(ctx)
		},
	}

	cmd.AddChild(NewCmdCreate(ctx, nil))
	return cmd
}
