// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package definitions

import (
	"fmt"
	"os"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2024-11-22/client/waypoint_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2024-11-22/models"
	"github.com/hashicorp/hcp/internal/commands/waypoint/internal"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/flagvalue"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/pkg/errors"
)

func NewCmdAddOnDefinitionUpdate(ctx *cmd.Context, opts *AddOnDefinitionOpts) *cmd.Command {
	cmd := &cmd.Command{
		Name:      "update",
		ShortHelp: "Update a HCP Waypoint add-on definition.",
		LongHelp: heredoc.New(ctx.IO).Must(`
The {{ template "mdCodeOrBold" "hcp waypoint add-ons definitions update" }}
command lets you update an existing HCP Waypoint add-on definition.
`),
		Examples: []cmd.Example{
			{
				Preamble: "Update a HCP Waypoint add-on definition:",
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
$ hcp waypoint add-ons definitions update -n=my-add-on-definition \
  -s="My updated Add-on Definition summary." \
  -d="My updated Add-on Definition description." \
  --readme-markdown-template-file "README.tpl" \
  --tfc-project-name="my-tfc-project" \
  --tfc-project-id="prj-123456" \
  -l=label1 \
  -l=label2
`),
			},
		},
		RunF: func(c *cmd.Command, args []string) error {
			if opts.testFunc != nil {
				return opts.testFunc(c, args)
			}
			return addOnDefinitionUpdate(opts)
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
				{
					Name:         "summary",
					Shorthand:    "s",
					DisplayValue: "SUMMARY",
					Description:  "The summary of the add-on definition.",
					Value:        flagvalue.Simple("", &opts.Summary),
				},
				{
					Name:         "description",
					Shorthand:    "d",
					DisplayValue: "DESCRIPTION",
					Description:  "The description of the add-on definition.",
					Value:        flagvalue.Simple("", &opts.Description),
				},
				{
					Name:         "readme-markdown-template-file",
					DisplayValue: "README_MARKDOWN_TEMPLATE_FILE",
					Description:  "The README markdown template file.",
					Value:        flagvalue.Simple("", &opts.ReadmeMarkdownTemplateFile),
				},
				{
					Name:         "label",
					Shorthand:    "l",
					DisplayValue: "LABEL",
					Description:  "A label to apply to the add-on definition.",
					Repeatable:   true,
					Value:        flagvalue.SimpleSlice(nil, &opts.Labels),
				},
				{
					Name:         "tfc-project-id",
					DisplayValue: "TFC_PROJECT_ID",
					Description:  "The Terraform Cloud project ID.",
					Value:        flagvalue.Simple("", &opts.TerraformCloudProjectID),
				},
				{
					Name:         "variable-options-file",
					DisplayValue: "VARIABLE_OPTIONS_FILE",
					Description:  "The file containing the HCL definition of Variable Options.",
					Value:        flagvalue.Simple("", &opts.VariableOptionsFile),
				},
				{
					Name:         "tfc-project-name",
					DisplayValue: "TFC_PROJECT_NAME",
					Description:  "The Terraform Cloud project name.",
					Value:        flagvalue.Simple("", &opts.TerraformCloudProjectName),
				},
				{
					Name:         "tf-execution-mode",
					DisplayValue: "TF_EXECUTION_MODE",
					Description: "The execution mode of the HCP Terraform " +
						"workspaces for add-ons using this add-on definition.",
					Value: flagvalue.Simple("remote", &opts.TerraformExecutionMode),
				},
				{
					Name:         "tf-agent-pool-id",
					DisplayValue: "TF_AGENT_POOL_ID",
					Description: "The ID of the Terraform agent pool to use for " +
						"running Terraform operations. This is only applicable " +
						"when the execution mode is set to 'agent'.",
					Value: flagvalue.Simple("", &opts.TerraformAgentPoolID),
				},
			},
		},
	}

	return cmd
}

func addOnDefinitionUpdate(opts *AddOnDefinitionOpts) error {
	var (
		readmeTpl []byte
		err       error
	)
	if opts.ReadmeMarkdownTemplateFile != "" {
		readmeTpl, err = os.ReadFile(opts.ReadmeMarkdownTemplateFile)
		if err != nil {
			return errors.Wrapf(err, "%s failed to read README markdown template file %q",
				opts.IO.ColorScheme().FailureIcon(),
				opts.ReadmeMarkdownTemplateFile,
			)
		}
	}

	// read variable options file and parse hcl
	var variables []*models.HashicorpCloudWaypointV20241122TFModuleVariable
	if opts.VariableOptionsFile != "" {
		vars, err := internal.ParseVariableOptionsFile(opts.VariableOptionsFile)
		if err != nil {
			return errors.Wrapf(err, "%s failed to read Variable Options hcl file %q",
				opts.IO.ColorScheme().FailureIcon(),
				opts.VariableOptionsFile,
			)
		}
		variables = vars
	}

	_, err = opts.WS2024Client.WaypointServiceUpdateAddOnDefinition2(
		&waypoint_service.WaypointServiceUpdateAddOnDefinition2Params{
			NamespaceLocationOrganizationID: opts.Profile.OrganizationID,
			NamespaceLocationProjectID:      opts.Profile.ProjectID,
			Context:                         opts.Ctx,
			ExistingAddOnDefinitionName:     opts.Name,
			Body: &models.HashicorpCloudWaypointV20241122WaypointServiceUpdateAddOnDefinitionBody{
				AddOnDefinition: &models.HashicorpCloudWaypointV20241122AddOnDefinition{
					Summary:                opts.Summary,
					Description:            opts.Description,
					ReadmeMarkdownTemplate: readmeTpl,
					Labels:                 opts.Labels,
					TerraformCloudWorkspaceDetails: &models.HashicorpCloudWaypointV20241122TerraformCloudWorkspaceDetails{
						ProjectID: opts.TerraformCloudProjectID,
						Name:      opts.TerraformCloudProjectName,
					},
					TfExecutionMode: opts.TerraformExecutionMode,
					TfAgentPoolID:   opts.TerraformAgentPoolID,
					VariableOptions: variables,
				},
			},
		}, nil,
	)
	if err != nil {
		return errors.Wrapf(err, "%s failed to update add-on definition %q",
			opts.IO.ColorScheme().FailureIcon(),
			opts.Name,
		)
	}

	fmt.Fprintf(opts.IO.Err(), "%s Add-on definition %q updated.\n",
		opts.IO.ColorScheme().SuccessIcon(),
		opts.Name)

	return nil
}
