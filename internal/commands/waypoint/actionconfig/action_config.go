package actionconfig

import (
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
)

var (
	actionConfigFields = []format.Field{
		format.NewField("Name", "{{ .Name }}"),
		format.NewField("ID", "{{ .ID }}"),
		format.NewField("Description", "{{ .Description }}"),
	}
)

func NewCmdActionConfig(ctx *cmd.Context) *cmd.Command {
	cmd := &cmd.Command{
		Name:      "action-config",
		ShortHelp: "Manage action configuration options for HCP Waypoint.",
		LongHelp: heredoc.New(ctx.IO).Must(`
Manage all action configuration options for HCP Waypoint. An action
configuration is a set of options that define how an action is executed. This
includes the action request type, and the action name. The action
configuration is used to launch action runs depending on the Request type.
		`),
	}

	cmd.AddChild(NewCmdCreate(ctx))
	cmd.AddChild(NewCmdDelete(ctx))
	cmd.AddChild(NewCmdList(ctx))
	cmd.AddChild(NewCmdRead(ctx))

	return cmd
}
