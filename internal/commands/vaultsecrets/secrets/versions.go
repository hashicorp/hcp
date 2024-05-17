// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package secrets

import (
	"context"
	"fmt"

	preview_secret_service "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/preview/2023-11-28/client/secret_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-06-13/client/secret_service"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/flagvalue"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
)

func NewCmdVersions(ctx *cmd.Context, runF func(*VersionsOpts) error) *cmd.Command {
	opts := &VersionsOpts{
		Ctx:           ctx.ShutdownCtx,
		Profile:       ctx.Profile,
		IO:            ctx.IO,
		Output:        ctx.Output,
		PreviewClient: preview_secret_service.New(ctx.HCP, nil),
		Client:        secret_service.New(ctx.HCP, nil),
	}

	cmd := &cmd.Command{
		Name:      "read",
		ShortHelp: "Read a static secret.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp vault-secrets secrets read" }} command reads a static secret from the Vault Secrets application.
		`),
		Examples: []cmd.Example{
			{
				Preamble: `Read secret's metadata:`,
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				$ hcp vault-secrets secret read "test_secret"
				`),
			},
			{
				Preamble: `Read plaintext secret:`,
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				$ hcp vault-secrets secret read "test_secret" --plaintext
				`),
			},
		},
		Args: cmd.PositionalArguments{
			Args: []cmd.PositionalArgument{
				{
					Name:          "NAME",
					Documentation: "The name of the secret to read.",
				},
			},
		},
		Flags: cmd.Flags{
			Local: []*cmd.Flag{
				{
					Name:          "plaintext",
					Description:   "Display the secret's value in plaintext.",
					Value:         flagvalue.Simple(false, &opts.OpenSecret),
					IsBooleanFlag: true,
				},
			},
		},
		RunF: func(c *cmd.Command, args []string) error {
			opts.AppName = appName
			opts.SecretName = args[0]

			if runF != nil {
				return runF(opts)
			}
			return versionsRun(opts)
		},
	}

	return cmd
}

type VersionsOpts struct {
	Ctx     context.Context
	Profile *profile.Profile
	IO      iostreams.IOStreams
	Output  *format.Outputter

	AppName       string
	SecretName    string
	OpenSecret    bool
	PreviewClient preview_secret_service.ClientService
	Client        secret_service.ClientService
}

func versionsRun(opts *VersionsOpts) error {
	req := secret_service.NewGetAppSecretParamsWithContext(opts.Ctx)
	req.LocationOrganizationID = opts.Profile.OrganizationID
	req.LocationProjectID = opts.Profile.ProjectID
	req.AppName = opts.AppName
	req.SecretName = opts.SecretName

	resp, err := opts.Client.GetAppSecret(req, nil)
	if err != nil {
		return fmt.Errorf("failed to read the secret %q: %w", opts.SecretName, err)
	}
	return opts.Output.Display(newDisplayer(true).Secrets(resp.Payload.Secret))
}
