package waypoint

import (
	"github.com/hashicorp/hcp/internal/commands/waypoint/tfcconfig"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
)

func NewCmdWaypoint(ctx *cmd.Context) *cmd.Command {
	cmd := &cmd.Command{
		Name:      "waypoint",
		ShortHelp: "Manage Waypoint",
		LongHelp:  "Managing HCP Waypoint with CLI commands",
		Examples: []cmd.Example{
			{
				Title:   "TFC Config Set",
				Command: "$ hcp waypoint tfc-config set",
			},
		},
	}

	cmd.AddChild(tfcconfig.NewCmdTFCConfig(ctx))
	return cmd
}
