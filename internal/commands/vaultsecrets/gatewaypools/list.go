// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package gatewaypools

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-11-28/client/secret_service"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
)

type ListOpts struct {
	Ctx     context.Context
	Profile *profile.Profile
	Output  *format.Outputter
	IO      iostreams.IOStreams

	Client secret_service.ClientService
}

func NewCmdList(ctx *cmd.Context, runF func(*ListOpts) error) *cmd.Command {
	opts := &ListOpts{
		Ctx:     ctx.ShutdownCtx,
		Profile: ctx.Profile,
		IO:      ctx.IO,
		Output:  ctx.Output,
		Client:  secret_service.New(ctx.HCP, nil),
	}

	cmd := &cmd.Command{
		Name:      "list",
		ShortHelp: "List Vault Secrets gateway pools.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp vault-secrets gateway-pools list" }} command lists all Vault Secrets gateway pools.
		`),
		Examples: []cmd.Example{
			{
				Preamble: `List gateway-pools:`,
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				$ hcp vault-secrets gateway-pools list
				`),
			},
		},
		RunF: func(c *cmd.Command, args []string) error {
			if runF != nil {
				return runF(opts)
			}
			return listRun(opts)
		},
	}
	return cmd
}

func listFields() []format.Field {
	return []format.Field{
		{
			Name:        "Gateway Pool Name",
			ValueFormat: "{{ .Name }}",
		},
		{
			Name:        "Gateway Pool Resource Name",
			ValueFormat: "{{ .ResourceName }}",
		},
		{
			Name:        "Gateway Pool Resource ID",
			ValueFormat: "{{ .ResourceID }}",
		},
		{
			Name:        "Description",
			ValueFormat: "{{ .Description }}",
		},
	}
}

func listRun(opts *ListOpts) error {
	params := &secret_service.ListGatewayPoolsParams{
		Context:        opts.Ctx,
		ProjectID:      opts.Profile.ProjectID,
		OrganizationID: opts.Profile.OrganizationID,
	}

	resp, err := opts.Client.ListGatewayPools(params, nil)
	if err != nil {
		return fmt.Errorf("failed to list gateway pools: %w", err)
	}

	return opts.Output.Display(newDisplayer(false).GatewayPools(resp.Payload.GatewayPools...).AddFields(listFields()...))
}
