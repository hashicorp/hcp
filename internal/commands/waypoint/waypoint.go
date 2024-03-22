package waypoint

import (
	"github.com/hashicorp/hcp/internal/commands/waypoint/actionconfig"
	"github.com/hashicorp/hcp/internal/commands/waypoint/agent"
	"github.com/hashicorp/hcp/internal/commands/waypoint/template"
	"github.com/hashicorp/hcp/internal/commands/waypoint/tfcconfig"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
)

func NewCmdWaypoint(ctx *cmd.Context) *cmd.Command {
	cmd := &cmd.Command{
		Name:      "waypoint",
		ShortHelp: "Manage Waypoint.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp waypoint" }} command group lets you
		manage HCP Waypoint resources through the CLI. These commands let you to interact
		with their HCP Waypoint instance to manage their application deployment process.
		`),
	}

	cmd.AddChild(tfcconfig.NewCmdTFCConfig(ctx))
	cmd.AddChild(actionconfig.NewCmdActionConfig(ctx))
	cmd.AddChild(agent.NewCmdAgent(ctx))
	cmd.AddChild(template.NewCmdTemplate(ctx))

	return cmd
}
