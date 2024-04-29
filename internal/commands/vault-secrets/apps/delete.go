// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package actionconfig

import (
	"fmt"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2023-08-18/client/waypoint_service"
	"github.com/hashicorp/hcp/internal/commands/waypoint/opts"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/flagvalue"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
)

type DeleteOpts struct {
	opts.WaypointOpts

	Name string
}

func NewCmdDelete(ctx *cmd.Context) *cmd.Command {
	opts := &DeleteOpts{
		WaypointOpts: opts.New(ctx),
	}

	cmd := &cmd.Command{
		Name:      "delete",
		ShortHelp: "Delete an existing action configuration.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp waypoint action-config delete" }}
		command deletes an existing action configuration. This will remove the
		config completely from HCP Waypoint.
		`),
		RunF: func(c *cmd.Command, args []string) error {
			return deleteActionConfig(c, args, opts)
		},
		PersistentPreRun: func(c *cmd.Command, args []string) error {
			return cmd.RequireOrgAndProject(ctx)
		},
		Flags: cmd.Flags{
			Local: []*cmd.Flag{
				{
					Name:        "name",
					Shorthand:   "n",
					Description: "The name of the action configuration to delete.",
					Value:       flagvalue.Simple("", &opts.Name),
					Required:    true,
				},
			},
		},
	}

	return cmd
}

func deleteActionConfig(c *cmd.Command, args []string, opts *DeleteOpts) error {
	ns, err := opts.Namespace()
	if err != nil {
		return err
	}

	// Make action name a string pointer
	actionName := &opts.Name
	_, err = opts.WS.WaypointServiceDeleteActionConfig(&waypoint_service.WaypointServiceDeleteActionConfigParams{
		NamespaceID: ns.ID,
		Context:     opts.Ctx,
		ActionName:  actionName,
	}, nil)
	if err != nil {
		return fmt.Errorf("failed to delete action config %q: %w", opts.Name, err)
	}

	fmt.Fprintf(opts.IO.Err(), "Action config %q deleted.", opts.Name)
	return nil
}
