package roles

import (
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
)

func NewCmdRoles(ctx *cmd.Context) *cmd.Command {
	cmd := &cmd.Command{
		Name:      "roles",
		ShortHelp: "Interact with an organization's roles.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ Bold "hcp iam roles" }} command group lets you interact with an HCP organization's roles.
		`),
	}

	cmd.AddChild(NewCmdList(ctx, nil))
	return cmd
}
