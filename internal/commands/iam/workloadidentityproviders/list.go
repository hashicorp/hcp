package workloadidentityproviders

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/client/service_principals_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/models"
	"github.com/hashicorp/hcp/internal/commands/iam/serviceprincipals/helper"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
)

func NewCmdList(ctx *cmd.Context, runF func(*ListOpts) error) *cmd.Command {
	opts := &ListOpts{
		Ctx:     ctx.ShutdownCtx,
		Profile: ctx.Profile,
		IO:      ctx.IO,
		Output:  ctx.Output,
		Client:  service_principals_service.New(ctx.HCP, nil),
	}

	cmd := &cmd.Command{
		Name:      "list",
		ShortHelp: "List workload identity providers.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ Bold "hcp iam workload-identity-providers list" }} command lists the workload
		identity providers for a given service principal.
		`),
		Examples: []cmd.Example{
			{
				Preamble: `List workload identity provider given the service principal's resource name suffix:`,
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				$ hcp iam workload-identity-provider list example-sp
				`),
			},
			{
				Preamble: `List workload identity provider given the service principal's resource name:`,
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				$ hcp iam workload-identity-provider list \
				  iam/project/example-project/service-principal/example-sp
				`),
			},
		},
		Args: cmd.PositionalArguments{
			Args: []cmd.PositionalArgument{
				{
					Name:          "SP_NAME",
					Documentation: heredoc.New(ctx.IO).Mustf(helper.SPNameArgDoc, "list workload identity providers for."),
				},
			},
		},
		RunF: func(c *cmd.Command, args []string) error {
			opts.SP = args[0]
			if runF != nil {
				return runF(opts)
			}

			return listRun(opts)
		},
		PersistentPreRun: func(c *cmd.Command, args []string) error {
			return cmd.RequireOrgAndProject(ctx)
		},
	}

	// Setup the autocomplete for the service principalname argument
	cmd.Args.Autocomplete = helper.PredictSPResourceNameSuffix(ctx, cmd, opts.Client)

	return cmd
}

type ListOpts struct {
	Ctx     context.Context
	Profile *profile.Profile
	IO      iostreams.IOStreams
	Output  *format.Outputter

	SP     string
	Client service_principals_service.ClientService
}

func listRun(opts *ListOpts) error {
	req := service_principals_service.NewServicePrincipalsServiceListWorkloadIdentityProviderParamsWithContext(opts.Ctx)
	req.ParentResourceName = helper.ResourceName(opts.SP, opts.Profile.OrganizationID, opts.Profile.ProjectID)

	var wips []*models.HashicorpCloudIamWorkloadIdentityProvider
	for {

		resp, err := opts.Client.ServicePrincipalsServiceListWorkloadIdentityProvider(req, nil)
		if err != nil {
			return fmt.Errorf("failed to list workload identity providers: %w", err)
		}
		wips = append(wips, resp.Payload.Providers...)
		if resp.Payload.Pagination == nil || resp.Payload.Pagination.NextPageToken == "" {
			break
		}

		next := resp.Payload.Pagination.NextPageToken
		req.PaginationNextPageToken = &next
	}

	return opts.Output.Display(newDisplayer(format.Pretty, false, wips...))
}
