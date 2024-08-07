// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package applications

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

func TestNewCmdApplicationsUpdate(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Name    string
		Args    []string
		Profile func(t *testing.T) *profile.Profile
		Error   string
		Expect  *ApplicationOpts
	}{
		{
			Name:    "No Org",
			Profile: profile.TestProfile,
			Args:    []string{},
			Error:   "Organization ID must be configured",
		},
		{
			Name:    "No Name",
			Profile: profile.TestProfile,
			Args:    []string{},
			Error:   "The name of the application is required",
		},
		{
			Name:    "Happy",
			Profile: profile.TestProfile,
			Args: []string{
				"-n",
				"app-name",
				"--action-config-name",
				"config-1",
				"--action-config-name",
				"config-2",
				"--readme-markdown-file",
				"readme.md",
			},
			Expect: &ApplicationOpts{
				Name: "app-name",
				ActionConfigNames: []string{
					"config-1",
					"config-2",
				},
				ReadmeMarkdownFile: "readme.md",
			},
		},
	}

	for _, c := range cases {
		c := c

		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()

			r := require.New(t)

			// Create a context.
			io := iostreams.Test()
			ctx := &cmd.Context{
				IO:          io,
				Profile:     c.Profile(t),
				HCP:         &client.Runtime{},
				ShutdownCtx: context.Background(),
				Output:      format.New(io),
			}

			var appOpts ApplicationOpts
			appOpts.testFunc = func(c *cmd.Command, args []string) error {
				return nil
			}
			cmd := NewCmdApplicationsUpdate(ctx, &appOpts)
			cmd.SetIO(io)

			cmd.Run(c.Args)

			if c.Expect != nil {
				r.Equal(c.Expect.Name, appOpts.Name)
				r.Equal(c.Expect.ActionConfigNames, appOpts.ActionConfigNames)
				r.Equal(c.Expect.ReadmeMarkdownFile, appOpts.ReadmeMarkdownFile)
			}
		})
	}
}
