// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package integrations

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/hcl/v2/hclsimple"
	preview_models "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/preview/2023-11-28/models"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/flagvalue"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"

	preview_secret_service "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/preview/2023-11-28/client/secret_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-06-13/client/secret_service"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
)

type UpdateOpts struct {
	Ctx     context.Context
	Profile *profile.Profile
	IO      iostreams.IOStreams
	Output  *format.Outputter

	IntegrationName string
	ConfigFilePath  string
	PreviewClient   preview_secret_service.ClientService
	Client          secret_service.ClientService
}

func NewCmdUpdate(ctx *cmd.Context, runF func(*UpdateOpts) error) *cmd.Command {
	opts := &UpdateOpts{
		Ctx:     ctx.ShutdownCtx,
		Profile: ctx.Profile,
		IO:      ctx.IO,
		Output:  ctx.Output,

		PreviewClient: preview_secret_service.New(ctx.HCP, nil),
		Client:        secret_service.New(ctx.HCP, nil),
	}

	cmd := &cmd.Command{
		Name:      "update",
		ShortHelp: "Update an integration.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp vault-secrets integrations update" }} command updates a Vault Secrets integration.
		The configuration for updating your integration will be read from the provided HCL config file. The following fields are 
		required: [type details]. For help populating the details for an integration type, please refer to the 
		{{ Link "API reference documentation" "https://developer.hashicorp.com/hcp/api-docs/vault-secrets/2023-11-28" }}.
		`),
		Examples: []cmd.Example{
			{
				Preamble: `Update a Vault Secrets integration:`,
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				$ hcp vault-secrets integrations update sample-integration --config-file=path-to-file/config.hcl
				`),
			},
		},
		Args: cmd.PositionalArguments{
			Args: []cmd.PositionalArgument{
				{
					Name:          "NAME",
					Documentation: "The name of the integration to update.",
				},
			},
		},
		Flags: cmd.Flags{
			Local: []*cmd.Flag{
				{
					Name:         "config-file",
					DisplayValue: "CONFIG_FILE",
					Description:  "File path to read integration config data.",
					Value:        flagvalue.Simple("", &opts.ConfigFilePath),
					Required:     true,
				},
			},
		},
		RunF: func(c *cmd.Command, args []string) error {
			opts.IntegrationName = args[0]

			if runF != nil {
				return runF(opts)
			}
			return updateRun(opts)
		},
	}

	return cmd
}

func updateRun(opts *UpdateOpts) error {
	var (
		config         IntegrationConfig
		internalConfig integrationConfigInternal
	)

	if err := hclsimple.DecodeFile(opts.ConfigFilePath, nil, &config); err != nil {
		return fmt.Errorf("failed to decode config file: %w", err)
	}

	detailsMap, err := CtyValueToMap(config.Details)
	if err != nil {
		return fmt.Errorf("failed to process config file: %w", err)
	}
	internalConfig.Details = detailsMap

	switch config.Type {
	case Twilio:
		req := preview_secret_service.NewUpdateTwilioIntegrationParamsWithContext(opts.Ctx)
		req.OrganizationID = opts.Profile.OrganizationID
		req.ProjectID = opts.Profile.ProjectID
		req.Name = opts.IntegrationName

		var twilioBody preview_models.SecretServiceUpdateTwilioIntegrationBody
		detailBytes, err := json.Marshal(internalConfig.Details)
		if err != nil {
			return fmt.Errorf("error marshaling details config: %w", err)
		}

		err = twilioBody.UnmarshalBinary(detailBytes)
		if err != nil {
			return fmt.Errorf("error marshaling details config: %w", err)
		}
		req.Body = &twilioBody

		_, err = opts.PreviewClient.UpdateTwilioIntegration(req, nil)
		if err != nil {
			return fmt.Errorf("failed to update Twilio integration: %w", err)
		}

	case MongoDBAtlas:
		req := preview_secret_service.NewUpdateMongoDBAtlasIntegrationParamsWithContext(opts.Ctx)
		req.OrganizationID = opts.Profile.OrganizationID
		req.ProjectID = opts.Profile.ProjectID
		req.Name = opts.IntegrationName

		var mongoDBBody preview_models.SecretServiceUpdateMongoDBAtlasIntegrationBody
		detailBytes, err := json.Marshal(internalConfig.Details)
		if err != nil {
			return fmt.Errorf("error marshaling details config: %w", err)
		}

		err = mongoDBBody.UnmarshalBinary(detailBytes)
		if err != nil {
			return fmt.Errorf("error marshaling details config: %w", err)
		}
		req.Body = &mongoDBBody

		_, err = opts.PreviewClient.UpdateMongoDBAtlasIntegration(req, nil)
		if err != nil {
			return fmt.Errorf("failed to update MongoDB Atlas integration: %w", err)
		}

	case AWS:
		req := preview_secret_service.NewUpdateAwsIntegrationParamsWithContext(opts.Ctx)
		req.OrganizationID = opts.Profile.OrganizationID
		req.ProjectID = opts.Profile.ProjectID
		req.Name = opts.IntegrationName

		var awsBody preview_models.SecretServiceUpdateAwsIntegrationBody
		detailBytes, err := json.Marshal(internalConfig.Details)
		if err != nil {
			return fmt.Errorf("error marshaling details config: %w", err)
		}

		err = awsBody.UnmarshalBinary(detailBytes)
		if err != nil {
			return fmt.Errorf("error marshaling details config: %w", err)
		}
		req.Body = &awsBody

		_, err = opts.PreviewClient.UpdateAwsIntegration(req, nil)
		if err != nil {
			return fmt.Errorf("failed to update AWS integration: %w", err)
		}

	case GCP:
		req := preview_secret_service.NewUpdateGcpIntegrationParamsWithContext(opts.Ctx)
		req.OrganizationID = opts.Profile.OrganizationID
		req.ProjectID = opts.Profile.ProjectID
		req.Name = opts.IntegrationName

		var gcpBody preview_models.SecretServiceUpdateGcpIntegrationBody
		detailBytes, err := json.Marshal(internalConfig.Details)
		if err != nil {
			return fmt.Errorf("error marshaling details config: %w", err)
		}

		err = gcpBody.UnmarshalBinary(detailBytes)
		if err != nil {
			return fmt.Errorf("error marshaling details config: %w", err)
		}
		req.Body = &gcpBody

		_, err = opts.PreviewClient.UpdateGcpIntegration(req, nil)
		if err != nil {
			return fmt.Errorf("failed to update GCP integration: %w", err)
		}

	case Postgres:
		req := preview_secret_service.NewUpdatePostgresIntegrationParamsWithContext(opts.Ctx)
		req.OrganizationID = opts.Profile.OrganizationID
		req.ProjectID = opts.Profile.ProjectID
		req.Name = opts.IntegrationName

		var body preview_models.SecretServiceUpdatePostgresIntegrationBody
		detailBytes, err := json.Marshal(internalConfig.Details)
		if err != nil {
			return fmt.Errorf("error marshaling details config: %w", err)
		}

		err = body.UnmarshalBinary(detailBytes)
		if err != nil {
			return fmt.Errorf("error marshaling details config: %w", err)
		}
		req.Body = &body

		_, err = opts.PreviewClient.UpdatePostgresIntegration(req, nil)
		if err != nil {
			return fmt.Errorf("failed to update MongoDB Atlas integration: %w", err)
		}
	}

	fmt.Fprintln(opts.IO.Err())
	fmt.Fprintf(opts.IO.Err(), "%s Successfully updated integration with name %q\n", opts.IO.ColorScheme().SuccessIcon(), opts.IntegrationName)

	return nil
}
