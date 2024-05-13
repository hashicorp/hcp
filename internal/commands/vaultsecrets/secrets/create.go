// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package secrets

import (
	"bufio"
	"context"
	"errors"
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
		The {{ template "mdCodeOrBold" "hcp vault-secrets secrets create" }} command creates a new static secret under an Vault Secrets App.

		Once the secret is created, it can be read using
		{{ template "mdCodeOrBold" "hcp vault-secrets secrets read" }} subcommand.
		`),
		Examples: []cmd.Example{
			{
				Preamble: `Create new secret in Vault Secrets application on active profile:`,
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				$ hcp vault-secrets secrets create secret-1 --data-file=/tmp/secrets1.txt
				`),
			},
			{
				Preamble: `Create secret in different Vault Secrets application, not active profile:`,
				Command: heredoc.New(ctx.IO, heredoc.WithNoWrap()).Must(`
				$ hcp vault-secrets secrets create secret-2 --app-name test-app --secret_file=/tmp/secrets2.txt
				`),
			},
		},
		Args: cmd.PositionalArguments{
			Args: []cmd.PositionalArgument{
				{
					Name:          "SECRET_NAME",
					Documentation: "The name of the secret to create.",
				},
			},
		},
		Flags: cmd.Flags{
			Local: []*cmd.Flag{
				{
					Name:         "data-file",
					DisplayValue: "DATA_FILE_PATH",
					Description:  "Absolute path to the secrets data file.",
					Value:        flagvalue.Simple("", &opts.SecretFilePath),
				},
			},
		},
		RunF: func(c *cmd.Command, args []string) error {
			opts.AppName = appName
			opts.SecretName = args[0]

			if runF != nil {
				return runF(opts)
			}
			return createRun(opts)
		},
		PersistentPreRun: func(c *cmd.Command, args []string) error {
			// Check if there's input piped to the command
			fi, err := opts.IO.InStat()
			if err != nil {
				c.Logger().Debug("failed to read whether input is piped or not", "error", err)
				return nil
			}
			if fi.Mode()&os.ModeCharDevice == 0 {
				scanner := bufio.NewScanner(opts.IO.In())
				for scanner.Scan() {
					opts.SecretFilePath = scanner.Text()
				}
			}
			return nil
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
	PreviewClient        preview_secret_service.ClientService
	Client               secret_service.ClientService
}

func createRun(opts *CreateOpts) error {
	if err := readPlainTextSecret(opts); err != nil {
		return err
	}

	req := preview_secret_service.NewCreateAppKVSecretParamsWithContext(opts.Ctx)
	req.OrganizationID = opts.Profile.OrganizationID
	req.ProjectID = opts.Profile.ProjectID
	req.AppName = opts.AppName

	req.Body = preview_secret_service.CreateAppKVSecretBody{
		Name:  opts.SecretName,
		Value: opts.SecretValuePlaintext,
	}

	resp, err := opts.PreviewClient.CreateAppKVSecret(req, nil)
	if err != nil {
		return fmt.Errorf("failed to create secret with name: %s - %w", opts.SecretName, err)
	}
	return opts.Output.Display(newDisplayer(true, resp.Payload.Secret))
}

func readPlainTextSecret(opts *CreateOpts) error {
	if opts.SecretValuePlaintext != "" {
		return nil
	}
	if opts.SecretFilePath == "" && opts.SecretValuePlaintext == "" {
		fmt.Fprintln(opts.IO.Err(), "Please enter the plaintext secret:")
		data, err := opts.IO.ReadSecret()
		if err != nil {
			return fmt.Errorf("failed to read the plaintext secret: %w", err)
		}

		if len(data) == 0 {
			return errors.New("secret value cannot be empty")
		}
		opts.SecretValuePlaintext = string(data)
		return nil
	}

	fileInfo, err := os.Stat(opts.SecretFilePath)
	if err != nil {
		return fmt.Errorf("%s failed to get data file info: %w", opts.IO.ColorScheme().FailureIcon(), err)
	}

	if fileInfo.Size() == 0 {
		return fmt.Errorf("%s data file cannot be empty", opts.IO.ColorScheme().FailureIcon())
	}

	data, err := os.ReadFile(opts.SecretFilePath)
	if err != nil {
		return fmt.Errorf("%s unable to read the data file: %w", opts.IO.ColorScheme().FailureIcon(), err)
	}
	opts.SecretValuePlaintext = string(data)
	return nil
}
