// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package secrets

import (
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-11-28/client/secret_service"
	"github.com/hashicorp/hcp/internal/commands/vaultsecrets/apps/helper"
	"github.com/hashicorp/hcp/internal/commands/vaultsecrets/secrets/appname"
	"github.com/hashicorp/hcp/internal/commands/vaultsecrets/secrets/versions"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
)

func NewCmdSecrets(ctx *cmd.Context) *cmd.Command {
	cmd := &cmd.Command{
		Name:      "secrets",
		ShortHelp: "Manage Vault Secrets application secrets.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp vault-secrets secrets" }} command group lets you
		manage Vault Secrets application secrets.
		`),
		Aliases: []string{"s"},
		Flags: cmd.Flags{
			Persistent: []*cmd.Flag{
				{
					Name:         "app",
					DisplayValue: "NAME",
					Description:  "The name of the Vault Secrets application. If not specified, the value from the active profile will be used.",
					Value:        appname.Flag(),
				},
			},
		},
		PersistentPreRun: func(c *cmd.Command, args []string) error {
			return appname.Require(ctx)
		},
	}

	// Autocomplete the persistent flag.
	for _, f := range cmd.Flags.Persistent {
		if f.Name == "app" {
			f.Autocomplete = helper.PredictAppName(ctx, cmd, secret_service.New(ctx.HCP, nil))
		}
	}

	cmd.AddChild(NewCmdCreate(ctx, nil))
	cmd.AddChild(NewCmdRead(ctx, nil))
	cmd.AddChild(NewCmdDelete(ctx, nil))
	cmd.AddChild(NewCmdList(ctx, nil))
	cmd.AddChild(NewCmdMigrate(ctx, nil))
	cmd.AddChild(NewCmdOpen(ctx, nil))
	cmd.AddChild(NewCmdRotate(ctx, nil))
	cmd.AddChild(NewCmdUpdate(ctx, nil))

	cmd.AddChild(versions.NewCmdVersions(ctx))
	return cmd
}
