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
		ShortHelp: "Manage an organization's IAM policy.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp organizations iam" }} command group is used
		to manage an organization's IAM Policy.
		`),
	}

	cmd.AddChild(NewCmdAddBinding(ctx, nil))
	cmd.AddChild(NewCmdDeleteBinding(ctx, nil))
	cmd.AddChild(NewCmdReadPolicy(ctx, nil))
	cmd.AddChild(NewCmdSetPolicy(ctx, nil))
	return cmd
}
