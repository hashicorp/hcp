package actionconfig

import "github.com/hashicorp/hcp/internal/pkg/cmd"

func NewCmdList(ctx *cmd.Context) *cmd.Command {
	cmd := &cmd.Command{
		Name:      "list",
		ShortHelp: "List all known action configurations.",
		LongHelp:  "List all known action configuration.",
		RunF:      listActionConfig,
	}

	return cmd
}

func listActionConfig(c *cmd.Command, args []string) error {
	return nil
}
