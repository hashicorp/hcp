// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: MPL-2.0

package versions

import (
	"context"
	"fmt"

	secret_service "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-11-28/client/secret_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-11-28/models"
	"github.com/hashicorp/hcp/internal/commands/vaultsecrets/secrets/appname"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
)

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
		ShortHelp: "List a secret's versions.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp vault-secrets secrets versions list" }} command lists all versions for a secret under a Vault Secrets application.
		`),
		Args: cmd.PositionalArguments{
			Args: []cmd.PositionalArgument{
				{
					Name:          "NAME",
					Documentation: "The name of the secret.",
				},
			},
		},
		Examples: []cmd.Example{
			{
				Preamble: `List all versions of a secret under the Vault Secrets application on active profile:`,
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				$ hcp vault-secrets secrets versions test_secret
				`),
			},
			{
				Preamble: `List all versions of a secret under the specified Vault Secrets application:`,
				Command: heredoc.New(ctx.IO, heredoc.WithNoWrap()).Must(`
				$ hcp vault-secrets secrets versions test_secret --app test-app
				`),
			},
		},
		RunF: func(c *cmd.Command, args []string) error {
			opts.SecretName = args[0]
			opts.AppName = appname.Get()
			if runF != nil {
				return runF(opts)
			}
			return versionsRun(opts)
		},
	}

	return cmd
}

type ListOpts struct {
	Ctx     context.Context
	Profile *profile.Profile
	IO      iostreams.IOStreams
	Output  *format.Outputter

	AppName    string
	SecretName string
	Client     secret_service.ClientService
}

func versionsRun(opts *ListOpts) error {
	req := secret_service.NewListAppSecretVersionsParamsWithContext(opts.Ctx)
	req.OrganizationID = opts.Profile.OrganizationID
	req.ProjectID = opts.Profile.ProjectID
	req.AppName = opts.AppName
	req.SecretName = opts.SecretName

	var secrets []*models.Secrets20231128SecretStaticVersion
	for {
		resp, err := opts.Client.ListAppSecretVersions(req, nil)
		if err != nil {
			return fmt.Errorf("failed to versions secrets: %w", err)
		}

		if resp.GetPayload().StaticVersions != nil {
			secrets = append(secrets, resp.Payload.StaticVersions.Versions...)
			if resp.Payload.Pagination == nil || resp.Payload.Pagination.NextPageToken == "" {
				break
			}
		}
		next := resp.Payload.Pagination.NextPageToken
		req.PaginationNextPageToken = &next
	}
	return opts.Output.Display(newDisplayer(false).StaticVersions(secrets...))
}
