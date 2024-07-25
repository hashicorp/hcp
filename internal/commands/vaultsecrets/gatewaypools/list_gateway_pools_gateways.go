// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package gatewaypools

import (
	"context"
	"fmt"

	preview_secret_service "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/preview/2023-11-28/client/secret_service"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/flagvalue"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
)

type ListGatewaysOpts struct {
	Ctx     context.Context
	Profile *profile.Profile
	Output  *format.Outputter
	IO      iostreams.IOStreams

	PreviewClient   preview_secret_service.ClientService
	GatewayPoolName string
	ShowAll         bool
}

func NewCmdListGatewayPoolsGateway(ctx *cmd.Context, runF func(*ListGatewaysOpts) error) *cmd.Command {
	opts := &ListGatewaysOpts{
		Ctx:           ctx.ShutdownCtx,
		Profile:       ctx.Profile,
		IO:            ctx.IO,
		Output:        ctx.Output,
		PreviewClient: preview_secret_service.New(ctx.HCP, nil),
	}

	cmd := &cmd.Command{
		Name:      "list-gateways",
		ShortHelp: "List Vault Secrets gateway pools gateways.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp vault-secrets gateway-pools list-gateways" }} command lists all Vault Secrets gateway pools gateways.
		`),
		Examples: []cmd.Example{
			{
				Preamble: `List gateway-pools gateways:`,
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				$ hcp vault-secrets gateway-pools list-gateways company-tunnel
				`),
			},
		},
		Flags: cmd.Flags{
			Local: []*cmd.Flag{
				{
					Name:          "show-all",
					Shorthand:     "a",
					Description:   "Show all fields.",
					IsBooleanFlag: true,
					Value:         flagvalue.Simple(false, &opts.ShowAll),
					Required:      false,
				},
			},
		},
		Args: cmd.PositionalArguments{
			Args: []cmd.PositionalArgument{
				{
					Name:          "NAME",
					Documentation: "The name of the gateway pool to list its gateways.",
				},
			},
		},
		RunF: func(c *cmd.Command, args []string) error {
			opts.GatewayPoolName = args[0]

			if runF != nil {
				return runF(opts)
			}
			return listGatewaysRun(opts)
		},
	}
	return cmd
}

func listGatewaysRun(opts *ListGatewaysOpts) error {
	params := &preview_secret_service.ListGatewayPoolGatewaysParams{
		Context:         opts.Ctx,
		ProjectID:       opts.Profile.ProjectID,
		OrganizationID:  opts.Profile.OrganizationID,
		GatewayPoolName: opts.GatewayPoolName,
	}

	resp, err := opts.PreviewClient.ListGatewayPoolGateways(params, nil)
	if err != nil {
		return fmt.Errorf("failed to list gateway pools: %w", err)
	}

	d := newDisplayer(false).Gateways(resp.Payload.Gateways...)

	return opts.Output.Display(d.AddFields(listGatewaysFields(opts.ShowAll)...))
}

func listGatewaysFields(showAll bool) []format.Field {
	fields := []format.Field{
		{
			Name:        "Gateway ID",
			ValueFormat: "{{ .ID }}",
		},
		{
			Name:        "Gateway Status",
			ValueFormat: "{{ .Status }}",
		},
		{
			Name:        "Gateway Version",
			ValueFormat: "{{ .Version }}",
		},
	}
	if showAll {
		ComplimentaryFields := []format.Field{
			{
				Name:        "Gateway Hostname",
				ValueFormat: "{{ .Hostname }}",
			},
			{
				Name:        "Gateway Os",
				ValueFormat: "{{ .Os }}",
			},
			{
				Name:        "Gateway Metadata",
				ValueFormat: "{{ .Metadata }}",
			},
			{
				Name:        "Gateway Status Message",
				ValueFormat: "{{ .StatusMessage }}",
			},
		}
		fields = append(fields, ComplimentaryFields...)
	}

	return fields
}
