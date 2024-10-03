// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package secrets

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/posener/complete"
	"github.com/zclconf/go-cty/cty"

	"github.com/hashicorp/hcl/v2/hclsimple"
	preview_secret_service "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/preview/2023-11-28/client/secret_service"
	preview_models "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/preview/2023-11-28/models"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-06-13/client/secret_service"
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
		Ctx:           ctx.ShutdownCtx,
		Profile:       ctx.Profile,
		IO:            ctx.IO,
		Output:        ctx.Output,
		PreviewClient: preview_secret_service.New(ctx.HCP, nil),
		Client:        secret_service.New(ctx.HCP, nil),
	}

	cmd := &cmd.Command{
		Name:      "create",
		ShortHelp: "Create a new static secret.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp vault-secrets secrets create" }} command creates a new static, rotating, or dynamic secret under a Vault Secrets application.
		For rotating and dynamic secrets, the following fields are required in the config file: [type integration_name details].
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
				Preamble: `Create a new rotating secret in the Vault Secrets application on your active profile:`,
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				$ hcp vault-secrets secrets create secret_1 --secret-type=rotating --data-file=path/to/file/config.hcl
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
					Description:  "File path to read secret data from. Set this to '-' to read the secret data from stdin.",
					Value:        flagvalue.Simple("", &opts.SecretFilePath),
					Required:     true,
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
	PreviewClient        preview_secret_service.ClientService
	Client               secret_service.ClientService
}

type secretConfigInternal struct {
	Details map[string]any
}

type SecretConfig struct {
	Type            integrations.IntegrationType `hcl:"type"`
	IntegrationName string                       `hcl:"integration_name"`
	Details         cty.Value                    `hcl:"details"`
}

func createRun(opts *CreateOpts) error {
	switch opts.Type {
	case secretTypeKV, "":
		if err := readPlainTextSecret(opts); err != nil {
			return err
		}

		req := secret_service.NewCreateAppKVSecretParamsWithContext(opts.Ctx)
		req.LocationOrganizationID = opts.Profile.OrganizationID
		req.LocationProjectID = opts.Profile.ProjectID
		req.AppName = opts.AppName

		req.Body = secret_service.CreateAppKVSecretBody{
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
		secretConfig, internalConfig, err := readConfigFile(opts.SecretFilePath)
		if err != nil {
			return fmt.Errorf("failed to process config file: %w", err)
		}

		missingFields := validateSecretConfig(secretConfig)

		if len(missingFields) > 0 {
			return fmt.Errorf("missing required field(s) in the config file: %s", missingFields)
		}

		switch secretConfig.Type {
		case integrations.Twilio:
			req := preview_secret_service.NewCreateTwilioRotatingSecretParamsWithContext(opts.Ctx)
			req.OrganizationID = opts.Profile.OrganizationID
			req.ProjectID = opts.Profile.ProjectID
			req.AppName = opts.AppName

			var twilioBody preview_models.SecretServiceCreateTwilioRotatingSecretBody
			detailBytes, err := json.Marshal(internalConfig.Details)
			if err != nil {
				return fmt.Errorf("error marshaling details config: %w", err)
			}

			err = twilioBody.UnmarshalBinary(detailBytes)
			if err != nil {
				return fmt.Errorf("error marshaling details config: %w", err)
			}

			twilioBody.IntegrationName = secretConfig.IntegrationName
			twilioBody.SecretName = opts.SecretName
			req.Body = &twilioBody

			resp, err := opts.PreviewClient.CreateTwilioRotatingSecret(req, nil)
			if err != nil {
				return fmt.Errorf("failed to create secret with name %q: %w", opts.SecretName, err)
			}

			if err := opts.Output.Display(newRotatingSecretsDisplayer(true).PreviewRotatingSecrets(resp.Payload.Config)); err != nil {
				return err
			}

		case integrations.MongoDBAtlas:

			req := preview_secret_service.NewCreateMongoDBAtlasRotatingSecretParamsWithContext(opts.Ctx)
			req.OrganizationID = opts.Profile.OrganizationID
			req.ProjectID = opts.Profile.ProjectID
			req.AppName = opts.AppName

			var mongoDBBody preview_models.SecretServiceCreateMongoDBAtlasRotatingSecretBody
			detailBytes, err := json.Marshal(internalConfig.Details)
			if err != nil {
				return fmt.Errorf("error marshaling details config: %w", err)
			}

			err = mongoDBBody.UnmarshalBinary(detailBytes)
			if err != nil {
				return fmt.Errorf("error marshaling details config: %w", err)
			}

			mongoDBBody.IntegrationName = secretConfig.IntegrationName
			mongoDBBody.SecretName = opts.SecretName
			req.Body = &mongoDBBody

			resp, err := opts.PreviewClient.CreateMongoDBAtlasRotatingSecret(req, nil)
			if err != nil {
				return fmt.Errorf("failed to create secret with name %q: %w", opts.SecretName, err)
			}

			if err := opts.Output.Display(newRotatingSecretsDisplayer(true).PreviewRotatingSecrets(resp.Payload.Config)); err != nil {
				return err
			}

		case integrations.AWS:
			req := preview_secret_service.NewCreateAwsIAMUserAccessKeyRotatingSecretParamsWithContext(opts.Ctx)
			req.OrganizationID = opts.Profile.OrganizationID
			req.ProjectID = opts.Profile.ProjectID
			req.AppName = opts.AppName

			var awsBody preview_models.SecretServiceCreateAwsIAMUserAccessKeyRotatingSecretBody
			detailBytes, err := json.Marshal(internalConfig.Details)
			if err != nil {
				return fmt.Errorf("error marshaling details config: %w", err)
			}

			err = awsBody.UnmarshalBinary(detailBytes)
			if err != nil {
				return fmt.Errorf("error marshaling details config: %w", err)
			}

			awsBody.IntegrationName = secretConfig.IntegrationName
			awsBody.Name = opts.SecretName
			req.Body = &awsBody

			_, err = opts.PreviewClient.CreateAwsIAMUserAccessKeyRotatingSecret(req, nil)
			if err != nil {
				return fmt.Errorf("failed to create secret with name %q: %w", opts.SecretName, err)
			}

		case integrations.GCP:
			req := preview_secret_service.NewCreateGcpServiceAccountKeyRotatingSecretParamsWithContext(opts.Ctx)
			req.OrganizationID = opts.Profile.OrganizationID
			req.ProjectID = opts.Profile.ProjectID
			req.AppName = opts.AppName

			var gcpBody preview_models.SecretServiceCreateGcpServiceAccountKeyRotatingSecretBody
			detailBytes, err := json.Marshal(internalConfig.Details)
			if err != nil {
				return fmt.Errorf("error marshaling details config: %w", err)
			}

			err = gcpBody.UnmarshalBinary(detailBytes)
			if err != nil {
				return fmt.Errorf("error marshaling details config: %w", err)
			}

			gcpBody.IntegrationName = secretConfig.IntegrationName
			gcpBody.Name = opts.SecretName
			req.Body = &gcpBody

			_, err = opts.PreviewClient.CreateGcpServiceAccountKeyRotatingSecret(req, nil)
			if err != nil {
				return fmt.Errorf("failed to create secret with name %q: %w", opts.SecretName, err)
			}

		default:
			return fmt.Errorf("unsupported rotating secret provider type")
		}

	case secretTypeDynamic:
		secretConfig, internalConfig, err := readConfigFile(opts.SecretFilePath)
		if err != nil {
			return fmt.Errorf("failed to process config file: %w", err)
		}
		missingFields := validateSecretConfig(secretConfig)

		if len(missingFields) > 0 {
			return fmt.Errorf("missing required field(s) in the config file: %s", missingFields)
		}

		switch secretConfig.Type {
		case integrations.AWS:
			req := preview_secret_service.NewCreateAwsDynamicSecretParamsWithContext(opts.Ctx)
			req.OrganizationID = opts.Profile.OrganizationID
			req.ProjectID = opts.Profile.ProjectID
			req.AppName = opts.AppName

			var awsBody preview_models.SecretServiceCreateAwsDynamicSecretBody
			detailBytes, err := json.Marshal(internalConfig.Details)
			if err != nil {
				return fmt.Errorf("error marshaling details config: %w", err)
			}

			err = awsBody.UnmarshalBinary(detailBytes)
			if err != nil {
				return fmt.Errorf("error marshaling details config: %w", err)
			}

			awsBody.IntegrationName = secretConfig.IntegrationName
			awsBody.Name = opts.SecretName
			req.Body = &awsBody

			_, err = opts.PreviewClient.CreateAwsDynamicSecret(req, nil)
			if err != nil {
				return fmt.Errorf("failed to create secret with name %q: %w", opts.SecretName, err)
			}

		case integrations.GCP:
			req := preview_secret_service.NewCreateGcpDynamicSecretParamsWithContext(opts.Ctx)
			req.OrganizationID = opts.Profile.OrganizationID
			req.ProjectID = opts.Profile.ProjectID
			req.AppName = opts.AppName

			var gcpBody preview_models.SecretServiceCreateGcpDynamicSecretBody
			detailBytes, err := json.Marshal(internalConfig.Details)
			if err != nil {
				return fmt.Errorf("error marshaling details config: %w", err)
			}

			err = gcpBody.UnmarshalBinary(detailBytes)
			if err != nil {
				return fmt.Errorf("error marshaling details config: %w", err)
			}

			gcpBody.IntegrationName = secretConfig.IntegrationName
			gcpBody.Name = opts.SecretName
			req.Body = &gcpBody

			_, err = opts.PreviewClient.CreateGcpDynamicSecret(req, nil)
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

	if secretConfig.IntegrationName == "" {
		missingKeys = append(missingKeys, "integration_name")
	}

	return missingKeys
}
