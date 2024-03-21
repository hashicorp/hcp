package iam

import (
	"context"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/client/iam_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/organization_service"
	"github.com/hashicorp/hcp/internal/pkg/api/iampolicy"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
)

func NewCmdReadPolicy(ctx *cmd.Context, runF func(*ReadPolicyOpts) error) *cmd.Command {
	opts := &ReadPolicyOpts{
		Ctx:                ctx.ShutdownCtx,
		Profile:            ctx.Profile,
		IO:                 ctx.IO,
		Output:             ctx.Output,
		OrganizationClient: organization_service.New(ctx.HCP, nil),
		IAMClient:          iam_service.New(ctx.HCP, nil),
	}

	cmd := &cmd.Command{
		Name:      "read-policy",
		ShortHelp: "Read the IAM policy for the organization.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp organizations iam read-policy" }} command reads the IAM policy for the organization.
		`),
		Examples: []cmd.Example{
			{
				Preamble: "Read the IAM Policy for the organization:",
				Command:  "$ hcp organizations iam read-policy",
			},
		},
		RunF: func(c *cmd.Command, args []string) error {
			if runF != nil {
				return runF(opts)
			}

			return readPolicyRun(opts)
		},
		PersistentPreRun: func(c *cmd.Command, args []string) error {
			return cmd.RequireOrganization(ctx)
		},
	}

	return cmd
}

type ReadPolicyOpts struct {
	Ctx     context.Context
	Profile *profile.Profile
	IO      iostreams.IOStreams
	Output  *format.Outputter

	OrganizationClient organization_service.ClientService
	IAMClient          iam_service.ClientService
}

func readPolicyRun(opts *ReadPolicyOpts) error {
	// Create our IAM Updater
	u := &iamUpdater{
		orgID:  opts.Profile.OrganizationID,
		client: opts.OrganizationClient,
	}

	// Get the existing policy
	p, err := u.GetIamPolicy(opts.Ctx)
	if err != nil {
		return err
	}

	// Get the displayer
	d, err := iampolicy.NewDisplayer(opts.Ctx, opts.Profile.OrganizationID, p, opts.IAMClient)
	if err != nil {
		return err
	}

	return opts.Output.Display(d)
}
