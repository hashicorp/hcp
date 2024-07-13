// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package gateway_pools

import (
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
)

func NewCmdGatewayPools(ctx *cmd.Context) *cmd.Command {
	command := &cmd.Command{
		Name:      "gateway-pools",
		ShortHelp: "Manage Vault Secrets gateway pools.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp vault-secrets gateway-pools" }} command group lets you
		manage Vault Secrets gateway pools.
		`),
		PersistentPreRun: func(c *cmd.Command, args []string) error {
			return cmd.RequireOrgAndProject(ctx)
		},
	}

	command.AddChild(NewCmdList(ctx, nil))
	command.AddChild(NewCmdCreate(ctx, nil))
	command.AddChild(NewCmdUpdate(ctx, nil))
	command.AddChild(NewCmdDelete(ctx, nil))
	command.AddChild(NewCmdRead(ctx, nil))
	return command
}
