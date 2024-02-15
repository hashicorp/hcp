package profiles

import (
	"fmt"

	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/flagvalue"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
)

func NewCmdCreate(ctx *cmd.Context) *cmd.Command {
	opts := &CreateOpts{
		IO: ctx.IO,
	}
	cmd := &cmd.Command{
		Name:      "create",
		ShortHelp: "Create a new HCP profile.",
		LongHelp: heredoc.New(ctx.IO).Mustf(`
		Creates a new named profile.

		Profile names start with a letter and may contain lower case letters a-z,
		upper case letters A-Z, digits 0-9, and hyphens '-'. The maximum length for
		a profile name is 64 characters.
		`),
		Examples: []cmd.Example{
			{
				Preamble: "To create a new profile, run:",
				Command:  "$ hcp profile profiles create my-profile",
			},
		},
		Args: cmd.PositionalArguments{
			Args: []cmd.PositionalArgument{
				{
					Name:          "NAME",
					Documentation: "The name of the profile to create.",
				},
			},
		},
		Flags: cmd.Flags{
			Local: []*cmd.Flag{
				{
					Name:          "no-activate",
					Description:   "Disables automatic activation of the newly created profile.",
					Value:         flagvalue.Simple(false, &opts.NoActivate),
					IsBooleanFlag: true,
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
			return createRun(opts)
		},
	}

	return cmd
}

type CreateOpts struct {
	IO iostreams.IOStreams

	Profiles   *profile.Loader
	Name       string
	NoActivate bool
}

func createRun(opts *CreateOpts) error {
	// Get the existing profiles
	profiles, err := opts.Profiles.ListProfiles()
	if err != nil {
		return fmt.Errorf("failed to list existing profiles: %w", err)
	}

	// Validate a profile with the given name doesn't already exist.
	for _, p := range profiles {
		if p == opts.Name {
			return fmt.Errorf("profile with name %q already exists", opts.Name)
		}
	}

	// Create the new profile
	p, err := opts.Profiles.NewProfile(opts.Name)
	if err != nil {
		return err
	}

	// Save the profile
	if err := p.Write(); err != nil {
		return fmt.Errorf("failed to save new profile: %w", err)
	}

	cs := opts.IO.ColorScheme()
	fmt.Fprintf(opts.IO.Err(), "%s Profile %q created.\n", cs.SuccessIcon(), p.Name)

	if !opts.NoActivate {
		// Update the active profile.
		active, err := opts.Profiles.GetActiveProfile()
		if err != nil {
			return fmt.Errorf("failed to retrieve active profile: %w", err)
		}

		active.Name = p.Name
		if err := active.Write(); err != nil {
			return fmt.Errorf("failed to update active profile: %w", err)
		}

		fmt.Fprintf(opts.IO.Err(), "%s Profile %q activated.\n", cs.SuccessIcon(), p.Name)
	}

	fmt.Fprintln(opts.IO.Err())
	fmt.Fprintln(opts.IO.Err(), heredoc.New(opts.IO).Must(`
		To initialize the newly created profile, run:

		  {{ Bold "$ hcp profile init" }}
		`))
	fmt.Fprintln(opts.IO.Err())

	return nil
}
