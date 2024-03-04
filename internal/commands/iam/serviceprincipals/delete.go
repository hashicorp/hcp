package serviceprincipals

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/client/service_principals_service"
	"github.com/hashicorp/hcp/internal/commands/iam/serviceprincipals/helper"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
)

func NewCmdDelete(ctx *cmd.Context, runF func(*DeleteOpts) error) *cmd.Command {
	opts := &DeleteOpts{
		Ctx:     ctx.ShutdownCtx,
		Profile: ctx.Profile,
		IO:      ctx.IO,
		Client:  service_principals_service.New(ctx.HCP, nil),
	}

	cmd := &cmd.Command{
		Name:      "delete",
		ShortHelp: "Delete a service principal.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp iam service-principals delete" }} command deletes a service principal.

		Once the service-principal is deleted, all IAM policy that bound the service principal will be updated.

		To delete an organization service principal, pass the service principal's resource name or set the --project
		flag to "-" and pass its resource name suffix.
		`),
		Examples: []cmd.Example{
			{
				Preamble: `Delete a service principal using its name suffix:`,
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				$ hcp iam service-principals delete example-sp
				`),
			},
			{
				Preamble: `Delete a service principal using its resource name:`,
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				$ hcp iam service-principals delete \
				  iam/project/example-project/service-principal/example-sp
				`),
			},
		},
		Args: cmd.PositionalArguments{
			Args: []cmd.PositionalArgument{
				{
					Name:          "SP_NAME",
					Documentation: heredoc.New(ctx.IO).Mustf(helper.SPNameArgDoc, "delete"),
				},
			},
		},
		RunF: func(c *cmd.Command, args []string) error {
			opts.Name = args[0]

			if runF != nil {
				return runF(opts)
			}
			return deleteRun(opts)
		},
		PersistentPreRun: func(c *cmd.Command, args []string) error {
			return cmd.RequireOrganization(ctx)
		},
	}

	// Setup the autocomplete for the name argument
	cmd.Args.Autocomplete = helper.PredictSPResourceNameSuffix(ctx, cmd, opts.Client)

	return cmd
}

type DeleteOpts struct {
	Ctx     context.Context
	IO      iostreams.IOStreams
	Profile *profile.Profile

	Name   string
	Client service_principals_service.ClientService
}

func deleteRun(opts *DeleteOpts) error {
	rn := helper.ResourceName(opts.Name, opts.Profile.OrganizationID, opts.Profile.ProjectID)
	if opts.IO.CanPrompt() {
		ok, err := opts.IO.PromptConfirm("The service principal will be deleted.\n\nDo you want to continue")
		if err != nil {
			return fmt.Errorf("failed to retrieve confirmation: %w", err)
		}

		if !ok {
			return nil
		}
	}

	req := service_principals_service.NewServicePrincipalsServiceDeleteServicePrincipalParamsWithContext(opts.Ctx)
	req.ResourceName = rn

	_, err := opts.Client.ServicePrincipalsServiceDeleteServicePrincipal(req, nil)
	if err != nil {
		return fmt.Errorf("failed to delete service principal: %w", err)
	}

	fmt.Fprintf(opts.IO.Err(), "%s Service principal %q deleted\n",
		opts.IO.ColorScheme().SuccessIcon(), rn)
	return nil
}
