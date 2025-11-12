// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: MPL-2.0

package groups

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/client/groups_service"
	"github.com/hashicorp/hcp/internal/commands/iam/groups/helper"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
)

func NewCmdDelete(ctx *cmd.Context, runF func(*DeleteOpts) error) *cmd.Command {
	opts := &DeleteOpts{
		Ctx:     ctx.ShutdownCtx,
		Profile: ctx.Profile,
		IO:      ctx.IO,
		Client:  groups_service.New(ctx.HCP, nil),
	}

	cmd := &cmd.Command{
		Name:      "delete",
		ShortHelp: "Delete a group.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp iam groups delete" }} command deletes a group.

		Once the group is deleted, all permissions granted to members based on group membership will also be revoked.
		`),
		Examples: []cmd.Example{
			{
				Preamble: `Delete a group using its name suffix:`,
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				$ hcp iam groups delete team-platform
				`),
			},
			{
				Preamble: `Delete a group using its resource name:`,
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				$ hcp iam groups delete iam/organization/example-org/group/team-platform
				`),
			},
		},
		Args: cmd.PositionalArguments{
			Args: []cmd.PositionalArgument{
				{
					Name:          "GROUP_NAME",
					Documentation: heredoc.New(ctx.IO).Mustf(helper.GroupNameArgDoc, "delete"),
				},
			},
			Autocomplete: helper.PredictGroupResourceNameSuffix(opts.Ctx, opts.Profile.OrganizationID, opts.Client),
		},
		RunF: func(c *cmd.Command, args []string) error {
			opts.Name = args[0]

			if runF != nil {
				return runF(opts)
			}
			return deleteRun(opts)
		},
		PersistentPreRun: func(c *cmd.Command, args []string) error {
			return cmd.RequireOrganization(ctx)
		},
	}

	return cmd
}

type DeleteOpts struct {
	Ctx     context.Context
	IO      iostreams.IOStreams
	Profile *profile.Profile

	Name   string
	Client groups_service.ClientService
}

func deleteRun(opts *DeleteOpts) error {
	rn := helper.ResourceName(opts.Name, opts.Profile.OrganizationID)
	if opts.IO.CanPrompt() {
		ok, err := opts.IO.PromptConfirm("The group will be deleted.\n\nDo you want to continue")
		if err != nil {
			return fmt.Errorf("failed to retrieve confirmation: %w", err)
		}

		if !ok {
			return nil
		}
	}

	req := groups_service.NewGroupsServiceDeleteGroupParamsWithContext(opts.Ctx)
	req.ResourceName = rn

	_, err := opts.Client.GroupsServiceDeleteGroup(req, nil)
	if err != nil {
		return fmt.Errorf("failed to delete group: %w", err)
	}

	fmt.Fprintf(opts.IO.Err(), "%s Group %q deleted\n",
		opts.IO.ColorScheme().SuccessIcon(), rn)
	return nil
}
