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

type DeleteOpts struct {
	Ctx     context.Context
	Profile *profile.Profile
	Output  *format.Outputter
	IO      iostreams.IOStreams

	AppName       string
	Client        secret_service.ClientService
	PreviewClient preview_secret_service.ClientService
}

func NewCmdDelete(ctx *cmd.Context, runF func(*DeleteOpts) error) *cmd.Command {
	opts := &DeleteOpts{
		Ctx:           ctx.ShutdownCtx,
		Profile:       ctx.Profile,
		Output:        ctx.Output,
		IO:            ctx.IO,
		Client:        secret_service.New(ctx.HCP, nil),
		PreviewClient: preview_secret_service.New(ctx.HCP, nil),
	}

	cmd := &cmd.Command{
		Name:      "delete",
		ShortHelp: "Delete a Vault Secrets application.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp vault-secrets apps delete" }} command deletes a Vault Secrets application.
		`),
		Examples: []cmd.Example{
			{
				Preamble: `Delete an application:`,
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				$ hcp vault-secrets apps delete company-card
				`),
			},
		},
		Args: cmd.PositionalArguments{
			Args: []cmd.PositionalArgument{
				{
					Name:          "NAME",
					Documentation: "The name of the app to delete.",
				},
			},
			Autocomplete: helper.PredictAppName(ctx, ctx.Profile.OrganizationID, ctx.Profile.ProjectID, opts.PreviewClient),
		},
		RunF: func(c *cmd.Command, args []string) error {
			opts.AppName = args[0]

			if runF != nil {
				return runF(opts)
			}
			return deleteRun(opts)
		},
	}

	return cmd
}

func deleteRun(opts *DeleteOpts) error {
	_, err := opts.Client.DeleteApp(&secret_service.DeleteAppParams{
		Context:                opts.Ctx,
		LocationOrganizationID: opts.Profile.OrganizationID,
		LocationProjectID:      opts.Profile.ProjectID,
		Name:                   opts.AppName,
	}, nil)

	if err != nil {
		return fmt.Errorf("failed to delete application: %w", err)
	}

	fmt.Fprintf(opts.IO.Err(), "%s Successfully deleted application with name %q\n", opts.IO.ColorScheme().SuccessIcon(), opts.AppName)
	return nil
}
