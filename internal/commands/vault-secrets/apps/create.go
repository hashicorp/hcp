// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apps

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-06-13/client/secret_service"
	"github.com/hashicorp/hcp/internal/commands/vault-secrets/helper"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/flagvalue"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
)

func NewCmdCreate(ctx *cmd.Context, runF func(*CreateOpts) error) *cmd.Command {
	opts := &CreateOpts{
		Ctx:     ctx.ShutdownCtx,
		Profile: ctx.Profile,
		Output:  ctx.Output,
		IO:      ctx.IO,
		Client:  secret_service.New(ctx.HCP, nil),
	}

	cmd := &cmd.Command{
		Name:      "create",
		ShortHelp: "Create a new vault-secrets application.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp vault-secrets apps create" }} command creates a new vault-secrets application.

		Once an application is created, secrets lifecycle can be managed using the
		{{ template "mdCodeOrBold" "hcp vault-secrets secret" }} command group.
		`),
		Examples: []cmd.Example{
			{
				Preamble: `Create a new application under HCP Vault Secrets:`,
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				$ hcp vault-secrets apps create test-app \
				  --description "This is a test app."
				`),
			},
			{
				Preamble: `Set your active profile to use with the newly created application:`,
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				$ hcp profile init --vault-secrets`),
			},
			{
				Preamble: `Create a new secret under the test-app:`,
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				$ hcp vault-secrets secret create foo=bar`),
			},
		},
		Args: cmd.PositionalArguments{
			Args: []cmd.PositionalArgument{
				{
					Name:          "APP_NAME",
					Documentation: "The name of the application to create.",
				},
			},
		},
		Flags: cmd.Flags{
			Local: []*cmd.Flag{
				{
					Name:         "description",
					DisplayValue: "DESCRIPTION",
					Description:  "An optional description for the application.",
					Value:        flagvalue.Simple("", &opts.Description),
				},
			},
		},
		RunF: func(c *cmd.Command, args []string) error {
			opts.AppName = args[0]
			opts.Logger = c.Logger()
			if runF != nil {
				return runF(opts)
			}
			return createRun(opts)
		},
		PersistentPreRun: func(c *cmd.Command, args []string) error {
			return cmd.RequireOrgAndProject(ctx)
		},
	}

	return cmd
}

type CreateOpts struct {
	Ctx     context.Context
	Profile *profile.Profile
	Output  *format.Outputter
	IO      iostreams.IOStreams
	Logger  hclog.Logger

	AppName     string
	Description string
	Client      secret_service.ClientService
}

func createRun(opts *CreateOpts) error {
	req := secret_service.NewCreateAppParamsWithContext(opts.Ctx)
	req.LocationOrganizationID = opts.Profile.OrganizationID
	req.LocationProjectID = opts.Profile.ProjectID

	req.Body = secret_service.CreateAppBody{
		Name:        opts.AppName,
		Description: opts.Description,
	}

	resp, err := opts.Client.CreateApp(req, nil)
	if err != nil {
		opts.Logger.Info("XXXX", helper.FmtErr(err))
		return fmt.Errorf("%s failed to create application with name: %s - %s", opts.IO.ColorScheme().FailureIcon(), opts.AppName, helper.FmtErr(err))
		//return errors.Wrapf(err, "%s failed to create application", opts.IO.ColorScheme().FailureIcon())
	}

	if opts.Output.GetFormat() == format.Unset {
		fmt.Fprintf(opts.IO.Err(), "%s Created application with name: %s\n", opts.IO.ColorScheme().SuccessIcon(), opts.AppName)
		return nil
	}

	return opts.Output.Display(newDisplayer(true, resp.Payload.App))
}
