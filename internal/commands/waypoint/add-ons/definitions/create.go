// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package definitions

import (
	"fmt"
	"os"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2023-08-18/client/waypoint_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2023-08-18/models"
	"github.com/hashicorp/hcp/internal/commands/waypoint/internal"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/flagvalue"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/pkg/errors"
)

func NewCmdAddOnDefinitionCreate(ctx *cmd.Context, opts *AddOnDefinitionOpts) *cmd.Command {
	cmd := &cmd.Command{
		Name:      "create",
		ShortHelp: "Create a new HCP Waypoint add-on definition.",
		LongHelp: heredoc.New(ctx.IO).Must(`
The {{ template "mdCodeOrBold" "hcp waypoint add-ons definitions create" }}
command lets you create HCP Waypoint add-on definitions.
`),
		Examples: []cmd.Example{
			{
				Preamble: "Create a new HCP Waypoint add-on definition:",
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
$ hcp waypoint add-ons definitions create -n=my-add-on-definition \
  -s="My Add-on Definition summary." \
  -d="My Add-on Definition description." \
  --readme-markdown-template-file="README.tpl" \
  --tfc-no-code-module-source="app.terraform.io/hashicorp/dir/template" \
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
			return addOnDefinitionCreate(opts)
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
					DisplayValue: "README_MARKDOWN_TEMPLATE_FILE_PATH",
					Description:  "The file containing the README markdown template.",
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
					Name:         "tfc-no-code-module-source",
					DisplayValue: "TFC_NO_CODE_MODULE_SOURCE",
					Description: heredoc.New(ctx.IO).Must(`
			The source of the Terraform no-code module. 
			The expected format is "NAMESPACE/NAME/PROVIDER". An
			optional "HOSTNAME/" can be added at the beginning for
			a private registry.
					`),
					Value:    flagvalue.Simple("", &opts.TerraformNoCodeModuleSource),
					Required: true,
				},
				{
					Name:         "tfc-project-name",
					DisplayValue: "TFC_PROJECT_NAME",
					Description: "The name of the Terraform Cloud project where" +
						" applications using this add-on definition will be created.",
					Value:    flagvalue.Simple("", &opts.TerraformCloudProjectName),
					Required: true,
				},
				{
					Name:         "tfc-project-id",
					DisplayValue: "TFC_PROJECT_ID",
					Description: "The ID of the Terraform Cloud project where" +
						" applications using this add-on definition will be created.",
					Value:    flagvalue.Simple("", &opts.TerraformCloudProjectID),
					Required: true,
				},
				{
					Name:         "variable-options-file",
					DisplayValue: "VARIABLE_OPTIONS_FILE",
					Description:  "The file containing the HCL definition of Variable Options.",
					Value:        flagvalue.Simple("", &opts.VariableOptionsFile),
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

func addOnDefinitionCreate(opts *AddOnDefinitionOpts) error {
	ns, err := opts.Namespace()
	if err != nil {
		return err
	}

	var readmeTpl []byte
	if opts.ReadmeMarkdownTemplateFile != "" {
		readmeTpl, err = os.ReadFile(opts.ReadmeMarkdownTemplateFile)
		if err != nil {
			return errors.Wrapf(err, "%s failed to read README markdown template file %q",
				opts.IO.ColorScheme().FailureIcon(),
				opts.ReadmeMarkdownTemplateFile)
		}
	}

	// read variable options file and parse hcl
	var variables []*models.HashicorpCloudWaypointTFModuleVariable
	if opts.VariableOptionsFile != "" {
		variables, err = internal.ParseVariableOptionsFile(opts.VariableOptionsFile)
		if err != nil {
			return errors.Wrapf(err, "%s failed to read Variable Options hcl file %q",
				opts.IO.ColorScheme().FailureIcon(),
				opts.VariableOptionsFile,
			)
		}
	}

	_, err = opts.WS.WaypointServiceCreateAddOnDefinition(
		&waypoint_service.WaypointServiceCreateAddOnDefinitionParams{
			NamespaceID: ns.ID,
			Body: &models.HashicorpCloudWaypointWaypointServiceCreateAddOnDefinitionBody{
				Name:                   opts.Name,
				Summary:                opts.Summary,
				Description:            opts.Description,
				ReadmeMarkdownTemplate: readmeTpl,
				Labels:                 opts.Labels,
				TerraformCloudWorkspaceDetails: &models.HashicorpCloudWaypointTerraformCloudWorkspaceDetails{
					Name:      opts.TerraformCloudProjectName,
					ProjectID: opts.TerraformCloudProjectID,
				},
				ModuleSource:    opts.TerraformNoCodeModuleSource,
				VariableOptions: variables,
			},
			Context: opts.Ctx,
		}, nil)
	if err != nil {
		return errors.Wrapf(err, "%s failed to create add-on definition %q",
			opts.IO.ColorScheme().FailureIcon(),
			opts.Name,
		)
	}

	fmt.Fprintf(opts.IO.Err(), "%s Add-on definition %q created.\n",
		opts.IO.ColorScheme().SuccessIcon(),
		opts.Name,
	)

	return nil
}
