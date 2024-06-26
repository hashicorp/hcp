// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package auth

import (
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
)

func NewCmdAuth(ctx *cmd.Context) *cmd.Command {
	cmd := &cmd.Command{
		Name:      "auth",
		ShortHelp: "Authenticate to HCP.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp auth" }} command group lets you manage authentication to HCP.
		`),
		Examples: []cmd.Example{
			{
				Preamble: "Login interactively using a browser:",
				Command:  "$ hcp auth login",
			},
			{
				Preamble: "Login using service principal credentials:",
				Command:  "$ hcp auth login --client-id=spID --client-secret=spSecret",
			},
			{
				Preamble: "Logout the CLI:",
				Command:  "$ hcp auth logout",
			},
		},
	}

	cmd.AddChild(NewCmdLogin(ctx))
	cmd.AddChild(NewCmdLogout(ctx))
	cmd.AddChild(NewCmdPrintAccessToken(ctx))
	return cmd
}
