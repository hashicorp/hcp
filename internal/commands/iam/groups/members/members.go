package members

import (
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
)

func NewCmdMembers(ctx *cmd.Context) *cmd.Command {
	cmd := &cmd.Command{
		Name:      "members",
		ShortHelp: "Manage group membership.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp iam groups members" }} command group allows you to manage group membership.
		`),
	}

	cmd.AddChild(NewCmdList(ctx, nil))
	cmd.AddChild(NewCmdAdd(ctx, nil))
	cmd.AddChild(NewCmdDelete(ctx, nil))

	return cmd
}
