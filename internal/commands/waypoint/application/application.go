package application

import (
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
)

func NewCmdApplications(ctx *cmd.Context) *cmd.Command {
	cmd := &cmd.Command{
		Name:      "applications",
		ShortHelp: "Manage HCP Waypoint applications.",
		LongHelp: heredoc.New(ctx.IO).Must(`
The {{ Bold "hcp waypoint applications" }} command group lets you manage
HCP Waypoint applications.
		`),
	}

	return cmd
}
