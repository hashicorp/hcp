// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package gatewaypools

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-11-28/client/secret_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-11-28/models"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/flagvalue"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
)

type UpdateOpts struct {
	Ctx     context.Context
	Profile *profile.Profile
	Output  *format.Outputter
	IO      iostreams.IOStreams

	GatewayPoolName string
	Description     string
	Client          secret_service.ClientService
}

func NewCmdUpdate(ctx *cmd.Context, runF func(*UpdateOpts) error) *cmd.Command {
	opts := &UpdateOpts{
		Ctx:     ctx.ShutdownCtx,
		Profile: ctx.Profile,
		Output:  ctx.Output,
		IO:      ctx.IO,
		Client:  secret_service.New(ctx.HCP, nil),
	}

	cmd := &cmd.Command{
		Name:      "update",
		ShortHelp: "Update a Vault Secrets gateway pool.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp vault-secrets gateway-pools update" }} command updates a Vault Secrets gateway pool.
		`),
		Examples: []cmd.Example{
			{
				Preamble: `Update a gateway pool:`,
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				$ hcp vault-secrets gateway-pools update company-tunnel --description "Extra secure tunnel for company secrets."
				`),
			},
		},
		Args: cmd.PositionalArguments{
			Args: []cmd.PositionalArgument{
				{
					Name:          "NAME",
					Documentation: "The name of the gateway pool to update.",
				},
			},
		},
		Flags: cmd.Flags{
			Local: []*cmd.Flag{
				{
					Name:         "description",
					DisplayValue: "DESCRIPTION",
					Description:  "The updated gateway pool description.",
					Value:        flagvalue.Simple("", &opts.Description),
					Required:     false,
				},
			},
		},
		RunF: func(c *cmd.Command, args []string) error {
			opts.GatewayPoolName = args[0]
			if opts.Description == "" {
				return fmt.Errorf("no fields specified for update")
			}

			if runF != nil {
				return runF(opts)
			}
			return updateRun(opts)
		},
	}

	return cmd
}

func updateRun(opts *UpdateOpts) error {
	_, err := opts.Client.UpdateGatewayPool(&secret_service.UpdateGatewayPoolParams{
		Context:         opts.Ctx,
		OrganizationID:  opts.Profile.OrganizationID,
		ProjectID:       opts.Profile.ProjectID,
		GatewayPoolName: opts.GatewayPoolName,
		Body: &models.SecretServiceUpdateGatewayPoolBody{
			Description: opts.Description,
		},
	}, nil)

	if err != nil {
		return fmt.Errorf("failed to update gateway pool: %w", err)
	}

	fmt.Fprintf(opts.IO.Err(), "%s Successfully updated gateway pool with name %q\n", opts.IO.ColorScheme().SuccessIcon(), opts.GatewayPoolName)
	return nil
}
