// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/hcp/internal/commands/hcp"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Define the flags
	var outputDir string
	var outputNavJSON string
	var linkPrefix string

	flag.StringVar(&outputDir, "output-dir", "web-docs", "The output directory for the generated documentation")
	flag.StringVar(&outputNavJSON, "output-nav-json", "web-docs/nav.json", "The output path for the generated nav json")
	flag.StringVar(&linkPrefix, "cmd-link-prefix", "/hcp/docs/cli/commands/", "Link prefix for the commands")

	// Parse the flags
	flag.Parse()

	// Create the command context
	l, err := profile.NewLoader()
	if err != nil {
		return fmt.Errorf("failed to create profile loader: %w", err)
	}

	io := iostreams.MD()
	ctx := &cmd.Context{
		IO:          io,
		Profile:     l.DefaultProfile(),
		Output:      format.New(io),
		ShutdownCtx: context.Background(),
	}

	// Get the root command
	rootCmd := hcp.NewCmdHcp(ctx)

	// Create the link handler
	linkHandler := func(cmd string) string {
		return linkPrefix + strings.TrimPrefix(cmd, "hcp/")
	}

	// Generate the markdown
	if err := cmd.GenMarkdownTree(rootCmd, outputDir, linkHandler); err != nil {
		return fmt.Errorf("failed to generate markdown: %w", err)
	}

	if outputNavJSON != "" {
		// Create the nav JSON file
		f, err := os.Create(outputNavJSON)
		if err != nil {
			return fmt.Errorf("failed to create nav JSON file: %w", err)
		}

		if err := cmd.GenNavJSON(rootCmd, f); err != nil {
			return fmt.Errorf("failed to generate nav JSON: %w", err)
		}

		if err := f.Close(); err != nil {
			return fmt.Errorf("failed to close nav JSON file: %w", err)
		}
	}

	return nil
}
