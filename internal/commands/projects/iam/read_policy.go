// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/client/iam_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/project_service"
	"github.com/hashicorp/hcp/internal/pkg/api/iampolicy"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
)

func NewCmdReadPolicy(ctx *cmd.Context, runF func(*ReadPolicyOpts) error) *cmd.Command {
	opts := &ReadPolicyOpts{
		Ctx:           ctx.ShutdownCtx,
		Profile:       ctx.Profile,
		IO:            ctx.IO,
		Output:        ctx.Output,
		ProjectClient: project_service.New(ctx.HCP, nil),
		IAMClient:     iam_service.New(ctx.HCP, nil),
	}

	cmd := &cmd.Command{
		Name:      "read-policy",
		ShortHelp: "Read the IAM policy for a project.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp projects iam read-policy" }} command reads the IAM policy for a project.
		`),
		Examples: []cmd.Example{
			{
				Preamble: "Read the IAM Policy for a project:",
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				$ hcp projects iam read-policy \
				  --project=8647ae06-ca65-467a-b72d-edba1f908fc8
				`),
			},
		},
		RunF: func(c *cmd.Command, args []string) error {
			if runF != nil {
				return runF(opts)
			}

			return readPolicyRun(opts)
		},
		PersistentPreRun: func(c *cmd.Command, args []string) error {
			return cmd.RequireOrgAndProject(ctx)
		},
	}

	return cmd
}

type ReadPolicyOpts struct {
	Ctx     context.Context
	Profile *profile.Profile
	IO      iostreams.IOStreams
	Output  *format.Outputter

	ProjectClient project_service.ClientService
	IAMClient     iam_service.ClientService
}

func readPolicyRun(opts *ReadPolicyOpts) error {
	// Create our project IAM Updater
	u := &iamUpdater{
		projectID: opts.Profile.ProjectID,
		client:    opts.ProjectClient,
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
