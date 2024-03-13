package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

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
)

func main() {
	os.Exit(realMain())
}

func realMain() int {
	args := os.Args[1:]

	// Listen for interupts
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

	// Create the profile loader
	profiles, err := profile.NewLoader()
	if err != nil {
		fmt.Fprintf(io.Err(), "failed to create profile loader: %v\n", err)
		return 1
	}

	// Load the active profile
	activeProfile, err := profiles.GetActiveProfile()
	if err != nil {
		if !errors.Is(err, profile.ErrNoActiveProfileFilePresent) && !errors.Is(err, profile.ErrActiveProfileFileEmpty) {
			fmt.Fprintf(io.Err(), "failed to read active profile: %v\n", err)
			return 1
		}

		if err := profiles.DefaultActiveProfile().Write(); err != nil {
			fmt.Fprintf(io.Err(), "failed to save default active profile config: %v\n", err)
			return 1
		}

		if err := profiles.DefaultProfile().Write(); err != nil {
			fmt.Fprintf(io.Err(), "failed to save default profile config: %v\n", err)
			return 1
		}

		activeProfile, err = profiles.GetActiveProfile()
		if err != nil {
			fmt.Fprintf(io.Err(), "failed to save default active profile config: %v\n", err)
			return 1
		}
	}

	// Get the active profile
	p, err := profiles.LoadProfile(activeProfile.Name)
	if err != nil {
		p = profiles.DefaultProfile()
		p.Name = activeProfile.Name
		if err := p.Write(); err != nil {
			fmt.Fprintf(io.Err(), "failed to save default profile config: %v\n", err)
			return 1
		}
	}

	// If the profile has disabled color, disable on the iostream.
	if p.Core != nil && p.Core.NoColor != nil && *p.Core.NoColor {
		io.ForceNoColor()
	}

	// Create the HCP Config
	hcpCfg, err := auth.GetHCPConfig(hcpconf.WithoutBrowserLogin())
	if err != nil {
		fmt.Fprintf(io.Err(), "failed to instantiate HCP config: %v\n", err)
		return 1
	}

	hconfig := httpclient.Config{
		HCPConfig:     hcpCfg,
		SourceChannel: getSourceChannel(),
	}

	hcpClient, err := httpclient.New(hconfig)
	if err != nil {
		fmt.Fprintf(io.Err(), "failed to create HCP API client: %v\n", err)
		return 1
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

// getSourceChannel returns the source channel for the CLI
func getSourceChannel() string {
	return fmt.Sprintf("hcp-cli/%s", version.FullVersion())
}
