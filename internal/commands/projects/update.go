package projects

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/project_service"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/flagvalue"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
)

func NewCmdUpdate(ctx *cmd.Context, runF func(*UpdateOpts) error) *cmd.Command {
	opts := &UpdateOpts{
		Ctx:     ctx.ShutdownCtx,
		Profile: ctx.Profile,
		IO:      ctx.IO,
		Client:  project_service.New(ctx.HCP, nil),
	}

	cmd := &cmd.Command{
		Name:      "update",
		ShortHelp: "Update an existing project.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp products update" }} command shows metadata for the project.
		`),
		Examples: []cmd.Example{
			{
				Preamble: "Update a project's name and description:",
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				$ hcp projects update --project=cd3d34d5-ceeb-493d-b004-9297365a01af \
				  --name=new-name --description="updated description"
				`),
			},
		},
		Flags: cmd.Flags{
			Local: []*cmd.Flag{
				{
					Name:         "description",
					DisplayValue: "NEW_DESCRIPTION",
					Description:  "New description for the project.",
					Value:        flagvalue.Simple((*string)(nil), &opts.Description),
				},
				{
					Name:         "name",
					DisplayValue: "NEW_NAME",
					Description:  "New name for the project.",
					Value:        flagvalue.Simple((*string)(nil), &opts.Name),
				},
			},
		},
		RunF: func(c *cmd.Command, args []string) error {
			if runF != nil {
				return runF(opts)
			}

			return updateRun(opts)
		},
		PersistentPreRun: func(c *cmd.Command, args []string) error {
			return cmd.RequireOrgAndProject(ctx)
		},
	}

	return cmd
}

type UpdateOpts struct {
	Ctx         context.Context
	Profile     *profile.Profile
	IO          iostreams.IOStreams
	Description *string
	Name        *string

	Client project_service.ClientService
}

func updateRun(opts *UpdateOpts) error {
	if opts.Name == nil && opts.Description == nil {
		return fmt.Errorf("either name or description must be specified")
	}

	if opts.Name != nil {
		req := project_service.NewProjectServiceSetNameParamsWithContext(opts.Ctx)
		req.ID = opts.Profile.ProjectID
		req.Body.Name = *opts.Name

		_, err := opts.Client.ProjectServiceSetName(req, nil)
		if err != nil {
			return fmt.Errorf("failed to update project name: %w", err)
		}
	}

	if opts.Description != nil {
		req := project_service.NewProjectServiceSetDescriptionParamsWithContext(opts.Ctx)
		req.ID = opts.Profile.ProjectID
		req.Body.Description = *opts.Description

		_, err := opts.Client.ProjectServiceSetDescription(req, nil)
		if err != nil {
			return fmt.Errorf("failed to update project description: %w", err)
		}
	}

	fmt.Fprintf(opts.IO.Err(), "%s Project %q updated\n",
		opts.IO.ColorScheme().SuccessIcon(), opts.Profile.ProjectID)
	return nil
}
