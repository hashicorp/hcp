// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: MPL-2.0

package applications

import (
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2024-11-22/client/waypoint_service"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/pkg/errors"
)

func NewCmdApplicationsList(ctx *cmd.Context, opts *ApplicationOpts) *cmd.Command {
	c := &cmd.Command{
		Name:      "list",
		ShortHelp: "List all HCP Waypoint applications.",
		LongHelp: heredoc.New(ctx.IO).Must(`
			The {{ template "mdCodeOrBold" "hcp waypoint applications list" }} command lists all
			HCP Waypoint applications.
		`),
		RunF: func(cmd *cmd.Command, args []string) error {
			if opts.testFunc != nil {
				return opts.testFunc(cmd, args)
			}
			return applicationsList(opts)
		},
	}

	return c
}

func applicationsList(opts *ApplicationOpts) error {
	resp, err := opts.WS2024Client.WaypointServiceListApplications(
		&waypoint_service.WaypointServiceListApplicationsParams{
			NamespaceLocationOrganizationID: opts.Profile.OrganizationID,
			NamespaceLocationProjectID:      opts.Profile.ProjectID,
			Context:                         opts.Ctx,
		}, nil,
	)
	if err != nil {
		return errors.Wrapf(err, "%s failed to list applications",
			opts.IO.ColorScheme().FailureIcon(),
		)
	}

	return opts.Output.Show(resp.GetPayload().Applications, format.Pretty)
}
