// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package gatewaypools

import (
	"context"
	"fmt"

	preview_secret_service "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/preview/2023-11-28/client/secret_service"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
)

type ReadOpts struct {
	Ctx     context.Context
	Profile *profile.Profile
	Output  *format.Outputter
	IO      iostreams.IOStreams

	GatewayPoolName string
	PreviewClient   preview_secret_service.ClientService
}

func NewCmdRead(ctx *cmd.Context, runF func(*ReadOpts) error) *cmd.Command {
	opts := &ReadOpts{
		Ctx:           ctx.ShutdownCtx,
		Profile:       ctx.Profile,
		Output:        ctx.Output,
		IO:            ctx.IO,
		PreviewClient: preview_secret_service.New(ctx.HCP, nil),
	}

	cmd := &cmd.Command{
		Name:      "read",
		ShortHelp: "Read a Vault Secrets gateway pool.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp vault-secrets gateway-pools read" }} command gets a Vault Secrets gateway pool.
		`),
		Examples: []cmd.Example{
			{
				Preamble: `Read a gateway pool:`,
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				$ hcp vault-secrets gateway-pools read company-tunnel
				`),
			},
		},
		Args: cmd.PositionalArguments{
			Args: []cmd.PositionalArgument{
				{
					Name:          "NAME",
					Documentation: "The name of the gateway pool to read.",
				},
			},
		},
		RunF: func(c *cmd.Command, args []string) error {
			opts.GatewayPoolName = args[0]

			if runF != nil {
				return runF(opts)
			}
			return readRun(opts)
		},
	}
	cmd.Args.Autocomplete = PredictGatewayPoolName(ctx, cmd, preview_secret_service.New(ctx.HCP, nil))

	return cmd
}

func readRun(opts *ReadOpts) error {
	resp, err := opts.PreviewClient.GetGatewayPool(&preview_secret_service.GetGatewayPoolParams{
		Context:         opts.Ctx,
		ProjectID:       opts.Profile.ProjectID,
		OrganizationID:  opts.Profile.OrganizationID,
		GatewayPoolName: opts.GatewayPoolName,
	}, nil)

	if err != nil {
		return fmt.Errorf("failed to read gateway pool: %w", err)
	}

	if resp.Payload == nil || resp.Payload.GatewayPool == nil {
		return fmt.Errorf("gateway pool not found")
	}

	integList, err := opts.PreviewClient.ListGatewayPoolIntegrations(&preview_secret_service.ListGatewayPoolIntegrationsParams{
		Context:         opts.Ctx,
		ProjectID:       opts.Profile.ProjectID,
		OrganizationID:  opts.Profile.OrganizationID,
		GatewayPoolName: resp.Payload.GatewayPool.Name,
	}, nil)
	if err != nil {
		return fmt.Errorf("failed to list gateway pool integrations: %w", err)
	}

	gwIntegrations := &gatewayPoolWithIntegrations{
		GatewayPool:  resp.Payload.GatewayPool,
		Integrations: integList.Payload.Integrations,
	}

	return opts.Output.Display(newDisplayer(true, true, gwIntegrations))
}
