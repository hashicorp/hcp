package groups

import (
	"github.com/hashicorp/hcp/internal/commands/iam/groups/members"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
)

func NewCmdGroups(ctx *cmd.Context) *cmd.Command {
	cmd := &cmd.Command{
		Name:      "groups",
		ShortHelp: "Manage HCP Groups.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ Bold "hcp iam groups" }} command group allows you to manage HCP
		groups as well as their memberships.

		Groups help manage users and their access at scale. Each member of a
		group inherits the roles granted to that group. This allows assigning
		many users the same roles by adding them to a group, rather than
		creating role bindings for all individuals that need the same access separately.
		`),
	}

	cmd.AddChild(NewCmdList(ctx, nil))
	cmd.AddChild(NewCmdRead(ctx, nil))
	cmd.AddChild(NewCmdCreate(ctx, nil))
	cmd.AddChild(NewCmdDelete(ctx, nil))
	cmd.AddChild(members.NewCmdMembers(ctx))
	return cmd
}
