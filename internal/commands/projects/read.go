package projects

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/project_service"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
)

func NewCmdRead(ctx *cmd.Context, runF func(*ReadOpts) error) *cmd.Command {
	opts := &ReadOpts{
		Ctx:     ctx.ShutdownCtx,
		Profile: ctx.Profile,
		IO:      ctx.IO,
		Output:  ctx.Output,
		Client:  project_service.New(ctx.HCP, nil),
	}

	cmd := &cmd.Command{
		Name:      "read",
		ShortHelp: "Show metadata for the project.",
		LongHelp:  "Show metadata for the project.",
		Examples: []cmd.Example{
			{
				Preamble: "Read a project:",
				Command:  "$ hcp projects read --project=cd3d34d5-ceeb-493d-b004-9297365a01af",
			},
		},
		RunF: func(c *cmd.Command, args []string) error {
			if runF != nil {
				return runF(opts)
			}
			return readRun(opts)
		},
		PersistentPreRun: func(c *cmd.Command, args []string) error {
			return cmd.RequireOrgAndProject(ctx)
		},
	}

	return cmd
}

type ReadOpts struct {
	Ctx     context.Context
	Profile *profile.Profile
	IO      iostreams.IOStreams
	Output  *format.Outputter
	Client  project_service.ClientService
}

func readRun(opts *ReadOpts) error {
	req := project_service.NewProjectServiceGetParamsWithContext(opts.Ctx)
	req.ID = opts.Profile.ProjectID

	resp, err := opts.Client.ProjectServiceGet(req, nil)
	if err != nil {
		return fmt.Errorf("failed to read project: %w", err)
	}

	d := newDisplayer(format.Pretty, true, resp.Payload.Project)
	return opts.Output.Display(d)
}
