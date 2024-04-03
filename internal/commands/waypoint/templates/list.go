package templates

import (
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2023-08-18/client/waypoint_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2023-08-18/models"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/pkg/errors"
)

func NewCmdList(ctx *cmd.Context, opts *TemplateOpts) *cmd.Command {
	cmd := &cmd.Command{
		Name:      "list",
		ShortHelp: "List all known HCP Waypoint templates.",
		LongHelp: heredoc.New(ctx.IO).Must(`
The {{ template "mdCodeOrBold" "hcp waypoint templates list" }} command lets you list
existing HCP Waypoint templates.
		`),
		RunF: func(c *cmd.Command, args []string) error {
			return listTemplates(opts)
		},
		PersistentPreRun: func(c *cmd.Command, args []string) error {
			return cmd.RequireOrgAndProject(ctx)
		},
	}

	return cmd
}

func listTemplates(opts *TemplateOpts) error {
	ns, err := opts.Namespace()
	if err != nil {
		return errors.Wrapf(err, "unable to access HCP project")
	}

	var templates []*models.HashicorpCloudWaypointApplicationTemplate

	resp, err := opts.WS.WaypointServiceListApplicationTemplates(
		&waypoint_service.WaypointServiceListApplicationTemplatesParams{
			NamespaceID: ns.ID,
			Context:     opts.Ctx,
		}, nil)
	if err != nil {
		return errors.Wrapf(err, "failed to list templates")
	}

	templates = append(templates, resp.GetPayload().ApplicationTemplates...)

	for resp.GetPayload().Pagination.NextPageToken != "" {
		resp, err = opts.WS.WaypointServiceListApplicationTemplates(
			&waypoint_service.WaypointServiceListApplicationTemplatesParams{
				NamespaceID:             ns.ID,
				Context:                 opts.Ctx,
				PaginationNextPageToken: &resp.GetPayload().Pagination.NextPageToken,
			}, nil)
		if err != nil {
			return errors.Wrapf(err, "failed to list paginated templates")
		}

		templates = append(templates, resp.GetPayload().ApplicationTemplates...)
	}

	return opts.Output.Show(templates, format.Pretty)
}
