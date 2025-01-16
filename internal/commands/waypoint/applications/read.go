// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package applications

import (
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2024-11-22/client/waypoint_service"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/flagvalue"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/pkg/errors"
)

func NewCmdApplicationsRead(ctx *cmd.Context, opts *ApplicationOpts) *cmd.Command {
	c := &cmd.Command{
		Name:      "read",
		ShortHelp: "Read details about an HCP Waypoint application.",
		LongHelp: heredoc.New(ctx.IO).Must(`
The {{ template "mdCodeOrBold" "hcp waypoint applications read" }} command lets you read
details about an HCP Waypoint application.
`),
		Examples: []cmd.Example{
			{
				Preamble: "Read an HCP Waypoint application:",
				Command:  "$ hcp waypoint applications read -n=my-application",
			},
		},
		RunF: func(c *cmd.Command, args []string) error {
			if opts.testFunc != nil {
				return opts.testFunc(c, args)
			}
			return applicationRead(opts)
		},
		PersistentPreRun: func(c *cmd.Command, args []string) error {
			return cmd.RequireOrgAndProject(ctx)
		},
		Flags: cmd.Flags{
			Local: []*cmd.Flag{
				{
					Name:         "name",
					Shorthand:    "n",
					DisplayValue: "NAME",
					Description:  "The name of the HCP Waypoint application.",
					Value:        flagvalue.Simple("", &opts.Name),
					Required:     true,
				},
			},
		},
	}

	return c
}

func applicationRead(opts *ApplicationOpts) error {
	getResp, err := opts.WS2024Client.WaypointServiceGetApplication2(
		&waypoint_service.WaypointServiceGetApplication2Params{
			NamespaceLocationOrganizationID: opts.Profile.OrganizationID,
			NamespaceLocationProjectID:      opts.Profile.ProjectID,
			Context:                         opts.Ctx,
			ApplicationName:                 opts.Name,
		}, nil,
	)
	if err != nil {
		return errors.Wrapf(err, "%s failed to get application %q",
			opts.IO.ColorScheme().FailureIcon(),
			opts.Name,
		)
	}

	return opts.Output.Show(getResp.GetPayload().Application, format.Pretty)
}
