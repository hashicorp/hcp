package hcp

import (
	"github.com/hashicorp/hcp/internal/commands/auth"
	"github.com/hashicorp/hcp/internal/commands/iam"
	"github.com/hashicorp/hcp/internal/commands/organizations"
	"github.com/hashicorp/hcp/internal/commands/profile"
	"github.com/hashicorp/hcp/internal/commands/projects"
	"github.com/hashicorp/hcp/internal/commands/version"
	"github.com/hashicorp/hcp/internal/commands/waypoint"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
)

func NewCmdHcp(ctx *cmd.Context) *cmd.Command {
	c := &cmd.Command{
		Name:      "hcp",
		ShortHelp: "Interact with HCP.",
		LongHelp:  "The HCP Command Line Interface is a unified tool to manage your HCP services.",
	}

	//  _   _  ___ _____ _____
	// | \ | |/ _ \_   _| ____|
	// |  \| | | | || | |  _|
	// | |\  | |_| || | | |___
	// |_| \_|\___/ |_| |_____|
	//
	// When adding a top level command group, be sure to regenerate the
	// screenshot in the README by running `make gen/screenshot`.

	// Add the subcommands
	c.AddChild(version.NewCmdVersion(ctx))
	c.AddChild(auth.NewCmdAuth(ctx))
	c.AddChild(projects.NewCmdProjects(ctx))
	c.AddChild(profile.NewCmdProfile(ctx))
	c.AddChild(organizations.NewCmdOrganizations(ctx))
	c.AddChild(iam.NewCmdIam(ctx))
	c.AddChild(waypoint.NewCmdWaypoint(ctx))

	// Configure the command as the root command.
	cmd.ConfigureRootCommand(ctx, c)

	return c
}
