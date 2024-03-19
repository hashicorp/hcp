package addon

import "github.com/hashicorp/hcp/internal/pkg/cmd"

func NewCmdAddOn(ctx *cmd.Context) *cmd.Command {
	cmd := &cmd.Command{
		Name:      "add-ons",
		ShortHelp: "Manage HCP Waypoint add-ons and add-on definitions.",
		LongHelp: "Manage HCP Waypoint add-ons. Add-ons are units of infrastructure" +
			" that can be deployed alongside applications. They can be used to deploy" +
			" databases, caches, and other infrastructure alongside applications.",
	}

	cmd.AddChild(NewCmdAddOnDefinition(ctx))

	return cmd
}
