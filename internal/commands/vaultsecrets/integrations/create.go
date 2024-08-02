// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package integrations

import (
	"context"
	"fmt"
	"os"
	"slices"

	"golang.org/x/exp/maps"
	"gopkg.in/yaml.v3"

	preview_secret_service "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/preview/2023-11-28/client/secret_service"
	preview_models "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/preview/2023-11-28/models"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-06-13/client/secret_service"
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
	IO      iostreams.IOStreams
	Output  *format.Outputter

	IntegrationName string
	ConfigFilePath  string
	PreviewClient   preview_secret_service.ClientService
	Client          secret_service.ClientService
}

func NewCmdCreate(ctx *cmd.Context, runF func(*CreateOpts) error) *cmd.Command {
	opts := &CreateOpts{
		Ctx:     ctx.ShutdownCtx,
		Profile: ctx.Profile,
		IO:      ctx.IO,
		Output:  ctx.Output,

		PreviewClient: preview_secret_service.New(ctx.HCP, nil),
		Client:        secret_service.New(ctx.HCP, nil),
	}

	cmd := &cmd.Command{
		Name:      "create",
		ShortHelp: "Create a new integration.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp vault-secrets integrations create" }} command creates a new Vault Secrets integration.
		`),
		Examples: []cmd.Example{
			{
				Preamble: `Create a new Vault Secrets integration:`,
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				$ hcp vault-secrets integrations create sample-integration --config-file=path-to-file/config.yaml
				`),
			},
		},
		Args: cmd.PositionalArguments{
			Args: []cmd.PositionalArgument{
				{
					Name:          "NAME",
					Documentation: "The name of the integration to create.",
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
			return createRun(opts)
		},
	}

	return cmd
}

type IntegrationConfig struct {
	Version string
	Type    IntegrationType
	Details map[string]string
}

var (
	TwilioKeys = []string{"account_sid", "api_key_secret", "api_key_sid"}
	MongoKeys  = []string{"private_key", "public_key"}
	AWSKeys    = []string{"audience", "role_arn"}
	GCPKeys    = []string{"audience", "service_account_email"}
)

func createRun(opts *CreateOpts) error {
	// Open the file
	f, err := os.ReadFile(opts.ConfigFilePath)
	if err != nil {
		return fmt.Errorf("failed to open config file: %w", err)
	}

	var i IntegrationConfig
	err = yaml.Unmarshal(f, &i)
	if err != nil {
		return fmt.Errorf("failed to unmarshal config file: %w", err)
	}

	switch i.Type {
	case Twilio:
		missingField := validateDetails(i.Details, TwilioKeys)

		if missingField != "" {
			return fmt.Errorf("missing required field in the config file: %s", missingField)
		}

		body := &preview_models.SecretServiceCreateTwilioIntegrationBody{
			Name:               opts.IntegrationName,
			TwilioAccountSid:   i.Details["account_sid"],
			TwilioAPIKeySecret: i.Details["api_key_secret"],
			TwilioAPIKeySid:    i.Details["api_key_sid"],
		}

		resp, err := opts.PreviewClient.CreateTwilioIntegration(&preview_secret_service.CreateTwilioIntegrationParams{
			Context:        opts.Ctx,
			ProjectID:      opts.Profile.ProjectID,
			OrganizationID: opts.Profile.OrganizationID,
			Body:           body,
		}, nil)

		if err != nil {
			return fmt.Errorf("failed to create Twilio integration: %w", err)
		}

		fmt.Fprintln(opts.IO.Err())
		fmt.Fprintf(opts.IO.Err(), "%s Successfully created integration with name %q\n", opts.IO.ColorScheme().SuccessIcon(), resp.Payload.Integration.Name)

	case MongoDBAtlas:
		missingField := validateDetails(i.Details, MongoKeys)

		if missingField != "" {
			return fmt.Errorf("missing required field in the config file: %s", missingField)
		}

		body := &preview_models.SecretServiceCreateMongoDBAtlasIntegrationBody{
			Name:                 opts.IntegrationName,
			MongodbAPIPrivateKey: i.Details["private_key"],
			MongodbAPIPublicKey:  i.Details["public_key"],
		}

		resp, err := opts.PreviewClient.CreateMongoDBAtlasIntegration(&preview_secret_service.CreateMongoDBAtlasIntegrationParams{
			Context:        opts.Ctx,
			ProjectID:      opts.Profile.ProjectID,
			OrganizationID: opts.Profile.OrganizationID,
			Body:           body,
		}, nil)

		if err != nil {
			return fmt.Errorf("failed to create MongoDB Atlas integration: %w", err)
		}

		fmt.Fprintln(opts.IO.Err())
		fmt.Fprintf(opts.IO.Err(), "%s Successfully created integration with name %q\n", opts.IO.ColorScheme().SuccessIcon(), resp.Payload.Integration.Name)

	case AWS:
		missingField := validateDetails(i.Details, AWSKeys)

		if missingField != "" {
			return fmt.Errorf("missing required field in the config file: %s", missingField)
		}

		body := &preview_models.SecretServiceCreateAwsIntegrationBody{
			Name: opts.IntegrationName,
			FederatedWorkloadIdentity: &preview_models.Secrets20231128AwsFederatedWorkloadIdentityRequest{
				Audience: i.Details["audience"],
				RoleArn:  i.Details["role_arn"],
			},
		}

		resp, err := opts.PreviewClient.CreateAwsIntegration(&preview_secret_service.CreateAwsIntegrationParams{
			Context:        opts.Ctx,
			ProjectID:      opts.Profile.ProjectID,
			OrganizationID: opts.Profile.OrganizationID,
			Body:           body,
		}, nil)

		if err != nil {
			return fmt.Errorf("failed to create AWS integration: %w", err)
		}

		fmt.Fprintln(opts.IO.Err())
		fmt.Fprintf(opts.IO.Err(), "%s Successfully created integration with name %q\n", opts.IO.ColorScheme().SuccessIcon(), resp.Payload.Integration.Name)

	case GCP:
		missingField := validateDetails(i.Details, GCPKeys)

		if missingField != "" {
			return fmt.Errorf("missing required field in the config file: %s", missingField)
		}

		body := &preview_models.SecretServiceCreateGcpIntegrationBody{
			Name: opts.IntegrationName,
			FederatedWorkloadIdentity: &preview_models.Secrets20231128GcpFederatedWorkloadIdentityRequest{
				Audience:            i.Details["audience"],
				ServiceAccountEmail: i.Details["service_account_email"],
			},
		}

		resp, err := opts.PreviewClient.CreateGcpIntegration(&preview_secret_service.CreateGcpIntegrationParams{
			Context:        opts.Ctx,
			ProjectID:      opts.Profile.ProjectID,
			OrganizationID: opts.Profile.OrganizationID,
			Body:           body,
		}, nil)

		if err != nil {
			return fmt.Errorf("failed to create GCP integration: %w", err)
		}

		fmt.Fprintln(opts.IO.Err())
		fmt.Fprintf(opts.IO.Err(), "%s Successfully created integration with name %q\n", opts.IO.ColorScheme().SuccessIcon(), resp.Payload.Integration.Name)
	}

	return nil
}

func validateDetails(details map[string]string, requiredKeys []string) string {
	detailsKeys := maps.Keys(details)

	for _, r := range requiredKeys {
		if !slices.Contains(detailsKeys, r) {
			return r
		}
	}
	return ""
}
