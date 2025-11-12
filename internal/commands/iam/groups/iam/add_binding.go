// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/client/groups_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/client/iam_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/organization_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/resource_service"
	"github.com/hashicorp/hcp/internal/commands/iam/groups/helper"
	"github.com/hashicorp/hcp/internal/pkg/api/iampolicy"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/flagvalue"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
)

func NewCmdAddBinding(ctx *cmd.Context, runF func(*AddBindingOpts) error) *cmd.Command {
	opts := &AddBindingOpts{
		Ctx:     ctx.ShutdownCtx,
		Profile: ctx.Profile,
		IO:      ctx.IO,

		GroupsClient:   groups_service.New(ctx.HCP, nil),
		ResourceClient: resource_service.New(ctx.HCP, nil),
	}

	cmd := &cmd.Command{
		Name:      "add-binding",
		ShortHelp: "Add an IAM policy binding for a group.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp iam groups iam add-binding" }}
		command adds an IAM policy binding for the given group. A binding grants the
		specified principal the given role on the group.

		To view the available roles to bind, run {{ template "mdCodeOrBold" "hcp iam roles list" }}.

		Currently, the only supported role on a principal in a group is {{ template "mdCodeOrBold" "roles/iam.group-manager" }}.

		A group manager can add/remove members from the group and update the group name/description.
		`),
		Examples: []cmd.Example{
			{
				Preamble: heredoc.New(ctx.IO).Must(`Bind a principal to role {{ template "mdCodeOrBold" "roles/iam.group-manager" }}:`),
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				$ hcp iam groups iam add-binding \
				  --group=Group-Name \
				  --member=ef938a22-09cf-4be9-b4d0-1f4587f80f53 \
				  --role=roles/iam.group-manager
				`),
			},
		},
		Flags: cmd.Flags{
			Local: []*cmd.Flag{
				{
					Name:         "group",
					Shorthand:    "g",
					DisplayValue: "NAME",
					Description:  "The name of the group to add the role binding to.",
					Value:        flagvalue.Simple("", &opts.GroupName),
					Autocomplete: helper.PredictGroupResourceNameSuffix(opts.Ctx, opts.Profile.OrganizationID, opts.GroupsClient),
					Required:     true,
				},
				{
					Name:         "member",
					Shorthand:    "m",
					DisplayValue: "PRINCIPAL_ID",
					Description:  "The ID of the principal to add the role binding to.",
					Value:        flagvalue.Simple("", &opts.PrincipalID),
					Required:     true,
				},
				{
					Name:         "role",
					Shorthand:    "r",
					DisplayValue: "ROLE_ID",
					Description:  `The role ID (e.g. "roles/iam.group-manager") to bind the member to.`,
					Value:        flagvalue.Simple("", &opts.Role),
					Required:     true,
					Autocomplete: iampolicy.AutocompleteRoles(opts.Ctx, ctx.Profile.OrganizationID, organization_service.New(ctx.HCP, nil)),
				},
			},
		},
		RunF: func(c *cmd.Command, args []string) error {
			resourceName := helper.ResourceName(opts.GroupName, ctx.Profile.OrganizationID)

			// Create the group IAM Updater
			u := &iamUpdater{
				resourceName: resourceName,
				client:       opts.ResourceClient,
			}

			// Create the policy setter
			opts.Setter = iampolicy.NewSetter(
				ctx.Profile.OrganizationID,
				u,
				iam_service.New(ctx.HCP, nil),
				c.Logger())

			if runF != nil {
				return runF(opts)
			}

			return addBindingRun(opts)
		},
		PersistentPreRun: func(c *cmd.Command, args []string) error {
			return cmd.RequireOrganization(ctx)
		},
	}

	return cmd
}

type AddBindingOpts struct {
	Ctx     context.Context
	Profile *profile.Profile
	IO      iostreams.IOStreams

	Setter         iampolicy.Setter
	GroupName      string
	PrincipalID    string
	Role           string
	GroupsClient   groups_service.ClientService
	ResourceClient resource_service.ClientService
}

func addBindingRun(opts *AddBindingOpts) error {
	_, err := opts.Setter.AddBinding(opts.Ctx, opts.PrincipalID, opts.Role)
	if err != nil {
		return err
	}

	fmt.Fprintf(opts.IO.Err(), "%s Principal %q bound to role %q.\n",
		opts.IO.ColorScheme().SuccessIcon(), opts.PrincipalID, opts.Role)
	return nil
}
