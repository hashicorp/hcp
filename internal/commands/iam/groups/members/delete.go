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

func NewCmdDelete(ctx *cmd.Context, runF func(*DeleteOpts) error) *cmd.Command {
	opts := &DeleteOpts{
		Ctx:     ctx.ShutdownCtx,
		IO:      ctx.IO,
		Profile: ctx.Profile,
		Client:  groups_service.New(ctx.HCP, nil),
	}

	cmd := &cmd.Command{
		Name:      "delete",
		ShortHelp: "Delete a membership from a group.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ Bold "hcp iam groups members delete" }} command deletes a membership from a group.

		All members that are deleted will no longer inherit any roles that have been granted to the group.
		`),
		Examples: []cmd.Example{
			{
				Preamble: `Delete members from the "platform-team"`,
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				$ hcp iam groups members delete --group=team-platform \
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
					Description:  heredoc.New(ctx.IO).Mustf(helper.GroupNameArgDoc, "delete a membership from"),
					Value:        flagvalue.Simple("", &opts.GroupName),
					Autocomplete: helper.PredictGroupResourceNameSuffix(opts.Ctx, opts.Profile.OrganizationID, opts.Client),
				},
				{
					Name:         "member",
					Shorthand:    "m",
					DisplayValue: "ID",
					Description:  "The ID of the user principal to remove membership from the group.",
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

	GroupName string
	Members   []string
	Client    groups_service.ClientService
}

func deleteRun(opts *DeleteOpts) error {
	req := groups_service.NewGroupsServiceUpdateGroupMembersParamsWithContext(opts.Ctx)
	req.ResourceName = helper.ResourceName(opts.GroupName, opts.Profile.OrganizationID)
	req.Body = groups_service.GroupsServiceUpdateGroupMembersBody{
		MemberPrincipalIdsToRemove: opts.Members,
	}

	_, err := opts.Client.GroupsServiceUpdateGroupMembers(req, nil)
	if err != nil {
		return fmt.Errorf("failed to update group membership: %w", err)
	}

	fmt.Fprintf(opts.IO.Err(), "%s Membership to group %q updated\n",
		opts.IO.ColorScheme().SuccessIcon(), req.ResourceName)
	return nil
}
