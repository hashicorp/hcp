// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package run

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/hashicorp/consul-template/child"
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

type RunOpts struct {
	Ctx     context.Context
	Profile *profile.Profile
	IO      iostreams.IOStreams
	Output  *format.Outputter

	App           string
	Command       string
	PreviewClient preview_secret_service.ClientService
	Client        secret_service.ClientService
}

func NewCmdRun(ctx *cmd.Context, runF func(*RunOpts) error) *cmd.Command {
	opts := &RunOpts{
		Ctx:           ctx.ShutdownCtx,
		Profile:       ctx.Profile,
		IO:            ctx.IO,
		Output:        ctx.Output,
		PreviewClient: preview_secret_service.New(ctx.HCP, nil),
		Client:        secret_service.New(ctx.HCP, nil),
	}

	cmd := &cmd.Command{
		Name:      "run",
		ShortHelp: "Run a process with secrets from a Vault Secrets app.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp vault-secrets run" }} command lets you
		run the provided command as a child process while injecting
        all of the app's secrets as environment variables, with all secret names
        converted to upper-case. The stdout/stderr from the child
        process are forwarded to the top level 'hcp vault-secrets run' command.
		`),
		Examples: []cmd.Example{
			{
				Preamble: `Inject secrets as an environment variable:`,
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				$ hcp vault-secrets run --command="env"
				`),
			},
		},
		Flags: cmd.Flags{
			Local: []*cmd.Flag{
				{
					Name:         "command",
					DisplayValue: "COMMAND",
					Description:  "Defines the invocation of the child process to inject secrets to.",
					Value:        flagvalue.Simple("", &opts.Command),
					Required:     true,
				},
				{
					Name:         "app",
					DisplayValue: "APP",
					Description:  "The application you want to pull all secrets from.",
					Value:        appname.Flag(),
				},
			},
		},
		PersistentPreRun: func(c *cmd.Command, args []string) error {
			return appname.Require(ctx)
		},
		RunF: func(c *cmd.Command, args []string) error {
			opts.App = appname.Get()

			if runF != nil {
				return runF(opts)
			}
			return runRun(opts)
		},
	}

	return cmd
}

func runRun(opts *RunOpts) (err error) {
	envSecrets, err := getAllSecretsForEnv(opts)
	if err != nil {
		return fmt.Errorf("failed to run with secrets in app %q: %w", opts.App, err)
	}

	childProcess, err := setupChildProcess(opts.Command, envSecrets)
	if err != nil {
		return fmt.Errorf("failed to run with secrets in app %q: %w", opts.App, err)
	}

	if err := childProcess.Start(); err != nil {
		return fmt.Errorf("failed to run with secrets in app %q: %w", opts.App, err)

	}

	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, os.Interrupt, syscall.SIGTERM)
	for {
		select {
		case <-sigC:
			childProcess.Stop()
			return fmt.Errorf("failed to run with secrets in app %q: %w", opts.App, err)
		case <-childProcess.ExitCh():
			return nil
		}
	}
}

func getAllSecretsForEnv(opts *RunOpts) ([]string, error) {
	params := preview_secret_service.NewOpenAppSecretsParamsWithContext(opts.Ctx)
	params.OrganizationID = opts.Profile.OrganizationID
	params.ProjectID = opts.Profile.ProjectID
	params.AppName = opts.App

	res, err := opts.PreviewClient.OpenAppSecrets(params, nil)
	if err != nil {
		return nil, err
	}

	// get existing environment variables detectable by runtime
	result := os.Environ()

	for _, secret := range res.Payload.Secrets {
		// we need to append results in case of duplicates we want secrets to override
		// only supporting static secrets
		if secret.StaticVersion != nil {
			result = append(result, strings.ToUpper(secret.Name)+"="+secret.StaticVersion.Value)
		}
	}

	return result, nil
}

func setupChildProcess(command string, envVars []string) (*child.Child, error) {
	pieces := strings.Split(command, " ")
	cmd := pieces[0]
	var args []string
	if len(pieces) > 1 {
		args = pieces[1:]
	}

	input := &child.NewInput{
		Stdin:       os.Stdin,
		Stdout:      os.Stdout,
		Stderr:      os.Stderr,
		Command:     cmd,
		Args:        args,
		Env:         envVars,
		KillSignal:  os.Kill,
		KillTimeout: 30 * time.Second,
		Splay:       0,
		Setpgid:     true,
	}

	return child.New(input)
}
