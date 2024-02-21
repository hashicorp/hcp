package actionconfig

import "github.com/hashicorp/hcp/internal/pkg/cmd"

func NewCmdShow(ctx *cmd.Context) *cmd.Command {
	cmd := &cmd.Command{
		Name:      "show",
		ShortHelp: "Show more details about an action configurations.",
		LongHelp:  "Show more details about an action configuration.",
		RunF:      showActionConfig,
	}

	return cmd
}

func showActionConfig(c *cmd.Command, args []string) error {
	return nil
}
