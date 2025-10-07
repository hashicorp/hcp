// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
)

func NewCmdIAM(ctx *cmd.Context) *cmd.Command {
	cmd := &cmd.Command{
		Name:      "iam",
		ShortHelp: "Manage a group's IAM policy.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp iam groups iam" }} command group lets you manage a group's IAM Policy.
		`),
		Examples: []cmd.Example{
			{
				Preamble: heredoc.New(ctx.IO).Must(`To set a member as a group manager, you can use the {{ template "mdCodeOrBold" "add-binding" }} subcommand with the {{ template "mdCodeOrBold" "roles/iam.group-manager" }} role:`),
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				$ hcp iam groups iam add-binding \
				  --group=Group-Name \
				  --member=ef938a22-09cf-4be9-b4d0-1f4587f80f53 \
				  --role=roles/iam.group-manager
				`),
			},
		},
	}

	cmd.AddChild(NewCmdAddBinding(ctx, nil))
	cmd.AddChild(NewCmdDeleteBinding(ctx, nil))
	cmd.AddChild(NewCmdReadPolicy(ctx, nil))
	cmd.AddChild(NewCmdSetPolicy(ctx, nil))
	return cmd
}
