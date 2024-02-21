package waypoint

import (
    "github.com/hashicorp/hcp/internal/commands/waypoint/tfcconfig"
    "github.com/hashicorp/hcp/internal/pkg/cmd"
    "github.com/hashicorp/hcp/internal/pkg/heredoc"
)

func NewCmdWaypoint(ctx *cmd.Context) *cmd.Command {
    cmd := &cmd.Command{
        Name:      "waypoint",
        ShortHelp: "Manage Waypoint.",
        LongHelp: heredoc.New(ctx.IO).Must(`The {{ Bold "hcp waypoint" }} command group allows users to manage HCP Waypoint resources through the CLI.
These commands allow the user to interact with their HCP Waypoint instance to manage their application deployment process.`),
    }

    cmd.AddChild(tfcconfig.NewCmdTFCConfig(ctx))
    return cmd
}
