package actionconfig

import (
	"github.com/hashicorp/hcp/internal/commands/waypoint/opts"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
)

type CreateOpts struct {
	opts.WaypointOpts
}

func NewCmdCreate(ctx *cmd.Context) *cmd.Command {
	cmd := &cmd.Command{
		Name:      "create",
		ShortHelp: "Create a new action configuration.",
		LongHelp:  "Create a new action configuration.",
		RunF:      createActionConfig,
	}

	return cmd
}

func createActionConfig(c *cmd.Command, args []string) error {
	return nil
}
