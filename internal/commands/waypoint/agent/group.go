package agent

import (
	"github.com/hashicorp/hcp/internal/commands/waypoint/opts"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
)

type GroupOpts struct {
	opts.WaypointOpts

	Name        string
	Description string

	testFunc func(c *cmd.Command, args []string) error
}

func NewCmdGroup(ctx *cmd.Context) *cmd.Command {
	opts := &GroupOpts{
		WaypointOpts: opts.New(ctx),
	}

	cmd := &cmd.Command{
		Name:      "group",
		ShortHelp: "Manage HCP Waypoint Agent groups.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp waypoint agent group" }} command group manages agent groups.
		`),
		PersistentPreRun: func(c *cmd.Command, args []string) error {
			return cmd.RequireOrgAndProject(ctx)
		},
	}

	cmd.AddChild(NewCmdGroupCreate(ctx, opts))
	cmd.AddChild(NewCmdGroupList(ctx, opts))
	cmd.AddChild(NewCmdGroupDelete(ctx, opts))

	return cmd
}
