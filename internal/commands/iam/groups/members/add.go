package members

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/client/groups_service"
	"github.com/hashicorp/hcp/internal/commands/iam/groups/helper"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/flagvalue"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
)

func NewCmdAdd(ctx *cmd.Context, runF func(*AddOpts) error) *cmd.Command {
	opts := &AddOpts{
		Ctx:     ctx.ShutdownCtx,
		IO:      ctx.IO,
		Profile: ctx.Profile,
		Client:  groups_service.New(ctx.HCP, nil),
	}

	cmd := &cmd.Command{
		Name:      "add",
		ShortHelp: "Add members to a group.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp iam groups members add" }} command adds members to a group.

		All added members will inherit any roles that have been granted to the group.
		`),
		Examples: []cmd.Example{
			{
				Preamble: `Add members to the "platform-team":`,
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				$ hcp iam groups members add --group=team-platform \
				  --member=7f8a81b2-1320-4e49-a2e5-44f628ec74c3 \
				  --member=f74f44b9-414a-409e-a257-72805d2c067b
				`),
			},
		},
		Flags: cmd.Flags{
			Local: []*cmd.Flag{
				{
					Name:         "group",
					Shorthand:    "g",
					DisplayValue: "NAME",
					Description:  heredoc.New(ctx.IO).Mustf(helper.GroupNameArgDoc, "add a member to"),
					Value:        flagvalue.Simple("", &opts.GroupName),
					Autocomplete: helper.PredictGroupResourceNameSuffix(opts.Ctx, opts.Profile.OrganizationID, opts.Client),
				},
				{
					Name:         "member",
					Shorthand:    "m",
					DisplayValue: "ID",
					Description:  "The ID of the user principal to add to the group.",
					Value:        flagvalue.SimpleSlice(nil, &opts.Members),
					Repeatable:   true,
				},
			},
		},
		RunF: func(c *cmd.Command, args []string) error {
			if len(opts.Members) == 0 {
				return fmt.Errorf("at least one member must be specified")
			}

			if runF != nil {
				return runF(opts)
			}
			return addRun(opts)
		},
		PersistentPreRun: func(c *cmd.Command, args []string) error {
			return cmd.RequireOrganization(ctx)
		},
	}

	return cmd
}

type AddOpts struct {
	Ctx     context.Context
	IO      iostreams.IOStreams
	Profile *profile.Profile

	GroupName string
	Members   []string
	Client    groups_service.ClientService
}

func addRun(opts *AddOpts) error {
	req := groups_service.NewGroupsServiceUpdateGroupMembersParamsWithContext(opts.Ctx)
	req.ResourceName = helper.ResourceName(opts.GroupName, opts.Profile.OrganizationID)
	req.Body = groups_service.GroupsServiceUpdateGroupMembersBody{
		MemberPrincipalIdsToAdd: opts.Members,
	}

	_, err := opts.Client.GroupsServiceUpdateGroupMembers(req, nil)
	if err != nil {
		return fmt.Errorf("failed to update group membership: %w", err)
	}

	fmt.Fprintf(opts.IO.Err(), "%s Membership to group %q updated\n",
		opts.IO.ColorScheme().SuccessIcon(), req.ResourceName)
	return nil
}
