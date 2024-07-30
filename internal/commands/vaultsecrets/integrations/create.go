// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package integrations

import (
	"context"
	"fmt"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/preview/2023-11-28/models"

	preview_secret_service "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/preview/2023-11-28/client/secret_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-06-13/client/secret_service"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/flagvalue"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
	"gopkg.in/yaml.v3"
	"os"
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
	Type    string
	Details map[string]any
}

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

		accountSid := i.Details["account_sid"].(string)
		apiKeySecret := i.Details["api_key_secret"].(string)
		apiKeySid := i.Details["api_key_sid"].(string)

		body := &models.SecretServiceCreateTwilioIntegrationBody{
			IntegrationName:    opts.IntegrationName,
			TwilioAccountSid:   accountSid,
			TwilioAPIKeySecret: apiKeySecret,
			TwilioAPIKeySid:    apiKeySid,
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
		fmt.Fprintf(opts.IO.Err(), "%s Successfully created integration with name %q\n", opts.IO.ColorScheme().SuccessIcon(), resp.Payload.Integration.IntegrationName)

	case MongoDB:
		privkey := i.Details["private_key"].(string)
		pubKey := i.Details["public_key"].(string)
		body := &models.SecretServiceCreateMongoDBAtlasIntegrationBody{
			IntegrationName:      opts.IntegrationName,
			MongodbAPIPrivateKey: privkey,
			MongodbAPIPublicKey:  pubKey,
		}

		resp, err := opts.PreviewClient.CreateMongoDBAtlasIntegration(&preview_secret_service.CreateMongoDBAtlasIntegrationParams{
			Context:        opts.Ctx,
			ProjectID:      opts.Profile.ProjectID,
			OrganizationID: opts.Profile.OrganizationID,
			Body:           body,
		}, nil)

		if err != nil {
			return fmt.Errorf("failed to create MongoDB integration: %w", err)
		}

		fmt.Fprintln(opts.IO.Err())
		fmt.Fprintf(opts.IO.Err(), "%s Successfully created integration with name %q\n", opts.IO.ColorScheme().SuccessIcon(), resp.Payload.Integration.IntegrationName)
	}

	return nil
}
