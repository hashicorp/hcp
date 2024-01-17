package groups

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/client/groups_service"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/flagvalue"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/hashicorp/hcp/internal/pkg/profile"
)

func NewCmdCreate(ctx *cmd.Context, runF func(*CreateOpts) error) *cmd.Command {
	opts := &CreateOpts{
		Ctx:     ctx.ShutdownCtx,
		Profile: ctx.Profile,
		Output:  ctx.Output,
		Client:  groups_service.New(ctx.HCP, nil),
	}

	cmd := &cmd.Command{
		Name:      "create",
		ShortHelp: "Create a new group.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ Bold "hcp iam groups create" }} command creates a new group.

		Once a group is created, membership can be managed using the {{ Bold "hcp iam groups members" }} command group.
		`),
		Examples: []cmd.Example{
			{
				Preamble: `Create a new group for the platform engineering team`,
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				$ hcp iam groups create team-platform
				  --description "Team Platform engineering group"
				`),
			},
			{
				Preamble: `Create a new group and specify the initial members`,
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				$ hcp iam groups create team-platform
				  --description "Team Platform engineering group"
				  --member=7f8a81b2-1320-4e49-a2e5-44f628ec74c3
				  --member=f74f44b9-414a-409e-a257-72805d2c067b
				`),
			},
		},
		Args: cmd.PositionalArguments{
			Args: []cmd.PositionalArgument{
				{
					Name:          "GROUP_NAME",
					Documentation: "The name of the group to create.",
				},
			},
		},
		Flags: cmd.Flags{
			Local: []*cmd.Flag{
				{
					Name:         "description",
					DisplayValue: "DESCRIPTION",
					Description:  "An optional description for the group.",
					Value:        flagvalue.Simple("", &opts.Description),
				},
				{
					Name:         "member",
					DisplayValue: "ID",
					Description:  "The ID of the principal to add to the group.",
					Value:        flagvalue.SimpleSlice(nil, &opts.Members),
					Repeatable:   true,
				},
			},
		},
		RunF: func(c *cmd.Command, args []string) error {
			opts.Name = args[0]

			if runF != nil {
				return runF(opts)
			}
			return createRun(opts)
		},
		PersistentPreRun: func(c *cmd.Command, args []string) error {
			return cmd.RequireOrganization(ctx)
		},
	}

	return cmd
}

type CreateOpts struct {
	Ctx     context.Context
	Profile *profile.Profile
	Output  *format.Outputter

	Name        string
	Description string
	Members     []string
	Client      groups_service.ClientService
}

func createRun(opts *CreateOpts) error {
	req := groups_service.NewGroupsServiceCreateGroupParamsWithContext(opts.Ctx)
	req.ParentResourceName = opts.Profile.GetOrgResourcePart().String()
	req.Body = groups_service.GroupsServiceCreateGroupBody{
		Name:               opts.Name,
		Description:        opts.Description,
		MemberPrincipalIds: opts.Members,
	}

	resp, err := opts.Client.GroupsServiceCreateGroup(req, nil)
	if err != nil {
		return fmt.Errorf("failed to create group: %w", err)
	}

	return opts.Output.Display(newDisplayer(format.Pretty, true, resp.Payload.Group))
}
