// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: MPL-2.0

package actions

import (
	"fmt"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2024-11-22/client/waypoint_service"

	"github.com/hashicorp/hcp/internal/commands/waypoint/opts"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/flagvalue"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
)

type ReadOpts struct {
	opts.WaypointOpts

	Name string
}

func NewCmdRead(ctx *cmd.Context) *cmd.Command {
	opts := &ReadOpts{
		WaypointOpts: opts.New(ctx),
	}

	cmd := &cmd.Command{
		Name:      "read",
		ShortHelp: "Read more details about an action.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp waypoint actions read" }}
		command returns more details about an action configurations.
		`),
		RunF: func(c *cmd.Command, args []string) error {
			return readAction(c, args, opts)
		},
		PersistentPreRun: func(c *cmd.Command, args []string) error {
			return cmd.RequireOrgAndProject(ctx)
		},
		Flags: cmd.Flags{
			Local: []*cmd.Flag{
				{
					Name:        "name",
					Shorthand:   "n",
					Description: "The name of the action.",
					Value:       flagvalue.Simple("", &opts.Name),
					Required:    true,
				},
			},
		},
	}

	return cmd
}

func readAction(c *cmd.Command, args []string, opts *ReadOpts) error {
	// Make action name a string pointer
	actionName := &opts.Name
	resp, err := opts.WS2024Client.WaypointServiceGetActionConfig(&waypoint_service.WaypointServiceGetActionConfigParams{
		NamespaceLocationOrganizationID: opts.Profile.OrganizationID,
		NamespaceLocationProjectID:      opts.Profile.ProjectID,
		Context:                         opts.Ctx,
		ActionName:                      actionName,
	}, nil)
	if err != nil {
		return fmt.Errorf("error getting action for %q: %w",
			opts.Name, err)
	}

	respPayload := resp.GetPayload()
	return opts.Output.Show(respPayload.ActionConfig, format.Pretty)
}
