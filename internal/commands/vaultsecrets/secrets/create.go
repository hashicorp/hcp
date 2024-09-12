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

type DetailsInternal struct {
	Details map[string]any
}

type SecretConfig struct {
	Type            integrations.IntegrationType `hcl:"type"`
	IntegrationName string                       `hcl:"integration_name"`
	Details         cty.Value                    `hcl:"details"`
}

var rotationPolicies = map[string]string{
	"30": "built-in:30-days-2-active",
	"60": "built-in:60-days-2-active",
	"90": "built-in:90-days-2-active",
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
		sc, di, err := readConfigFile(opts)
		if err != nil {
			return fmt.Errorf("failed to process config file: %w", err)
		}

		missingFields := validateSecretConfig(sc)

		if len(missingFields) > 0 {
			return fmt.Errorf("missing required field(s) in the config file: %s", missingFields)
		}

		switch sc.Type {
		case integrations.Twilio:
			req := preview_secret_service.NewCreateTwilioRotatingSecretParamsWithContext(opts.Ctx)
			req.OrganizationID = opts.Profile.OrganizationID
			req.ProjectID = opts.Profile.ProjectID
			req.AppName = opts.AppName

			var twilioBody preview_models.SecretServiceCreateTwilioRotatingSecretBody
			detailBytes, err := json.Marshal(di.Details)
			if err != nil {
				return fmt.Errorf("error marshaling details config: %w", err)
			}

			err = twilioBody.UnmarshalBinary(detailBytes)
			if err != nil {
				return fmt.Errorf("error marshaling details config: %w", err)
			}

			twilioBody.IntegrationName = sc.IntegrationName
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
			detailBytes, err := json.Marshal(di.Details)
			if err != nil {
				return fmt.Errorf("error marshaling details config: %w", err)
			}

			err = mongoDBBody.UnmarshalBinary(detailBytes)
			if err != nil {
				return fmt.Errorf("error marshaling details config: %w", err)
			}

			mongoDBBody.IntegrationName = sc.IntegrationName
			mongoDBBody.SecretName = opts.SecretName
			req.Body = &mongoDBBody

			resp, err := opts.PreviewClient.CreateMongoDBAtlasRotatingSecret(req, nil)
			if err != nil {
				return fmt.Errorf("failed to create secret with name %q: %w", opts.SecretName, err)
			}

			if err := opts.Output.Display(newRotatingSecretsDisplayer(true).PreviewRotatingSecrets(resp.Payload.Config)); err != nil {
				return err
			}

		default:
			return fmt.Errorf("unsupported rotating secret provider type")
		}

	case secretTypeDynamic:
		sc, di, err := readConfigFile(opts)
		if err != nil {
			return fmt.Errorf("failed to process config file: %w", err)
		}
		missingFields := validateSecretConfig(sc)

		if len(missingFields) > 0 {
			return fmt.Errorf("missing required field(s) in the config file: %s", missingFields)
		}

		switch sc.Type {
		case integrations.AWS:
			req := preview_secret_service.NewCreateAwsDynamicSecretParamsWithContext(opts.Ctx)
			req.OrganizationID = opts.Profile.OrganizationID
			req.ProjectID = opts.Profile.ProjectID
			req.AppName = opts.AppName

			var awsBody preview_models.SecretServiceCreateAwsDynamicSecretBody
			detailBytes, err := json.Marshal(di.Details)
			if err != nil {
				return fmt.Errorf("error marshaling details config: %w", err)
			}

			err = awsBody.UnmarshalBinary(detailBytes)
			if err != nil {
				return fmt.Errorf("error marshaling details config: %w", err)
			}

			awsBody.IntegrationName = sc.IntegrationName
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
			detailBytes, err := json.Marshal(di.Details)
			if err != nil {
				return fmt.Errorf("error marshaling details config: %w", err)
			}

			err = gcpBody.UnmarshalBinary(detailBytes)
			if err != nil {
				return fmt.Errorf("error marshaling details config: %w", err)
			}

			gcpBody.IntegrationName = sc.IntegrationName
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

func readConfigFile(opts *CreateOpts) (SecretConfig, DetailsInternal, error) {
	var (
		sc SecretConfig
		di DetailsInternal
	)

	if err := hclsimple.DecodeFile(opts.SecretFilePath, nil, &sc); err != nil {
		return sc, di, fmt.Errorf("failed to decode config file: %w", err)
	}

	detailsMap, err := ctyValueToMap(sc.Details)
	if err != nil {
		return sc, di, err
	}
	di.Details = detailsMap

	return sc, di, nil
}

func ctyValueToMap(value cty.Value) (map[string]any, error) {
	varMapNow := make(map[string]any)
	for k, v := range value.AsValueMap() {
		if v.Type() == cty.String {
			if k == "rotation_policy_name" {
				varMapNow[k] = rotationPolicies[v.AsString()]
			} else {
				varMapNow[k] = v.AsString()
			}
		} else if v.Type().IsObjectType() {
			nestedMap, err := ctyValueToMap(v)
			if err != nil {
				return nil, err
			}
			varMapNow[k] = nestedMap
		} else if v.Type().IsTupleType() {
			var roles []map[string]any
			for _, val := range v.AsValueSlice() {
				nestedMap, err := ctyValueToMap(val)
				if err != nil {
					return nil, err
				}
				roles = append(roles, nestedMap)
			}
			varMapNow[k] = roles
		} else {
			return nil, fmt.Errorf("found unsupported value type")
		}
	}

	return varMapNow, nil
}

func validateSecretConfig(sc SecretConfig) []string {
	var missingKeys []string

	if sc.Type == "" {
		missingKeys = append(missingKeys, "type")
	}

	if sc.IntegrationName == "" {
		missingKeys = append(missingKeys, "integration_name")
	}

	return missingKeys
}
