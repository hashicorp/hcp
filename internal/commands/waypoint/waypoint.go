package waypoint

import (
	"github.com/hashicorp/hcp/internal/commands/waypoint/tfcconfig"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
)

func NewCmdWaypoint(ctx *cmd.Context) *cmd.Command {
	cmd := &cmd.Command{
		Name:      "waypoint",
		ShortHelp: "Manage Waypoint",
		LongHelp:  heredoc.New(ctx.IO).Must(`"Managing HCP Waypoint with CLI commands"`),
	}

	cmd.AddChild(tfcconfig.NewCmdTFCConfig(ctx))
	return cmd
}
