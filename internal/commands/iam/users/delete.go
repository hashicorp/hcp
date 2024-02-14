package users

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/client/iam_service"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
)

func NewCmdDelete(ctx *cmd.Context, runF func(*DeleteOpts) error) *cmd.Command {
	opts := &DeleteOpts{
		Ctx:     ctx.ShutdownCtx,
		Profile: ctx.Profile,
		IO:      ctx.IO,
		Client:  iam_service.New(ctx.HCP, nil),
	}

	cmd := &cmd.Command{
		Name:      "delete",
		ShortHelp: "Delete a user from the organization.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ Bold "hcp iam users delete" }} command deletes a user from the organization.
		`),
		Examples: []cmd.Example{
			{
				Preamble: `Delete a user:`,
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				$ hcp iam users delete example-id-123
				`),
			},
		},
		Args: cmd.PositionalArguments{
			Args: []cmd.PositionalArgument{
				{
					Name:          "ID",
					Documentation: "The ID of the user to delete.",
				},
			},
		},
		RunF: func(c *cmd.Command, args []string) error {
			opts.ID = args[0]
			if runF != nil {
				return runF(opts)
			}
			return deleteRun(opts)
		},
		PersistentPreRun: func(c *cmd.Command, args []string) error {
			return cmd.RequireOrganization(ctx)
		},
	}

	return cmd
}

type DeleteOpts struct {
	Ctx     context.Context
	IO      iostreams.IOStreams
	Profile *profile.Profile

	ID     string
	Client iam_service.ClientService
}

func deleteRun(opts *DeleteOpts) error {
	if opts.IO.CanPrompt() {
		ok, err := opts.IO.PromptConfirm("The user will be deleted from the organization.\n\nDo you want to continue")
		if err != nil {
			return fmt.Errorf("failed to retrieve confirmation: %w", err)
		}

		if !ok {
			return nil
		}
	}

	req := iam_service.NewIamServiceDeleteOrganizationMembershipParamsWithContext(opts.Ctx)
	req.OrganizationID = opts.Profile.OrganizationID
	req.UserPrincipalID = opts.ID

	_, err := opts.Client.IamServiceDeleteOrganizationMembership(req, nil)
	if err != nil {
		return fmt.Errorf("failed to delete user principal from organization: %w", err)
	}

	fmt.Fprintf(opts.IO.Err(), "%s User %q deleted from organization\n",
		opts.IO.ColorScheme().SuccessIcon(), opts.ID)
	return nil
}
