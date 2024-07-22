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

type ListOpts struct {
	Ctx     context.Context
	Profile *profile.Profile
	Output  *format.Outputter
	IO      iostreams.IOStreams

	PreviewClient preview_secret_service.ClientService
}

func NewCmdList(ctx *cmd.Context, runF func(*ListOpts) error) *cmd.Command {
	opts := &ListOpts{
		Ctx:           ctx.ShutdownCtx,
		Profile:       ctx.Profile,
		IO:            ctx.IO,
		Output:        ctx.Output,
		PreviewClient: preview_secret_service.New(ctx.HCP, nil),
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

func listRun(opts *ListOpts) error {
	params := &preview_secret_service.ListGatewayPoolsParams{
		Context:        opts.Ctx,
		ProjectID:      opts.Profile.ProjectID,
		OrganizationID: opts.Profile.OrganizationID,
	}

	resp, err := opts.PreviewClient.ListGatewayPools(params, nil)
	if err != nil {
		return fmt.Errorf("failed to list gateway pools: %w", err)
	}
	if resp.Payload == nil || resp.Payload.GatewayPools == nil {
		return fmt.Errorf("failed to list gateway pools: empty response")
	}

	return opts.Output.Display(newDisplayer(false, resp.Payload.GatewayPools...))
}
