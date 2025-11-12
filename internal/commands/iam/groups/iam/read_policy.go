// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/client/groups_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/client/iam_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/resource_service"
	"github.com/hashicorp/hcp/internal/commands/iam/groups/helper"
	"github.com/hashicorp/hcp/internal/pkg/api/iampolicy"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/flagvalue"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
)

func NewCmdReadPolicy(ctx *cmd.Context, runF func(*ReadPolicyOpts) error) *cmd.Command {
	opts := &ReadPolicyOpts{
		Ctx:            ctx.ShutdownCtx,
		Profile:        ctx.Profile,
		IO:             ctx.IO,
		Output:         ctx.Output,
		ResourceClient: resource_service.New(ctx.HCP, nil),
		GroupsClient:   groups_service.New(ctx.HCP, nil),
		IAMClient:      iam_service.New(ctx.HCP, nil),
	}

	cmd := &cmd.Command{
		Name:      "read-policy",
		ShortHelp: "Read the IAM policy for a group.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp iam groups iam read-policy" }} command reads the IAM policy for a group.
		`),
		Examples: []cmd.Example{
			{
				Preamble: "Read the IAM Policy for a group:",
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				$ hcp iam groups iam read-policy \
				  --group=iam/organization/cf8ef907-b9b9-4f2f-b675-e290448f0000/group/Group-Name
				`),
			},
		},
		Flags: cmd.Flags{
			Local: []*cmd.Flag{
				{
					Name:         "group",
					Shorthand:    "g",
					DisplayValue: "NAME",
					Description:  heredoc.New(ctx.IO).Mustf(helper.GroupNameArgDoc, "read the IAM policy for"),
					Value:        flagvalue.Simple("", &opts.GroupName),
					Autocomplete: helper.PredictGroupResourceNameSuffix(opts.Ctx, opts.Profile.OrganizationID, opts.GroupsClient),
				},
			},
		},
		RunF: func(c *cmd.Command, args []string) error {
			if opts.GroupName == "" {
				return fmt.Errorf("a group resource name must be specified")
			}

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

	GroupName      string
	ResourceClient resource_service.ClientService
	GroupsClient   groups_service.ClientService
	IAMClient      iam_service.ClientService
}

func readPolicyRun(opts *ReadPolicyOpts) error {
	resourceName := helper.ResourceName(opts.GroupName, opts.Profile.OrganizationID)

	// Create the group IAM Updater
	u := &iamUpdater{
		resourceName: resourceName,
		client:       opts.ResourceClient,
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
