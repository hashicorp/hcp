// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package definitions

import (
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2024-11-22/client/waypoint_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2024-11-22/models"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/pkg/errors"
)

func NewCmdAddOnDefinitionList(ctx *cmd.Context, opts *AddOnDefinitionOpts) *cmd.Command {
	cmd := &cmd.Command{
		Name:      "list",
		ShortHelp: "List all known HCP Waypoint add-on definitions.",
		LongHelp: heredoc.New(ctx.IO).Must(`
The {{ template "mdCodeOrBold" "hcp waypoint add-ons definitions list" }}
command lets you list all existing HCP Waypoint add-on definitions.
`),
		Examples: []cmd.Example{
			{
				Preamble: "List all known HCP Waypoint add-on definitions:",
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
$ hcp waypoint add-ons definitions list
`),
			},
		},
		RunF: func(c *cmd.Command, args []string) error {
			if opts.testFunc != nil {
				return opts.testFunc(c, args)
			}
			return addOnDefinitionsList(opts)
		},
		PersistentPreRun: func(c *cmd.Command, args []string) error {
			return cmd.RequireOrgAndProject(ctx)
		},
	}

	return cmd
}

func addOnDefinitionsList(opts *AddOnDefinitionOpts) error {
	var addOnDefinitions []*models.HashicorpCloudWaypointAddOnDefinition

	listResp, err := opts.WS2024Client.WaypointServiceListAddOnDefinitions(
		&waypoint_service.WaypointServiceListAddOnDefinitionsParams{
			NamespaceLocationOrganizationID: opts.Profile.OrganizationID,
			NamespaceLocationProjectID:      opts.Profile.ProjectID,
			Context:                         opts.Ctx,
		}, nil,
	)
	if err != nil {
		return errors.Wrapf(err, "%s failed to list add-on definitions",
			opts.IO.ColorScheme().FailureIcon(),
		)
	}

	addOnDefinitions = append(addOnDefinitions, listResp.GetPayload().AddOnDefinitions...)

	for listResp.GetPayload().Pagination.NextPageToken != "" {
		listResp, err = opts.WS2024Client.WaypointServiceListAddOnDefinitions(
			&waypoint_service.WaypointServiceListAddOnDefinitionsParams{
				NamespaceLocationOrganizationID: opts.Profile.OrganizationID,
				NamespaceLocationProjectID:      opts.Profile.ProjectID,
				Context:                         opts.Ctx,
				PaginationNextPageToken:         &listResp.GetPayload().Pagination.NextPageToken,
			}, nil)
		if err != nil {
			return errors.Wrapf(err, "%s failed to list paginated add-on definitions",
				opts.IO.ColorScheme().FailureIcon(),
			)
		}

		addOnDefinitions = append(addOnDefinitions, listResp.GetPayload().AddOnDefinitions...)
	}

	return opts.Output.Show(addOnDefinitions, format.Pretty)
}
