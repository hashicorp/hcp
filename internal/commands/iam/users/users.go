package users

import (
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
)

func NewCmdUsers(ctx *cmd.Context) *cmd.Command {
	cmd := &cmd.Command{
		Name:      "users",
		ShortHelp: "Manage an organization's users",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ Bold "hcp iam users" }} command group allows you to manage the users of an HCP organization.
		`),
	}

	cmd.AddChild(NewCmdList(ctx, nil))
	cmd.AddChild(NewCmdRead(ctx, nil))
	cmd.AddChild(NewCmdDelete(ctx, nil))
	return cmd
}
