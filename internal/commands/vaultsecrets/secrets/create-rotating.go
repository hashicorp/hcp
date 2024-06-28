package secrets

import (
	"context"
	"fmt"
	"os"

	preview_secret_service "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/preview/2023-11-28/client/secret_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/preview/2023-11-28/models"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-06-13/client/secret_service"
	"github.com/hashicorp/hcp/internal/commands/vaultsecrets/secrets/appname"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/flagvalue"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
	"github.com/mitchellh/mapstructure"
	"github.com/posener/complete"
	"gopkg.in/yaml.v3"
)

func NewCmdCreateRotating(ctx *cmd.Context, runF func(*CreateRotatingOpts) error) *cmd.Command {
	opts := &CreateRotatingOpts{
		Ctx:           ctx.ShutdownCtx,
		Profile:       ctx.Profile,
		IO:            ctx.IO,
		Output:        ctx.Output,
		PreviewClient: preview_secret_service.New(ctx.HCP, nil),
		Client:        secret_service.New(ctx.HCP, nil),
	}

	cmd := &cmd.Command{
		Name:      "create-rotating",
		ShortHelp: "Create a new rotating secret.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp vault-secrets secrets create-rotating" }} command creates a new rotating secret under a Vault Secrets application.
		`),
		Flags: cmd.Flags{
			Local: []*cmd.Flag{
				{
					Name:         "config-file",
					DisplayValue: "CONFIG_FILE_PATH",
					Description:  "File path to a secret config",
					Value:        flagvalue.Simple("", &opts.SecretConfigFile),
					Autocomplete: complete.PredictOr(
						complete.PredictFiles("*"),
						complete.PredictSet("-"),
					),
				},
			},
		},
		RunF: func(c *cmd.Command, args []string) error {
			opts.AppName = appname.Get()

			if runF != nil {
				return runF(opts)
			}
			return createRotatingRun(opts)
		},
	}

	return cmd
}

type CreateRotatingOpts struct {
	Ctx     context.Context
	Profile *profile.Profile
	IO      iostreams.IOStreams
	Output  *format.Outputter

	AppName          string
	SecretConfigFile string
	PreviewClient    preview_secret_service.ClientService
	Client           secret_service.ClientService
}

type RotatingSecretConfig struct {
	Version                 string
	SecretName              string `yaml:"secret_name"`
	RotationIntegrationType string `yaml:"rotation_integration_type"`
	RotationIntegrationName string `yaml:"rotation_integration_name"`
	RotationPolicyName      string `yaml:"rotation_policy_name"`
	Details                 map[string]interface{}
}

type MongoDBRole struct {
	RoleName       string `mapstructure:"role_name"`
	DatabaseName   string `mapstructure:"database_name"`
	CollectionName string `mapstructure:"collection_name"`
}

func createRotatingRun(opts *CreateRotatingOpts) error {
	fmt.Fprintf(opts.IO.Err(), "starting")
	f, err := os.ReadFile(opts.SecretConfigFile)
	if err != nil {
		return fmt.Errorf("failed to open config file: %w", err)
	}

	var rsc RotatingSecretConfig
	err = yaml.Unmarshal(f, &rsc)
	if err != nil {
		return fmt.Errorf("failed to unmarshal policy file: %w", err)
	}
	fmt.Fprintf(opts.IO.Err(), "read config %+v\n", rsc)
	if rsc.RotationIntegrationType == "mongodb" {
		roles := rsc.Details["roles"].([]interface{})
		req := preview_secret_service.NewCreateMongoDBAtlasRotatingSecretParamsWithContext(opts.Ctx)
		req.OrganizationID = opts.Profile.OrganizationID
		req.ProjectID = opts.Profile.ProjectID
		req.AppName = opts.AppName

		var reqRoles []*models.Secrets20231128MongoDBRole
		for _, r := range roles {
			var castRole MongoDBRole
			decoder, _ := mapstructure.NewDecoder(&mapstructure.DecoderConfig{WeaklyTypedInput: true, Result: &castRole})
			if err := decoder.Decode(r); err != nil {
				return fmt.Errorf("invalid decoding to mongodbrole")
			}

			reqRole := &models.Secrets20231128MongoDBRole{
				CollectionName: castRole.CollectionName,
				RoleName:       castRole.RoleName,
				DatabaseName:   castRole.DatabaseName,
			}
			reqRoles = append(reqRoles, reqRole)
		}
		req.Body = preview_secret_service.CreateMongoDBAtlasRotatingSecretBody{
			SecretName:              rsc.SecretName,
			RotationIntegrationName: rsc.RotationIntegrationName,
			RotationPolicyName:      rsc.RotationPolicyName,
			MongodbGroupID:          rsc.Details["group_id"].(string),
			MongodbRoles:            reqRoles,
		}
		fmt.Fprintf(opts.IO.Err(), "creating secret with body %+v\n", req.Body)
		_, err := opts.PreviewClient.CreateMongoDBAtlasRotatingSecret(req, nil)
		if err != nil {
			return fmt.Errorf("failed to create secret with name %q: %w", rsc.SecretName, err)
		}

		fmt.Fprintln(opts.IO.Err())
		fmt.Fprintf(opts.IO.Err(), "%s Successfully created secret with name %q\n", opts.IO.ColorScheme().SuccessIcon(), rsc.SecretName)
	} else {
		return fmt.Errorf("invalid type")
	}
	return nil
}
