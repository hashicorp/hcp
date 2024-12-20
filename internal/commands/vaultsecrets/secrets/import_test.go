package secrets

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

func TestNewCmdImport(t *testing.T) {
	t.Parallel()

	testProfile := func(t *testing.T) *profile.Profile {
		tp := profile.TestProfile(t).SetOrgID("123").SetProjectID("456")
		return tp
	}

	cases := []struct {
		Name    string
		Args    []string
		Profile func(t *testing.T) *profile.Profile
		Error   string
		Expect  *ImportOpts
	}{
		{
			Name:    "Good",
			Profile: testProfile,
			Args:    []string{"--config-file", "path/to/file"},
			Expect: &ImportOpts{
				ConfigFilePath: "path/to/file",
			},
		},
		{
			Name:    "Failed: No config file specified",
			Profile: testProfile,
			Expect: &ImportOpts{
				ConfigFilePath: "path/to/file",
			},
			Error: "ERROR: missing required flag: --config-file=CONFIG_FILE",
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

			var gotOpts *ImportOpts
			importCmd := NewCmdImport(ctx, func(o *ImportOpts) error {
				gotOpts = o
				return nil
			})
			importCmd.SetIO(io)

			code := importCmd.Run(c.Args)
			if c.Error != "" {
				r.NotZero(code)
				r.Contains(io.Error.String(), c.Error)
				return
			}

			r.Zero(code, io.Error.String())
			r.NotNil(gotOpts)
			r.Equal(c.Expect.ConfigFilePath, gotOpts.ConfigFilePath)
		})
	}
}
