package application

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

func TestNewCmdCreateApplication(t *testing.T) {
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
			Args:    []string{"-t", "template-name"},
			Error:   "The name of the application is required",
		},
		{
			Name:    "No Template Name",
			Profile: profile.TestProfile,
			Args:    []string{"-n", "app-name"},
			Error:   "The name of the template to use for the application is required",
		},
		{
			Name:    "Happy",
			Profile: profile.TestProfile,
			Args:    []string{"-n", "app-name", "-t", "template-name"},
			Expect: &ApplicationOpts{
				Name:         "app-name",
				TemplateName: "template-name",
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
				Output:      format.New(io),
				HCP:         &client.Runtime{},
				ShutdownCtx: context.Background(),
			}

			var appOpts ApplicationOpts
			appOpts.testFunc = func(c *cmd.Command, args []string) error {
				return nil
			}
			cmd := NewCmdCreateApplication(ctx, &appOpts)
			cmd.SetIO(io)

			cmd.Run(c.Args)

			if c.Expect != nil {
				r.Equal(c.Expect.Name, appOpts.Name)
				r.Equal(c.Expect.TemplateName, appOpts.TemplateName)
			}
		})
	}
}
