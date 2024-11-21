// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package integrations

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

type ListOpts struct {
	Ctx     context.Context
	Profile *profile.Profile
	Output  *format.Outputter
	IO      iostreams.IOStreams

	Type   IntegrationType
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
		ShortHelp: "List Vault Secrets integrations.",
		LongHelp: heredoc.New(ctx.IO).Must(fmt.Sprintf(`
		The {{ template "mdCodeOrBold" "hcp vault-secrets integrations list" }} command lists Vault Secrets generic integrations.
		The optional {{ template "mdCodeOrBold" "--type" }} flag may be any of the following: %v
		`, IntegrationProviders)),
		Examples: []cmd.Example{
			{
				Preamble: `List twilio integrations:`,
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				$ hcp vault-secrets integrations list --type "twilio"
				`),
			},
			{
				Preamble: `List all generic integrations:`,
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				$ hcp vault-secrets integrations list
				`),
			},
		},
		Flags: cmd.Flags{
			Local: []*cmd.Flag{
				{
					Name:         "type",
					DisplayValue: "TYPE",
					Description:  "The optional type of integration to list.",
					Value:        flagvalue.Simple("", &opts.Type),
				},
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
	if opts.Type == "" {
		var integrations []*models.Secrets20231128Integration
		params := &secret_service.ListIntegrationsParams{
			Context:        opts.Ctx,
			ProjectID:      opts.Profile.ProjectID,
			OrganizationID: opts.Profile.OrganizationID,
			Capabilities:   []string{"ROTATION", "DYNAMIC"},
		}
		for {
			resp, err := opts.Client.ListIntegrations(params, nil)
			if err != nil {
				return fmt.Errorf("failed to list integrations: %w", err)
			}

			integrations = append(integrations, resp.Payload.Integrations...)
			if resp.Payload.Pagination == nil || resp.Payload.Pagination.NextPageToken == "" {
				break
			}

			next := resp.Payload.Pagination.NextPageToken
			params.PaginationNextPageToken = &next
		}
		return opts.Output.Display(newGenericDisplayer(false, integrations...))

	}

	switch opts.Type {
	case Twilio:
		var integrations []*models.Secrets20231128TwilioIntegration

		params := &secret_service.ListTwilioIntegrationsParams{
			Context:        opts.Ctx,
			ProjectID:      opts.Profile.ProjectID,
			OrganizationID: opts.Profile.OrganizationID,
		}

		for {
			resp, err := opts.Client.ListTwilioIntegrations(params, nil)
			if err != nil {
				return fmt.Errorf("failed to list Twilio integrations: %w", err)
			}

			integrations = append(integrations, resp.Payload.Integrations...)
			if resp.Payload.Pagination == nil || resp.Payload.Pagination.NextPageToken == "" {
				break
			}

			next := resp.Payload.Pagination.NextPageToken
			params.PaginationNextPageToken = &next
		}

		return opts.Output.Display(newTwilioDisplayer(false, integrations...))

	case MongoDBAtlas:
		var integrations []*models.Secrets20231128MongoDBAtlasIntegration

		params := &secret_service.ListMongoDBAtlasIntegrationsParams{
			Context:        opts.Ctx,
			ProjectID:      opts.Profile.ProjectID,
			OrganizationID: opts.Profile.OrganizationID,
		}

		for {
			resp, err := opts.Client.ListMongoDBAtlasIntegrations(params, nil)
			if err != nil {
				return fmt.Errorf("failed to list MongoDB Atlas integrations: %w", err)
			}

			integrations = append(integrations, resp.Payload.Integrations...)
			if resp.Payload.Pagination == nil || resp.Payload.Pagination.NextPageToken == "" {
				break
			}

			next := resp.Payload.Pagination.NextPageToken
			params.PaginationNextPageToken = &next
		}

		return opts.Output.Display(newMongoDBDisplayer(false, integrations...))

	case AWS:
		var integrations []*models.Secrets20231128AwsIntegration

		params := &secret_service.ListAwsIntegrationsParams{
			Context:        opts.Ctx,
			ProjectID:      opts.Profile.ProjectID,
			OrganizationID: opts.Profile.OrganizationID,
		}

		for {
			resp, err := opts.Client.ListAwsIntegrations(params, nil)
			if err != nil {
				return fmt.Errorf("failed to list AWS integrations: %w", err)
			}

			integrations = append(integrations, resp.Payload.Integrations...)
			if resp.Payload.Pagination == nil || resp.Payload.Pagination.NextPageToken == "" {
				break
			}

			next := resp.Payload.Pagination.NextPageToken
			params.PaginationNextPageToken = &next
		}

		return opts.Output.Display(newAwsDisplayer(false, false, integrations...))

	case GCP:
		var integrations []*models.Secrets20231128GcpIntegration

		params := &secret_service.ListGcpIntegrationsParams{
			Context:        opts.Ctx,
			ProjectID:      opts.Profile.ProjectID,
			OrganizationID: opts.Profile.OrganizationID,
		}

		for {
			resp, err := opts.Client.ListGcpIntegrations(params, nil)
			if err != nil {
				return fmt.Errorf("failed to list GCP integrations: %w", err)
			}

			integrations = append(integrations, resp.Payload.Integrations...)
			if resp.Payload.Pagination == nil || resp.Payload.Pagination.NextPageToken == "" {
				break
			}

			next := resp.Payload.Pagination.NextPageToken
			params.PaginationNextPageToken = &next
		}

		return opts.Output.Display(newGcpDisplayer(false, false, integrations...))

	case Postgres:
		var integrations []*models.Secrets20231128PostgresIntegration

		params := &secret_service.ListPostgresIntegrationsParams{
			Context:        opts.Ctx,
			ProjectID:      opts.Profile.ProjectID,
			OrganizationID: opts.Profile.OrganizationID,
		}

		for {
			resp, err := opts.Client.ListPostgresIntegrations(params, nil)
			if err != nil {
				return fmt.Errorf("failed to list Postgres integrations: %w", err)
			}

			integrations = append(integrations, resp.Payload.Integrations...)
			if resp.Payload.Pagination == nil || resp.Payload.Pagination.NextPageToken == "" {
				break
			}

			next := resp.Payload.Pagination.NextPageToken
			params.PaginationNextPageToken = &next
		}

		return opts.Output.Display(newPostgresDisplayer(false, integrations...))

	default:
		return fmt.Errorf("not a valid integration type")
	}
}
