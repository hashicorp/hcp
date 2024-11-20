// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package integrations

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-06-13/client/secret_service"
	preview_secret_service "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-11-28/client/secret_service"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/flagvalue"
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

	IntegrationName string
	Type            IntegrationType
	Client          secret_service.ClientService
	PreviewClient   preview_secret_service.ClientService
}

func NewCmdDelete(ctx *cmd.Context, runF func(*DeleteOpts) error) *cmd.Command {
	opts := &DeleteOpts{
		Ctx:           ctx.ShutdownCtx,
		Profile:       ctx.Profile,
		Output:        ctx.Output,
		IO:            ctx.IO,
		Client:        secret_service.New(ctx.HCP, nil),
		PreviewClient: preview_secret_service.New(ctx.HCP, nil),
	}

	cmd := &cmd.Command{
		Name:      "delete",
		ShortHelp: "Delete a Vault Secrets integration.",
		LongHelp: heredoc.New(ctx.IO).Must(fmt.Sprintf(`
      The {{ template "mdCodeOrBold" "hcp vault-secrets integrations delete" }} command deletes a Vault Secrets integration.
      The required {{ template "mdCodeOrBold" "--type" }} flag may be any of the following: %v
      `, IntegrationProviders)),
		Examples: []cmd.Example{
			{
				Preamble: `Delete an integration:`,
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
            $ hcp vault-secrets integrations delete sample-integration --type twilio
            `),
			},
		},
		Args: cmd.PositionalArguments{
			Args: []cmd.PositionalArgument{
				{
					Name:          "NAME",
					Documentation: "The name of the integration to delete.",
				},
			},
		},
		Flags: cmd.Flags{
			Local: []*cmd.Flag{
				{
					Name:         "type",
					DisplayValue: "TYPE",
					Description:  "The type of the integration to delete.",
					Value:        flagvalue.Simple("", &opts.Type),
					Required:     true,
				},
			},
		},
		RunF: func(c *cmd.Command, args []string) error {
			opts.IntegrationName = args[0]

			if runF != nil {
				return runF(opts)
			}
			return deleteRun(opts)
		},
	}

	return cmd
}

func deleteRun(opts *DeleteOpts) error {
	switch opts.Type {
	case Twilio:
		_, err := opts.PreviewClient.DeleteTwilioIntegration(&preview_secret_service.DeleteTwilioIntegrationParams{
			Context:        opts.Ctx,
			ProjectID:      opts.Profile.ProjectID,
			OrganizationID: opts.Profile.OrganizationID,
			Name:           opts.IntegrationName,
		}, nil)
		if err != nil {
			return fmt.Errorf("failed to delete integration: %w", err)
		}

	case MongoDBAtlas:
		_, err := opts.PreviewClient.DeleteMongoDBAtlasIntegration(&preview_secret_service.DeleteMongoDBAtlasIntegrationParams{
			Context:        opts.Ctx,
			ProjectID:      opts.Profile.ProjectID,
			OrganizationID: opts.Profile.OrganizationID,
			Name:           opts.IntegrationName,
		}, nil)
		if err != nil {
			return fmt.Errorf("failed to delete integration: %w", err)
		}

	case AWS:
		_, err := opts.PreviewClient.DeleteAwsIntegration(&preview_secret_service.DeleteAwsIntegrationParams{
			Context:        opts.Ctx,
			ProjectID:      opts.Profile.ProjectID,
			OrganizationID: opts.Profile.OrganizationID,
			Name:           opts.IntegrationName,
		}, nil)
		if err != nil {
			return fmt.Errorf("failed to delete integration: %w", err)
		}

	case GCP:
		_, err := opts.PreviewClient.DeleteGcpIntegration(&preview_secret_service.DeleteGcpIntegrationParams{
			Context:        opts.Ctx,
			ProjectID:      opts.Profile.ProjectID,
			OrganizationID: opts.Profile.OrganizationID,
			Name:           opts.IntegrationName,
		}, nil)
		if err != nil {
			return fmt.Errorf("failed to delete integration: %w", err)
		}

	case Postgres:
		_, err := opts.PreviewClient.DeletePostgresIntegration(&preview_secret_service.DeletePostgresIntegrationParams{
			Context:        opts.Ctx,
			ProjectID:      opts.Profile.ProjectID,
			OrganizationID: opts.Profile.OrganizationID,
			Name:           opts.IntegrationName,
		}, nil)
		if err != nil {
			return fmt.Errorf("failed to delete integration: %w", err)
		}

	default:
		return fmt.Errorf("not a valid integration type")
	}

	fmt.Fprintf(opts.IO.Err(), "%s Successfully deleted integration with name %q\n", opts.IO.ColorScheme().SuccessIcon(), opts.IntegrationName)

	return nil
}
