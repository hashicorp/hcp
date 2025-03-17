// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package actions

import (
	"fmt"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2024-11-22/client/waypoint_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2024-11-22/models"

	"github.com/hashicorp/hcp/internal/commands/waypoint/opts"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
)

type ListOpts struct {
	opts.WaypointOpts
}

func NewCmdList(ctx *cmd.Context) *cmd.Command {
	opts := &ListOpts{
		WaypointOpts: opts.New(ctx),
	}

	cmd := &cmd.Command{
		Name:      "list",
		ShortHelp: "List all known actions.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp waypoint actions list" }} command
		lists all known actions from HCP Waypoint.
		`),
		RunF: func(c *cmd.Command, args []string) error {
			return listActions(c, args, opts)
		},
		PersistentPreRun: func(c *cmd.Command, args []string) error {
			return cmd.RequireOrgAndProject(ctx)
		},
	}

	return cmd
}

func listActions(c *cmd.Command, args []string, opts *ListOpts) error {
	var actionsList actionsListDisplayer

	resp, err := opts.WS2024Client.WaypointServiceListActionConfigs(&waypoint_service.WaypointServiceListActionConfigsParams{
		NamespaceLocationOrganizationID: opts.Profile.OrganizationID,
		NamespaceLocationProjectID:      opts.Profile.ProjectID,
		Context:                         opts.Ctx,
	}, nil)
	if err != nil {
		return fmt.Errorf("error listing actions: %w", err)
	}

	respPayload := resp.GetPayload()

	actionsList = append(actionsList, respPayload.ActionConfigs...)

	return opts.Output.Show(actionsList, format.Pretty)
}

type actionsListDisplayer []*models.HashicorpCloudWaypointActionConfig

func (d actionsListDisplayer) DefaultFormat() format.Format {
	return format.Table
}

func (d actionsListDisplayer) Payload() any {
	return d
}

func (d actionsListDisplayer) FieldTemplates() []format.Field {
	//TODO(henry): fix this format (maybe don't include URL)
	// Also write tests for this
	return []format.Field{
		{
			Name:        "Name",
			ValueFormat: "{{ .Name }}",
		},
		{
			Name:        "ActionURL",
			ValueFormat: "{{ .ActionURL }}",
		},
		{
			Name:        "Description",
			ValueFormat: "{{ .Description }}",
		},
	}
}
