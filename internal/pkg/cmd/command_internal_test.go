// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: MPL-2.0

package cmd

import (
	"testing"

	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/stretchr/testify/require"
)

func TestAuthErrorHelp(t *testing.T) {
	t.Parallel()
	r := require.New(t)
	io := iostreams.Test()

	commandPath := "hcp example"
	args := []string{"simple", "'single-quote'", `escaped \"inner\"`}

	// Get the help text
	helpText := authErrorHelp(io, commandPath, args)
	r.Contains(helpText, `$ hcp example simple 'single-quote' "escaped \\\"inner\\\""`)
}
