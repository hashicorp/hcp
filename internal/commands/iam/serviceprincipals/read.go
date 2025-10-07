// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: MPL-2.0

package serviceprincipals

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/client/service_principals_service"
	"github.com/hashicorp/hcp/internal/commands/iam/serviceprincipals/helper"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/hashicorp/hcp/internal/pkg/profile"
)

func NewCmdRead(ctx *cmd.Context, runF func(*ReadOpts) error) *cmd.Command {
	opts := &ReadOpts{
		Ctx:     ctx.ShutdownCtx,
		Profile: ctx.Profile,
		Output:  ctx.Output,
		Client:  service_principals_service.New(ctx.HCP, nil),
	}

	cmd := &cmd.Command{
		Name:      "read",
		ShortHelp: "Show metadata for the given service principal.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp iam service-principals read" }} command reads details about the given
		service principal.

		To read an organization service principal, pass the service principal's 
		resource name or set the {{ template "mdCodeOrBold" "--project" }}
		flag to {{ template "mdCodeOrBold" "-" }} and pass its resource name suffix.
		`),
		Examples: []cmd.Example{
			{
				Preamble: `Read the service principal using the resource name suffix "example-sp":`,
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				$ hcp iam service-principals read example-sp
				`),
			},
			{
				Preamble: `Read the service principal using the service principal's resource name:`,
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				$ hcp iam service-principals read \
				  iam/project/example-project/service-principal/example-sp
				`),
			},
		},
		Args: cmd.PositionalArguments{
			Args: []cmd.PositionalArgument{
				{
					Name:          "SP_NAME",
					Documentation: heredoc.New(ctx.IO).Mustf(helper.SPNameArgDoc, "read"),
				},
			},
		},
		RunF: func(c *cmd.Command, args []string) error {
			opts.Name = args[0]

			if runF != nil {
				return runF(opts)
			}
			return readRun(opts)
		},
		PersistentPreRun: func(c *cmd.Command, args []string) error {
			return cmd.RequireOrganization(ctx)
		},
	}

	// Setup the autocomplete for the name argument
	cmd.Args.Autocomplete = helper.PredictSPResourceNameSuffix(ctx, cmd, opts.Client)

	return cmd
}

type ReadOpts struct {
	Ctx     context.Context
	Profile *profile.Profile
	Output  *format.Outputter

	Name   string
	Client service_principals_service.ClientService
}

func readRun(opts *ReadOpts) error {
	rn := helper.ResourceName(opts.Name, opts.Profile.OrganizationID, opts.Profile.ProjectID)
	req := service_principals_service.NewServicePrincipalsServiceGetServicePrincipalParamsWithContext(opts.Ctx)
	req.ResourceName = rn

	resp, err := opts.Client.ServicePrincipalsServiceGetServicePrincipal(req, nil)
	if err != nil {
		return fmt.Errorf("failed to read service principal: %w", err)
	}

	return opts.Output.Display(newDisplayer(format.Pretty, true, resp.Payload.ServicePrincipal))
}
