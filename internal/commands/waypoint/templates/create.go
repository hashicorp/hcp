// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package templates

import (
	"fmt"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2023-08-18/client/waypoint_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2023-08-18/models"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/flagvalue"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/pkg/errors"
)

func NewCmdCreate(ctx *cmd.Context, opts *TemplateOpts) *cmd.Command {
	cmd := &cmd.Command{
		Name:      "create",
		ShortHelp: "Create a new HCP Waypoint template.",
		LongHelp: heredoc.New(ctx.IO).Must(`
The {{ template "mdCodeOrBold" "hcp waypoint templates create" }} command lets you create
HCP Waypoint templates.
		`),
		Examples: []cmd.Example{
			{
				Preamble: "Create a new HCP Waypoint template:",
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
$ hcp waypoint templates create -n=my-template \
  -s="My Template Summary" \
  -d="My Template Description" \
  --readme-markdown-template-file "README.tpl" \
  --tfc-no-code-module-source="app.terraform.io/hashicorp/dir/template" \
  --tfc-project-name="my-tfc-project" \
  --tfc-project-id="prj-123456" \
  -l="label1" \
  -l="label2"
`),
			},
		},
		RunF: func(c *cmd.Command, args []string) error {
			if opts.testFunc != nil {
				return opts.testFunc(c, args)
			}
			return templateCreate(opts)
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
				{
					Name:         "summary",
					Shorthand:    "s",
					DisplayValue: "SUMMARY",
					Description:  "The summary of the template.",
					Value:        flagvalue.Simple("", &opts.Summary),
					Required:     true,
				},
				{
					Name:         "description",
					Shorthand:    "d",
					DisplayValue: "DESCRIPTION",
					Description:  "The description of the template.",
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
					Description:  "A label to apply to the template.",
					Repeatable:   true,
					Value:        flagvalue.SimpleSlice(nil, &opts.Labels),
				},
				{
					Name:         "tag",
					Shorthand:    "t",
					DisplayValue: "KEY=VALUE",
					Description:  "A tag to apply to the template.",
					Repeatable:   true,
					Value:        flagvalue.SimpleMap(nil, &opts.Tags),
					Hidden:       true,
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
						" applications using this template will be created.",
					Value:    flagvalue.Simple("", &opts.TerraformCloudProjectName),
					Required: true,
				},
				{
					Name:         "tfc-project-id",
					DisplayValue: "TFC_PROJECT_ID",
					Description: "The ID of the HCP Terraform project where" +
						" applications using this template will be created.",
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
						"workspaces for applications using this template.",
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

func templateCreate(opts *TemplateOpts) error {
	ns, err := opts.Namespace()
	if err != nil {
		return err
	}

	var tags []*models.HashicorpCloudWaypointTag
	for k, v := range opts.Tags {
		tags = append(tags, &models.HashicorpCloudWaypointTag{
			Key:   k,
			Value: v,
		})
	}

	var readmeTpl []byte
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
	var variables []*models.HashicorpCloudWaypointTFModuleVariable
	if opts.VariableOptionsFile != "" {
		variables, err = parseVariableOptionsFile(opts.VariableOptionsFile)
		if err != nil {
			return errors.Wrapf(err, "%s failed to read Variable Options hcl file %q",
				opts.IO.ColorScheme().FailureIcon(),
				opts.VariableOptionsFile,
			)
		}
	}

	_, err = opts.WS.WaypointServiceCreateApplicationTemplate(
		&waypoint_service.WaypointServiceCreateApplicationTemplateParams{
			NamespaceID: ns.ID,
			Body: &models.HashicorpCloudWaypointWaypointServiceCreateApplicationTemplateBody{
				ApplicationTemplate: &models.HashicorpCloudWaypointApplicationTemplate{
					Name:                   opts.Name,
					Summary:                opts.Summary,
					Description:            opts.Description,
					ReadmeMarkdownTemplate: readmeTpl,
					Labels:                 opts.Labels,
					Tags:                   tags,
					TerraformCloudWorkspaceDetails: &models.HashicorpCloudWaypointTerraformCloudWorkspaceDetails{
						Name:      opts.TerraformCloudProjectName,
						ProjectID: opts.TerraformCloudProjectID,
					},
					ModuleSource:    opts.TerraformNoCodeModuleSource,
					VariableOptions: variables,
					TfExecutionMode: opts.TerraformExecutionMode,
					TfAgentPoolID:   opts.TerraformAgentPoolID,
				},
			},
			Context: opts.Ctx,
		}, nil)
	if err != nil {
		return errors.Wrapf(
			err,
			"%s failed to create template %q",
			opts.IO.ColorScheme().FailureIcon(),
			opts.Name,
		)
	}

	fmt.Fprintf(opts.IO.Err(), "%s Template %q created.\n",
		opts.IO.ColorScheme().SuccessIcon(),
		opts.Name,
	)

	return nil
}

// parseVariableOptions reads the input bytes and parses the HCL file to extract
// the variable options. Note that we intentionally do not provide much in terms
// of validation of the HCL file, as we expect the HCL to be validated by the
// server side.
//
// # Example contents of a vars.hcl file
//
//	variable_option "string_variable" {
//	  type = "string"
//	  options = [
//	    "a string value",
//	  ]
//	  user_editable = false
//	}
//
//	variable_option "misc_variable" {
//	  type = "string"
//	  options = [
//	    "another string value",
//	  ]
//	  user_editable = false
//	}
func parseVariableOptions(filename string, input []byte) ([]*models.HashicorpCloudWaypointTFModuleVariable, error) {
	var hc hclVariableOptionsFile
	var ctx hcl.EvalContext
	// the Decode method expects a filename to provide context to the error; it
	// does not actually load anything from the file system
	if err := hclsimple.Decode(filename, input, &ctx, &hc); err != nil {
		return nil, err
	}

	// var variables []*models.HashicorpCloudWaypointTFModuleVariable
	variables := make([]*models.HashicorpCloudWaypointTFModuleVariable, 0)
	if len(hc.VariableOptions) > 0 {
		for _, v := range hc.VariableOptions {
			variables = append(variables, &models.HashicorpCloudWaypointTFModuleVariable{
				Name:         v.Name,
				VariableType: v.Type,
				Options:      v.Options,
				UserEditable: v.UserEditable,
			})
		}
	}
	return variables, nil
}

func parseVariableOptionsFile(path string) ([]*models.HashicorpCloudWaypointTFModuleVariable, error) {
	input, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return parseVariableOptions(path, input)
}

type hclVariableOption struct {
	Name         string   `hcl:",label"`
	Type         string   `hcl:"type"`
	Options      []string `hcl:"options"`
	UserEditable bool     `hcl:"user_editable,optional"`
}

type hclVariableOptionsFile struct {
	VariableOptions []*hclVariableOption `hcl:"variable_option,block"`
}
