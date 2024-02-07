package agent

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

func TestCmdGroupCreate(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Name    string
		Args    []string
		Profile func(t *testing.T) *profile.Profile
		Error   string
		Expect  *GroupOpts
	}{
		{
			Name:    "No Org",
			Profile: profile.TestProfile,
			Args:    []string{},
			Error:   "Organization ID must be configured",
		},
		{
			Name: "No name",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123")
			},
			Args:  []string{},
			Error: "accepts 1 arg(s), received 0",
		},
		{
			Name: "Too many args",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123")
			},
			Args:  []string{"foo", "bar"},
			Error: "accepts 0 arg(s), received 2",
		},
		{
			Name: "Good",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123")
			},
			Args: []string{"-n", "foo", "--description=test"},
			Expect: &GroupOpts{
				Name:        "foo",
				Description: "test",
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

			var gotOpts GroupOpts
			gotOpts.testFunc = func(c *cmd.Command, args []string) error {
				return nil
			}
			cmd := NewCmdGroupCreate(ctx, &gotOpts)
			cmd.SetIO(io)

			cmd.Run(c.Args)

			if c.Expect != nil {
				r.NotNil(c.Expect)

				r.Equal(c.Expect.Name, gotOpts.Name)
				r.Equal(c.Expect.Description, gotOpts.Description)
			}
		})
	}

}
