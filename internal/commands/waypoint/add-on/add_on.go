package addon

import (
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
)

func NewCmdAddOn(ctx *cmd.Context) *cmd.Command {
	cmd := &cmd.Command{
		Name:      "add-ons",
		ShortHelp: "Manage HCP Waypoint add-ons and add-on definitions.",
		LongHelp: heredoc.New(ctx.IO).Must(`
The {{ template "mdCodeOrBold" "hcp waypoint add-ons" }} command group lets you
manage HCP Waypoint add-ons and add-on definitions.
`),
	}

	cmd.AddChild(NewCmdAddOnDefinition(ctx))

	return cmd
}
