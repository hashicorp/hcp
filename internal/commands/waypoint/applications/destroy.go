// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package applications

import (
	"fmt"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2024-11-22/client/waypoint_service"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/flagvalue"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/pkg/errors"
)

func NewCmdApplicationsDestroy(ctx *cmd.Context, opts *ApplicationOpts) *cmd.Command {
	c := &cmd.Command{
		Name:      "destroy",
		ShortHelp: "Destroy an HCP Waypoint application and its infrastructure.",
		LongHelp: heredoc.New(ctx.IO).Must(`
The {{ template "mdCodeOrBold" "hcp waypoint applications destroy" }} command lets you destroy
an HCP Waypoint application and its infrastructure.
		`),
		Examples: []cmd.Example{
			{
				Preamble: "Destroy an HCP Waypoint application:",
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
$ hcp waypoint applications destroy -n=my-application
`),
			},
		},
		RunF: func(c *cmd.Command, args []string) error {
			if opts.testFunc != nil {
				return opts.testFunc(c, args)
			}
			return applicationDestroy(opts)
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
					Description:  "The name of the application to destroy.",
					Value:        flagvalue.Simple("", &opts.Name),
					Required:     true,
				},
			},
		},
	}

	return c
}

func applicationDestroy(opts *ApplicationOpts) error {
	if opts.IO.CanPrompt() {
		ok, err := opts.IO.PromptConfirm(
			"The HCP Waypoint application will be deleted.\n\n" +
				"Do you want to continue")
		if err != nil {
			return errors.Wrapf(err, "%s failed to prompt for confirmation",
				opts.IO.ColorScheme().FailureIcon(),
			)
		}
		if !ok {
			return nil
		}
	}

	_, err := opts.WS2024Client.WaypointServiceDestroyApplication2(
		&waypoint_service.WaypointServiceDestroyApplication2Params{
			NamespaceLocationOrganizationID: opts.Profile.OrganizationID,
			NamespaceLocationProjectID:      opts.Profile.ProjectID,
			Context:                         opts.Ctx,
			ApplicationName:                 opts.Name,
		}, nil,
	)
	if err != nil {
		return errors.Wrapf(err, "%s failed to destroy application %q",
			opts.IO.ColorScheme().FailureIcon(),
			opts.Name,
		)
	}

	fmt.Fprintf(opts.IO.Err(), "%s Application %q destroyed.\n",
		opts.IO.ColorScheme().SuccessIcon(),
		opts.Name)

	return nil
}
