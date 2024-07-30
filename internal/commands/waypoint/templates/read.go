// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package templates

import (
	"strings"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2023-08-18/client/waypoint_service"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/flagvalue"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/pkg/errors"
)

func NewCmdRead(ctx *cmd.Context, opts *TemplateOpts) *cmd.Command {
	cmd := &cmd.Command{
		Name:      "read",
		ShortHelp: "Read more details about an HCP Waypoint template.",
		LongHelp: heredoc.New(ctx.IO).Must(`
The {{ template "mdCodeOrBold" "hcp waypoint templates read" }} command lets you read
an existing HCP Waypoint template.
		`),
		Examples: []cmd.Example{
			{
				Preamble: "Read an HCP Waypoint template:",
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
$ hcp waypoint templates read -n=my-template
`),
			},
		},
		RunF: func(c *cmd.Command, args []string) error {
			if opts.testFunc != nil {
				return opts.testFunc(c, args)
			}
			return templateRead(opts)
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
					Description:  "The name of the template.",
					Value:        flagvalue.Simple("", &opts.Name),
					Required:     true,
				},
			},
		},
	}

	return cmd
}

func templateRead(opts *TemplateOpts) error {
	ns, err := opts.Namespace()
	if err != nil {
		return err
	}

	resp, err := opts.WS.WaypointServiceGetApplicationTemplate2(
		&waypoint_service.WaypointServiceGetApplicationTemplate2Params{
			NamespaceID:             ns.ID,
			Context:                 opts.Ctx,
			ApplicationTemplateName: opts.Name,
		}, nil,
	)
	if err != nil {
		return errors.Wrapf(err, "%s failed to get template %q",
			opts.IO.ColorScheme().FailureIcon(),
			opts.Name,
		)
	}

	respPayload := resp.GetPayload()
	if respPayload.ApplicationTemplate == nil {
		return errors.Wrapf(err, "%s empty template returned for name %q",
			opts.IO.ColorScheme().FailureIcon(),
			opts.Name,
		)
	}
	template := respPayload.ApplicationTemplate

	// Create the fields. The fields allow setting the outputting name directly
	// and the value is a text/template which allows additional formatting.

	// for now we just flatten the Variable Option names
	var optionNames []string
	for _, variableOption := range template.VariableOptions {
		optionNames = append(optionNames, variableOption.Name)
	}
	optionNamesStr := strings.Join(optionNames, ", ")
	projectFields := []format.Field{
		format.NewField("ID", "{{ .ID }}"),
		format.NewField("Name", "{{ .Name }}"),
		format.NewField("Description", "{{ .Description }}"),
		format.NewField("Labels", "{{ .Labels }}"),
		format.NewField("Readme Template", "{{ .ReadmeTemplate }}"),
		format.NewField("Tags", "{{ .Tags}}"),
		format.NewField("Terraform Nocode Module ID", "{{ .TerraformNocodeModule.ModuleID}}"),
		format.NewField("Terraform Nocode Source", "{{ .TerraformNocodeModule.Source}}"),
		format.NewField("Variable Options", optionNamesStr),
	}

	// Display the created project
	return opts.Output.Display(format.NewDisplayer(template, format.Pretty, projectFields))
}
