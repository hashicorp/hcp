package addons

import (
	"github.com/hashicorp/hcp/internal/commands/waypoint/add-ons/definitions"
	"github.com/hashicorp/hcp/internal/commands/waypoint/opts"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
)

type AddOnOpts struct {
	opts.WaypointOpts

	Name                string
	AddOnDefinitionName string
	ApplicationName     string

	testFunc func(c *cmd.Command, args []string) error
}

func NewCmdAddOn(ctx *cmd.Context) *cmd.Command {
	cmd := &cmd.Command{
		Name:      "add-ons",
		ShortHelp: "Manage HCP Waypoint add-ons and add-on definitions.",
		LongHelp: heredoc.New(ctx.IO).Must(`
The {{ template "mdCodeOrBold" "hcp waypoint add-ons" }} command group lets you
manage HCP Waypoint add-ons and add-on definitions.
`),
	}

	cmd.AddChild(definitions.NewCmdAddOnDefinition(ctx))

	opts := &AddOnOpts{
		WaypointOpts: opts.New(ctx),
	}

	cmd.AddChild(NewCmdCreate(ctx, opts))
	cmd.AddChild(NewCmdDestroy(ctx, opts))

	return cmd
}
