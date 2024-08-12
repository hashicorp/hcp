// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package secrets

import (
	"context"
	"fmt"

	preview_secret_service "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/preview/2023-11-28/client/secret_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-06-13/client/secret_service"
	"github.com/hashicorp/hcp/internal/commands/vaultsecrets/secrets/appname"
	"github.com/hashicorp/hcp/internal/commands/vaultsecrets/secrets/helper"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
)

type RotateOpts struct {
	Ctx     context.Context
	Profile *profile.Profile
	Output  *format.Outputter
	IO      iostreams.IOStreams

	AppName       string
	SecretName    string
	Client        secret_service.ClientService
	PreviewClient preview_secret_service.ClientService
}

func NewCmdRotate(ctx *cmd.Context, runF func(*RotateOpts) error) *cmd.Command {
	opts := &RotateOpts{
		Ctx:           ctx.ShutdownCtx,
		Profile:       ctx.Profile,
		IO:            ctx.IO,
		Output:        ctx.Output,
		PreviewClient: preview_secret_service.New(ctx.HCP, nil),
		Client:        secret_service.New(ctx.HCP, nil),
	}

	cmd := &cmd.Command{
		Name:      "rotate",
		ShortHelp: "Rotate a rotating secret.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp vault-secrets secrets rotate" }} command rotates a rotating secret from the Vault Secrets application.
		`),
		Examples: []cmd.Example{
			{
				Preamble: `Rotate a secret:`,
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				$ hcp vault-secrets secret rotate "test_secret"
				`),
			},
			{
				Preamble: `Read a secret under a specified Vault Secrets application:`,
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				$ hcp vault-secrets secret rotate "test_secret" --app test-app
				`),
			},
		},
		Args: cmd.PositionalArguments{
			Args: []cmd.PositionalArgument{
				{
					Name:          "NAME",
					Documentation: "The name of the secret to rotate.",
				},
			},
		},
		RunF: func(c *cmd.Command, args []string) error {
			opts.AppName = appname.Get()
			opts.SecretName = args[0]

			if runF != nil {
				return runF(opts)
			}
			return rotateRun(opts)
		},
	}
	cmd.Args.Autocomplete = helper.PredictSecretName(ctx, cmd, opts.PreviewClient)

	return cmd
}

func rotateRun(opts *RotateOpts) error {
	params := &preview_secret_service.RotateSecretParams{
		Context:        opts.Ctx,
		OrganizationID: opts.Profile.OrganizationID,
		ProjectID:      opts.Profile.ProjectID,
		AppName:        opts.AppName,
		SecretName:     opts.SecretName,
	}

	_, err := opts.PreviewClient.RotateSecret(params, nil)
	if err != nil {
		return fmt.Errorf("failed to rotate the secret %q: %w", opts.SecretName, err)
	}

	fmt.Fprintf(opts.IO.Err(), "%s Successfully rotated secret with name %q\n", opts.IO.ColorScheme().SuccessIcon(), opts.SecretName)
	return nil
}
