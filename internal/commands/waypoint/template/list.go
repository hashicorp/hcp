package template

import (
	"fmt"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2023-08-18/client/waypoint_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2023-08-18/models"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/format"
)

func NewCmdList(ctx *cmd.Context, opts *TemplateOpts) *cmd.Command {
	cmd := &cmd.Command{
		Name:      "list",
		ShortHelp: "List all known HCP Waypoint templates.",
		LongHelp:  "List all known templates for HCP Waypoint.",
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
		return err
	}

	var templates []*models.HashicorpCloudWaypointApplicationTemplate

	resp, err := opts.WS.WaypointServiceListApplicationTemplates(
		&waypoint_service.WaypointServiceListApplicationTemplatesParams{
			NamespaceID: ns.ID,
			Context:     opts.Ctx,
		}, nil)
	if err != nil {
		return fmt.Errorf("error listing templates: %w", err)
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
			return fmt.Errorf("error listing templates: %w", err)
		}

		templates = append(templates, resp.GetPayload().ApplicationTemplates...)
	}
	return opts.Output.Show(templates, format.Pretty)
}
