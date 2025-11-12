// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: MPL-2.0

package versions

import (
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
)

func NewCmdVersions(ctx *cmd.Context) *cmd.Command {
	cmd := &cmd.Command{
		Name:      "versions",
		ShortHelp: "Manage Vault Secrets application secret's versions.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp vault-secrets secrets versions" }} command group lets you
		manage a secret's versions.
		`),
	}

	cmd.AddChild(NewCmdList(ctx, nil))
	return cmd
}
