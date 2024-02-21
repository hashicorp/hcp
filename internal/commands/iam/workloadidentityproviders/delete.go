package workloadidentityproviders

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/client/service_principals_service"
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
		ShortHelp: "Delete a workload identity provider.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ Bold "hcp iam workload-identity-providers delete" }} command deletes a workload identity provider.
		`),
		Examples: []cmd.Example{
			{
				Preamble: `Delete a workload identity provider:`,
				Command: heredoc.New(ctx.IO, heredoc.WithNoWrap(), heredoc.WithPreserveNewlines()).Must(`
				$ hcp iam workload-identity-provider delete \
				  iam/project/my-project/service-principal/my-sp/workload-identity-provider/example-wip
				`),
			},
		},
		Args: cmd.PositionalArguments{
			Args: []cmd.PositionalArgument{
				{
					Name:          "WIP_NAME",
					Documentation: heredoc.New(ctx.IO).Mustf(WIPNameArgDoc, "delete"),
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
			return cmd.RequireOrgAndProject(ctx)
		},
	}

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
	if opts.IO.CanPrompt() {
		ok, err := opts.IO.PromptConfirm("The workload identity provider will be deleted.\n\nDo you want to continue")
		if err != nil {
			return fmt.Errorf("failed to retrieve confirmation: %w", err)
		}

		if !ok {
			return nil
		}
	}

	req := service_principals_service.NewServicePrincipalsServiceDeleteWorkloadIdentityProviderParamsWithContext(opts.Ctx)
	req.ResourceName4 = opts.Name

	_, err := opts.Client.ServicePrincipalsServiceDeleteWorkloadIdentityProvider(req, nil)
	if err != nil {
		return fmt.Errorf("failed to delete workload identity provider: %w", err)
	}

	fmt.Fprintf(opts.IO.Err(), "%s Workload identity provider %q deleted\n",
		opts.IO.ColorScheme().SuccessIcon(), opts.Name)
	return nil
}
