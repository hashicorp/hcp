// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: MPL-2.0

package keys

import (
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
)

func NewCmdKeys(ctx *cmd.Context) *cmd.Command {
	cmd := &cmd.Command{
		Name:      "keys",
		ShortHelp: "Create and manage service principals keys.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp iam service-principals keys" }} command group lets you create
		and manage service principals keys.

		A service principal key is the credential used by a service principal to authenticate with HCP
		and should be treated as a secret.
		`),
	}

	cmd.AddChild(NewCmdCreate(ctx, nil))
	cmd.AddChild(NewCmdList(ctx, nil))
	cmd.AddChild(NewCmdDelete(ctx, nil))

	return cmd
}
