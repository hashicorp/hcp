// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apps

import (
	"context"
	"fmt"

	preview_secret_service "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/preview/2023-11-28/client/secret_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-06-13/client/secret_service"
	"github.com/hashicorp/hcp/internal/commands/vaultsecrets/apps/helper"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
)

type ReadOpts struct {
	Ctx     context.Context
	Profile *profile.Profile
	Output  *format.Outputter
	IO      iostreams.IOStreams

	AppName       string
	Client        secret_service.ClientService
	PreviewClient preview_secret_service.ClientService
}

func NewCmdRead(ctx *cmd.Context, runF func(*ReadOpts) error) *cmd.Command {
	opts := &ReadOpts{
		Ctx:           ctx.ShutdownCtx,
		Profile:       ctx.Profile,
		Output:        ctx.Output,
		IO:            ctx.IO,
		Client:        secret_service.New(ctx.HCP, nil),
		PreviewClient: preview_secret_service.New(ctx.HCP, nil),
	}

	cmd := &cmd.Command{
		Name:      "read",
		ShortHelp: "Read a Vault Secrets application.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp vault-secrets apps read" }} command gets a Vault Secrets application.
		`),
		Examples: []cmd.Example{
			{
				Preamble: `Read an application:`,
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				$ hcp vault-secrets apps read company-card
				`),
			},
		},
		Args: cmd.PositionalArguments{
			Args: []cmd.PositionalArgument{
				{
					Name:          "NAME",
					Documentation: "The name of the app to read.",
				},
			},
		},
		RunF: func(c *cmd.Command, args []string) error {
			opts.AppName = args[0]

			if runF != nil {
				return runF(opts)
			}
			return readRun(opts)
		},
	}
	cmd.Args.Autocomplete = helper.PredictAppName(ctx, cmd, preview_secret_service.New(ctx.HCP, nil))

	return cmd
}

func readRun(opts *ReadOpts) error {
	resp, err := opts.Client.GetApp(&secret_service.GetAppParams{
		Context:                opts.Ctx,
		LocationProjectID:      opts.Profile.ProjectID,
		LocationOrganizationID: opts.Profile.OrganizationID,
		Name:                   opts.AppName,
	}, nil)

	if err != nil {
		return fmt.Errorf("failed to read application: %w", err)
	}

	return opts.Output.Display(newDisplayer(true, resp.Payload.App))
}
