// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package integrations

import (
	"context"
	"fmt"

	preview_secret_service "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/preview/2023-11-28/client/secret_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-06-13/client/secret_service"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/flagvalue"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
)

type ReadOpts struct {
	Ctx     context.Context
	Profile *profile.Profile
	Output  *format.Outputter
	IO      iostreams.IOStreams

	IntegrationName string
	Type            IntegrationType
	Client          secret_service.ClientService
	PreviewClient   preview_secret_service.ClientService
}

func NewCmdRead(ctx *cmd.Context, runF func(*ReadOpts) error) *cmd.Command {
	opts := &ReadOpts{
		Ctx:           ctx.ShutdownCtx,
		Profile:       ctx.Profile,
		Output:        ctx.Output,
		IO:            ctx.IO,
		Client:        secret_service.New(ctx.HCP, nil),
		PreviewClient: preview_secret_service.New(ctx.HCP, nil),
	}

	cmd := &cmd.Command{
		Name:      "read",
		ShortHelp: "Read a Vault Secrets integration.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp vault-secrets integrations read" }} command gets a Vault Secrets integration.
		`),
		Examples: []cmd.Example{
			{
				Preamble: `Read an integration:`,
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				$ hcp vault-secrets integrations read sample-integration --type twilio
				`),
			},
		},
		Args: cmd.PositionalArguments{
			Args: []cmd.PositionalArgument{
				{
					Name:          "NAME",
					Documentation: "The name of the integration to read.",
				},
			},
		},
		Flags: cmd.Flags{
			Local: []*cmd.Flag{
				{
					Name:         "type",
					DisplayValue: "TYPE",
					Description:  "The type of the integration to read.",
					Value:        flagvalue.Simple("", &opts.Type),
				},
			},
		},
		RunF: func(c *cmd.Command, args []string) error {
			opts.IntegrationName = args[0]

			if runF != nil {
				return runF(opts)
			}
			return readRun(opts)
		},
	}

	return cmd
}

func readRun(opts *ReadOpts) error {
	switch opts.Type {
	case "":
		resp, err := opts.PreviewClient.GetIntegration(&preview_secret_service.GetIntegrationParams{
			Context:        opts.Ctx,
			ProjectID:      opts.Profile.ProjectID,
			OrganizationID: opts.Profile.OrganizationID,
			Name:           opts.IntegrationName,
		}, nil)
		if err != nil {
			return fmt.Errorf("failed to read integration: %w", err)
		}

		return opts.Output.Display(newGenericDisplayer(true, resp.Payload.Integration))

	case Twilio:
		resp, err := opts.PreviewClient.GetTwilioIntegration(&preview_secret_service.GetTwilioIntegrationParams{
			Context:        opts.Ctx,
			ProjectID:      opts.Profile.ProjectID,
			OrganizationID: opts.Profile.OrganizationID,
			Name:           opts.IntegrationName,
		}, nil)
		if err != nil {
			return fmt.Errorf("failed to read integration: %w", err)
		}

		return opts.Output.Display(newTwilioDisplayer(true, resp.Payload.Integration))

	case MongoDBAtlas:
		resp, err := opts.PreviewClient.GetMongoDBAtlasIntegration(&preview_secret_service.GetMongoDBAtlasIntegrationParams{
			Context:        opts.Ctx,
			ProjectID:      opts.Profile.ProjectID,
			OrganizationID: opts.Profile.OrganizationID,
			Name:           opts.IntegrationName,
		}, nil)
		if err != nil {
			return fmt.Errorf("failed to read integration: %w", err)
		}

		return opts.Output.Display(newMongoDBDisplayer(true, resp.Payload.Integration))

	case AWS:
		resp, err := opts.PreviewClient.GetAwsIntegration(&preview_secret_service.GetAwsIntegrationParams{
			Context:        opts.Ctx,
			ProjectID:      opts.Profile.ProjectID,
			OrganizationID: opts.Profile.OrganizationID,
			Name:           opts.IntegrationName,
		}, nil)
		if err != nil {
			return fmt.Errorf("failed to read integration: %w", err)
		}

		return opts.Output.Display(newAwsDisplayer(true, resp.Payload.Integration.FederatedWorkloadIdentity != nil, resp.Payload.Integration))

	case GCP:
		resp, err := opts.PreviewClient.GetGcpIntegration(&preview_secret_service.GetGcpIntegrationParams{
			Context:        opts.Ctx,
			ProjectID:      opts.Profile.ProjectID,
			OrganizationID: opts.Profile.OrganizationID,
			Name:           opts.IntegrationName,
		}, nil)
		if err != nil {
			return fmt.Errorf("failed to read integration: %w", err)
		}

		return opts.Output.Display(newGcpDisplayer(true, resp.Payload.Integration.FederatedWorkloadIdentity != nil, resp.Payload.Integration))

	default:
		return fmt.Errorf("not a valid integration type")
	}
}
