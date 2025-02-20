// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package addons

import (
	"fmt"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2024-11-22/client/waypoint_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2024-11-22/models"
	"github.com/hashicorp/hcp/internal/commands/waypoint/internal"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/flagvalue"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/pkg/errors"
)

func NewCmdCreate(ctx *cmd.Context, opts *AddOnOpts) *cmd.Command {
	c := &cmd.Command{
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
					DisplayValue: "NAME",
					Description:  "The name of the application to which the add-on will be added.",
					Value:        flagvalue.Simple("", &opts.ApplicationName),
					Required:     true,
				},
				{
					Name:         "var",
					DisplayValue: "KEY=VALUE",
					Description: "A variable to be used in the application. The" +
						" flag can be repeated to specify multiple variables. " +
						"Variables specified with the flag will override " +
						"variables specified in a file.",
					Value:      flagvalue.SimpleMap(nil, &opts.Variables),
					Required:   false,
					Repeatable: true,
				},
				{
					Name:         "var-file",
					DisplayValue: "FILE",
					Description: "A file containing variables to be used in the " +
						"application. The file should be in HCL format Variables" +
						" in the file will be overridden by variables specified" +
						" with the --var flag.",
					Value:    flagvalue.Simple("", &opts.VariablesFile),
					Required: false,
				},
			},
		},
	}

	return c
}

func addOnCreate(opts *AddOnOpts) error {
	// Variable Processing

	// a map is used with the key being the variable name, so that
	// flags can override file values.
	ivs := make(map[string]*models.HashicorpCloudWaypointInputVariable)
	if opts.VariablesFile != "" {
		variables, err := internal.ParseInputVariablesFile(opts.VariablesFile)
		if err != nil {
			return errors.Wrapf(err, "%s failed to parse input variables file %q",
				opts.IO.ColorScheme().FailureIcon(),
				opts.VariablesFile,
			)
		}
		for _, v := range variables {
			ivs[v.Name] = &models.HashicorpCloudWaypointInputVariable{
				Name:  v.Name,
				Value: v.Value,
			}
		}
	}

	// Flags are processed second, so that they can override file values.
	// Flags take precedence over file values.
	for k, v := range opts.Variables {
		ivs[k] = &models.HashicorpCloudWaypointInputVariable{
			Name:  k,
			Value: v,
		}
	}

	var vars []*models.HashicorpCloudWaypointInputVariable
	for _, v := range ivs {
		vars = append(vars, v)
	}

	// End Variable Processing

	_, err := opts.WS2024Client.WaypointServiceCreateAddOn(
		&waypoint_service.WaypointServiceCreateAddOnParams{
			NamespaceLocationOrganizationID: opts.Profile.OrganizationID,
			NamespaceLocationProjectID:      opts.Profile.ProjectID,
			Context:                         opts.Ctx,
			Body: &models.HashicorpCloudWaypointV20241122WaypointServiceCreateAddOnBody{
				Application: &models.HashicorpCloudWaypointRefApplication{
					Name: opts.ApplicationName,
				},
				Definition: &models.HashicorpCloudWaypointRefAddOnDefinition{
					Name: opts.AddOnDefinitionName,
				},
				Name:      opts.Name,
				Variables: vars,
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
