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

type DeleteOpts struct {
	Ctx     context.Context
	Profile *profile.Profile
	Output  *format.Outputter
	IO      iostreams.IOStreams

	GatewayPoolName string
	Client          secret_service.ClientService
}

func NewCmdDelete(ctx *cmd.Context, runF func(*DeleteOpts) error) *cmd.Command {
	opts := &DeleteOpts{
		Ctx:     ctx.ShutdownCtx,
		Profile: ctx.Profile,
		Output:  ctx.Output,
		IO:      ctx.IO,
		Client:  secret_service.New(ctx.HCP, nil),
	}

	cmd := &cmd.Command{
		Name:      "delete",
		ShortHelp: "Delete a Vault Secrets gateway pool.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp vault-secrets gateway-pools delete" }} command deletes a Vault Secrets gateway pool.
		`),
		Examples: []cmd.Example{
			{
				Preamble: `Delete a gateway pool:`,
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				$ hcp vault-secrets gateway-pools delete company-tunnel
				`),
			},
		},
		Args: cmd.PositionalArguments{
			Args: []cmd.PositionalArgument{
				{
					Name:          "NAME",
					Documentation: "The name of the gateway pool to delete.",
				},
			},
		},
		RunF: func(c *cmd.Command, args []string) error {
			opts.GatewayPoolName = args[0]

			if runF != nil {
				return runF(opts)
			}
			return deleteRun(opts)
		},
	}
	cmd.Args.Autocomplete = PredictGatewayPoolName(ctx, cmd, secret_service.New(ctx.HCP, nil))

	return cmd
}

func deleteRun(opts *DeleteOpts) error {
	_, err := opts.Client.DeleteGatewayPool(&secret_service.DeleteGatewayPoolParams{
		Context:         opts.Ctx,
		OrganizationID:  opts.Profile.OrganizationID,
		ProjectID:       opts.Profile.ProjectID,
		GatewayPoolName: opts.GatewayPoolName,
	}, nil)

	if err != nil {
		return fmt.Errorf("failed to delete gateway pool: %w", err)
	}

	fmt.Fprintf(opts.IO.Err(), "%s Successfully deleted the gateway pool with name %q\n", opts.IO.ColorScheme().SuccessIcon(), opts.GatewayPoolName)
	return nil
}
