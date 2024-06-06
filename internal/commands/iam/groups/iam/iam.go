// Copyright (c) HashiCorp, Inc.
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
		LongHelp: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
		The {{ template "mdCodeOrBold" "hcp iam groups iam" }} command group lets you manage a group's IAM Policy.

		To set a member as a group manager, you can use the {{ template "mdCodeOrBold" "add-binding" }} subcommand with the {{ template "mdCodeOrBold" "roles/iam.group-manager" }} role.

		$ hcp iam groups iam add-binding \
			--group=8647ae06-ca65-467a-b72d-edba1f908fc8 \
			--member=ef938a22-09cf-4be9-b4d0-1f4587f80f53 \
			--role=roles/iam.group-manager

		`),
	}

	// TODO: Uncomment as subcommands are added
	cmd.AddChild(NewCmdAddBinding(ctx, nil))
	// cmd.AddChild(NewCmdDeleteBinding(ctx, nil))
	cmd.AddChild(NewCmdReadPolicy(ctx, nil))
	// cmd.AddChild(NewCmdSetPolicy(ctx, nil))
	return cmd
}
