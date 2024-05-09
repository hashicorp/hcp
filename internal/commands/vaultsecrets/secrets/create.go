// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package secrets

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/preview/2023-11-28/client/secret_service"
	preview_secret_service "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/preview/2023-11-28/client/secret_service"

	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/flagvalue"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/hashicorp/hcp/internal/pkg/profile"
)

var (
	appName string
)

func NewCmdCreate(ctx *cmd.Context, runF func(*CreateOpts) error) *cmd.Command {
	opts := &CreateOpts{
		Ctx:           ctx.ShutdownCtx,
		Profile:       ctx.Profile,
		Output:        ctx.Output,
		PreviewClient: preview_secret_service.New(ctx.HCP, nil),
		Client:        secret_service.New(ctx.HCP, nil),
	}

	cmd := &cmd.Command{
		Name:      "create",
		ShortHelp: "Create a new secret.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp vault-secrets secret create" }} command creates a new secret under an Vault Secrets App.

		Once a secret is created, it can be managed using
		{{ template "mdCodeOrBold" "hcp vault-secrets secret" }} command group.
		`),
		Examples: []cmd.Example{
			{
				Preamble: `Create a new secret`,
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				$ hcp vault-secret secret create foo=$(cat /tmp/secret.txt)
				`),
			},
		},
		Flags: cmd.Flags{
			Persistent: []*cmd.Flag{
				{
					Name:         "app",
					DisplayValue: "NAME",
					Value:        flagvalue.Simple("", &appName),
				},
			},
		},
		RunF: func(c *cmd.Command, args []string) error {
			if runF != nil {
				return runF(opts)
			}
			return createRun(opts)
		},
		PersistentPreRun: func(c *cmd.Command, args []string) error {
			fmt.Println("XXXX presistent pre runs")
			return cmd.RequireVaultSecretsAppName(ctx)
		},
	}

	return cmd
}

type CreateOpts struct {
	Ctx     context.Context
	Profile *profile.Profile
	Output  *format.Outputter

	PreviewClient preview_secret_service.ClientService
	Client        secret_service.ClientService
}

func createRun(opts *CreateOpts) error {
	return nil
}
