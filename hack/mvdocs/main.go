// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: MPL-2.0

// This tool is used to move the generated commands documentation from the `web-docs` directory to the `web-unified-docs` repository.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"io/fs"

	"github.com/hashicorp/hcp/internal/pkg/cmd"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Define the flags
	var srcCommandsDir string
	var destCommandsDir string
	var srcNavJSON string
	var destNavJSON string

	flag.StringVar(&srcCommandsDir, "generated-commands-dir", "web-docs/commands", "The generated commands documentation to move to web-unified-docs repository")
	flag.StringVar(&destCommandsDir, "dest-commands-dir", "../web-unified-docs/content/docs/cli/commands/", "The destination directory for the generated commands documentation")
	flag.StringVar(&srcNavJSON, "generated-nav-json", "web-docs/nav.json", "The output path for the generated nav json")
	flag.StringVar(&destNavJSON, "dest-nav-json", "../web-unified-docs/data/docs-nav-data.json", "Path to `web-unified-docs` nav json file")

	// Parse the flags
	flag.Parse()

	// Delete the existing commands directory
	if err := os.RemoveAll(destCommandsDir); err != nil {
		return fmt.Errorf("failed to remove destination commands directory: %w", err)
	}

	// Move the commands directory
	if err := CopyDir(destCommandsDir, srcCommandsDir); err != nil {
		return fmt.Errorf("failed to copy commands directory: %w", err)
	}

	// Open the existing nav JSON for both read and writing.
	dstNavFD, err := os.OpenFile(destNavJSON, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("failed to open destination nav JSON: %w", err)
	}

	// Parse the JSON
	var navData []cmd.DocNavItem
	if err := json.NewDecoder(dstNavFD).Decode(&navData); err != nil {
		return fmt.Errorf("failed to decode destination nav JSON: %w", err)
	}

	// Find the HCP CLI section to inject into
	var hcpCommandsSection *cmd.DocNavItem

OUTER:
	for _, root := range navData {
		if root.Title != "HCP CLI" {
			continue
		}

		// Find the generated commands section
		for _, route := range root.Routes {
			if route.Title != "Commands (CLI)" {
				continue
			}

			hcpCommandsSection = route
			break OUTER
		}
	}
	if hcpCommandsSection == nil {
		return fmt.Errorf("failed to find HCP CLI section in destination nav JSON")
	}

	// Open the generated nav JSON
	srcNavFD, err := os.Open(srcNavJSON)
	if err != nil {
		return fmt.Errorf("failed to open source nav JSON: %w", err)
	}

	// Parse the JSON
	var srcNavData cmd.DocNavItem
	if err := json.NewDecoder(srcNavFD).Decode(&srcNavData); err != nil {
		return fmt.Errorf("failed to decode source nav JSON: %w", err)
	}

	// Inject the HCP CLI section
	*hcpCommandsSection = srcNavData

	// Serialize the JSON
	if _, err := dstNavFD.Seek(0, 0); err != nil {
		return fmt.Errorf("failed to seek destination nav JSON: %w", err)
	}

	e := json.NewEncoder(dstNavFD)
	e.SetIndent("", "  ")
	e.SetEscapeHTML(false)
	if err := e.Encode(navData); err != nil {
		return fmt.Errorf("failed to encode destination nav JSON: %w", err)
	}

	return nil
}

// CopyDir copies the content of src to dst. src should be a full path.
func CopyDir(dst, src string) error {
	return filepath.Walk(src, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// copy to this path
		outpath := filepath.Join(dst, strings.TrimPrefix(path, src))
		if info.IsDir() {
			if err := os.MkdirAll(outpath, info.Mode()); err != nil {
				return err
			}
			return nil // means recursive
		}

		// handle irregular files
		if !info.Mode().IsRegular() {
			if info.Mode().Type()&os.ModeType == os.ModeSymlink {
				link, err := os.Readlink(path)
				if err != nil {
					return err
				}
				return os.Symlink(link, outpath)
			}
			return nil
		}

		// copy contents of regular file efficiently

		// open input
		in, err := os.Open(path)
		if err != nil {
			return err
		}
		defer in.Close()

		// create output
		fh, err := os.Create(outpath)
		if err != nil {
			return err
		}
		defer fh.Close()

		// make it the same
		if err := fh.Chmod(info.Mode()); err != nil {
			return err
		}

		// copy content
		_, err = io.Copy(fh, in)
		return err
	})
}
