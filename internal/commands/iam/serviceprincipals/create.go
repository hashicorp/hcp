package serviceprincipals

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/client/service_principals_service"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/hashicorp/hcp/internal/pkg/profile"
)

func NewCmdCreate(ctx *cmd.Context, runF func(*CreateOpts) error) *cmd.Command {
	opts := &CreateOpts{
		Ctx:     ctx.ShutdownCtx,
		Profile: ctx.Profile,
		Output:  ctx.Output,
		Client:  service_principals_service.New(ctx.HCP, nil),
	}

	cmd := &cmd.Command{
		Name:      "create",
		ShortHelp: "Create a new service principal.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp iam service-principals create" }} command creates a
		new service principal.

		Once a service principal is created, access to the service principal can be granted by
		generating keys using the {{ template "mdCodeOrBold" "hcp iam service-principals keys create" }}
		command or by federating access using an external workload identity provider using
		{{ template "mdCodeOrBold" "hcp iam service-principals workload-identity-provider create" }}.

		Service principals can be created at the organization scope or project
		scope. It is recommended to create service principals at the project
		scope to limit access to the service principal and to locate the service
		principal near the resources it will be accessing.

		To create an organization service principal, set the 
		{{ template "mdCodeOrBold" "--project" }}
		flag to {{ template "mdCodeOrBold" "-" }}.
		`),
		Examples: []cmd.Example{
			{
				Preamble: `Create a new service principal:`,
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				$ hcp iam service-principals create example-sp
				`),
			},
			{
				Preamble: `Create a new organization service principal:`,
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				$ hcp iam service-principals create example-sp --project="-"
				`),
			},
		},
		Args: cmd.PositionalArguments{
			Args: []cmd.PositionalArgument{
				{
					Name:          "SP_NAME",
					Documentation: "The name of the service principal to create.",
				},
			},
		},
		RunF: func(c *cmd.Command, args []string) error {
			opts.Name = args[0]

			if runF != nil {
				return runF(opts)
			}
			return createRun(opts)
		},
		PersistentPreRun: func(c *cmd.Command, args []string) error {
			return cmd.RequireOrganization(ctx)
		},
	}

	return cmd
}

type CreateOpts struct {
	Ctx     context.Context
	Profile *profile.Profile
	Output  *format.Outputter

	Name   string
	Client service_principals_service.ClientService
}

func createRun(opts *CreateOpts) error {
	req := service_principals_service.NewServicePrincipalsServiceCreateServicePrincipalParamsWithContext(opts.Ctx)
	req.ParentResourceName = opts.Profile.GetOrgResourcePart().String()
	if opts.Profile.ProjectID != "-" && opts.Profile.ProjectID != "" {
		req.ParentResourceName = opts.Profile.GetProjectResourcePart().String()
	}
	req.Body = service_principals_service.ServicePrincipalsServiceCreateServicePrincipalBody{
		Name: opts.Name,
	}

	resp, err := opts.Client.ServicePrincipalsServiceCreateServicePrincipal(req, nil)
	if err != nil {
		return fmt.Errorf("failed to create service principal: %w", err)
	}

	return opts.Output.Display(newDisplayer(format.Pretty, true, resp.Payload.ServicePrincipal))
}
