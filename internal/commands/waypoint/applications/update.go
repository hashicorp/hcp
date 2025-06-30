// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package applications

import (
	"fmt"
	"os"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2024-11-22/client/waypoint_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2024-11-22/models"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/flagvalue"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/pkg/errors"
)

func NewCmdApplicationsUpdate(ctx *cmd.Context, opts *ApplicationOpts) *cmd.Command {
	c := &cmd.Command{
		Name:      "update",
		ShortHelp: "Update an existing HCP Waypoint application.",
		LongHelp: heredoc.New(ctx.IO).Must(`
The {{ template "mdCodeOrBold" "hcp waypoint applications update" }} command lets you update
an existing HCP Waypoint application.
		`),
		Examples: []cmd.Example{
			{
				Preamble: "Update an existing HCP Waypoint application:",
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
$ hcp waypoint applications update -n=my-application --action-config-name my-action-config
`),
			},
		},
		RunF: func(c *cmd.Command, args []string) error {
			if opts.testFunc != nil {
				return opts.testFunc(c, args)
			}
			return applicationUpdate(opts)
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
					Description:  "The name of the HCP Waypoint application to update.",
					Value:        flagvalue.Simple("", &opts.Name),
					Required:     true,
				},
				{
					Name:         "action-config-name",
					DisplayValue: "ACTION_CONFIG_NAME",
					Description:  "The name of the action configuration to be added to the application.",
					Value:        flagvalue.SimpleSlice(nil, &opts.ActionConfigNames),
					Required:     false,
					Repeatable:   true,
				},
				{
					Name:         "readme-markdown-file",
					DisplayValue: "README_MARKDOWN_FILE",
					Description:  "The path to the README markdown file to be used for the application.",
					Value:        flagvalue.Simple("", &opts.ReadmeMarkdownFile),
					Required:     false,
				},
			},
		},
	}

	return c
}

func applicationUpdate(opts *ApplicationOpts) error {
	var (
		acrs []*models.HashicorpCloudWaypointV20241122ActionCfgRef
		err  error
	)
	for _, acn := range opts.ActionConfigNames {
		acrs = append(acrs, &models.HashicorpCloudWaypointV20241122ActionCfgRef{
			Name: acn,
		})
	}

	var readme []byte
	if opts.ReadmeMarkdownFile != "" {
		readme, err = os.ReadFile(opts.ReadmeMarkdownFile)
		if err != nil {
			return errors.Wrapf(err,
				"%s failed to read README markdown file %q",
				opts.IO.ColorScheme().FailureIcon(),
				opts.ReadmeMarkdownFile)
		}
	}

	_, err = opts.WS2024Client.WaypointServiceUpdateApplication2(
		&waypoint_service.WaypointServiceUpdateApplication2Params{
			NamespaceLocationOrganizationID: opts.Profile.OrganizationID,
			NamespaceLocationProjectID:      opts.Profile.ProjectID,
			Context:                         opts.Ctx,
			ApplicationName:                 opts.Name,
			Body: &models.HashicorpCloudWaypointV20241122WaypointServiceUpdateApplicationBody{
				ActionCfgRefs:  acrs,
				ReadmeMarkdown: readme,
			},
		}, nil,
	)
	if err != nil {
		return errors.Wrapf(err, "%s failed to update application %q",
			opts.IO.ColorScheme().FailureIcon(),
			opts.Name,
		)
	}

	fmt.Fprintf(opts.IO.Err(), "%s Application %q updated.\n",
		opts.IO.ColorScheme().SuccessIcon(),
		opts.Name,
	)

	return nil
}
