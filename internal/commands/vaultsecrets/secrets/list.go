// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package secrets

import (
	"context"
	"fmt"

	preview_secret_service "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/preview/2023-11-28/client/secret_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/preview/2023-11-28/models"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-06-13/client/secret_service"

	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
)

func NewCmdList(ctx *cmd.Context, runF func(*ListOpts) error) *cmd.Command {
	opts := &ListOpts{
		Ctx:           ctx.ShutdownCtx,
		Profile:       ctx.Profile,
		IO:            ctx.IO,
		Output:        ctx.Output,
		PreviewClient: preview_secret_service.New(ctx.HCP, nil),
		Client:        secret_service.New(ctx.HCP, nil),
	}

	cmd := &cmd.Command{
		Name:      "list",
		ShortHelp: "List all active secrets.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp vault-secrets secrets list" }} command list all active secrets under an Vault Secrets application.

		Once the secrets are listed, it can be read using
		{{ template "mdCodeOrBold" "hcp vault-secrets secrets read" }} subcommand.
		`),
		Examples: []cmd.Example{
			{
				Preamble: `List all secrets under Vault Secrets application on active profile:`,
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				$ hcp vault-secrets secrets list
				`),
			},
			{
				Preamble: `List all secrets under the specified Vault Secrets application:`,
				Command: heredoc.New(ctx.IO, heredoc.WithNoWrap()).Must(`
				$ hcp vault-secrets secrets list --app test-app
				`),
			},
		},
		RunF: func(c *cmd.Command, args []string) error {
			opts.AppName = appName

			if runF != nil {
				return runF(opts)
			}
			return listRun(opts)
		},
	}

	return cmd
}

type ListOpts struct {
	Ctx     context.Context
	Profile *profile.Profile
	IO      iostreams.IOStreams
	Output  *format.Outputter

	AppName       string
	PreviewClient preview_secret_service.ClientService
	Client        secret_service.ClientService
}

func listRun(opts *ListOpts) error {
	req := preview_secret_service.NewListAppSecretsParamsWithContext(opts.Ctx)
	req.OrganizationID = opts.Profile.OrganizationID
	req.ProjectID = opts.Profile.ProjectID
	req.AppName = opts.AppName

	var secrets []*models.Secrets20231128Secret
	listResp, err := opts.PreviewClient.ListAppSecrets(req, nil)
	if err != nil {
		return fmt.Errorf("failed to list secrets: %w", err)
	}
	secrets = append(secrets, listResp.Payload.Secrets...)

	for listResp.GetPayload().Pagination != nil && listResp.GetPayload().Pagination.NextPageToken != "" {
		req.PaginationNextPageToken = &listResp.Payload.Pagination.NextPageToken
		listResp, err = opts.PreviewClient.ListAppSecrets(req, nil)
		if err != nil {
			return fmt.Errorf("failed to list paginated secrets: %w", err)
		}
		secrets = append(secrets, listResp.GetPayload().Secrets...)
	}
	return opts.Output.Display(newDisplayerPreview(false, secrets...))
}
