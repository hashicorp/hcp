// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package secrets

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	preview_secret_service "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/preview/2023-11-28/client/secret_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-06-13/client/secret_service"
	"github.com/hashicorp/hcp/internal/commands/vaultsecrets/secrets/appname"
	"github.com/hashicorp/hcp/internal/commands/vaultsecrets/secrets/helper"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/flagvalue"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
)

func NewCmdOpen(ctx *cmd.Context, runF func(*OpenOpts) error) *cmd.Command {
	opts := &OpenOpts{
		Ctx:           ctx.ShutdownCtx,
		Profile:       ctx.Profile,
		IO:            ctx.IO,
		Output:        ctx.Output,
		PreviewClient: preview_secret_service.New(ctx.HCP, nil),
		Client:        secret_service.New(ctx.HCP, nil),
	}

	cmd := &cmd.Command{
		Name:      "open",
		ShortHelp: "Open a secret.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp vault-secrets secrets open" }} command reads the plaintext value of a static, rotating, or dynamic secret from the Vault Secrets application.
		`),
		Examples: []cmd.Example{
			{
				Preamble: `Open plaintext secret:`,
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				$ hcp vault-secrets secret open "test_secret"
				`),
			},
		},
		Args: cmd.PositionalArguments{
			Args: []cmd.PositionalArgument{
				{
					Name:          "NAME",
					Documentation: "The name of the secret to open.",
				},
			},
		},
		Flags: cmd.Flags{
			Local: []*cmd.Flag{
				{
					Name:         "out-file",
					DisplayValue: "OUTPUT_FILE_PATH",
					Shorthand:    "o",
					Description:  "File path where the secret value should be written.",
					Value:        flagvalue.Simple("", &opts.OutputFilePath),
				},
			},
		},
		RunF: func(c *cmd.Command, args []string) error {
			opts.AppName = appname.Get()
			opts.SecretName = args[0]

			if runF != nil {
				return runF(opts)
			}
			return openRun(opts)
		},
	}
	cmd.Args.Autocomplete = helper.PredictSecretName(ctx, cmd, opts.PreviewClient)

	return cmd
}

type OpenOpts struct {
	Ctx     context.Context
	Profile *profile.Profile
	IO      iostreams.IOStreams
	Output  *format.Outputter

	AppName        string
	SecretName     string
	OutputFilePath string
	PreviewClient  preview_secret_service.ClientService
	Client         secret_service.ClientService
}

func openRun(opts *OpenOpts) (err error) {
	var fd *os.File
	if opts.OutputFilePath != "" {
		fd, err = os.OpenFile(opts.OutputFilePath, os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			return fmt.Errorf("failed to open the outout file %q: %w", opts.OutputFilePath, err)
		}
	}
	defer func() {
		if opts.OutputFilePath != "" {
			err = errors.Join(err, fd.Close())
		}
	}()

	req := preview_secret_service.NewOpenAppSecretParamsWithContext(opts.Ctx)
	req.OrganizationID = opts.Profile.OrganizationID
	req.ProjectID = opts.Profile.ProjectID
	req.AppName = opts.AppName
	req.SecretName = opts.SecretName

	resp, err := opts.PreviewClient.OpenAppSecret(req, nil)
	if err != nil {
		return fmt.Errorf("failed to read the secret %q: %w", opts.SecretName, err)
	}

	var (
		secretValue string
		field       format.Field
	)

	switch {
	case resp.Payload.Secret.RotatingVersion != nil:
		secretValue, err = formatSecretMap(resp.Payload.Secret.RotatingVersion.Values)
		if err != nil {
			secretValue = "<< COULD NOT ENCODE TO JSON >>"
		}

		field = format.Field{
			Name:        "Values",
			ValueFormat: `{{ range $key, $value := .RotatingVersion.Values }}{{printf "%s: %s\n" $key $value}}{{ end }}`,
		}
	case resp.Payload.Secret.DynamicInstance != nil:
		secretValue, err = formatSecretMap(resp.Payload.Secret.DynamicInstance.Values)
		if err != nil {
			secretValue = "<< COULD NOT ENCODE TO JSON >>"
		}

		field = format.Field{
			Name:        "Values",
			ValueFormat: `{{ range $key, $value := .DynamicInstance.Values }}{{printf "%s: %s\n" $key $value}}{{ end }}`,
		}
	case resp.Payload.Secret.StaticVersion != nil:
		secretValue = resp.Payload.Secret.StaticVersion.Value

		field = format.Field{
			Name:        "Value",
			ValueFormat: "{{ .StaticVersion.Value }}",
		}
	default:
		secretValue = "<< SECRET TYPE NOT SUPPORTED >>"
	}

	if opts.OutputFilePath != "" {
		_, err = fd.WriteString(secretValue)
		if err != nil {
			return fmt.Errorf("failed to write the secret value to the output file %q: %w", opts.OutputFilePath, err)
		}
		fmt.Fprintf(opts.IO.Err(), "%s Successfully wrote plaintext secret with name %q to path %q\n", opts.IO.ColorScheme().SuccessIcon(), opts.SecretName, opts.OutputFilePath)
		return nil
	}

	displayer := newDisplayer(true).OpenAppSecrets(resp.Payload.Secret).
		SetDefaultFormat(format.Pretty).AddFields(field)
	return opts.Output.Display(displayer)
}

func formatSecretMap(secretMap map[string]string) (string, error) {
	var buf strings.Builder
	encoder := json.NewEncoder(&buf)
	if err := encoder.Encode(secretMap); err != nil {
		return "", err
	}
	return buf.String(), nil
}
