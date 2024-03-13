package users

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/client/iam_service"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/hashicorp/hcp/internal/pkg/profile"
)

func NewCmdRead(ctx *cmd.Context, runF func(*ReadOpts) error) *cmd.Command {
	opts := &ReadOpts{
		Ctx:     ctx.ShutdownCtx,
		Profile: ctx.Profile,
		Output:  ctx.Output,
		Client:  iam_service.New(ctx.HCP, nil),
	}

	cmd := &cmd.Command{
		Name:      "read",
		ShortHelp: "Show metadata for the given user.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp iam users read" }} command reads details about the given user.
		`),
		Examples: []cmd.Example{
			{
				Preamble: `Read the user principal with ID "example-id-123":`,
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				$ hcp iam users read example-id-123
				`),
			},
		},
		Args: cmd.PositionalArguments{
			Args: []cmd.PositionalArgument{
				{
					Name:          "ID",
					Documentation: "The ID of the user to read.",
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
		PersistentPreRun: func(c *cmd.Command, args []string) error {
			return cmd.RequireOrganization(ctx)
		},
	}

	return cmd
}

type ReadOpts struct {
	Ctx     context.Context
	Profile *profile.Profile
	Output  *format.Outputter

	ID     string
	Client iam_service.ClientService
}

func readRun(opts *ReadOpts) error {
	req := iam_service.NewIamServiceGetUserPrincipalByIDInOrganizationParamsWithContext(opts.Ctx)
	req.OrganizationID = opts.Profile.OrganizationID
	req.UserPrincipalID = opts.ID

	resp, err := opts.Client.IamServiceGetUserPrincipalByIDInOrganization(req, nil)
	if err != nil {
		return fmt.Errorf("failed to read user principal: %w", err)
	}

	return opts.Output.Display(newDisplayer(format.Pretty, true, resp.Payload.UserPrincipal))
}
