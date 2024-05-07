// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package actions

import (
	"fmt"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2023-08-18/client/waypoint_service"
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
	ns, err := opts.Namespace()
	if err != nil {
		return err
	}

	resp, err := opts.WS.WaypointServiceListActionConfigs(&waypoint_service.WaypointServiceListActionConfigsParams{
		NamespaceID: ns.ID,
		Context:     opts.Ctx,
	}, nil)
	if err != nil {
		return fmt.Errorf("error listing actions: %w", err)
	}

	respPayload := resp.GetPayload()
	return opts.Output.Show(respPayload.ActionConfigs, format.Pretty)
}
