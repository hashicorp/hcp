package profiles

import (
	"fmt"

	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
)

func NewCmdDelete(ctx *cmd.Context) *cmd.Command {
	opts := &DeleteOpts{
		IO: ctx.IO,
	}
	cmd := &cmd.Command{
		Name:      "delete",
		ShortHelp: "delete an existing HCP profile.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		Deletes an existing HCP profiles. If the profile is the active profile, it may not be deleted.

		To delete the current active profile, first run {{ Bold "hcp profile profiles activate" }} another one.
		`),
		Examples: []cmd.Example{
			{
				Title:   "Delete a profile",
				Command: "$ hcp profile profiles delete my-profile",
			},
			{
				Title:   "Delete multiple profiles",
				Command: "$ hcp profile profiles delete my-profile-1 my-profile-2 my-profile-3",
			},
			{
				Title:    "Delete the active profile",
				Preamble: "To delete the active profile, my-profile, run:",
				Command: heredoc.New(ctx.IO).Must(`
				$ hcp profile profiles active my-other-profile
				$ hcp profile profiles delete my-profile
				`),
			},
		},
		Args: cmd.PositionalArguments{
			Autocomplete: predictProfiles(true, false),
			Args: []cmd.PositionalArgument{
				{
					Name:          "profile_names",
					Documentation: "The name of the profile to delete. May not be the active profile.",
					Repeatable:    true,
				},
			},
		},
		RunF: func(c *cmd.Command, args []string) error {
			l, err := profile.NewLoader()
			if err != nil {
				return err
			}
			opts.Profiles = l
			opts.Names = args
			return deleteRun(opts)
		},
	}

	return cmd
}

type DeleteOpts struct {
	IO       iostreams.IOStreams
	Profiles *profile.Loader

	Names []string
}

func deleteRun(opts *DeleteOpts) error {
	// Get the active profile
	active, err := opts.Profiles.GetActiveProfile()
	if err != nil {
		return fmt.Errorf("failed to get active profile: %w", err)
	}

	profileNames, err := opts.Profiles.ListProfiles()
	if err != nil {
		return fmt.Errorf("failed to list profiles: %w", err)
	}

	// Validate that the given profiles to delete aren't active and that they
	// all exist.
	existing := make(map[string]struct{}, len(profileNames))
	for _, p := range profileNames {
		existing[p] = struct{}{}
	}

	cs := opts.IO.ColorScheme()
	for _, toDelete := range opts.Names {
		if toDelete == active.Name {
			return fmt.Errorf("profile %q is the active profile and may not be deleted. Use %s to change the active configuration",
				toDelete, cs.String("hcp profile profiles activate").Bold())
		}
		if _, ok := existing[toDelete]; !ok {
			return fmt.Errorf("profile %q does not exist", toDelete)
		}
	}

	if opts.IO.CanPrompt() {
		fmt.Fprintln(opts.IO.Err(), "The following profiles will be deleted:")
		for _, toDelete := range opts.Names {
			fmt.Fprintf(opts.IO.Err(), "  - %s\n", toDelete)
		}

		fmt.Fprintln(opts.IO.Err())
		ok, err := opts.IO.PromptConfirm("Do you want to continue")
		if err != nil {
			return err
		}

		if !ok {
			return nil
		}
	}

	for _, toDelete := range opts.Names {
		if err := opts.Profiles.DeleteProfile(toDelete); err != nil {
			return fmt.Errorf("failed to delete profile %q: %w", toDelete, err)
		}

		fmt.Fprintf(opts.IO.Err(), "%s Profile %q deleted.\n", cs.SuccessIcon(), toDelete)
	}

	return nil
}
