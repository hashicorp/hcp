// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package groups

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/client/groups_service"
	"github.com/hashicorp/hcp/internal/commands/iam/groups/helper"
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
		Client:  groups_service.New(ctx.HCP, nil),
	}

	cmd := &cmd.Command{
		Name:      "read",
		ShortHelp: "Show metadata for the given group.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp iam groups read" }} command reads details about the given group.
		`),
		Examples: []cmd.Example{
			{
				Preamble: `Read the group using the resource name suffix "example-group":`,
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				$ hcp iam groups read example-group
				`),
			},
			{
				Preamble: `Read the group using the group's resource name:`,
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				$ hcp iam groups read iam/organization/example-org/group/example-group
				`),
			},
		},
		Args: cmd.PositionalArguments{
			Args: []cmd.PositionalArgument{
				{
					Name:          "GROUP_NAME",
					Documentation: heredoc.New(ctx.IO).Mustf(helper.GroupNameArgDoc, "read"),
				},
			},
			Autocomplete: helper.PredictGroupResourceNameSuffix(opts.Ctx, opts.Profile.OrganizationID, opts.Client),
		},
		RunF: func(c *cmd.Command, args []string) error {
			opts.Name = args[0]

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

	Name   string
	Client groups_service.ClientService
}

func readRun(opts *ReadOpts) error {
	rn := helper.ResourceName(opts.Name, opts.Profile.OrganizationID)
	req := groups_service.NewGroupsServiceGetGroupParamsWithContext(opts.Ctx)
	req.ResourceName = rn

	resp, err := opts.Client.GroupsServiceGetGroup(req, nil)
	if err != nil {
		return fmt.Errorf("failed to read group: %w", err)
	}

	return opts.Output.Display(newDisplayer(format.Pretty, true, resp.Payload.Group))
}
