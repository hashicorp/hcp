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
		ShortHelp: "Manage HCP Vault Secrets Apps.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp vault-secrets secret" }} command group lets you
		manage HCP Vault Secrets application lifefcycle.
		`),
	}

	cmd.AddChild(NewCmdCreate(ctx, nil))
	return cmd
}