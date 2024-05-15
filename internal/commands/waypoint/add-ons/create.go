// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package addons

import (
	"fmt"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2023-08-18/client/waypoint_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2023-08-18/models"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/flagvalue"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/pkg/errors"
)

func NewCmdCreate(ctx *cmd.Context, opts *AddOnOpts) *cmd.Command {
	cmd := &cmd.Command{
		Name:      "create",
		ShortHelp: "Create a new HCP Waypoint add-on.",
		LongHelp: heredoc.New(ctx.IO).Must(`
			The {{ template "mdCodeOrBold" "hcp waypoint add-ons create" }} command creates a new HCP Waypoint add-on.
		`),
		Examples: []cmd.Example{
			{
				Preamble: "Create a new HCP Waypoint add-on:",
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
$ hcp waypoint add-ons create -n=my-addon -a=my-application -d=my-addon-definition
`),
			},
		},
		AdditionalDocs: nil,
		PersistentPreRun: func(c *cmd.Command, args []string) error {
			return cmd.RequireOrgAndProject(ctx)
		},
		RunF: func(c *cmd.Command, args []string) error {
			if opts.testFunc != nil {
				return opts.testFunc(c, args)
			}
			return addOnCreate(opts)
		},
		Flags: cmd.Flags{
			Local: []*cmd.Flag{
				{
					Name:         "name",
					Shorthand:    "n",
					DisplayValue: "NAME",
					Description: "The name of the add-on. If no name is provided," +
						" a name will be generated.",
					Value:    flagvalue.Simple("", &opts.Name),
					Required: false,
				},
				{
					Name:         "add-on-definition-name",
					DisplayValue: "NAME",
					Description:  "The name of the add-on definition to use.",
					Value:        flagvalue.Simple("", &opts.AddOnDefinitionName),
					Required:     true,
				},
				{
					Name:         "app",
					DisplayValue: "APP-NAME",
					Description:  "The name of the application to which the add-on will be added.",
					Value:        flagvalue.Simple("", &opts.ApplicationName),
					Required:     true,
				},
			},
		},
	}

	return cmd
}

func addOnCreate(opts *AddOnOpts) error {
	ns, err := opts.Namespace()
	if err != nil {
		return err
	}

	_, err = opts.WS.WaypointServiceCreateAddOn(
		&waypoint_service.WaypointServiceCreateAddOnParams{
			NamespaceID: ns.ID,
			Context:     opts.Ctx,
			Body: &models.HashicorpCloudWaypointWaypointServiceCreateAddOnBody{
				Application: &models.HashicorpCloudWaypointRefApplication{
					Name: opts.ApplicationName,
				},
				Definition: &models.HashicorpCloudWaypointRefAddOnDefinition{
					Name: opts.AddOnDefinitionName,
				},
				Name: opts.Name,
			},
		}, nil)
	if err != nil {
		return errors.Wrapf(err, "%s failed to create add-on %q",
			opts.IO.ColorScheme().FailureIcon(),
			opts.Name)
	}

	fmt.Fprintf(opts.IO.Err(), "%s Add-on %q created!\n", opts.IO.ColorScheme().SuccessIcon(), opts.Name)

	return nil
}
