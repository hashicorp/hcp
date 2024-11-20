// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apps

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-06-13/client/secret_service"
	preview_secret_service "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-11-28/client/secret_service"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/flagvalue"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
)

type UpdateOpts struct {
	Ctx     context.Context
	Profile *profile.Profile
	Output  *format.Outputter
	IO      iostreams.IOStreams

	AppName       string
	Description   string
	Client        secret_service.ClientService
	PreviewClient preview_secret_service.ClientService
}

func NewCmdUpdate(ctx *cmd.Context, runF func(*UpdateOpts) error) *cmd.Command {
	opts := &UpdateOpts{
		Ctx:           ctx.ShutdownCtx,
		Profile:       ctx.Profile,
		Output:        ctx.Output,
		IO:            ctx.IO,
		Client:        secret_service.New(ctx.HCP, nil),
		PreviewClient: preview_secret_service.New(ctx.HCP, nil),
	}

	cmd := &cmd.Command{
		Name:      "update",
		ShortHelp: "Update a Vault Secrets application.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp vault-secrets apps update" }} command updates the description of a Vault Secrets application.
		`),
		Examples: []cmd.Example{
			{
				Preamble: `Update an application:`,
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				$ hcp vault-secrets apps update company-card --description "Visa card info"
				`),
			},
		},
		Args: cmd.PositionalArguments{
			Args: []cmd.PositionalArgument{
				{
					Name:          "NAME",
					Documentation: "The name of the app to update.",
				},
			},
		},
		Flags: cmd.Flags{
			Local: []*cmd.Flag{
				{
					Name:         "description",
					DisplayValue: "DESCRIPTION",
					Description:  "The updated app description.",
					Value:        flagvalue.Simple("", &opts.Description),
					Required:     true,
				},
			},
		},
		RunF: func(c *cmd.Command, args []string) error {
			opts.AppName = args[0]

			if runF != nil {
				return runF(opts)
			}
			return updateRun(opts)
		},
	}

	return cmd
}

func updateRun(opts *UpdateOpts) error {
	_, err := opts.Client.UpdateApp(&secret_service.UpdateAppParams{
		Context:                opts.Ctx,
		LocationProjectID:      opts.Profile.ProjectID,
		Name:                   opts.AppName,
		LocationOrganizationID: opts.Profile.OrganizationID,
		Body: secret_service.UpdateAppBody{
			Description: opts.Description,
		},
	}, nil)

	if err != nil {
		return fmt.Errorf("failed to update application: %w", err)
	}

	fmt.Fprintf(opts.IO.Err(), "%s Successfully updated application with name %q\n", opts.IO.ColorScheme().SuccessIcon(), opts.AppName)
	return nil
}
