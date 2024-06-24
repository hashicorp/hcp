package integrations

import (
	"context"
	"fmt"
	"os"

	preview_secret_service "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/preview/2023-11-28/client/secret_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-06-13/client/secret_service"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/flagvalue"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
	"github.com/posener/complete"
	"gopkg.in/yaml.v3"
)

type CreateOpts struct {
	Ctx     context.Context
	Profile *profile.Profile
	Output  *format.Outputter
	IO      iostreams.IOStreams

	ConfigFilePath string
	Client         secret_service.ClientService
	PreviewClient  preview_secret_service.ClientService
}

func NewCmdCreate(ctx *cmd.Context, runF func(*CreateOpts) error) *cmd.Command {
	opts := &CreateOpts{
		Ctx:           ctx.ShutdownCtx,
		Profile:       ctx.Profile,
		Output:        ctx.Output,
		IO:            ctx.IO,
		Client:        secret_service.New(ctx.HCP, nil),
		PreviewClient: preview_secret_service.New(ctx.HCP, nil),
	}

	cmd := &cmd.Command{
		Name:      "create",
		ShortHelp: "help",
		LongHelp:  heredoc.New(ctx.IO).Must(`help`),
		Flags: cmd.Flags{
			Local: []*cmd.Flag{
				{
					Name:         "config-file",
					DisplayValue: "PATH",
					Description:  "The path to a file containing an IAM policy object.",
					Value:        flagvalue.Simple("", &opts.ConfigFilePath),
					Required:     true,
					Autocomplete: complete.PredictFiles("*.json"),
				},
			},
		},
		RunF: func(c *cmd.Command, args []string) error {

			if runF != nil {
				return runF(opts)
			}

			return createIntegration(opts)
		},
		PersistentPreRun: func(c *cmd.Command, args []string) error {
			return cmd.RequireOrgAndProject(ctx)
		},
	}

	return cmd
}

type IntegrationConfig struct {
	Version string
	Type    string
	Name    string
	Details map[string]interface{}
}

type MongoDBDetails struct {
}

func createIntegration(opts *CreateOpts) error {
	// Open the file
	f, err := os.ReadFile(opts.ConfigFilePath)
	if err != nil {
		return fmt.Errorf("failed to open config file: %w", err)
	}

	var i IntegrationConfig
	err = yaml.Unmarshal(f, &i)
	if err != nil {
		return fmt.Errorf("failed to unmarshal policy file: %w", err)
	}

	if i.Type == "mongodb" {
		privkey := i.Details["private_key"].(string)
		fmt.Fprintf(opts.IO.Err(), "got privkey %q", privkey)
		resp, err := opts.PreviewClient.CreateMongoDBAtlasRotationIntegration(&preview_secret_service.CreateMongoDBAtlasRotationIntegrationParams{
			Context:        opts.Ctx,
			ProjectID:      opts.Profile.ProjectID,
			OrganizationID: opts.Profile.OrganizationID,
			Body: preview_secret_service.CreateMongoDBAtlasRotationIntegrationBody{
				IntegrationName:      i.Name,
				MongodbAPIPrivateKey: i.Details["private_key"].(string),
				MongodbAPIPublicKey:  i.Details["public_key"].(string),
			},
		}, nil)

		if err != nil {
			return fmt.Errorf("failed to create integration: %w", err)
		}

		fmt.Fprintln(opts.IO.Err())
		fmt.Fprintf(opts.IO.Err(), "%s Successfully created integration with name %q\n", opts.IO.ColorScheme().SuccessIcon(), resp.Payload.RotationIntegration.IntegrationName)
	} else {
		fmt.Errorf("invalid type given")
	}

	return nil
}
