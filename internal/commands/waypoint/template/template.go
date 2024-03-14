package template

import (
	"github.com/hashicorp/hcp/internal/commands/waypoint/opts"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
)

type TemplateOpts struct {
	opts.WaypointOpts

	Name                       string
	Summary                    string
	Description                string
	ReadmeMarkdownTemplateFile string
	Labels                     []string
	Tags                       map[string]string

	TerraformNoCodeModuleSource  string
	TerraformNoCodeModuleVersion string
	TerraformCloudProjectName    string
	TerraformCloudProjectID      string

	// testFunc is used for testing, so that the command can be tested without
	// using the real API.
	testFunc func(c *cmd.Command, args []string) error
}

func NewCmdTemplate(ctx *cmd.Context) *cmd.Command {
	opts := &TemplateOpts{
		WaypointOpts: opts.New(ctx),
	}

	cmd := &cmd.Command{
		Name:      "templates",
		ShortHelp: "Manage HCP Waypoint templates.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ Bold "hcp waypoint template" }} commands manage templates. A
		template is a reusable configuration for creating applications.
		`),
	}

	cmd.AddChild(NewCmdCreate(ctx, opts))

	return cmd
}
