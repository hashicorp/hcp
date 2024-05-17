// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/client/iam_service"
	hcpconf "github.com/hashicorp/hcp-sdk-go/config"
	"github.com/hashicorp/hcp-sdk-go/httpclient"
	"github.com/hashicorp/hcp/internal/commands/hcp"
	"github.com/hashicorp/hcp/internal/pkg/auth"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
	"github.com/hashicorp/hcp/version"
	"github.com/mitchellh/cli"
	"github.com/posener/complete"
	"golang.org/x/oauth2"
)

func main() {
	os.Exit(realMain())
}

func realMain() int {
	args := os.Args[1:]

	// Listen for interrupts
	shutdownCtx, shutdown := context.WithCancelCause(context.Background())
	defer shutdown(nil)
	go func() {
		signalCh := make(chan os.Signal, 1)
		signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)
		sig := <-signalCh
		shutdown(fmt.Errorf("command received signal: %s", sig))
	}()

	// Create our iostreams
	io, err := iostreams.System(shutdownCtx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to configure iostreams: %v\n", err)
		return 1
	}
	defer func() {
		if err := io.RestoreConsole(); err != nil {
			fmt.Fprintf(os.Stderr, "failed to restore console output: %v\n", err)
		}
	}()

	// Create the HCP Config
	hcpCfg, err := auth.GetHCPConfig(hcpconf.WithoutBrowserLogin())
	if err != nil {
		fmt.Fprintf(io.Err(), "failed to instantiate HCP config: %v\n", err)
		return 1
	}

	hconfig := httpclient.Config{
		HCPConfig:     hcpCfg,
		SourceChannel: version.GetSourceChannel(),
	}

	hcpClient, err := httpclient.New(hconfig)
	if err != nil {
		fmt.Fprintf(io.Err(), "failed to create HCP API client: %v\n", err)
		return 1
	}

	// Load the profile
	p, err := loadProfile(shutdownCtx, iam_service.New(hcpClient, nil), hcpCfg)
	if err != nil {
		fmt.Fprintln(io.Err(), err)
		return 1
	}

	// If the profile has disabled color, disable on the iostream.
	if p.Core != nil && p.Core.NoColor != nil && *p.Core.NoColor {
		io.ForceNoColor()
	}

	// Create the command context
	cCtx := &cmd.Context{
		IO:          io,
		Profile:     p,
		Output:      format.New(io),
		HCP:         hcpClient,
		ShutdownCtx: shutdownCtx,
	}

	// Get the HCP Root command
	hcpCmd := hcp.NewCmdHcp(cCtx)
	cmdMap := cmd.ToCommandMap(hcpCmd)

	c := cli.CLI{
		Name:                       hcpCmd.Name,
		Args:                       args,
		Commands:                   cmdMap,
		HelpFunc:                   cmd.RootHelpFunc(hcpCmd),
		Version:                    "0.0.1",
		Autocomplete:               true,
		AutocompleteNoDefaultFlags: true,
		AutocompleteGlobalFlags: map[string]complete.Predictor{
			"--help": complete.PredictNothing,
		},
	}

	status, err := c.Run()
	if err != nil {
		fmt.Fprintf(io.Err(), "Error executing hcp: %s\n", err.Error())
	}

	return status
}

// loadProfile loads the active profile and if one doesn't exist, a default
// profile is created.
func loadProfile(ctx context.Context, iam iam_service.ClientService, tokenSource oauth2.TokenSource) (*profile.Profile, error) {
	// Create the profile loader
	profiles, err := profile.NewLoader()
	if err != nil {
		return nil, fmt.Errorf("failed to create profile loader: %w", err)
	}

	// Load the active profile
	activeProfile, err := profiles.GetActiveProfile()
	if err != nil {
		if !errors.Is(err, profile.ErrNoActiveProfileFilePresent) && !errors.Is(err, profile.ErrActiveProfileFileEmpty) {
			return nil, fmt.Errorf("failed to read active profile: %w", err)
		}

		if err := profiles.DefaultActiveProfile().Write(); err != nil {
			return nil, fmt.Errorf("failed to save default active profile config: %w", err)
		}

		if err := profiles.DefaultProfile().Write(); err != nil {
			return nil, fmt.Errorf("failed to save default profile config: %w", err)
		}

		activeProfile, err = profiles.GetActiveProfile()
		if err != nil {
			return nil, fmt.Errorf("failed to save default active profile config: %w", err)
		}
	}

	// Get the active profile
	p, err := profiles.LoadProfile(activeProfile.Name)
	if err != nil {
		p = profiles.DefaultProfile()
		p.Name = activeProfile.Name
		if err := p.Write(); err != nil {
			return nil, fmt.Errorf("failed to save default profile config: %w", err)
		}
	}

	// If the profile has an org/project, or we don't have a valid access
	// token, skip trying to infer the organization and project.
	tkn, err := tokenSource.Token()
	if p.OrganizationID != "" || p.ProjectID != "" || err != nil || !tkn.Expiry.After(time.Now()) {
		return p, nil
	}

	// Get the caller identity. If it is a service principal, we can set the
	// organization and potentially project automatically. This is particularly
	// useful when authenticating the CLI with a service principal and running
	// one off commands, where the profile have not been set interactively.
	callerIdentityParams := iam_service.NewIamServiceGetCallerIdentityParamsWithContext(ctx)
	ident, err := iam.IamServiceGetCallerIdentity(callerIdentityParams, nil)
	if err != nil {
		return p, nil
	}

	// Skip if the caller isn't a service principal
	if ident.Payload == nil || ident.Payload.Principal == nil || ident.Payload.Principal.Service == nil {
		return p, nil
	}

	// Set the organization and project. Project may be empty.
	p.OrganizationID = ident.Payload.Principal.Service.OrganizationID
	p.ProjectID = ident.Payload.Principal.Service.ProjectID

	// Save the profile.
	if err := p.Write(); err != nil {
		return nil, fmt.Errorf("failed to save default profile: %w", err)
	}

	return p, nil
}
