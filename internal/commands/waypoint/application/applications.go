package application

import (
	"github.com/hashicorp/hcp/internal/commands/waypoint/opts"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
)

type ApplicationOpts struct {
	opts.WaypointOpts

	ID                string
	Name              string
	TemplateName      string
	ActionConfigNames []string
}

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
