package keys

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

func NewCmdList(ctx *cmd.Context, runF func(*ListOpts) error) *cmd.Command {
	opts := &ListOpts{
		Ctx:     ctx.ShutdownCtx,
		Profile: ctx.Profile,
		Output:  ctx.Output,
		Client:  service_principals_service.New(ctx.HCP, nil),
	}

	cmd := &cmd.Command{
		Name:      "list",
		ShortHelp: "List a service principal's keys.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp iam service-principals keys list" }} command lists a service principal's keys.

		To list keys for an organization service principal, pass the service principal's resource name or set the --project
		flag to "-" and pass its resource name suffix.
		`),
		Examples: []cmd.Example{
			{
				Preamble: `List a service principal's keys:`,
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				$ hcp iam service-principals keys list my-service-principal
				`),
			},
			{
				Preamble: `List a service principal's keys specifying the resource name of the service principal:`,
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				$ hcp iam service-principals keys list \
				  iam/project/123/service-principal/my-service-principal
				`),
			},
		},
		Args: cmd.PositionalArguments{
			Args: []cmd.PositionalArgument{
				{
					Name:          "SP_NAME",
					Documentation: heredoc.New(ctx.IO).Mustf(helper.SPNameArgDoc, "list keys for"),
				},
			},
		},
		RunF: func(c *cmd.Command, args []string) error {
			opts.Name = args[0]

			if runF != nil {
				return runF(opts)
			}
			return listRun(opts)
		},
		PersistentPreRun: func(c *cmd.Command, args []string) error {
			return cmd.RequireOrganization(ctx)
		},
	}

	// Setup the autocomplete for the name argument
	cmd.Args.Autocomplete = helper.PredictSPResourceNameSuffix(ctx, cmd, opts.Client)

	return cmd
}

type ListOpts struct {
	Ctx     context.Context
	Profile *profile.Profile
	Output  *format.Outputter

	Name   string
	Client service_principals_service.ClientService
}

func listRun(opts *ListOpts) error {
	req := service_principals_service.NewServicePrincipalsServiceGetServicePrincipalParamsWithContext(opts.Ctx)
	req.ResourceName = helper.ResourceName(opts.Name, opts.Profile.OrganizationID, opts.Profile.ProjectID)

	resp, err := opts.Client.ServicePrincipalsServiceGetServicePrincipal(req, nil)
	if err != nil {
		return fmt.Errorf("failed to read service principal to retrieve keys: %w", err)
	}

	return opts.Output.Display(newDisplayer(format.Pretty, resp.Payload.Keys...))
}
