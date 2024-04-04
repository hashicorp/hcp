// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package keys

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
		ShortHelp: "Delete a service principal key.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp iam service-principals keys delete" }} command deletes a service principal key.
		`),
		Examples: []cmd.Example{
			{
				Preamble: `Create a new service principal key:`,
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				$ hcp iam service-principals keys delete \
				  iam/project/example/service-principal/example-sp/key/3KgtSLWTSs
				`),
			},
		},
		Args: cmd.PositionalArguments{
			Args: []cmd.PositionalArgument{
				{
					Name:          "KEY_NAME",
					Documentation: "The resource name of the service principal key to delete.",
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

	return cmd
}

type DeleteOpts struct {
	Ctx     context.Context
	Profile *profile.Profile
	IO      iostreams.IOStreams

	Name   string
	Client service_principals_service.ClientService
}

func deleteRun(opts *DeleteOpts) error {
	if opts.IO.CanPrompt() {
		ok, err := opts.IO.PromptConfirm("The service principal key will be deleted. " +
			"Any workload using this service principal will no longer be able to authenticate." +
			"\n\nDo you want to continue")
		if err != nil {
			return fmt.Errorf("failed to retrieve confirmation: %w", err)
		}

		if !ok {
			return nil
		}
	}
	req := service_principals_service.NewServicePrincipalsServiceDeleteServicePrincipalKeyParamsWithContext(opts.Ctx)
	req.ResourceName2 = opts.Name

	_, err := opts.Client.ServicePrincipalsServiceDeleteServicePrincipalKey(req, nil)
	if err != nil {
		return fmt.Errorf("failed to delete service principal key: %w", err)
	}

	fmt.Fprintf(opts.IO.Err(), "%s Service principal key %q deleted\n",
		opts.IO.ColorScheme().SuccessIcon(), opts.Name)
	return nil
}
