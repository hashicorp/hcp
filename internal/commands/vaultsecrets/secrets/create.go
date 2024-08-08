// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package secrets

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"

	"github.com/mitchellh/mapstructure"
	"github.com/posener/complete"
	"golang.org/x/exp/maps"
	"gopkg.in/yaml.v3"

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

type SecretType string

const (
	Static   SecretType = "static"
	Rotating SecretType = "rotating"
	Dynamic  SecretType = "dynamic"
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
		The {{ template "mdCodeOrBold" "hcp vault-secrets secrets create" }} command creates a new static secret under a Vault Secrets application.
		`),
		Examples: []cmd.Example{
			{
				Preamble: `Create a new secret in the Vault Secrets application on your active profile:`,
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
	Type                 SecretType
	PreviewClient        preview_secret_service.ClientService
	Client               secret_service.ClientService
}

type SecretConfig struct {
	Version         string
	Type            integrations.IntegrationType
	IntegrationName string `yaml:"rotation_integration_name"`
	Details         map[string]any
}

type MongoDBRole struct {
	RoleName       string `mapstructure:"role_name"`
	DatabaseName   string `mapstructure:"database_name"`
	CollectionName string `mapstructure:"collection_name"`
}

type MongoDBScope struct {
	Name string `mapstructure:"type"`
	Type string `mapstructure:"name"`
}

var (
	TwilioKeys = []string{"rotation_policy_name"}
	MongoKeys  = []string{"rotation_policy_name", "mongodb_group_id"}
)

var rotationPolicies = map[string]string{
	"30": "built-in:30-days-2-active",
	"60": "built-in:60-days-2-active",
	"90": "built-in:90-days-2-active",
}

func createRun(opts *CreateOpts) error {

	switch opts.Type {
	case Static, "":
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
	case Rotating:
		f, err := os.ReadFile(opts.SecretFilePath)
		if err != nil {
			return fmt.Errorf("failed to open config file: %w", err)
		}

		var sc SecretConfig
		err = yaml.Unmarshal(f, &sc)
		if err != nil {
			return fmt.Errorf("failed to unmarshal config file: %w", err)
		}

		if sc.Type == "" || sc.IntegrationName == "" {
			return fmt.Errorf("missing required field in the config file: type")
		}

		if sc.IntegrationName == "" {
			return fmt.Errorf("missing required field in the config file: rotation_integration_name")
		}

		switch sc.Type {
		case integrations.Twilio:
			missingField := validateDetails(sc.Details, TwilioKeys)

			if missingField != "" {
				return fmt.Errorf("missing required field in the config file: %s", missingField)
			}

			req := preview_secret_service.NewCreateTwilioRotatingSecretParamsWithContext(opts.Ctx)
			req.OrganizationID = opts.Profile.OrganizationID
			req.ProjectID = opts.Profile.ProjectID
			req.AppName = opts.AppName
			req.Body = &preview_models.SecretServiceCreateTwilioRotatingSecretBody{
				RotationIntegrationName: sc.IntegrationName,
				RotationPolicyName:      rotationPolicies[sc.Details["rotation_policy_name"].(string)],
				SecretName:              opts.SecretName,
			}

			resp, err := opts.PreviewClient.CreateTwilioRotatingSecret(req, nil)
			if err != nil {
				return fmt.Errorf("failed to create secret with name %q: %w", opts.SecretName, err)
			}

			if err := opts.Output.Display(newRotatingSecretsDisplayer(true).PreviewRotatingSecrets(resp.Payload.Config)); err != nil {
				return err
			}

		case integrations.MongoDBAtlas:
			missingField := validateDetails(sc.Details, MongoKeys)

			if missingField != "" {
				return fmt.Errorf("missing required field in the config file: %s", missingField)
			}

			req := preview_secret_service.NewCreateMongoDBAtlasRotatingSecretParamsWithContext(opts.Ctx)
			req.OrganizationID = opts.Profile.OrganizationID
			req.ProjectID = opts.Profile.ProjectID

			roles := sc.Details["mongodb_roles"].([]interface{})
			var reqRoles []*preview_models.Secrets20231128MongoDBRole
			for _, r := range roles {
				var role MongoDBRole
				decoder, _ := mapstructure.NewDecoder(&mapstructure.DecoderConfig{WeaklyTypedInput: true, Result: &role})
				if err := decoder.Decode(r); err != nil {
					return fmt.Errorf("unable to decode to a mongodb role")
				}

				reqRole := &preview_models.Secrets20231128MongoDBRole{
					CollectionName: role.CollectionName,
					RoleName:       role.RoleName,
					DatabaseName:   role.DatabaseName,
				}
				reqRoles = append(reqRoles, reqRole)
			}

			scopes := sc.Details["mongodb_scopes"].([]interface{})
			var reqScopes []*preview_models.Secrets20231128MongoDBScope
			for _, r := range scopes {
				var scope MongoDBScope
				decoder, _ := mapstructure.NewDecoder(&mapstructure.DecoderConfig{WeaklyTypedInput: true, Result: &scope})
				if err := decoder.Decode(r); err != nil {
					return fmt.Errorf("unable to decode to a mongodb role")
				}

				reqScope := &preview_models.Secrets20231128MongoDBScope{
					Name: scope.Name,
					Type: scope.Type,
				}
				reqScopes = append(reqScopes, reqScope)
			}

			req.Body = &preview_models.SecretServiceCreateMongoDBAtlasRotatingSecretBody{
				MongodbGroupID:          sc.Details["mongodb_group_id"].(string),
				MongodbRoles:            reqRoles,
				MongodbScopes:           reqScopes,
				RotationIntegrationName: sc.IntegrationName,
				RotationPolicyName:      rotationPolicies[sc.Details["rotation_policy_name"].(string)],
				SecretName:              opts.SecretName,
			}
			resp, err := opts.PreviewClient.CreateMongoDBAtlasRotatingSecret(req, nil)
			if err != nil {
				return fmt.Errorf("failed to create secret with name %q: %w", opts.SecretName, err)
			}

			if err := opts.Output.Display(newRotatingSecretsDisplayer(true).PreviewRotatingSecrets(resp.Payload.Config)); err != nil {
				return err
			}
		}
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

func validateDetails(details map[string]any, requiredKeys []string) string {
	detailsKeys := maps.Keys(details)

	for _, r := range requiredKeys {
		if !slices.Contains(detailsKeys, r) {
			return r
		}
	}
	return ""
}
