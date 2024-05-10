package apps

import (
	"context"
	"testing"

	"github.com/go-openapi/runtime/client"
	"github.com/stretchr/testify/require"
	
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
)

func TestNewCmdCreate(t *testing.T) {
	cases := []struct {
		Name    string
		Args    []string
		Profile func(t *testing.T) *profile.Profile
		Error   string
		Expect  *CreateOpts
	}{
		{
			Name: "Good",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123").SetProjectID("abc")
			},
			Args: []string{"company-card"},
			Expect: &CreateOpts{
				AppName: "company-card",
			},
		},
		{
			Name: "Good with description",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123").SetProjectID("abc")
			},
			Args: []string{"company-card", "--description", "Stores corporate card info."},
			Expect: &CreateOpts{
				AppName:     "company-card",
				Description: "Stores corporate card info.",
			},
		},
		{
			Name:    "No Org or Project ID",
			Profile: profile.TestProfile,
			Args:    []string{},
			Error:   "Organization ID and Project ID must be configured before running the command.",
		},
		{
			Name: "No args",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123").SetProjectID("abc")
			},
			Args:  []string{},
			Error: "accepts 1 arg(s), received 0",
		},
		{
			Name: "Too many args",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123").SetProjectID("abc")
			},
			Args:  []string{"company-card", "additional-arg"},
			Error: "ERROR: accepts 1 arg(s), received 2",
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
				ShutdownCtx: context.Background(),
				HCP:         &client.Runtime{},
				Output:      format.New(io),
			}

			var createOpts *CreateOpts
			createCmd := NewCmdCreate(ctx, func(o *CreateOpts) error {
				createOpts = o
				return nil
			})
			createCmd.SetIO(io)

			code := createCmd.Run(c.Args)
			if c.Error != "" {
				r.NotZero(code)
				r.Contains(io.Error.String(), c.Error)
				return
			}

			r.Zero(code, io.Error.String())
			r.NotNil(createOpts)
			r.Equal(c.Expect.AppName, createOpts.AppName)
			r.Equal(c.Expect.Description, createOpts.Description)
		})
	}
}
