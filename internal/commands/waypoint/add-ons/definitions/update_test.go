// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package definitions

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

func TestCmdAddOnDefinitionUpdate(t *testing.T) {
	t.Parallel()
	cases := []struct {
		Name    string
		Args    []string
		Profile func(t *testing.T) *profile.Profile
		Error   string
		Expect  *AddOnDefinitionOpts
	}{
		{
			Name:    "No Org",
			Profile: profile.TestProfile,
			Args:    []string{},
			Error:   "Organization ID must be configured",
		},
		{
			Name: "no args",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123")
			},
			Args:  []string{},
			Error: "accepts 1 arg(s), received 0",
		},
		{
			Name: "happy",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123")
			},
			Args: []string{
				"-n=cli-test",
				"-s", "An add-on definition created using the CLI.",
				"--tfc-project-id", "prj-abcdefghij",
				"--tfc-project-name", "test",
				"-l", "cli",
				"-d", "An add-on definition created with the CLI.",
				"--readme-markdown-template-file", "readme_test.txt",
				"--tf-execution-mode", "agent",
				"--tf-agent-pool-id", "pool-abc123",
				"--variable-options-file", "vars.hcl",
			},
			Expect: &AddOnDefinitionOpts{
				Name:                       "cli-test",
				Summary:                    "An add-on definition created using the CLI.",
				Description:                "An add-on definition created with the CLI.",
				TerraformCloudProjectID:    "prj-abcdefghij",
				TerraformCloudProjectName:  "test",
				Labels:                     []string{"cli"},
				ReadmeMarkdownTemplateFile: "readme_test.txt",
				TerraformExecutionMode:     "agent",
				TerraformAgentPoolID:       "pool-abc123",
				VariableOptionsFile:        "vars.hcl",
			},
		},
	}

	for _, c := range cases {
		c := c

		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()

			r := require.New(t)

			io := iostreams.Test()

			ctx := &cmd.Context{
				IO:          io,
				Profile:     c.Profile(t),
				Output:      format.New(io),
				HCP:         &client.Runtime{},
				ShutdownCtx: context.Background(),
			}

			var aodOpts AddOnDefinitionOpts
			aodOpts.testFunc = func(c *cmd.Command, args []string) error {
				return nil
			}
			cmd := NewCmdAddOnDefinitionUpdate(ctx, &aodOpts)
			cmd.SetIO(io)

			cmd.Run(c.Args)

			if c.Expect != nil {
				r.Equal(c.Expect.Name, aodOpts.Name)
				r.Equal(c.Expect.Summary, aodOpts.Summary)
				r.Equal(c.Expect.Description, aodOpts.Description)
				r.Equal(c.Expect.TerraformCloudProjectID, aodOpts.TerraformCloudProjectID)
				r.Equal(c.Expect.TerraformCloudProjectName, aodOpts.TerraformCloudProjectName)
				r.Equal(c.Expect.Labels, aodOpts.Labels)
				r.Equal(c.Expect.ReadmeMarkdownTemplateFile, aodOpts.ReadmeMarkdownTemplateFile)
				r.Equal(c.Expect.TerraformExecutionMode, aodOpts.TerraformExecutionMode)
				r.Equal(c.Expect.TerraformAgentPoolID, aodOpts.TerraformAgentPoolID)
				r.Equal(c.Expect.VariableOptionsFile, aodOpts.VariableOptionsFile)
			}
		})
	}
}
