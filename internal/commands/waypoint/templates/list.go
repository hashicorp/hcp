// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package templates

import (
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2024-11-22/client/waypoint_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2024-11-22/models"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/pkg/errors"
)

func NewCmdList(ctx *cmd.Context, opts *TemplateOpts) *cmd.Command {
	c := &cmd.Command{
		Name:      "list",
		ShortHelp: "List all HCP Waypoint templates.",
		LongHelp: heredoc.New(ctx.IO).Must(`
The {{ template "mdCodeOrBold" "hcp waypoint templates list" }} command lets you list
existing HCP Waypoint templates.
		`),
		Examples: []cmd.Example{
			{
				Preamble: "List all HCP Waypoint templates:",
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
$ hcp waypoint templates list
`),
			},
		},
		RunF: func(c *cmd.Command, args []string) error {
			return listTemplates(opts)
		},
		PersistentPreRun: func(c *cmd.Command, args []string) error {
			return cmd.RequireOrgAndProject(ctx)
		},
	}

	return c
}

func listTemplates(opts *TemplateOpts) error {
	var templates []*models.HashicorpCloudWaypointApplicationTemplate

	resp, err := opts.WS2024Client.WaypointServiceListApplicationTemplates(
		&waypoint_service.WaypointServiceListApplicationTemplatesParams{
			NamespaceLocationOrganizationID: opts.Profile.OrganizationID,
			NamespaceLocationProjectID:      opts.Profile.ProjectID,
			Context:                         opts.Ctx,
		}, nil)
	if err != nil {
		return errors.Wrapf(err, "%s failed to list templates",
			opts.IO.ColorScheme().FailureIcon(),
		)
	}

	templates = append(templates, resp.GetPayload().ApplicationTemplates...)

	for resp.GetPayload().Pagination.NextPageToken != "" {
		resp, err = opts.WS2024Client.WaypointServiceListApplicationTemplates(
			&waypoint_service.WaypointServiceListApplicationTemplatesParams{
				NamespaceLocationOrganizationID: opts.Profile.OrganizationID,
				NamespaceLocationProjectID:      opts.Profile.ProjectID,
				Context:                         opts.Ctx,
				PaginationNextPageToken:         &resp.GetPayload().Pagination.NextPageToken,
			}, nil)
		if err != nil {
			return errors.Wrapf(err, "%s failed to list paginated templates",
				opts.IO.ColorScheme().FailureIcon(),
			)
		}

		templates = append(templates, resp.GetPayload().ApplicationTemplates...)
	}

	templateFields := []format.Field{
		format.NewField("Name", "{{ .Name }}"),
		format.NewField("ID", "{{ .ID }}"),
		format.NewField("Summary", "{{ .Summary }}"),
	}

	return opts.Output.Display(format.NewDisplayer(templates, format.Table, templateFields))
}
