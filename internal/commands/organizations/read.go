package organizations

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/organization_service"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
)

func NewCmdRead(ctx *cmd.Context, runF func(*ReadOpts) error) *cmd.Command {
	opts := &ReadOpts{
		Ctx:     ctx.ShutdownCtx,
		Profile: ctx.Profile,
		IO:      ctx.IO,
		Output:  ctx.Output,
		Client:  organization_service.New(ctx.HCP, nil),
	}

	cmd := &cmd.Command{
		Name:      "read",
		ShortHelp: "Show metadata for the organization.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp organizations read" }} command shows metadata for the organization.
		`),
		Args: cmd.PositionalArguments{
			Args: []cmd.PositionalArgument{
				{
					Name:          "ID",
					Documentation: "ID of the organization to read.",
				},
			},
		},
		RunF: func(c *cmd.Command, args []string) error {
			opts.ID = args[0]
			if runF != nil {
				return runF(opts)
			}
			return readRun(opts)
		},
	}

	return cmd
}

type ReadOpts struct {
	Ctx     context.Context
	Profile *profile.Profile
	IO      iostreams.IOStreams
	Output  *format.Outputter

	ID     string
	Client organization_service.ClientService
}

func readRun(opts *ReadOpts) error {
	req := organization_service.NewOrganizationServiceGetParamsWithContext(opts.Ctx)
	req.ID = opts.ID

	resp, err := opts.Client.OrganizationServiceGet(req, nil)
	if err != nil {
		return fmt.Errorf("failed to read organization: %w", err)
	}

	d := newDisplayer(format.Pretty, true, resp.Payload.Organization)
	return opts.Output.Display(d)
}
