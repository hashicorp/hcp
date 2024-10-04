package secrets

import (
	"context"
	"encoding/json"
	"fmt"

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

type UpdateOpts struct {
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

func NewCmdUpdate(ctx *cmd.Context, runF func(*UpdateOpts) error) *cmd.Command {
	opts := &UpdateOpts{
		Ctx:           ctx.ShutdownCtx,
		Profile:       ctx.Profile,
		IO:            ctx.IO,
		Output:        ctx.Output,
		PreviewClient: preview_secret_service.New(ctx.HCP, nil),
		Client:        secret_service.New(ctx.HCP, nil),
	}

	cmd := &cmd.Command{
		Name:      "update",
		ShortHelp: "Update an existing dynamic or rotating secret.",
		LongHelp: heredoc.New(ctx.IO).Must(`
      The {{ template "mdCodeOrBold" "hcp vault-secrets secrets update" }} command updates an existing rotating or dynamic secret under a Vault Secrets application.
	  The configuration for updating your rotating or dynamic secret will be read from the provided HCL config file. The following fields are required in the config 
	  file: [type details]. For help populating the details for a dynamic or rotating secret, please refer to the 
	  {{ Link "API reference documentation" "https://developer.hashicorp.com/hcp/api-docs/vault-secrets/2023-11-28" }}.
      `),
		Examples: []cmd.Example{
			{
				Preamble: `Update a rotating secret in the Vault Secrets application on your active profile:`,
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
            $ hcp vault-secrets secrets update secret_1 --secret-type=rotating --data-file=tmp/secrets1.txt
            `),
			},
		},
		Args: cmd.PositionalArguments{
			Args: []cmd.PositionalArgument{
				{
					Name:          "NAME",
					Documentation: "The name of the secret to update.",
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
					Description:  "The type of secret to update: rotating or dynamic.",
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
			return updateRun(opts)
		},
	}

	return cmd
}

type SecretUpdateConfig struct {
	Type    integrations.IntegrationType `hcl:"type"`
	Details cty.Value                    `hcl:"details"`
}

func updateRun(opts *UpdateOpts) error {
	switch opts.Type {
	case secretTypeRotating:
		secretConfig, internalConfig, err := readUpdateConfigFile(opts.SecretFilePath)
		if err != nil {
			return fmt.Errorf("failed to process config file: %w", err)
		}

		missingFields := validateSecretUpdateConfig(secretConfig)

		if len(missingFields) > 0 {
			return fmt.Errorf("missing required field(s) in the config file: %s", missingFields)
		}

		switch secretConfig.Type {
		case integrations.Twilio:
			req := preview_secret_service.NewUpdateTwilioRotatingSecretParamsWithContext(opts.Ctx)
			req.OrganizationID = opts.Profile.OrganizationID
			req.ProjectID = opts.Profile.ProjectID
			req.AppName = opts.AppName
			req.SecretName = opts.SecretName

			var twilioBody preview_models.SecretServiceUpdateTwilioRotatingSecretBody
			detailBytes, err := json.Marshal(internalConfig.Details)
			if err != nil {
				return fmt.Errorf("error marshaling details config: %w", err)
			}

			err = twilioBody.UnmarshalBinary(detailBytes)
			if err != nil {
				return fmt.Errorf("error unmarshaling details config: %w", err)
			}
			req.Body = &twilioBody

			resp, err := opts.PreviewClient.UpdateTwilioRotatingSecret(req, nil)
			if err != nil {
				return fmt.Errorf("failed to update secret with name %q: %w", opts.SecretName, err)
			}

			if err := opts.Output.Display(newRotatingSecretsDisplayer(true).PreviewRotatingSecrets(resp.Payload.Config)); err != nil {
				return err
			}

		case integrations.MongoDBAtlas:
			req := preview_secret_service.NewUpdateMongoDBAtlasRotatingSecretParamsWithContext(opts.Ctx)
			req.OrganizationID = opts.Profile.OrganizationID
			req.ProjectID = opts.Profile.ProjectID
			req.AppName = opts.AppName
			req.SecretName = opts.SecretName

			var mongoDBBody preview_models.SecretServiceUpdateMongoDBAtlasRotatingSecretBody
			detailBytes, err := json.Marshal(internalConfig.Details)
			if err != nil {
				return fmt.Errorf("error marshaling details config: %w", err)
			}

			err = mongoDBBody.UnmarshalBinary(detailBytes)
			if err != nil {
				return fmt.Errorf("error unmarshaling details config: %w", err)
			}
			req.Body = &mongoDBBody

			resp, err := opts.PreviewClient.UpdateMongoDBAtlasRotatingSecret(req, nil)
			if err != nil {
				return fmt.Errorf("failed to update secret with name %q: %w", opts.SecretName, err)
			}

			if err := opts.Output.Display(newRotatingSecretsDisplayer(true).PreviewRotatingSecrets(resp.Payload.Config)); err != nil {
				return err
			}

		case integrations.AWS:
			req := preview_secret_service.NewUpdateAwsIAMUserAccessKeyRotatingSecretParamsWithContext(opts.Ctx)
			req.OrganizationID = opts.Profile.OrganizationID
			req.ProjectID = opts.Profile.ProjectID
			req.AppName = opts.AppName
			req.Name = opts.SecretName

			var awsBody preview_models.SecretServiceUpdateAwsIAMUserAccessKeyRotatingSecretBody
			detailBytes, err := json.Marshal(internalConfig.Details)
			if err != nil {
				return fmt.Errorf("error marshaling details config: %w", err)
			}

			err = awsBody.UnmarshalBinary(detailBytes)
			if err != nil {
				return fmt.Errorf("error unmarshaling details config: %w", err)
			}

			req.Body = &awsBody

			_, err = opts.PreviewClient.UpdateAwsIAMUserAccessKeyRotatingSecret(req, nil)
			if err != nil {
				return fmt.Errorf("failed to update secret with name %q: %w", opts.SecretName, err)
			}

		case integrations.GCP:
			req := preview_secret_service.NewUpdateGcpServiceAccountKeyRotatingSecretParamsWithContext(opts.Ctx)
			req.OrganizationID = opts.Profile.OrganizationID
			req.ProjectID = opts.Profile.ProjectID
			req.AppName = opts.AppName
			req.Name = opts.SecretName

			var gcpBody preview_models.SecretServiceUpdateGcpServiceAccountKeyRotatingSecretBody
			detailBytes, err := json.Marshal(internalConfig.Details)
			if err != nil {
				return fmt.Errorf("error marshaling details config: %w", err)
			}

			err = gcpBody.UnmarshalBinary(detailBytes)
			if err != nil {
				return fmt.Errorf("error unmarshaling details config: %w", err)
			}

			req.Body = &gcpBody

			_, err = opts.PreviewClient.UpdateGcpServiceAccountKeyRotatingSecret(req, nil)
			if err != nil {
				return fmt.Errorf("failed to update secret with name %q: %w", opts.SecretName, err)
			}
		}

	case secretTypeDynamic:
		secretConfig, internalConfig, err := readUpdateConfigFile(opts.SecretFilePath)
		if err != nil {
			return fmt.Errorf("failed to process config file: %w", err)
		}

		missingFields := validateSecretUpdateConfig(secretConfig)

		if len(missingFields) > 0 {
			return fmt.Errorf("missing required field(s) in the config file: %s", missingFields)
		}

		switch secretConfig.Type {
		case integrations.AWS:
			req := preview_secret_service.NewUpdateAwsDynamicSecretParamsWithContext(opts.Ctx)
			req.OrganizationID = opts.Profile.OrganizationID
			req.ProjectID = opts.Profile.ProjectID
			req.AppName = opts.AppName
			req.Name = opts.SecretName

			var awsBody preview_models.SecretServiceUpdateAwsDynamicSecretBody
			detailBytes, err := json.Marshal(internalConfig.Details)
			if err != nil {
				return fmt.Errorf("error marshaling details config: %w", err)
			}

			err = awsBody.UnmarshalBinary(detailBytes)
			if err != nil {
				return fmt.Errorf("error unmarshaling details config: %w", err)
			}

			req.Body = &awsBody

			_, err = opts.PreviewClient.UpdateAwsDynamicSecret(req, nil)
			if err != nil {
				return fmt.Errorf("failed to update secret with name %q: %w", opts.SecretName, err)
			}

		case integrations.GCP:
			req := preview_secret_service.NewUpdateGcpDynamicSecretParamsWithContext(opts.Ctx)
			req.OrganizationID = opts.Profile.OrganizationID
			req.ProjectID = opts.Profile.ProjectID
			req.AppName = opts.AppName
			req.Name = opts.SecretName

			var gcpBody preview_models.SecretServiceUpdateGcpDynamicSecretBody
			detailBytes, err := json.Marshal(internalConfig.Details)
			if err != nil {
				return fmt.Errorf("error marshaling details config: %w", err)
			}

			err = gcpBody.UnmarshalBinary(detailBytes)
			if err != nil {
				return fmt.Errorf("error unmarshaling details config: %w", err)
			}

			req.Body = &gcpBody

			_, err = opts.PreviewClient.UpdateGcpDynamicSecret(req, nil)
			if err != nil {
				return fmt.Errorf("failed to update secret with name %q: %w", opts.SecretName, err)
			}
		}

	default:
		return fmt.Errorf("%q is an unsupported secret type; \"rotating\" and \"dynamic\" are available types", opts.Type)
	}

	fmt.Fprintf(opts.IO.Err(), "%s Successfully updated secret with name %q\n", opts.IO.ColorScheme().SuccessIcon(), opts.SecretName)

	return nil
}

func readUpdateConfigFile(filePath string) (SecretUpdateConfig, secretConfigInternal, error) {
	var (
		secretConfig   SecretUpdateConfig
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

func validateSecretUpdateConfig(secretConfig SecretUpdateConfig) []string {
	var missingKeys []string

	if secretConfig.Type == "" {
		missingKeys = append(missingKeys, "type")
	}

	return missingKeys
}
