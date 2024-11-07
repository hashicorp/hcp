// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package integrations

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/manifoldco/promptui"
	"github.com/zclconf/go-cty/cty"
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

var IntegrationProviders = maps.Keys(providerToRequiredFields)

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
		When the {{ template "mdCodeOrBold" "--config-file" }} flag is specified, the configuration for your integration will be read
		from the provided HCL config file. The following fields are required: [type details]. For help populating the details for an 
		integration type, please refer to the 
		{{ Link "API reference documentation" "https://developer.hashicorp.com/hcp/api-docs/vault-secrets/2023-11-28" }}.
		When the {{ template "mdCodeOrBold" "--config-file" }} 
		flag is not specified, you will be prompted to create the integration interactively.
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

type integrationConfigInternal struct {
	Details map[string]any
}

type IntegrationConfig struct {
	Type    IntegrationType `hcl:"type"`
	Details cty.Value       `hcl:"details"`
}

var (
	TwilioKeys   = []string{"account_sid", "api_key_secret", "api_key_sid"}
	MongoKeys    = []string{"private_key", "public_key"}
	AWSKeys      = []string{"access_keys", "federated_workload_identity"}
	GCPKeys      = []string{"service_account_key", "federated_workload_identity"}
	PostgresKeys = []string{"connection_string"}
)

var providerToRequiredFields = map[string][]string{
	string(Twilio):       TwilioKeys,
	string(MongoDBAtlas): MongoKeys,
	string(AWS):          AWSKeys,
	string(GCP):          GCPKeys,
	string(Postgres):     PostgresKeys,
}

var awsAuthMethodsToReqKeys = map[string][]string{
	"federated_workload_identity": {"audience", "role_arn"},
	"access_keys":                 {"access_key_id", "secret_access_key"},
}

var gcpAuthMethodsToReqKeys = map[string][]string{
	"federated_workload_identity": {"audience", "service_account_email"},
	"service_account_key":         {"credentials"},
}

func createRun(opts *CreateOpts) error {
	var (
		config         IntegrationConfig
		internalConfig integrationConfigInternal
		err            error
	)

	if opts.ConfigFilePath == "" {
		config, internalConfig, err = promptUserForConfig(opts)
		if err != nil {
			return fmt.Errorf("failed to create integration via cli prompt: %w", err)
		}
	} else {
		if err = hclsimple.DecodeFile(opts.ConfigFilePath, nil, &config); err != nil {
			return fmt.Errorf("failed to decode config file: %w", err)
		}

		detailsMap, err := CtyValueToMap(config.Details)
		if err != nil {
			return fmt.Errorf("failed to process config file: %w", err)
		}
		internalConfig.Details = detailsMap
	}

	switch config.Type {
	case Twilio:
		req := preview_secret_service.NewCreateTwilioIntegrationParamsWithContext(opts.Ctx)
		req.OrganizationID = opts.Profile.OrganizationID
		req.ProjectID = opts.Profile.ProjectID

		var twilioBody preview_models.SecretServiceCreateTwilioIntegrationBody
		detailBytes, err := json.Marshal(internalConfig.Details)
		if err != nil {
			return fmt.Errorf("error marshaling details config: %w", err)
		}

		err = twilioBody.UnmarshalBinary(detailBytes)
		if err != nil {
			return fmt.Errorf("error marshaling details config: %w", err)
		}
		req.Body = &twilioBody
		req.Body.Name = opts.IntegrationName

		_, err = opts.PreviewClient.CreateTwilioIntegration(req, nil)
		if err != nil {
			return fmt.Errorf("failed to create Twilio integration: %w", err)
		}

	case MongoDBAtlas:
		req := preview_secret_service.NewCreateMongoDBAtlasIntegrationParamsWithContext(opts.Ctx)
		req.OrganizationID = opts.Profile.OrganizationID
		req.ProjectID = opts.Profile.ProjectID

		var mongoDBBody preview_models.SecretServiceCreateMongoDBAtlasIntegrationBody
		detailBytes, err := json.Marshal(internalConfig.Details)
		if err != nil {
			return fmt.Errorf("error marshaling details config: %w", err)
		}

		err = mongoDBBody.UnmarshalBinary(detailBytes)
		if err != nil {
			return fmt.Errorf("error marshaling details config: %w", err)
		}
		req.Body = &mongoDBBody
		req.Body.Name = opts.IntegrationName

		_, err = opts.PreviewClient.CreateMongoDBAtlasIntegration(req, nil)
		if err != nil {
			return fmt.Errorf("failed to create MongoDB Atlas integration: %w", err)
		}

	case AWS:
		req := preview_secret_service.NewCreateAwsIntegrationParamsWithContext(opts.Ctx)
		req.OrganizationID = opts.Profile.OrganizationID
		req.ProjectID = opts.Profile.ProjectID

		var awsBody preview_models.SecretServiceCreateAwsIntegrationBody
		detailBytes, err := json.Marshal(internalConfig.Details)
		if err != nil {
			return fmt.Errorf("error marshaling details config: %w", err)
		}

		err = awsBody.UnmarshalBinary(detailBytes)
		if err != nil {
			return fmt.Errorf("error marshaling details config: %w", err)
		}
		req.Body = &awsBody
		req.Body.Name = opts.IntegrationName

		_, err = opts.PreviewClient.CreateAwsIntegration(req, nil)
		if err != nil {
			return fmt.Errorf("failed to create AWS integration: %w", err)
		}

	case GCP:
		req := preview_secret_service.NewCreateGcpIntegrationParamsWithContext(opts.Ctx)
		req.OrganizationID = opts.Profile.OrganizationID
		req.ProjectID = opts.Profile.ProjectID

		var gcpBody preview_models.SecretServiceCreateGcpIntegrationBody
		detailBytes, err := json.Marshal(internalConfig.Details)
		if err != nil {
			return fmt.Errorf("error marshaling details config: %w", err)
		}

		err = gcpBody.UnmarshalBinary(detailBytes)
		if err != nil {
			return fmt.Errorf("error marshaling details config: %w", err)
		}
		req.Body = &gcpBody
		req.Body.Name = opts.IntegrationName

		_, err = opts.PreviewClient.CreateGcpIntegration(req, nil)
		if err != nil {
			return fmt.Errorf("failed to create GCP integration: %w", err)
		}

	case Postgres:
		req := preview_secret_service.NewCreatePostgresIntegrationParamsWithContext(opts.Ctx)
		req.OrganizationID = opts.Profile.OrganizationID
		req.ProjectID = opts.Profile.ProjectID

		detailBytes, err := json.Marshal(internalConfig.Details)
		if err != nil {
			return fmt.Errorf("error marshaling details config: %w", err)
		}

		var body preview_models.SecretServiceCreatePostgresIntegrationBody
		err = body.UnmarshalBinary(detailBytes)
		if err != nil {
			return fmt.Errorf("error marshaling details config: %w", err)
		}

		req.Body = &body
		req.Body.Name = opts.IntegrationName

		_, err = opts.PreviewClient.CreatePostgresIntegration(req, nil)
		if err != nil {
			return fmt.Errorf("failed to create MongoDB Atlas integration: %w", err)
		}

	default:
		return fmt.Errorf("unsupported integration provider type")
	}

	fmt.Fprintln(opts.IO.Err())
	fmt.Fprintf(opts.IO.Err(), "%s Successfully created integration with name %q\n", opts.IO.ColorScheme().SuccessIcon(), opts.IntegrationName)

	return nil
}

func promptUserForConfig(opts *CreateOpts) (IntegrationConfig, integrationConfigInternal, error) {
	var (
		config         IntegrationConfig
		internalConfig integrationConfigInternal
	)

	if !opts.IO.CanPrompt() {
		return config, internalConfig, fmt.Errorf("unable to create integration interactively")
	}

	providerPrompt := promptui.Select{
		Label:  "Please select the provider you would like to configure",
		Items:  IntegrationProviders,
		Stdin:  io.NopCloser(opts.IO.In()),
		Stdout: iostreams.NopWriteCloser(opts.IO.Err()),
	}

	_, provider, err := providerPrompt.Run()
	if err != nil {
		return config, internalConfig, fmt.Errorf("provider selection prompt failed: %w", err)
	}
	config.Type = IntegrationType(provider)

	var (
		fields     []string
		authMethod string
	)
	if config.Type == AWS {
		authPrompt := promptui.Select{
			Label:  "Please select an authentication method",
			Items:  providerToRequiredFields[provider],
			Stdin:  io.NopCloser(opts.IO.In()),
			Stdout: iostreams.NopWriteCloser(opts.IO.Err()),
		}

		_, authMethod, err = authPrompt.Run()
		if err != nil {
			return config, internalConfig, fmt.Errorf("authentication method selection prompt failed: %w", err)
		}

		fields = awsAuthMethodsToReqKeys[authMethod]
	} else if config.Type == GCP {
		authPrompt := promptui.Select{
			Label:  "Please select an authentication method",
			Items:  providerToRequiredFields[provider],
			Stdin:  io.NopCloser(opts.IO.In()),
			Stdout: iostreams.NopWriteCloser(opts.IO.Err()),
		}

		_, authMethod, err = authPrompt.Run()
		if err != nil {
			return config, internalConfig, fmt.Errorf("authentication method selection prompt failed: %w", err)
		}
		fields = gcpAuthMethodsToReqKeys[authMethod]

	} else {
		fields = providerToRequiredFields[provider]
	}

	var fieldPrompt promptui.Prompt
	fieldValues := make(map[string]any)
	for _, field := range fields {
		fieldPrompt = promptui.Prompt{
			Label: field,
			Mask:  '*',
		}

		input, err := fieldPrompt.Run()
		if err != nil {
			return config, internalConfig, fmt.Errorf("prompt for field %s failed: %w", field, err)
		}

		fieldValues[field] = input
	}

	if config.Type == AWS || config.Type == GCP {
		internalConfig.Details = map[string]any{authMethod: fieldValues}
		return config, internalConfig, err
	}

	internalConfig.Details = fieldValues
	return config, internalConfig, err
}

func CtyValueToMap(value cty.Value) (map[string]any, error) {
	fieldsMap := make(map[string]any)
	for k, v := range value.AsValueMap() {
		if v.Type() == cty.String {
			fieldsMap[k] = v.AsString()
		} else if v.Type() == cty.Bool {
			fieldsMap[k] = v.True()
		} else if v.Type().IsObjectType() {
			nestedMap, err := CtyValueToMap(v)
			if err != nil {
				return nil, err
			}
			fieldsMap[k] = nestedMap
		} else if v.Type().IsTupleType() {
			// Check the type of the first element in the slice
			// (we will assume all other elements in slice are of the same type)
			if v.AsValueSlice()[0].Type() == cty.String {
				var items []string
				for _, val := range v.AsValueSlice() {
					items = append(items, val.AsString())
				}
				fieldsMap[k] = items
			} else {
				var items []map[string]any
				for _, val := range v.AsValueSlice() {
					nestedMap, err := CtyValueToMap(val)
					if err != nil {
						return nil, err
					}
					items = append(items, nestedMap)
				}
				fieldsMap[k] = items
			}
		} else {
			return nil, fmt.Errorf("found unsupported value type")
		}
	}

	return fieldsMap, nil
}
