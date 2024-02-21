package actionconfig

import "github.com/hashicorp/hcp/internal/pkg/cmd"

func NewCmdDelete(ctx *cmd.Context) *cmd.Command {
	cmd := &cmd.Command{
		Name:      "delete",
		ShortHelp: "Delete an existing action configuration.",
		LongHelp:  "Delete an existing action configuration.",
		RunF:      deleteActionConfig,
	}

	return cmd
}

func deleteActionConfig(c *cmd.Command, args []string) error {
	return nil
}
