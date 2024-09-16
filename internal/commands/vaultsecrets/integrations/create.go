// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package integrations

import (
	"context"
	"fmt"
	"github.com/manifoldco/promptui"
	"io"
	"slices"

	"golang.org/x/exp/maps"

	"github.com/hashicorp/hcl/v2/hclsimple"
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
				$ hcp vault-secrets integrations create sample-integration --config-file=path-to-file/config.hcl
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
	Type    IntegrationType   `hcl:"type"`
	Details map[string]string `hcl:"details"`
}

var (
	TwilioKeys = []string{"account_sid", "api_key_secret", "api_key_sid"}
	MongoKeys  = []string{"private_key", "public_key"}
	AWSKeys    = []string{"audience", "role_arn"}
	GCPKeys    = []string{"audience", "service_account_email"}
)

var providerToRequiredFields = map[string][]string{
	string(Twilio):       TwilioKeys,
	string(MongoDBAtlas): MongoKeys,
	string(AWS):          AWSKeys,
	string(GCP):          GCPKeys,
}

func createRun(opts *CreateOpts) error {
	var (
		ic  IntegrationConfig
		err error
	)

	if opts.ConfigFilePath == "" {
		ic, err = promptUserForConfig(opts)
		if err != nil {
			return fmt.Errorf("failed to create integration via cli prompt: %w", err)
		}
	} else {
		if err = hclsimple.DecodeFile(opts.ConfigFilePath, nil, &ic); err != nil {
			return fmt.Errorf("failed to decode config file: %w", err)
		}
	}

	switch ic.Type {
	case Twilio:
		missingFields := validateDetails(ic.Details, TwilioKeys)

		if len(missingFields) > 0 {
			return fmt.Errorf("missing required field(s) in the config file: %s", missingFields)
		}

		body := &preview_models.SecretServiceCreateTwilioIntegrationBody{
			Name: opts.IntegrationName,
			StaticCredentialDetails: &preview_models.Secrets20231128TwilioStaticCredentialsRequest{
				AccountSid:   ic.Details[TwilioKeys[0]],
				APIKeySecret: ic.Details[TwilioKeys[1]],
				APIKeySid:    ic.Details[TwilioKeys[2]],
			},
			Capabilities: []*preview_models.Secrets20231128Capability{
				preview_models.Secrets20231128CapabilityROTATION.Pointer(),
			},
		}

		_, err := opts.PreviewClient.CreateTwilioIntegration(&preview_secret_service.CreateTwilioIntegrationParams{
			Context:        opts.Ctx,
			ProjectID:      opts.Profile.ProjectID,
			OrganizationID: opts.Profile.OrganizationID,
			Body:           body,
		}, nil)

		if err != nil {
			return fmt.Errorf("failed to create Twilio integration: %w", err)
		}

	case MongoDBAtlas:
		missingFields := validateDetails(ic.Details, MongoKeys)

		if len(missingFields) > 0 {
			return fmt.Errorf("missing required field(s) in the config file: %s", missingFields)
		}

		body := &preview_models.SecretServiceCreateMongoDBAtlasIntegrationBody{
			Name: opts.IntegrationName,
			StaticCredentialDetails: &preview_models.Secrets20231128MongoDBAtlasStaticCredentialsRequest{
				APIPrivateKey: ic.Details[MongoKeys[0]],
				APIPublicKey:  ic.Details[MongoKeys[1]],
			},
			Capabilities: []*preview_models.Secrets20231128Capability{
				preview_models.Secrets20231128CapabilityROTATION.Pointer(),
			},
		}

		_, err := opts.PreviewClient.CreateMongoDBAtlasIntegration(&preview_secret_service.CreateMongoDBAtlasIntegrationParams{
			Context:        opts.Ctx,
			ProjectID:      opts.Profile.ProjectID,
			OrganizationID: opts.Profile.OrganizationID,
			Body:           body,
		}, nil)

		if err != nil {
			return fmt.Errorf("failed to create MongoDB Atlas integration: %w", err)
		}

	case AWS:
		missingFields := validateDetails(ic.Details, AWSKeys)

		if len(missingFields) > 0 {
			return fmt.Errorf("missing required field(s) in the config file: %s", missingFields)
		}

		body := &preview_models.SecretServiceCreateAwsIntegrationBody{
			Name: opts.IntegrationName,
			FederatedWorkloadIdentity: &preview_models.Secrets20231128AwsFederatedWorkloadIdentityRequest{
				Audience: ic.Details[AWSKeys[0]],
				RoleArn:  ic.Details[AWSKeys[1]],
			},
			Capabilities: []*preview_models.Secrets20231128Capability{
				preview_models.Secrets20231128CapabilityDYNAMIC.Pointer(),
			},
		}

		_, err := opts.PreviewClient.CreateAwsIntegration(&preview_secret_service.CreateAwsIntegrationParams{
			Context:        opts.Ctx,
			ProjectID:      opts.Profile.ProjectID,
			OrganizationID: opts.Profile.OrganizationID,
			Body:           body,
		}, nil)

		if err != nil {
			return fmt.Errorf("failed to create AWS integration: %w", err)
		}

	case GCP:
		missingFields := validateDetails(ic.Details, GCPKeys)

		if len(missingFields) > 0 {
			return fmt.Errorf("missing required field(s) in the config file: %s", missingFields)
		}

		body := &preview_models.SecretServiceCreateGcpIntegrationBody{
			Name: opts.IntegrationName,
			FederatedWorkloadIdentity: &preview_models.Secrets20231128GcpFederatedWorkloadIdentityRequest{
				Audience:            ic.Details[GCPKeys[0]],
				ServiceAccountEmail: ic.Details[GCPKeys[1]],
			},
			Capabilities: []*preview_models.Secrets20231128Capability{
				preview_models.Secrets20231128CapabilityDYNAMIC.Pointer(),
			},
		}

		_, err := opts.PreviewClient.CreateGcpIntegration(&preview_secret_service.CreateGcpIntegrationParams{
			Context:        opts.Ctx,
			ProjectID:      opts.Profile.ProjectID,
			OrganizationID: opts.Profile.OrganizationID,
			Body:           body,
		}, nil)

		if err != nil {
			return fmt.Errorf("failed to create GCP integration: %w", err)
		}
	}

	fmt.Fprintln(opts.IO.Err())
	fmt.Fprintf(opts.IO.Err(), "%s Successfully created integration with name %q\n", opts.IO.ColorScheme().SuccessIcon(), opts.IntegrationName)

	return nil
}

func promptUserForConfig(opts *CreateOpts) (IntegrationConfig, error) {
	var config IntegrationConfig

	if !opts.IO.CanPrompt() {
		return config, fmt.Errorf("unable to creative integration interactively")
	}

	providerPrompt := promptui.Select{
		Label:  "Please select the provider you would like to configure",
		Items:  maps.Keys(providerToRequiredFields),
		Stdin:  io.NopCloser(opts.IO.In()),
		Stdout: iostreams.NopWriteCloser(opts.IO.Err()),
	}

	_, provider, err := providerPrompt.Run()
	if err != nil {
		return config, fmt.Errorf("provider selection prompt failed: %w", err)
	}

	config.Type = IntegrationType(provider)

	var fieldPrompt promptui.Prompt
	fieldValues := make(map[string]string)
	for _, field := range providerToRequiredFields[provider] {
		fieldPrompt = promptui.Prompt{
			Label: field,
			Mask:  '*',
		}

		input, err := fieldPrompt.Run()
		if err != nil {
			return config, fmt.Errorf("Prompt for field %s failed %v\n", field, err)
		}

		fieldValues[field] = input
	}

	config.Details = fieldValues
	return config, err
}

func validateDetails(details map[string]string, requiredKeys []string) []string {
	detailsKeys := maps.Keys(details)
	var missingKeys []string

	for _, r := range requiredKeys {
		if !slices.Contains(detailsKeys, r) {
			missingKeys = append(missingKeys, r)
		}
	}
	return missingKeys
}
