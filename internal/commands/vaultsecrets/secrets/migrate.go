// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package secrets

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-11-28/client/secret_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-11-28/models"
	"github.com/hashicorp/hcp/internal/commands/vaultsecrets/secrets/appname"

	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
	"github.com/hashicorp/vault/api"
)

func NewCmdMigrate(ctx *cmd.Context, runF func(opt *MigrateOpt) error) *cmd.Command {
	opts := &MigrateOpt{
		Ctx:     ctx.ShutdownCtx,
		Profile: ctx.Profile,
		IO:      ctx.IO,
		Output:  ctx.Output,
		Client:  secret_service.New(ctx.HCP, nil),
	}

	cmd := &cmd.Command{
		Name:      "migrate",
		ShortHelp: "Migrate an application's secrets.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp vault-secrets secrets migrate" }} command migrate all secrets under a Vault Secrets application to a Vault cluster under KV secret engine.
		Path will be "hvs".

		Individual secrets can be read using
		{{ template "mdCodeOrBold" "hcp vault-secrets secrets read" }} subcommand.
		`),
		Examples: []cmd.Example{
			{
				Preamble: `Migrate all secrets under the specified Vault Secrets application to Vault cluster under "hvs" path under KV secret engine:`,
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				$ hcp vault-secrets secrets migrate
				`),
			},
			{
				Preamble: `Migrate all secrets under the specified Vault Secrets application to Vault cluster under "hvs" path under KV secret engine:`,
				Command: heredoc.New(ctx.IO, heredoc.WithNoWrap()).Must(`
				$ hcp vault-secrets secrets migrate --app test-app
				`),
			},
		},
		RunF: func(c *cmd.Command, args []string) error {
			opts.AppName = appname.Get()
			if runF != nil {
				return runF(opts)
			}
			return migrateRun(opts)
		},
	}

	return cmd
}

type MigrateOpt struct {
	Ctx     context.Context
	Profile *profile.Profile
	IO      iostreams.IOStreams
	Output  *format.Outputter

	AppName string
	Client  secret_service.ClientService
}

func migrateRun(opts *MigrateOpt) error {
	req := secret_service.NewListAppSecretsParamsWithContext(opts.Ctx)
	req.OrganizationID = opts.Profile.OrganizationID
	req.ProjectID = opts.Profile.ProjectID
	req.AppName = opts.AppName

	fmt.Printf(fmt.Sprintf("Migrating secrets in App %s...\n", req.AppName))

	config := api.DefaultConfig()
	client, err := api.NewClient(config)
	if err != nil {
		return fmt.Errorf("failed to create Vault client: %w", err)
	}
	client.SetToken(os.Getenv("VAULT_TOKEN"))

	mountInput := &api.MountInput{
		Type:        "kv",
		Description: "Key-Value Secrets Engine",
		Options: map[string]string{
			"version": "2", // KV version 2
		},
	}

	mounts, err := client.Sys().ListMounts()
	if err != nil {
		return fmt.Errorf("failed to list mounts: %w", err)
	}

	if _, ok := mounts["hvs/"]; !ok {
		err = client.Sys().Mount("hvs/", mountInput)
		if err != nil {
			return fmt.Errorf("failed to mount secret engine")
		}
	}

	secretsData := make(map[string]interface{})

	var secrets []*models.Secrets20231128Secret
	for {
		resp, err := opts.Client.ListAppSecrets(req, nil)
		if err != nil {
			return fmt.Errorf("failed to list secrets: %w", err)
		}

		secrets = append(secrets, resp.Payload.Secrets...)
		next := resp.Payload.Pagination.NextPageToken
		req.PaginationNextPageToken = &next

		for _, secret := range secrets {
			openReq := secret_service.NewOpenAppSecretParamsWithContext(opts.Ctx)
			openReq.OrganizationID = opts.Profile.OrganizationID
			openReq.ProjectID = opts.Profile.ProjectID
			openReq.AppName = opts.AppName
			openReq.SecretName = secret.Name
			resp, err := opts.Client.OpenAppSecret(openReq, nil)
			if err != nil {
				return fmt.Errorf("failed to open the secret %q: %w", err)
			}

			var secretValue string

			switch resp.Payload.Secret.Type {
			case secretTypeRotating:
				secretValue, err = formatSecretMap(resp.Payload.Secret.RotatingVersion.Values)
				if err != nil {
					secretValue = "<< COULD NOT ENCODE TO JSON >>"
				}
			case secretTypeDynamic:
				secretValue, err = formatSecretMap(resp.Payload.Secret.DynamicInstance.Values)
				if err != nil {
					secretValue = "<< COULD NOT ENCODE TO JSON >>"
				}
			case secretTypeKV:
				secretValue = resp.Payload.Secret.StaticVersion.Value
			default:
				secretValue = "<< SECRET TYPE NOT SUPPORTED >>"
			}

			secretsData[secret.Name] = secretValue
		}

		if resp.Payload.Pagination == nil || resp.Payload.Pagination.NextPageToken == "" {
			break
		}
	}
	_, err = client.Logical().Write("hvs/data/"+req.AppName, map[string]interface{}{
		"data": secretsData,
	})
	if err != nil {
		log.Fatalf("Failed to write secret: %v", err)
	}
	return opts.Output.Display(newDisplayer().Secrets(secrets...))
}
