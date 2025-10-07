// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: MPL-2.0

package organizations

import (
	"github.com/hashicorp/hcp/internal/commands/organizations/iam"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
)

func NewCmdOrganizations(ctx *cmd.Context) *cmd.Command {
	cmd := &cmd.Command{
		Name:      "organizations",
		Aliases:   []string{"orgs"},
		ShortHelp: "Interact with an existing organization.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp organizations" }} command group lets you
		interact with an existing HCP organization.
		`),
	}

	cmd.AddChild(NewCmdRead(ctx, nil))
	cmd.AddChild(NewCmdList(ctx, nil))
	cmd.AddChild(iam.NewCmdIAM(ctx))
	return cmd
}
