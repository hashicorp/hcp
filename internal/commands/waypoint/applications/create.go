// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package applications

import (
	"fmt"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2023-08-18/client/waypoint_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2023-08-18/models"
	"github.com/hashicorp/hcp/internal/commands/waypoint/internal"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/flagvalue"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/pkg/errors"
)

func NewCmdApplicationsCreate(ctx *cmd.Context, opts *ApplicationOpts) *cmd.Command {
	cmd := &cmd.Command{
		Name:      "create",
		ShortHelp: "Create a new HCP Waypoint application.",
		LongHelp: heredoc.New(ctx.IO).Must(`
The {{ template "mdCodeOrBold" "hcp waypoint applications create" }} command lets you create
a new HCP Waypoint application.
		`),
		Examples: []cmd.Example{
			{
				Preamble: "Create a new HCP Waypoint application:",
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
$ hcp waypoint application create -n=my-application -t=my-template
`),
			},
		},
		RunF: func(c *cmd.Command, args []string) error {
			if opts.testFunc != nil {
				return opts.testFunc(c, args)
			}
			return applicationCreate(opts)
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
					Description:  "The name of the application.",
					Value:        flagvalue.Simple("", &opts.Name),
					Required:     true,
				},
				{
					Name:         "template-name",
					Shorthand:    "t",
					DisplayValue: "TEMPLATE_NAME",
					Description:  "The name of the template to use for the application.",
					Value:        flagvalue.Simple("", &opts.TemplateName),
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

	return cmd
}

func applicationCreate(opts *ApplicationOpts) error {
	ns, err := opts.Namespace()
	if err != nil {
		return err
	}

	actionConfigs := make([]*models.HashicorpCloudWaypointActionCfgRef, len(opts.ActionConfigNames))
	for i, name := range opts.ActionConfigNames {
		actionConfigs[i] = &models.HashicorpCloudWaypointActionCfgRef{
			Name: name,
		}
	}

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

	_, err = opts.WS.WaypointServiceCreateApplicationFromTemplate(
		&waypoint_service.WaypointServiceCreateApplicationFromTemplateParams{
			NamespaceID: ns.ID,
			Context:     opts.Ctx,
			Body: &models.HashicorpCloudWaypointWaypointServiceCreateApplicationFromTemplateBody{
				Name: opts.Name,
				ApplicationTemplate: &models.HashicorpCloudWaypointRefApplicationTemplate{
					Name: opts.TemplateName,
				},
				ActionCfgRefs: actionConfigs,
				Variables:     vars,
			},
		}, nil)
	if err != nil {
		return errors.Wrapf(err, "%s failed to create application %q",
			opts.IO.ColorScheme().FailureIcon(),
			opts.Name,
		)
	}

	fmt.Fprintf(opts.IO.Err(), "%s Application %q created.\n",
		opts.IO.ColorScheme().SuccessIcon(),
		opts.Name,
	)

	return nil
}
