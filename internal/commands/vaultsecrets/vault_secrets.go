// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vaultsecrets

import (
	"github.com/hashicorp/hcp/internal/commands/vaultsecrets/apps"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
)

func NewCmdVaultSecrets(ctx *cmd.Context) *cmd.Command {
	cmd := &cmd.Command{
		Name:      "vault-secrets",
		ShortHelp: "Manage Vault Secrets.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp vault-secrets" }} command group lets you manage Vault Secrets
		resources through the CLI (in Beta).
		`),
		Aliases: []string{
			"vs",
		},
	}

	cmd.AddChild(apps.NewCmdApps(ctx))
	return cmd
}
