// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package secrets

import (
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/flagvalue"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
)

var (
	appName string
)

func NewCmdSecrets(ctx *cmd.Context) *cmd.Command {
	cmd := &cmd.Command{
		Name:      "secrets",
		ShortHelp: "Manage Vault Secrets application secrets.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp vault-secrets secrets" }} command group lets you
		manage Vault Secrets application secrets.
		`),
		Flags: cmd.Flags{
			Persistent: []*cmd.Flag{
				{
					Name:         "app-name",
					DisplayValue: "NAME",
					Description:  "The name of the Vault Secrets application. If not specified, then the value from the active profile will be used.",
					Shorthand:    "a",
					Value:        flagvalue.Simple("", &appName),
				},
			},
		},
		PersistentPreRun: func(c *cmd.Command, args []string) error {
			if appName == "" && ctx.Profile.VaultSecrets != nil {
				appName = ctx.Profile.VaultSecrets.AppName
			}
			return cmd.RequireVaultSecretsAppName(ctx, appName)
		},
	}

	cmd.AddChild(NewCmdCreate(ctx, nil))
	return cmd
}
