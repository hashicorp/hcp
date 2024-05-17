// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package secrets

import (
<<<<<<< HEAD
	"github.com/hashicorp/hcp/internal/commands/vaultsecrets/secrets/appname"
	"github.com/hashicorp/hcp/internal/commands/vaultsecrets/secrets/versions"
=======
	preview_secret_service "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/preview/2023-11-28/client/secret_service"
	"github.com/hashicorp/hcp/internal/commands/vaultsecrets/apps/helper"
>>>>>>> 16ec8bc (allow things so as to pass on CI/CD)
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
					Shorthand:    "a",
					Value:        appname.Flag(),
				},
			},
		},
		PersistentPreRun: func(c *cmd.Command, args []string) error {
			return appname.Require(ctx)
		},
	}

	cmd.AddChild(NewCmdCreate(ctx, nil))
	cmd.AddChild(NewCmdRead(ctx, nil))
	cmd.AddChild(NewCmdDelete(ctx, nil))
	cmd.AddChild(NewCmdList(ctx, nil))
	cmd.AddChild(NewCmdOpen(ctx, nil))

	cmd.AddChild(versions.NewCmdVersions(ctx))
	return cmd
}
