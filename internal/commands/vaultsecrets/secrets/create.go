// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: MPL-2.0

package secrets

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"slices"

	"github.com/manifoldco/promptui"
	"github.com/posener/complete"
	"github.com/zclconf/go-cty/cty"
	"golang.org/x/exp/maps"

	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-11-28/client/secret_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-11-28/models"
	"github.com/hashicorp/hcp/internal/commands/vaultsecrets/integrations"
	"github.com/hashicorp/hcp/internal/commands/vaultsecrets/secrets/appname"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/flagvalue"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
)

func NewCmdCreate(ctx *cmd.Context, runF func(*CreateOpts) error) *cmd.Command {
	opts := &CreateOpts{
		Ctx:     ctx.ShutdownCtx,
		Profile: ctx.Profile,
		IO:      ctx.IO,
		Output:  ctx.Output,
		Client:  secret_service.New(ctx.HCP, nil),
	}

	cmd := &cmd.Command{
		Name:      "create",
		ShortHelp: "Create a new secret.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp vault-secrets secrets create" }} command creates a new static, rotating, or dynamic secret under a Vault Secrets application.
		The configuration for creating your rotating or dynamic secret will be read from the provided HCL config file. The following fields are required in the config 
		file: [type integration_name details]. For help populating the details for a dynamic or rotating secret, please refer to the 
		{{ Link "API reference documentation" "https://developer.hashicorp.com/hcp/api-docs/vault-secrets/2023-11-28" }}.
		`),
		Examples: []cmd.Example{
			{
				Preamble: `Create a new static secret in the Vault Secrets application on your active profile:`,
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				$ hcp vault-secrets secrets create secret_1 --data-file=tmp/secrets1.txt
				`),
			},
			{
				Preamble: `Create a new secret in a Vault Secrets application by piping the plaintext secret from a command output:`,
				Command: heredoc.New(ctx.IO, heredoc.WithNoWrap()).Must(`
				$ echo -n "my super secret" | hcp vault-secrets secrets create secret_2 --data-file=-
				`),
			},
			{
				Preamble: `Create a new rotating secret on your active profile from a config file:`,
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				$ hcp vault-secrets secrets create secret_1 --secret-type=rotating --data-file=path/to/file/config.hcl
				`),
			},
			{
				Preamble: `Create a new dynamic secret interactively on your active profile:`,
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				$ hcp vault-secrets secrets create secret_1 --secret-type=dynamic
				`),
			},
		},
		Args: cmd.PositionalArguments{
			Args: []cmd.PositionalArgument{
				{
					Name:          "NAME",
					Documentation: "The name of the secret to create.",
				},
			},
		},
		Flags: cmd.Flags{
			Local: []*cmd.Flag{
				{
					Name:         "data-file",
					DisplayValue: "DATA_FILE_PATH",
					Description:  "File path to read secret data from. Set this to '-' to read the secret data from stdin for a static secret.",
					Value:        flagvalue.Simple("", &opts.SecretFilePath),
					Autocomplete: complete.PredictOr(
						complete.PredictFiles("*"),
						complete.PredictSet("-"),
					),
				},
				{
					Name:         "secret-type",
					DisplayValue: "SECRET_TYPE",
					Description:  "The type of secret to create: static, rotating, or dynamic.",
					Value:        flagvalue.Simple("", &opts.Type),
				},
			},
		},
		RunF: func(c *cmd.Command, args []string) error {
			opts.AppName = appname.Get()
			opts.SecretName = args[0]

			if runF != nil {
				return runF(opts)
			}
			return createRun(opts)
		},
	}

	return cmd
}

type CreateOpts struct {
	Ctx     context.Context
	Profile *profile.Profile
	IO      iostreams.IOStreams
	Output  *format.Outputter

	AppName              string
	SecretName           string
	SecretValuePlaintext string
	SecretFilePath       string
	Type                 string
	Client               secret_service.ClientService
}

type secretConfigInternal struct {
	Details map[string]any
}

type SecretConfig struct {
	Type    integrations.IntegrationType `hcl:"type"`
	Details cty.Value                    `hcl:"details"`
}

var (
	twilioRotatingSecretTemplate = map[string]any{
		"integration_name":     "",
		"rotation_policy_name": "",
	}

	mongoDBAtlasRotatingSecretTemplate = map[string]any{
		"integration_name": "",
		"secret_details": map[string]any{
			"mongodb_roles": []map[string]string{
				{
					"collection_name": "",
					"database_name":   "",
					"role_name":       "",
				},
			},
			"mongodb_scopes": []map[string]string{
				{
					"name": "",
					"type": "",
				},
			},
			"mongodb_group_id": "",
		},
		"rotation_policy_name": "",
	}

	awsRotatingSecretTemplate = map[string]any{
		"integration_name":     "",
		"rotation_policy_name": "",
		"aws_iam_user_access_key_params": map[string]any{
			"username": "",
		},
	}

	gcpRotatingSecretTemplate = map[string]any{
		"integration_name":     "",
		"rotation_policy_name": "",
		"gcp_service_account_key_params": map[string]any{
			"service_account_email": "",
		},
	}

	postgresRotatingSecretTemplate = map[string]any{
		"integration_name":     "",
		"rotation_policy_name": "",
		"postgres_params": map[string]any{
			"usernames": []string{},
		},
	}

	awsDynamicSecretTemplate = map[string]any{
		"integration_name": "",
		"default_ttl":      "",
		"assume_role": map[string]any{
			"role_arn": "",
		},
	}

	gcpDynamicSecretTemplate = map[string]any{
		"integration_name": "",
		"default_ttl":      "",
		"service_account_impersonation": map[string]any{
			"service_account_email": "",
		},
	}

	optionalFields = []string{"mongodb_scopes", "collection_name"}
)

func createRun(opts *CreateOpts) error {
	switch opts.Type {
	case secretTypeKV, "":
		if err := readPlainTextSecret(opts); err != nil {
			return err
		}

		req := secret_service.NewCreateAppKVSecretParamsWithContext(opts.Ctx)
		req.OrganizationID = opts.Profile.OrganizationID
		req.ProjectID = opts.Profile.ProjectID
		req.AppName = opts.AppName

		req.Body = &models.SecretServiceCreateAppKVSecretBody{
			Name:  opts.SecretName,
			Value: opts.SecretValuePlaintext,
		}

		resp, err := opts.Client.CreateAppKVSecret(req, nil)
		if err != nil {
			return fmt.Errorf("failed to create secret with name %q: %w", opts.SecretName, err)
		}

		if err := opts.Output.Display(newDisplayer().Secrets(resp.Payload.Secret)); err != nil {
			return err
		}
	case secretTypeRotating:
		var (
			config         SecretConfig
			internalConfig secretConfigInternal
			err            error
		)

		if opts.SecretFilePath == "" {
			config, internalConfig, err = promptUserForConfig(opts)
			if err != nil {
				return fmt.Errorf("failed to create integration via cli prompt: %w", err)
			}
		} else {
			config, internalConfig, err = readConfigFile(opts.SecretFilePath)
			if err != nil {
				return fmt.Errorf("failed to process config file: %w", err)
			}
		}

		missingFields := validateSecretConfig(config)

		if len(missingFields) > 0 {
			return fmt.Errorf("missing required field(s) in the config file: %s", missingFields)
		}

		switch config.Type {
		case integrations.Twilio:
			req := secret_service.NewCreateTwilioRotatingSecretParamsWithContext(opts.Ctx)
			req.OrganizationID = opts.Profile.OrganizationID
			req.ProjectID = opts.Profile.ProjectID
			req.AppName = opts.AppName

			var twilioBody models.SecretServiceCreateTwilioRotatingSecretBody
			detailBytes, err := json.Marshal(internalConfig.Details)
			if err != nil {
				return fmt.Errorf("error marshaling details config: %w", err)
			}

			err = twilioBody.UnmarshalBinary(detailBytes)
			if err != nil {
				return fmt.Errorf("error marshaling details config: %w", err)
			}

			twilioBody.Name = opts.SecretName
			req.Body = &twilioBody

			resp, err := opts.Client.CreateTwilioRotatingSecret(req, nil)
			if err != nil {
				return fmt.Errorf("failed to create secret with name %q: %w", opts.SecretName, err)
			}

			if err := opts.Output.Display(newRotatingSecretsDisplayer(true).PreviewRotatingSecrets(resp.Payload.Config)); err != nil {
				return err
			}

		case integrations.MongoDBAtlas:

			req := secret_service.NewCreateMongoDBAtlasRotatingSecretParamsWithContext(opts.Ctx)
			req.OrganizationID = opts.Profile.OrganizationID
			req.ProjectID = opts.Profile.ProjectID
			req.AppName = opts.AppName

			var mongoDBBody models.SecretServiceCreateMongoDBAtlasRotatingSecretBody
			detailBytes, err := json.Marshal(internalConfig.Details)
			if err != nil {
				return fmt.Errorf("error marshaling details config: %w", err)
			}

			err = mongoDBBody.UnmarshalBinary(detailBytes)
			if err != nil {
				return fmt.Errorf("error marshaling details config: %w", err)
			}

			mongoDBBody.Name = opts.SecretName
			req.Body = &mongoDBBody

			resp, err := opts.Client.CreateMongoDBAtlasRotatingSecret(req, nil)
			if err != nil {
				return fmt.Errorf("failed to create secret with name %q: %w", opts.SecretName, err)
			}

			if err := opts.Output.Display(newRotatingSecretsDisplayer(true).PreviewRotatingSecrets(resp.Payload.Config)); err != nil {
				return err
			}

		case integrations.AWS:
			req := secret_service.NewCreateAwsIAMUserAccessKeyRotatingSecretParamsWithContext(opts.Ctx)
			req.OrganizationID = opts.Profile.OrganizationID
			req.ProjectID = opts.Profile.ProjectID
			req.AppName = opts.AppName

			var awsBody models.SecretServiceCreateAwsIAMUserAccessKeyRotatingSecretBody
			detailBytes, err := json.Marshal(internalConfig.Details)
			if err != nil {
				return fmt.Errorf("error marshaling details config: %w", err)
			}

			err = awsBody.UnmarshalBinary(detailBytes)
			if err != nil {
				return fmt.Errorf("error marshaling details config: %w", err)
			}

			awsBody.Name = opts.SecretName
			req.Body = &awsBody

			_, err = opts.Client.CreateAwsIAMUserAccessKeyRotatingSecret(req, nil)
			if err != nil {
				return fmt.Errorf("failed to create secret with name %q: %w", opts.SecretName, err)
			}

		case integrations.GCP:
			req := secret_service.NewCreateGcpServiceAccountKeyRotatingSecretParamsWithContext(opts.Ctx)
			req.OrganizationID = opts.Profile.OrganizationID
			req.ProjectID = opts.Profile.ProjectID
			req.AppName = opts.AppName

			var gcpBody models.SecretServiceCreateGcpServiceAccountKeyRotatingSecretBody
			detailBytes, err := json.Marshal(internalConfig.Details)
			if err != nil {
				return fmt.Errorf("error marshaling details config: %w", err)
			}

			err = gcpBody.UnmarshalBinary(detailBytes)
			if err != nil {
				return fmt.Errorf("error marshaling details config: %w", err)
			}

			gcpBody.Name = opts.SecretName
			req.Body = &gcpBody

			_, err = opts.Client.CreateGcpServiceAccountKeyRotatingSecret(req, nil)
			if err != nil {
				return fmt.Errorf("failed to create secret with name %q: %w", opts.SecretName, err)
			}

		case integrations.Postgres:
			req := secret_service.NewCreatePostgresRotatingSecretParamsWithContext(opts.Ctx)
			req.OrganizationID = opts.Profile.OrganizationID
			req.ProjectID = opts.Profile.ProjectID
			req.AppName = opts.AppName

			var postgresBody models.SecretServiceCreatePostgresRotatingSecretBody
			detailBytes, err := json.Marshal(internalConfig.Details)
			if err != nil {
				return fmt.Errorf("error marshaling details config: %w", err)
			}

			err = postgresBody.UnmarshalBinary(detailBytes)
			if err != nil {
				return fmt.Errorf("error marshaling details config: %w", err)
			}

			postgresBody.Name = opts.SecretName
			req.Body = &postgresBody

			_, err = opts.Client.CreatePostgresRotatingSecret(req, nil)
			if err != nil {
				return fmt.Errorf("failed to create secret with name %q: %w", opts.SecretName, err)
			}

		default:
			return fmt.Errorf("unsupported rotating secret provider type")
		}

	case secretTypeDynamic:
		var (
			config         SecretConfig
			internalConfig secretConfigInternal
			err            error
		)

		if opts.SecretFilePath == "" {
			config, internalConfig, err = promptUserForConfig(opts)
			if err != nil {
				return fmt.Errorf("failed to create integration via cli prompt: %w", err)
			}
		} else {
			config, internalConfig, err = readConfigFile(opts.SecretFilePath)
			if err != nil {
				return fmt.Errorf("failed to process config file: %w", err)
			}
		}

		missingFields := validateSecretConfig(config)

		if len(missingFields) > 0 {
			return fmt.Errorf("missing required field(s) in the config file: %s", missingFields)
		}

		switch config.Type {
		case integrations.AWS:
			req := secret_service.NewCreateAwsDynamicSecretParamsWithContext(opts.Ctx)
			req.OrganizationID = opts.Profile.OrganizationID
			req.ProjectID = opts.Profile.ProjectID
			req.AppName = opts.AppName

			var awsBody models.SecretServiceCreateAwsDynamicSecretBody
			detailBytes, err := json.Marshal(internalConfig.Details)
			if err != nil {
				return fmt.Errorf("error marshaling details config: %w", err)
			}

			err = awsBody.UnmarshalBinary(detailBytes)
			if err != nil {
				return fmt.Errorf("error marshaling details config: %w", err)
			}

			awsBody.Name = opts.SecretName
			req.Body = &awsBody

			_, err = opts.Client.CreateAwsDynamicSecret(req, nil)
			if err != nil {
				return fmt.Errorf("failed to create secret with name %q: %w", opts.SecretName, err)
			}

		case integrations.GCP:
			req := secret_service.NewCreateGcpDynamicSecretParamsWithContext(opts.Ctx)
			req.OrganizationID = opts.Profile.OrganizationID
			req.ProjectID = opts.Profile.ProjectID
			req.AppName = opts.AppName

			var gcpBody models.SecretServiceCreateGcpDynamicSecretBody
			detailBytes, err := json.Marshal(internalConfig.Details)
			if err != nil {
				return fmt.Errorf("error marshaling details config: %w", err)
			}

			err = gcpBody.UnmarshalBinary(detailBytes)
			if err != nil {
				return fmt.Errorf("error marshaling details config: %w", err)
			}

			gcpBody.Name = opts.SecretName
			req.Body = &gcpBody

			_, err = opts.Client.CreateGcpDynamicSecret(req, nil)
			if err != nil {
				return fmt.Errorf("failed to create secret with name %q: %w", opts.SecretName, err)
			}

		default:
			return fmt.Errorf("unsupported dynamic secret provider type")
		}

	default:
		return fmt.Errorf("%q is an unsupported secret type; \"static\", \"rotating\", \"dynamic\" are available types", opts.Type)
	}

	command := fmt.Sprintf(`$ hcp vault-secrets secrets read %s --app %s`, opts.SecretName, opts.AppName)
	fmt.Fprintln(opts.IO.Err())
	fmt.Fprintf(opts.IO.Err(), "%s Successfully created secret with name %q\n", opts.IO.ColorScheme().SuccessIcon(), opts.SecretName)
	fmt.Fprintln(opts.IO.Err())
	fmt.Fprintf(opts.IO.Err(), `To read your secret, run:
  %s`, opts.IO.ColorScheme().String(command).Bold())
	fmt.Fprintln(opts.IO.Err())
	return nil
}

func readPlainTextSecret(opts *CreateOpts) error {
	// If the secret value is provided, then we don't need to read it from the file
	// this is used for making testing easier without needing to create a file
	if opts.SecretValuePlaintext != "" {
		return nil
	}

	if opts.SecretFilePath == "" {
		return errors.New("data file path is required")
	}

	if opts.SecretFilePath == "-" {
		plaintextSecretBytes, err := io.ReadAll(opts.IO.In())
		if err != nil {
			return fmt.Errorf("failed to read the plaintext secret: %w", err)
		}

		if len(plaintextSecretBytes) == 0 {
			return errors.New("secret value cannot be empty")
		}
		opts.SecretValuePlaintext = string(plaintextSecretBytes)
		return nil
	}

	fileInfo, err := os.Stat(opts.SecretFilePath)
	if err != nil {
		return fmt.Errorf("failed to get data file info: %w", err)
	}

	if fileInfo.Size() == 0 {
		return errors.New("data file cannot be empty")
	}

	data, err := os.ReadFile(opts.SecretFilePath)
	if err != nil {
		return fmt.Errorf("unable to read the data file: %w", err)
	}
	opts.SecretValuePlaintext = string(data)
	return nil
}

func readConfigFile(filePath string) (SecretConfig, secretConfigInternal, error) {
	var (
		secretConfig   SecretConfig
		internalConfig secretConfigInternal
	)

	if err := hclsimple.DecodeFile(filePath, nil, &secretConfig); err != nil {
		return secretConfig, internalConfig, fmt.Errorf("failed to decode config file: %w", err)
	}

	detailsMap, err := integrations.CtyValueToMap(secretConfig.Details)
	if err != nil {
		return secretConfig, internalConfig, err
	}
	internalConfig.Details = detailsMap

	return secretConfig, internalConfig, nil
}

func validateSecretConfig(secretConfig SecretConfig) []string {
	var missingKeys []string

	if secretConfig.Type == "" {
		missingKeys = append(missingKeys, "type")
	}

	return missingKeys
}

var availableRotatingSecretProviders = map[string]map[string]any{
	string(integrations.Twilio):       twilioRotatingSecretTemplate,
	string(integrations.MongoDBAtlas): mongoDBAtlasRotatingSecretTemplate,
	string(integrations.AWS):          awsRotatingSecretTemplate,
	string(integrations.GCP):          gcpRotatingSecretTemplate,
	string(integrations.Postgres):     postgresRotatingSecretTemplate,
}

var availableDynamicSecretProviders = map[string]map[string]any{
	string(integrations.AWS): awsDynamicSecretTemplate,
	string(integrations.GCP): gcpDynamicSecretTemplate,
}

func promptUserForConfig(opts *CreateOpts) (SecretConfig, secretConfigInternal, error) {
	var (
		config         SecretConfig
		internalConfig secretConfigInternal
		err            error
		providerFields map[string]map[string]any
		items          []string
	)

	if !opts.IO.CanPrompt() {
		return config, internalConfig, fmt.Errorf("unable to create secret interactively")
	}

	if opts.Type == secretTypeDynamic {
		providerFields = availableDynamicSecretProviders
		items = maps.Keys(providerFields)
	} else if opts.Type == secretTypeRotating {
		providerFields = availableRotatingSecretProviders
		items = maps.Keys(providerFields)
	}

	providerPrompt := promptui.Select{
		Label:  "Please select the provider you would like to configure",
		Items:  items,
		Stdin:  io.NopCloser(opts.IO.In()),
		Stdout: iostreams.NopWriteCloser(opts.IO.Err()),
	}

	_, provider, err := providerPrompt.Run()
	if err != nil {
		return config, internalConfig, fmt.Errorf("provider selection prompt failed: %w", err)
	}
	config.Type = integrations.IntegrationType(provider)

	fieldValues, err := populateFieldValues(providerFields[provider], opts)
	if err != nil {
		return SecretConfig{}, secretConfigInternal{}, err
	}

	internalConfig.Details = fieldValues

	return config, internalConfig, nil
}

func populateFieldValues(providerFields map[string]any, opts *CreateOpts) (map[string]any, error) {
	var fieldPrompt promptui.Prompt
	fieldsMap := make(map[string]any)

	for field, value := range providerFields {
		switch reflect.ValueOf(value).Kind() {
		case reflect.String:
			label := field
			if slices.Contains(optionalFields, field) {
				label = field + " (optional)"
			}

			fieldPrompt = promptui.Prompt{
				Label: label,
				Mask:  '*',
			}

			input, err := fieldPrompt.Run()
			if err != nil {
				return nil, fmt.Errorf("prompt for field %s failed: %w", field, err)
			}

			fieldsMap[field] = input
		case reflect.Map:
			nestedMap, err := populateFieldValues(value.(map[string]any), opts)
			if err != nil {
				return nil, err
			}
			fieldsMap[field] = nestedMap
		case reflect.Slice:
			var sliceItems []map[string]any
			nestedMap := make(map[string]any)
			valueSlice := value.([]map[string]string)

			ok := true

			if slices.Contains(optionalFields, field) {
				ok = promptForOptionalField(field, opts)
			}

			for proceed := ok; proceed; proceed = promptForAdditionalField(field, opts) {

				// Since the valueSlice here is hardcoded in the provider field templates, we know there is
				// only a map at the first index of the slice. So instead of iterating through the slice,
				// let's just grab the first index and iterative over the key/value pairs in that map.
				for _, nestedField := range maps.Keys(valueSlice[0]) {
					label := any(nestedField).(string)
					if slices.Contains(optionalFields, nestedField) {
						label += " (optional)"
					}

					fieldPrompt = promptui.Prompt{
						Label: "Enter " + label + " for " + field,
						Mask:  '*',
					}

					input, err := fieldPrompt.Run()
					if err != nil {
						return nil, fmt.Errorf("prompt for field %s failed: %w", field, err)
					}

					nestedMap[any(nestedField).(string)] = input
				}
				sliceItems = append(sliceItems, nestedMap)
				fieldsMap[field] = sliceItems
			}
		}
	}

	return fieldsMap, nil
}

func promptForAdditionalField(field string, opts *CreateOpts) bool {
	proceed, err := opts.IO.PromptConfirm("Would you like to add configuration for another " + field + "?")
	if err != nil {
		return false
	}

	return proceed
}

func promptForOptionalField(field string, opts *CreateOpts) bool {
	proceed, err := opts.IO.PromptConfirm(field + " is optional. Would you like to configure " + field + "?")
	if err != nil {
		return false
	}

	return proceed
}
