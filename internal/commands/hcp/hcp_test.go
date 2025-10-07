// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: MPL-2.0

package hcp

import (
	"context"
	"testing"

	"github.com/go-openapi/runtime/client"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
	"github.com/stretchr/testify/require"
)

func TestHCP(t *testing.T) {
	t.Parallel()
	r := require.New(t)

	// Create a context.
	io := iostreams.Test()
	ctx := &cmd.Context{
		IO:          io,
		Profile:     profile.TestProfile(t).SetOrgID("123").SetProjectID("456"),
		Output:      format.New(io),
		HCP:         &client.Runtime{},
		ShutdownCtx: context.Background(),
	}

	c := NewCmdHcp(ctx)
	r.NoError(c.Validate())
}
