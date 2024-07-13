// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package gateway_pools

import (
	"context"
	"fmt"

	preview_secret_service "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/preview/2023-11-28/client/secret_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/preview/2023-11-28/models"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/flagvalue"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
)

type CreateOpts struct {
	Ctx     context.Context
	Profile *profile.Profile
	Output  *format.Outputter
	IO      iostreams.IOStreams

	GatewayPoolName string
	Description     string
	PreviewClient   preview_secret_service.ClientService
}

func NewCmdCreate(ctx *cmd.Context, runF func(*CreateOpts) error) *cmd.Command {
	opts := &CreateOpts{
		Ctx:           ctx.ShutdownCtx,
		Profile:       ctx.Profile,
		Output:        ctx.Output,
		IO:            ctx.IO,
		PreviewClient: preview_secret_service.New(ctx.HCP, nil),
	}

	cmd := &cmd.Command{
		Name:      "create",
		ShortHelp: "Create a new Vault Secrets Gateway Pool.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp vault-secrets gateway-pools create" }} command creates a new Vault Secrets gateway pool.
		`),
		Examples: []cmd.Example{
			{
				Preamble: `Create a new gateway pool:`,
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				$ hcp vault-secrets gateway-pools create company-tunnel \
				  --description "Tunnels to corporate network."
				`),
			},
		},
		Args: cmd.PositionalArguments{
			Args: []cmd.PositionalArgument{
				{
					Name:          "NAME",
					Documentation: "The name of the gateway pool to create.",
				},
			},
		},
		Flags: cmd.Flags{
			Local: []*cmd.Flag{
				{
					Name:         "description",
					DisplayValue: "DESCRIPTION",
					Description:  "An optional description for the gateway pool to create.",
					Value:        flagvalue.Simple("", &opts.Description),
					Required:     false,
				},
			},
		},
		RunF: func(c *cmd.Command, args []string) error {
			opts.GatewayPoolName = args[0]

			if runF != nil {
				return runF(opts)
			}
			return createRun(opts)
		},
	}

	return cmd
}

func createRun(opts *CreateOpts) error {
	resp, err := opts.PreviewClient.CreateGatewayPool(&preview_secret_service.CreateGatewayPoolParams{
		Context:        opts.Ctx,
		OrganizationID: opts.Profile.OrganizationID,
		ProjectID:      opts.Profile.ProjectID,
		Body: &models.SecretServiceCreateGatewayPoolBody{
			Name:        opts.GatewayPoolName,
			Description: opts.Description,
		},
	}, nil)

	if err != nil {
		return fmt.Errorf("failed to create gateway pool: %w", err)
	}

	return opts.Output.Display(newDisplayer(true, resp.Payload.GatewayPool))
}
