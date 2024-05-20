// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package secrets

import (
	"context"
	"fmt"

	preview_secret_service "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/preview/2023-11-28/client/secret_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-06-13/client/secret_service"
	"github.com/hashicorp/hcp/internal/commands/vaultsecrets/secrets/appname"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/flagvalue"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
)

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
		Name:      "delete",
		ShortHelp: "Update a static secret.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp vault-secrets secrets delete" }} command updates a static secret under an Vault Secrets application.`),
		Examples: []cmd.Example{
			{
				Preamble: `Update a new secret in Vault Secrets application on active profile:`,
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				$ hcp vault-secrets secrets update secret_1 --data-file=tmp/secrets1.txt
				`),
			},
			{
				Preamble: `Update a new secret in Vault Secrets application by piping the plaintext secret from a command output:`,
				Command: heredoc.New(ctx.IO, heredoc.WithNoWrap()).Must(`
				$ echo -n "my super secret" | hcp vault-secrets secrets update secret_2 --data-file=-
				`),
			},
			{
				Preamble: `Update a new secret in the specified Vault Secrets application:`,
				Command: heredoc.New(ctx.IO, heredoc.WithNoWrap()).Must(`
				$ hcp vault-secrets secrets update secret_3 --app test-app --secret_file=/tmp/secrets2.txt
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

type UpdateOpts struct {
	Ctx     context.Context
	Profile *profile.Profile
	IO      iostreams.IOStreams
	Output  *format.Outputter

	AppName              string
	SecretName           string
	SecretValuePlaintext string
	SecretFilePath       string
	PreviewClient        preview_secret_service.ClientService
	Client               secret_service.ClientService
}

func updateRun(opts *UpdateOpts) error {
	getReq := secret_service.NewGetAppSecretParamsWithContext(opts.Ctx)
	getReq.LocationOrganizationID = opts.Profile.OrganizationID
	getReq.LocationProjectID = opts.Profile.ProjectID
	getReq.AppName = opts.AppName
	getReq.SecretName = opts.SecretName

	_, err := opts.Client.GetAppSecret(getReq, nil)
	if err != nil {
		return fmt.Errorf("secret %q not found: %w", opts.SecretName, err)
	}

	opts.SecretValuePlaintext, err = readPlainTextSecret(opts.SecretValuePlaintext, opts.SecretFilePath, opts.IO.In())
	if err != nil {
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
		return fmt.Errorf("failed to update secret with name %q: %w", opts.SecretName, err)
	}
	if err := opts.Output.Display(newDisplayer(true).Secrets(resp.Payload.Secret)); err != nil {
		return err
	}

	command := fmt.Sprintf(`$ hcp vault-secrets secrets read %s --app %s`, opts.SecretName, req.AppName)
	fmt.Fprintln(opts.IO.Err())
	fmt.Fprintf(opts.IO.Err(), "%s Successfully updated secret with name %q\n", opts.IO.ColorScheme().SuccessIcon(), opts.SecretName)
	fmt.Fprintln(opts.IO.Err())
	fmt.Fprintf(opts.IO.Err(), `To read your secret, run:
  %s`, opts.IO.ColorScheme().String(command).Bold())
	fmt.Fprintln(opts.IO.Err())
	return nil
}
