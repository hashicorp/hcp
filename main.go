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
	"github.com/hashicorp/hcp/internal/pkg/versioncheck"
	"github.com/hashicorp/hcp/version"
	"github.com/mitchellh/cli"
	"github.com/posener/complete"
	"golang.org/x/oauth2"
)

const (
	// versionCheckStatePath is the path to store the result of checking for a new
	// version.
	versionCheckStatePath = "~/.config/hcp/version_check_state.json"
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

	// Create the version checker
	checker, err := versioncheck.New(io, versionCheckStatePath)
	if err != nil {
		// On error, a nil checker is returned but it is still safe to call
		// Check/Display.
		fmt.Fprintf(os.Stderr, "failed to configure version checker: %v\n", err)
	}

	// Start checking for a new version as soon as possible
	go func() {
		if err := checker.Check(shutdownCtx); err != nil && !errors.Is(err, context.Canceled) {
			fmt.Fprintf(os.Stderr, "failed to check for new version: %v\n", err)
		}
	}()

	// Load profile to get geography setting for HCP configuration
	var geography string
	profile, err := loadActiveProfile()
	if err != nil {
		fmt.Fprintln(io.Err(), err)
		return 1
	}
	geography = profile.GetGeography()

	// Validate geography match and warn about mismatches
	if err := validateGeographyMatch(io, geography); err != nil {
		fmt.Fprintf(io.Err(), "Warning: %v\n", err)
	}

	// Create the HCP Config with geography setting
	hcpCfg, err := auth.GetHCPConfigWithGeography(geography, hcpconf.WithoutBrowserLogin())
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

	// Display the check results if we aren't being run in autocomplete. The
	// check results will only be displayed if there is a new version and we
	// haven't prompted recently.
	if !isAutocomplete() {
		checker.Display()
	}

	return status
}

// loadActiveProfile loads the active profile
func loadActiveProfile() (*profile.Profile, error) {
	// Create the profile loader
	loader, err := profile.NewLoader()
	if err != nil {
		return nil, fmt.Errorf("failed to create profile loader: %w", err)
	}

	// Load the active profile
	activeProfile, err := loader.GetActiveProfile()
	if err != nil {
		if !errors.Is(err, profile.ErrNoActiveProfileFilePresent) && !errors.Is(err, profile.ErrActiveProfileFileEmpty) {
			return nil, fmt.Errorf("failed to read active profile: %w", err)
		}

		if err := loader.DefaultActiveProfile().Write(); err != nil {
			return nil, fmt.Errorf("failed to save default active profile config: %w", err)
		}

		if err := loader.DefaultProfile().Write(); err != nil {
			return nil, fmt.Errorf("failed to save default profile config: %w", err)
		}

		activeProfile, err = loader.GetActiveProfile()
		if err != nil {
			return nil, fmt.Errorf("failed to save default active profile config: %w", err)
		}
	}

	return loader.LoadProfile(activeProfile.Name)
}

// loadProfile loads the active profile and if one doesn't exist, a default
// profile is created.
func loadProfile(ctx context.Context, iam iam_service.ClientService, tokenSource oauth2.TokenSource) (*profile.Profile, error) {
	// Get the active profile
	p, err := loadActiveProfile()
	if err != nil {
		return nil, err
	}

	// If the profile has an org, or we don't have a valid access
	// token, skip trying to infer the organization and project.
	tkn, err := tokenSource.Token()
	if p.OrganizationID != "" || err != nil || !tkn.Expiry.After(time.Now()) {
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

	// Set the organization.
	p.OrganizationID = ident.Payload.Principal.Service.OrganizationID

	// Only set the project if it is not already set.
	if p.ProjectID == "" {
		p.ProjectID = ident.Payload.Principal.Service.ProjectID
	}

	// Save the profile.
	if err := p.Write(); err != nil {
		return nil, fmt.Errorf("failed to save default profile: %w", err)
	}

	return p, nil
}

// isAutocomplete returns true if the CLI is being run in an autocomplete
// context.
func isAutocomplete() bool {
	return os.Getenv("COMP_LINE") != "" && os.Getenv("COMP_POINT") != ""
}

// validateGeographyMatch checks if the profile geography matches the cached geography
// and provides helpful warnings when there's a mismatch.
func validateGeographyMatch(io iostreams.IOStreams, profileGeography string) error {
	// Skip validation if profile geography is not set
	if profileGeography == "" {
		return nil
	}

	// Check for geography mismatch
	isMatch, cachedGeo, err := auth.ValidateGeographyMatch(profileGeography)
	if err != nil {
		// Don't fail the entire command for validation errors
		return fmt.Errorf("could not validate geography match: %w", err)
	}

	// If there's no cached geography or they match, no warning needed
	if cachedGeo == "" || isMatch {
		return nil
	}

	// There's a mismatch - provide helpful guidance
	cs := io.ColorScheme()
	errorMsg := fmt.Sprintf(
		"Profile is set to %s but you are authenticated to %s. This may cause authentication errors.\n"+
			"To fix this mismatch, run: %s",
		cs.String(profileGeography).Bold(), cs.String(cachedGeo).Bold(), cs.String("hcp auth login").Bold())
	return fmt.Errorf("%s", cs.String(errorMsg).Color(cs.Yellow()))
}
