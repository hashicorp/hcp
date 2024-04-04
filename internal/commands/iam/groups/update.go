// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package groups

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/client/groups_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/models"
	"github.com/hashicorp/hcp/internal/commands/iam/groups/helper"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/flagvalue"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
)

func NewCmdUpdate(ctx *cmd.Context, runF func(*UpdateOpts) error) *cmd.Command {
	opts := &UpdateOpts{
		Ctx:     ctx.ShutdownCtx,
		Profile: ctx.Profile,
		IO:      ctx.IO,
		Client:  groups_service.New(ctx.HCP, nil),
	}

	cmd := &cmd.Command{
		Name:      "update",
		ShortHelp: "Update an existing group.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp iam groups update" }} command updates a group.

		Update can be used to update the display name or description of an existing group.
		`),
		Examples: []cmd.Example{
			{
				Preamble: "Update a group's description:",
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				$ hcp iam groups update example-group \
				  --description="updated description" \
				  --display-name="new display name"
				`),
			},
		},
		Args: cmd.PositionalArguments{
			Args: []cmd.PositionalArgument{
				{
					Name:          "GROUP_NAME",
					Documentation: heredoc.New(ctx.IO).Mustf(helper.GroupNameArgDoc, "update"),
				},
			},
			Autocomplete: helper.PredictGroupResourceNameSuffix(opts.Ctx, opts.Profile.OrganizationID, opts.Client),
		},
		Flags: cmd.Flags{
			Local: []*cmd.Flag{
				{
					Name:         "description",
					DisplayValue: "NEW_DESCRIPTION",
					Description:  "New description for the group.",
					Value:        flagvalue.Simple((*string)(nil), &opts.Description),
				},
				{
					Name:         "display-name",
					DisplayValue: "NEW_DISPLAY_NAME",
					Description:  "New display name for the group.",
					Value:        flagvalue.Simple((*string)(nil), &opts.DisplayName),
				},
			},
		},
		RunF: func(c *cmd.Command, args []string) error {
			opts.Name = args[0]
			if runF != nil {
				return runF(opts)
			}

			return updateRun(opts)
		},
		PersistentPreRun: func(c *cmd.Command, args []string) error {
			return cmd.RequireOrganization(ctx)
		},
	}

	return cmd
}

type UpdateOpts struct {
	Ctx     context.Context
	Profile *profile.Profile
	IO      iostreams.IOStreams

	Name        string
	DisplayName *string
	Description *string
	Client      groups_service.ClientService
}

func updateRun(opts *UpdateOpts) error {
	if opts.DisplayName == nil && opts.Description == nil {
		return fmt.Errorf("either display name or description must be specified")
	}

	rn := helper.ResourceName(opts.Name, opts.Profile.OrganizationID)
	req := groups_service.NewGroupsServiceUpdateGroup2ParamsWithContext(opts.Ctx)
	req.ResourceName = rn
	req.Group = &models.HashicorpCloudIamGroup{}

	if opts.DisplayName != nil {
		req.Group.DisplayName = *opts.DisplayName
	}
	if opts.Description != nil {
		req.Group.Description = *opts.Description
	}

	if _, err := opts.Client.GroupsServiceUpdateGroup2(req, nil); err != nil {
		return fmt.Errorf("failed to update group: %w", err)
	}

	fmt.Fprintf(opts.IO.Err(), "%s Group %q updated\n",
		opts.IO.ColorScheme().SuccessIcon(), rn)
	return nil
}
