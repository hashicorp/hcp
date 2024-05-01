// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package secret

import (
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
)

func NewCmdSecret(ctx *cmd.Context) *cmd.Command {
	cmd := &cmd.Command{
		Name:      "secret",
		ShortHelp: "Manage HCP Vault Secrets App secrets.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp vault-secrets secret" }} command group lets you
		manage secrets lifecycle under HCP Vault Secrets applications.
		`),
	}

	return cmd
}
