// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package definitions

import (
	"strings"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2024-11-22/client/waypoint_service"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/flagvalue"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/pkg/errors"
)

func NewCmdAddOnDefinitionRead(ctx *cmd.Context, opts *AddOnDefinitionOpts) *cmd.Command {
	cmd := &cmd.Command{
		Name:      "read",
		ShortHelp: "Read an HCP Waypoint add-on definition.",
		LongHelp: heredoc.New(ctx.IO).Must(`
The {{ template "mdCodeOrBold" "hcp waypoint add-ons definitions read" }}
command lets you read an existing HCP Waypoint add-on definition.
`),
		Examples: []cmd.Example{
			{
				Preamble: "Read an HCP Waypoint add-on definition:",
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
$ hcp waypoint add-ons definitions read -n=my-addon-definition
`),
			},
		},
		RunF: func(c *cmd.Command, args []string) error {
			if opts.testFunc != nil {
				return opts.testFunc(c, args)
			}
			return addOnDefinitionRead(opts)
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
					Description:  "The name of the add-on definition.",
					Value:        flagvalue.Simple("", &opts.Name),
					Required:     true,
				},
			},
		},
	}
	return cmd
}

func addOnDefinitionRead(opts *AddOnDefinitionOpts) error {
	getResp, err := opts.WS2024Client.WaypointServiceGetAddOnDefinition2(
		&waypoint_service.WaypointServiceGetAddOnDefinition2Params{
			NamespaceLocationOrganizationID: opts.Profile.OrganizationID,
			NamespaceLocationProjectID:      opts.Profile.ProjectID,
			Context:                         opts.Ctx,
			AddOnDefinitionName:             opts.Name,
		}, nil,
	)
	if err != nil {
		return errors.Wrapf(err, "%s failed to get add-on definition %q",
			opts.IO.ColorScheme().FailureIcon(),
			opts.Name,
		)
	}

	getRespPayload := getResp.GetPayload()
	if getRespPayload.AddOnDefinition == nil {
		return errors.Errorf(
			"%s empty add-on definition returned for name %q",
			opts.IO.ColorScheme().FailureIcon(),
			opts.Name,
		)
	}
	addOnDef := getRespPayload.AddOnDefinition
	var optionNames []string
	for _, option := range addOnDef.VariableOptions {
		optionNames = append(optionNames, option.Name)
	}
	optionNamesStr := strings.Join(optionNames, ", ")
	fields := []format.Field{
		format.NewField("ID", "{{ .ID }}"),
		format.NewField("Name", "{{ .Name }}"),
		format.NewField("Summary", "{{ .Summary }}"),
		format.NewField("Description", "{{ .Description }}"),
		format.NewField("Labels", "{{ .Labels }}"),
		format.NewField("Terraform Cloud Workspace Details", "{{ .TerraformCloudWorkspaceDetails }}"),
		format.NewField("Module Source", "{{ .ModuleSource }}"),
		format.NewField("Execution Mode", "{{ .TfExecutionMode }}"),
		format.NewField("Agent Pool ID", "{{ .TfAgentPoolID }}"),
		format.NewField("Variable Options", optionNamesStr),
		format.NewField("Terraform No-code Module ID", "{{ .ModuleID }}"),
	}

	return opts.Output.Display(format.NewDisplayer(addOnDef, format.Pretty, fields))
}
