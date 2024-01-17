package profiles

import (
	"testing"

	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
	"github.com/stretchr/testify/require"
)

func TestDelete(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Name    string
		Active  string
		Create  []string
		Delete  []string
		Prompt  bool
		Confirm bool
		Error   string
	}{
		{
			Name:   "bad profile name",
			Active: "foo",
			Create: []string{"foo"},
			Delete: []string{"bar"},
			Error:  "profile \"bar\" does not exist",
		},
		{
			Name:   "active profile name",
			Active: "foo",
			Create: []string{"foo"},
			Delete: []string{"foo"},
			Error:  "profile \"foo\" is the active profile and may not be deleted.",
		},
		{
			Name:   "single",
			Active: "foo",
			Create: []string{"foo", "bar"},
			Delete: []string{"bar"},
		},
		{
			Name:   "multiple",
			Active: "foo",
			Create: []string{"foo", "bar", "baz", "bam"},
			Delete: []string{"bar", "baz", "bam"},
		},
		{
			Name:    "prompt and decline",
			Active:  "foo",
			Create:  []string{"foo", "bar", "baz", "bam"},
			Delete:  []string{"bar", "baz", "bam"},
			Prompt:  true,
			Confirm: false,
		},
		{
			Name:    "prompt and accept",
			Active:  "foo",
			Create:  []string{"foo", "bar", "baz", "bam"},
			Delete:  []string{"bar", "baz", "bam"},
			Prompt:  true,
			Confirm: true,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			r := require.New(t)
			l := profile.TestLoader(t)
			io := iostreams.Test()

			// Create the profiles
			for _, name := range c.Create {
				p, err := l.NewProfile(name)
				r.NoError(err)
				r.NoError(p.Write())
			}

			// Mark the correct profile as active
			active, err := l.GetActiveProfile()
			r.NoError(err)
			active.Name = c.Active
			r.NoError(active.Write())

			opts := &DeleteOpts{
				IO:       io,
				Profiles: l,
				Names:    c.Delete,
			}

			if c.Prompt {
				io.InputTTY = true
				io.ErrorTTY = true

				resp := 'y'
				if !c.Confirm {
					resp = 'n'
				}

				// Write to stdin
				_, err := io.Input.WriteRune(resp)
				r.NoError(err)
			}

			err = deleteRun(opts)
			if c.Error != "" {
				r.ErrorContains(err, c.Error)
				return
			}

			r.NoError(err)

			// Load the profiles that now exist
			profiles, err := l.ListProfiles()
			r.NoError(err)

			if !c.Prompt || c.Confirm {
				// Ensure that any deleted profile does not exist in the set
				for _, d := range c.Delete {
					r.NotContains(profiles, d)
				}
			} else if c.Prompt && !c.Confirm {
				// If we prompted and didn't accept, ensure we didn't delete
				// anything.
				for _, p := range c.Create {
					r.Contains(profiles, p)
				}
			}
		})
	}
}
