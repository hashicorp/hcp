package application

import (
	"github.com/hashicorp/hcp/internal/commands/waypoint/opts"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
)

type ApplicationOpts struct {
	opts.WaypointOpts

	Name              string
	TemplateName      string
	ActionConfigNames []string

	testFunc func(c *cmd.Command, args []string) error
}

func NewCmdApplications(ctx *cmd.Context) *cmd.Command {
	opts := &ApplicationOpts{
		WaypointOpts: opts.New(ctx),
	}

	cmd := &cmd.Command{
		Name:      "applications",
		ShortHelp: "Manage HCP Waypoint applications.",
		LongHelp: heredoc.New(ctx.IO).Must(`
The {{ Bold "hcp waypoint applications" }} command group lets you manage
HCP Waypoint applications.
		`),
	}

	cmd.AddChild(NewCmdCreateApplication(ctx, opts))
	cmd.AddChild(NewCmdDestroyApplication(ctx, opts))

	return cmd
}
