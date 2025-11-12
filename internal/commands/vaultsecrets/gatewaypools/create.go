// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: MPL-2.0

package gatewaypools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/hashicorp/hcp-sdk-go/auth"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-11-28/client/secret_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-11-28/models"
	"github.com/hashicorp/hcp-sdk-go/config/files"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/flagvalue"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
)

const (
	CredsFilePath  = "./creds.json"
	ConfigFilePath = "config.hcl"
)

type CreateOpts struct {
	Ctx     context.Context
	Profile *profile.Profile
	Output  *format.Outputter
	IO      iostreams.IOStreams

	GatewayPoolName  string
	Description      string
	OutDirPath       string
	ShowClientSecret bool
	Client           secret_service.ClientService
}

func NewCmdCreate(ctx *cmd.Context, runF func(*CreateOpts) error) *cmd.Command {
	opts := &CreateOpts{
		Ctx:     ctx.ShutdownCtx,
		Profile: ctx.Profile,
		Output:  ctx.Output,
		IO:      ctx.IO,
		Client:  secret_service.New(ctx.HCP, nil),
	}

	cmd := &cmd.Command{
		Name:      "create",
		ShortHelp: "Create a new Vault Secrets Gateway Pool.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp vault-secrets gateway-pools create" }} command creates a new Vault Secrets gateway pool.
		`),
		Examples: []cmd.Example{
			{
				Preamble: `Create a new gateway pool:`,
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				$ hcp vault-secrets gateway-pools create company-tunnel \
				  --description "Tunnels to corporate network."
				`),
			},
		},
		Args: cmd.PositionalArguments{
			Args: []cmd.PositionalArgument{
				{
					Name:          "NAME",
					Documentation: "The name of the gateway pool to create.",
				},
			},
		},
		Flags: cmd.Flags{
			Local: []*cmd.Flag{
				{
					Name:         "description",
					DisplayValue: "DESCRIPTION",
					Description:  "An optional description for the gateway pool to create.",
					Value:        flagvalue.Simple("", &opts.Description),
					Required:     false,
				},
				{
					Name:         "output-dir",
					DisplayValue: "OUTPUT_DIR_PATH",
					Shorthand:    "o",
					Description:  "Directory path where the gateway credentials file and config file should be written.",
					Value:        flagvalue.Simple("", &opts.OutDirPath),
					Required:     false,
				},
				{
					Name:          "show-client-secret",
					DisplayValue:  "SHOW_CLIENT_SECRET",
					IsBooleanFlag: true,
					Description:   "Show the client secret in the output. If this is not set, OUTPUT_DIR_PATH should be set.",
					Shorthand:     "s",
					Value:         flagvalue.Simple(false, &opts.ShowClientSecret),
					Required:      false,
				},
			},
		},
		RunF: func(c *cmd.Command, args []string) error {
			opts.GatewayPoolName = args[0]

			if runF != nil {
				return runF(opts)
			}
			return createRun(opts)
		},
	}

	return cmd
}

func createFields(showOauth bool) []format.Field {
	fields := []format.Field{
		{
			Name:        "Gateway Pool Name",
			ValueFormat: "{{ .GatewayPool.Name }}",
		},
		{
			Name:        "Gateway Pool Resource Name",
			ValueFormat: "{{ .GatewayPool.ResourceName }}",
		},
		{
			Name:        "Gateway Pool Resource ID",
			ValueFormat: "{{ .GatewayPool.ResourceID }}",
		},
		{
			Name:        "Description",
			ValueFormat: "{{ .GatewayPool.Description }}",
		},
	}

	if showOauth {
		fields = append(fields, []format.Field{
			{
				Name:        "Client ID",
				ValueFormat: "{{ .Oauth.ClientID }}",
			},
			{
				Name:        "Client Secret",
				ValueFormat: "{{ .Oauth.ClientSecret }}",
			},
		}...)
	}

	return fields
}

func createRun(opts *CreateOpts) error {
	if !opts.ShowClientSecret && opts.OutDirPath == "" {
		return fmt.Errorf("either show-client-secret or output-dir should be set")
	}
	if opts.OutDirPath != "" {
		if err := os.Mkdir(opts.OutDirPath, 0o700); err != nil {
			return fmt.Errorf("failed to create the output directory: %w", err)
		}
	}
	resp, err := opts.Client.CreateGatewayPool(&secret_service.CreateGatewayPoolParams{
		Context:        opts.Ctx,
		OrganizationID: opts.Profile.OrganizationID,
		ProjectID:      opts.Profile.ProjectID,
		Body: &models.SecretServiceCreateGatewayPoolBody{
			Name:        opts.GatewayPoolName,
			Description: opts.Description,
		},
	}, nil)
	if err != nil {
		return fmt.Errorf("failed to create gateway pool: %w", err)
	}

	oauth := &auth.OauthConfig{
		ClientID:     resp.Payload.ClientID,
		ClientSecret: resp.Payload.ClientSecret,
	}
	if opts.OutDirPath != "" {
		creds := &gatewayCreds{
			ProjectID: resp.Payload.GatewayPool.ProjectID,
			Scheme:    auth.CredentialFileSchemeServicePrincipal,
			Oauth:     oauth,
		}
		if err := writeGatewayCredentialFile(filepath.Join(opts.OutDirPath, CredsFilePath), creds); err != nil {
			return fmt.Errorf("failed to write the gateway credential file: %w", err)
		}

		c := &cloud{
			CredFile:     CredsFilePath,
			ResourceName: resp.Payload.GatewayPool.ResourceName,
		}
		if err := WriteConfig(filepath.Join(opts.OutDirPath, ConfigFilePath), &config{Cloud: c}); err != nil {
			return fmt.Errorf("failed to write the gateway config file: %w", err)
		}
	}

	// Display Oauth in the output if explicitly asked for it
	displayerOpts := &gatewayPoolWithOauth{
		GatewayPool: resp.Payload.GatewayPool,
	}
	if opts.ShowClientSecret {
		displayerOpts.Oauth = oauth
	}

	d := newDisplayer(true).GatewayPoolsWithOauth(displayerOpts)
	return opts.Output.Display(d.SetDefaultFormat(format.Pretty).AddFields(createFields(opts.ShowClientSecret)...))
}

type gatewayCreds struct {
	ProjectID string `json:"project_id,omitempty"`

	// Scheme is the authentication scheme which is service_principal_creds
	Scheme string `json:"scheme,omitempty"`

	Oauth *auth.OauthConfig `json:"oauth,omitempty"`
}

// writeGatewayCredentialFile writes the given credential file to the path.
func writeGatewayCredentialFile(path string, cf *gatewayCreds) error {
	data, err := json.MarshalIndent(cf, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, files.FileMode)
}

type cloud struct {
	// CredFile is the path to the credential json file.
	CredFile string `hcl:"cred_file"`
	// ResourceName is the resource name of the gateway pool.
	ResourceName string `hcl:"resource_name"`
}

type config struct {
	Cloud *cloud `hcl:"cloud,block"`
}

// WriteConfig writes the config to disk.
func WriteConfig(path string, c *config) error {
	f := hclwrite.NewEmptyFile()
	gohcl.EncodeIntoBody(c, f.Body())
	return os.WriteFile(path, f.Bytes(), 0o700)
}
