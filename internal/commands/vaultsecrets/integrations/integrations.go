// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: MPL-2.0

package integrations

import (
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
)

type IntegrationType string

const (
	Twilio       IntegrationType = "twilio"
	MongoDBAtlas IntegrationType = "mongodb-atlas"
	AWS          IntegrationType = "aws"
	GCP          IntegrationType = "gcp"
	Postgres     IntegrationType = "postgres"
)

func NewCmdIntegrations(ctx *cmd.Context) *cmd.Command {
	cmd := &cmd.Command{
		Name:      "integrations",
		ShortHelp: "Manage Vault Secrets integrations.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp vault-secrets integrations" }} command group lets you
		manage Vault Secrets integrations.
		`),
		PersistentPreRun: func(c *cmd.Command, args []string) error {
			return cmd.RequireOrgAndProject(ctx)
		},
	}

	cmd.AddChild(NewCmdRead(ctx, nil))
	cmd.AddChild(NewCmdDelete(ctx, nil))
	cmd.AddChild(NewCmdList(ctx, nil))
	cmd.AddChild(NewCmdCreate(ctx, nil))
	cmd.AddChild(NewCmdUpdate(ctx, nil))
	return cmd
}
