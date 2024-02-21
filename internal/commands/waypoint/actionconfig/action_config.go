package actionconfig

import "github.com/hashicorp/hcp/internal/pkg/cmd"

func NewCmdActionConfig(ctx *cmd.Context) *cmd.Command {
	cmd := &cmd.Command{
		Name:      "action-config",
		ShortHelp: "Manage action configuration options for HCP Waypoint.",
		LongHelp:  "Manage action configuration options for HCP Waypoint.",
	}

	cmd.AddChild(NewCmdCreate(ctx))
	cmd.AddChild(NewCmdDelete(ctx))
	cmd.AddChild(NewCmdList(ctx))
	cmd.AddChild(NewCmdShow(ctx))

	return cmd
}
