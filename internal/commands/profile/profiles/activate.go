package profiles

import (
	"fmt"
	"slices"

	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
)

func NewCmdActivate(ctx *cmd.Context) *cmd.Command {
	opts := &ActivateOpts{
		IO: ctx.IO,
	}
	cmd := &cmd.Command{
		Name:      "activate",
		ShortHelp: "Activates an existing profile.",
		LongHelp:  "Activates an existing profile.",
		Examples: []cmd.Example{
			{
				Preamble: "To active profile my-profile, run:",
				Command:  "$ hcp profile profiles activate my-profile",
			},
		},
		Args: cmd.PositionalArguments{
			Autocomplete: predictProfiles(false, false),
			Args: []cmd.PositionalArgument{
				{
					Name:          "NAME",
					Documentation: "The name of the profile to activate.",
				},
			},
		},
		RunF: func(c *cmd.Command, args []string) error {
			opts.Name = args[0]
			l, err := profile.NewLoader()
			if err != nil {
				return err
			}
			opts.Profiles = l
			return activateRun(opts)
		},
	}

	return cmd
}

type ActivateOpts struct {
	IO       iostreams.IOStreams
	Profiles *profile.Loader
	Name     string
}

func activateRun(opts *ActivateOpts) error {
	// Get the active profile
	active, err := opts.Profiles.GetActiveProfile()
	if err != nil {
		return fmt.Errorf("failed to get active profile: %w", err)
	}

	// Ensure the given profile isn't already the active profile
	if active.Name == opts.Name {
		return fmt.Errorf("profile %q is already the active profile", opts.Name)
	}

	// Ensure the given profile exists.
	profileNames, err := opts.Profiles.ListProfiles()
	if err != nil {
		return fmt.Errorf("failed to list profiles: %w", err)
	}

	if !slices.Contains(profileNames, opts.Name) {
		return fmt.Errorf("profile %q does not exist", opts.Name)
	}

	// Save the new active profile
	active.Name = opts.Name
	if err := active.Write(); err != nil {
		return fmt.Errorf("failed to save active profile: %w", err)
	}

	fmt.Fprintf(opts.IO.Err(), "%s Profile %q activated.\n",
		opts.IO.ColorScheme().SuccessIcon(), opts.Name)
	return nil
}
