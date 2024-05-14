// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apps

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

type CreateOpts struct {
	Ctx     context.Context
	Profile *profile.Profile
	Output  *format.Outputter
	IO      iostreams.IOStreams

	AppName       string
	Description   string
	Client        secret_service.ClientService
	PreviewClient preview_secret_service.ClientService
}

func NewCmdCreate(ctx *cmd.Context, runF func(*CreateOpts) error) *cmd.Command {
	opts := &CreateOpts{
		Ctx:           ctx.ShutdownCtx,
		Profile:       ctx.Profile,
		Output:        ctx.Output,
		IO:            ctx.IO,
		Client:        secret_service.New(ctx.HCP, nil),
		PreviewClient: preview_secret_service.New(ctx.HCP, nil),
	}

	cmd := &cmd.Command{
		Name:      "create",
		ShortHelp: "Create a new HCP Vault Secrets application.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp vault-secrets apps create" }} command creates a new Vault Secrets application.

		Once an application is created, secrets lifecycle can be managed using the
		{{ template "mdCodeOrBold" "hcp vault-secrets secrets" }} command group.
		`),
		Examples: []cmd.Example{
			{
				Preamble: `Create a new application:`,
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				$ hcp vault-secrets apps create company-card \
				  --description "Stores corporate card info."
				`),
			},
		},
		Args: cmd.PositionalArguments{
			Args: []cmd.PositionalArgument{
				{
					Name:          "APP_NAME",
					Documentation: "The name of the app to create.",
				},
			},
		},
		Flags: cmd.Flags{
			Local: []*cmd.Flag{
				{
					Name:         "description",
					DisplayValue: "DESCRIPTION",
					Description:  "An optional description for the app to create.",
					Value:        flagvalue.Simple("", &opts.Description),
					Required:     false,
				},
			},
		},
		RunF: func(c *cmd.Command, args []string) error {
			opts.AppName = args[0]

			if runF != nil {
				return runF(opts)
			}
			return appCreate(opts)
		},
		PersistentPreRun: func(c *cmd.Command, args []string) error {
			return cmd.RequireOrgAndProject(ctx)
		},
	}

	return cmd
}

func appCreate(opts *CreateOpts) error {
	resp, err := opts.Client.CreateApp(&secret_service.CreateAppParams{
		Context:                opts.Ctx,
		LocationProjectID:      opts.Profile.ProjectID,
		LocationOrganizationID: opts.Profile.OrganizationID,
		Body: secret_service.CreateAppBody{
			Name:        opts.AppName,
			Description: opts.Description,
		},
	}, nil)

	if err != nil {
		return fmt.Errorf("failed to create application: %w", err)
	}

	if opts.Output.GetFormat() == format.Unset {
		fmt.Fprintf(opts.IO.Err(), "%s Successfully created application with name %s\n", opts.IO.ColorScheme().SuccessIcon(), opts.AppName)
		return nil
	}

	return opts.Output.Display(newDisplayer(true, resp.Payload.App))
}
