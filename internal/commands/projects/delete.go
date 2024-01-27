package projects

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/project_service"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
	"github.com/muesli/reflow/indent"
)

func NewCmdDelete(ctx *cmd.Context, runF func(*DeleteOpts) error) *cmd.Command {
	opts := &DeleteOpts{
		Ctx:     ctx.ShutdownCtx,
		IO:      ctx.IO,
		Profile: ctx.Profile,
		Client:  project_service.New(ctx.HCP, nil),
	}

	cmd := &cmd.Command{
		Name:      "delete",
		ShortHelp: "Delete a project.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		Delete the specified project. The project must be empty before it can be deleted.
		`),
		Examples: []cmd.Example{
			{
				Preamble: "Delete a project.",
				Command:  "$ hcp projects delete --project=cd3d34d5-ceeb-493d-b004-9297365a01af",
			},
		},
		RunF: func(c *cmd.Command, args []string) error {
			l, err := profile.NewLoader()
			if err != nil {
				return err
			}
			opts.Profiles = l
			if runF != nil {
				return runF(opts)
			}

			return deleteRun(opts)
		},
		PersistentPreRun: func(c *cmd.Command, args []string) error {
			return cmd.RequireOrgAndProject(ctx)
		},
	}

	return cmd
}

type DeleteOpts struct {
	Ctx      context.Context
	IO       iostreams.IOStreams
	Profile  *profile.Profile
	Profiles *profile.Loader

	Client project_service.ClientService
}

func deleteRun(opts *DeleteOpts) error {
	req := project_service.NewProjectServiceDeleteParamsWithContext(opts.Ctx)
	req.ID = opts.Profile.ProjectID

	if opts.IO.CanPrompt() {
		ok, err := opts.IO.PromptConfirm("Your project will be deleted.\n\nDo you want to continue")
		if err != nil {
			return fmt.Errorf("failed to retrieve confirmation: %w", err)
		}

		if !ok {
			return nil
		}
	}

	_, err := opts.Client.ProjectServiceDelete(req, nil)
	if err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	fmt.Fprintf(opts.IO.Err(), "%s Project %q deleted\n",
		opts.IO.ColorScheme().SuccessIcon(), opts.Profile.ProjectID)

	profiles, err := opts.Profiles.LoadProfiles()
	if err != nil {
		return fmt.Errorf("failed to load HCP profiles: %w", err)
	}

	var impacted []string
	for _, p := range profiles {
		if p.ProjectID == opts.Profile.ProjectID {
			impacted = append(impacted, fmt.Sprintf("* %s", p.Name))
		}
	}

	if len(impacted) > 0 {
		fmt.Fprintln(opts.IO.Err())
		fmt.Fprintln(opts.IO.Err(), heredoc.New(opts.IO).Mustf(`
%s The following profiles have their project_id property set to the deleted project:

{{ PreserveNewLines }}
%s
{{ PreserveNewLines }}

To update the profile's project_id property interactively, run:

  {{ Bold "$ hcp profile init --profile=PROFILE" }}

Or, to directly set the property, run:

  {{ Bold "$ hcp profile set --profile=PROFILE project_id PROJECT_ID" }}
		`, opts.IO.ColorScheme().WarningLabel(), indent.String(strings.Join(impacted, "\n"), 2)))
		fmt.Fprintln(opts.IO.Err())
	}

	return nil
}
