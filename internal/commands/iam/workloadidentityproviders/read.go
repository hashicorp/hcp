// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package workloadidentityproviders

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/client/service_principals_service"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
)

func NewCmdRead(ctx *cmd.Context, runF func(*ReadOpts) error) *cmd.Command {
	opts := &ReadOpts{
		Ctx:     ctx.ShutdownCtx,
		Profile: ctx.Profile,
		IO:      ctx.IO,
		Output:  ctx.Output,
		Client:  service_principals_service.New(ctx.HCP, nil),
	}

	cmd := &cmd.Command{
		Name:      "read",
		ShortHelp: "Show metadata about a workload identity provider.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp iam workload-identity-providers read" }} command
		shows metadata about the specified workload identity provider.
		`),
		Examples: []cmd.Example{
			{
				Preamble: `Read a workload identity provider:`,
				Command: heredoc.New(ctx.IO, heredoc.WithNoWrap(), heredoc.WithPreserveNewlines()).Must(`
				$ hcp iam workload-identity-provider read \
				  iam/project/my-project/service-principal/my-sp/workload-identity-provider/example-wip
				`),
			},
		},
		Args: cmd.PositionalArguments{
			Args: []cmd.PositionalArgument{
				{
					Name:          "WIP_NAME",
					Documentation: heredoc.New(ctx.IO).Mustf(WIPNameArgDoc, "read"),
				},
			},
		},
		RunF: func(c *cmd.Command, args []string) error {
			opts.WIP = args[0]
			if runF != nil {
				return runF(opts)
			}

			return readRun(opts)
		},
		PersistentPreRun: func(c *cmd.Command, args []string) error {
			return cmd.RequireOrgAndProject(ctx)
		},
	}

	return cmd
}

type ReadOpts struct {
	Ctx     context.Context
	Profile *profile.Profile
	IO      iostreams.IOStreams
	Output  *format.Outputter

	WIP    string
	Client service_principals_service.ClientService
}

func readRun(opts *ReadOpts) error {
	req := service_principals_service.NewServicePrincipalsServiceGetWorkloadIdentityProviderParamsWithContext(opts.Ctx)
	req.ResourceName2 = opts.WIP

	resp, err := opts.Client.ServicePrincipalsServiceGetWorkloadIdentityProvider(req, nil)
	if err != nil {
		return fmt.Errorf("failed to get workload identity provider: %w", err)
	}

	return opts.Output.Display(newDisplayer(format.Pretty, true, resp.Payload.Provider))
}
