package addon

import (
	"context"
	"testing"

	"github.com/go-openapi/runtime/client"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
)

func TestNewCmdAddOnDefinitionList(t *testing.T) {
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
		// there's no args for the list command right now, but if that changes,
		// we should add a test case here
	}

	for _, c := range cases {
		c := c

		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()

			// Create a context.
			io := iostreams.Test()
			ctx := &cmd.Context{
				IO:          io,
				Profile:     c.Profile(t),
				Output:      format.New(io),
				HCP:         &client.Runtime{},
				ShutdownCtx: context.Background(),
			}

			var tplOpts AddOnDefinitionOpts
			tplOpts.testFunc = func(c *cmd.Command, args []string) error {
				return nil
			}
			cmd := NewCmdAddOnDefinitionList(ctx, &tplOpts)
			cmd.SetIO(io)

			cmd.Run(c.Args)
		})
	}
}
