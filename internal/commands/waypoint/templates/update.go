// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package templates

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

func NewCmdUpdate(ctx *cmd.Context, opts *TemplateOpts) *cmd.Command {
	cmd := &cmd.Command{
		Name:      "update",
		ShortHelp: "Update an existing HCP Waypoint template.",
		LongHelp: heredoc.New(ctx.IO).Must(`
The {{ template "mdCodeOrBold" "hcp waypoint templates update" }} command lets you update
existing HCP Waypoint templates.
		`),
		Examples: []cmd.Example{
			{
				Preamble: "Update an HCP Waypoint template:",
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
$ hcp waypoint templates update -n=my-template \
  -s="My Template Summary" \
  -d="My Template Description" \
  -readme-markdown-template-file "README.tpl" \
  --tfc-project-name="my-tfc-project" \
  --tfc-project-id="prj-123456 \
  -l="label1" \
  -l="label2"
`),
			},
		},
		RunF: func(c *cmd.Command, args []string) error {
			if opts.testFunc != nil {
				return opts.testFunc(c, args)
			}
			return templateUpdate(opts)
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
					Description:  "The name of the template to be updated.",
					Value:        flagvalue.Simple("", &opts.Name),
					Required:     true,
				},
				{
					Name:         "new-name",
					DisplayValue: "NEW_NAME",
					Description:  "The new name of the template.",
					Value:        flagvalue.Simple("", &opts.UpdatedName),
					Hidden:       true,
				},
				{
					Name:         "summary",
					Shorthand:    "s",
					DisplayValue: "SUMMARY",
					Description:  "The summary of the template.",
					Value:        flagvalue.Simple("", &opts.Summary),
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
					Name:         "tfc-project-name",
					DisplayValue: "TFC_PROJECT_NAME",
					Description: "The name of the Terraform Cloud project where" +
						" applications using this template will be created.",
					Value: flagvalue.Simple("", &opts.TerraformCloudProjectName),
				},
				{
					Name:         "tfc-project-id",
					DisplayValue: "TFC_PROJECT_ID",
					Description: "The ID of the Terraform Cloud project where" +
						" applications using this template will be created.",
					Value: flagvalue.Simple("", &opts.TerraformCloudProjectID),
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

func templateUpdate(opts *TemplateOpts) error {
	ns, err := opts.Namespace()
	if err != nil {
		return err
	}

	// NOTE (clint): We have to get the existing template to get the collections
	// (labels, tags, actions, variables) so that we can either omit or post
	// updates to them based on the inputs here. This is because of how the
	// model is generated from swagger with collections NOT being omit empty. As
	// a result, the fieldmask is getting set even if nothing is set, and our
	// API is then removing the collection (variables, et.al) even if we don't
	// set it here in the request (ex: we don't intend to change things).
	//
	// Unfortunately even if the model had omitempty, this would then cause the
	// fieldmask to not get set, which is needed for the PATCH semantics, thus
	// no change would occur.
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

	// start our updated template with the existing template
	updatedTpl := &models.HashicorpCloudWaypointApplicationTemplate{
		Name:        opts.UpdatedName,
		Summary:     opts.Summary,
		Description: opts.Description,
		// Labels:   opts.Labels
		TerraformCloudWorkspaceDetails: &models.HashicorpCloudWaypointTerraformCloudWorkspaceDetails{
			Name:      opts.TerraformCloudProjectName,
			ProjectID: opts.TerraformCloudProjectID,
		},
		TfExecutionMode: opts.TerraformExecutionMode,
		TfAgentPoolID:   opts.TerraformAgentPoolID,
	}

	// grab the existing collection things
	updatedTpl.Labels = template.Labels
	// grab the existing template options
	updatedTpl.VariableOptions = template.VariableOptions

	// the HCP CLI doesn't support working with Action Cfgs yet, so we just copy
	// over
	updatedTpl.ActionCfgRefs = template.ActionCfgRefs

	// var tags []*models.HashicorpCloudWaypointTag
	tags := template.Tags
	if len(opts.Tags) > 0 {
		// clear out the existing tags to replace with new ones
		tags = []*models.HashicorpCloudWaypointTag{}
		for k, v := range opts.Tags {
			tags = append(tags, &models.HashicorpCloudWaypointTag{
				Key:   k,
				Value: v,
			})
		}
	}
	updatedTpl.Tags = tags

	if len(opts.Labels) > 0 {
		updatedTpl.Labels = opts.Labels
	}

	var readmeTpl []byte
	if opts.ReadmeMarkdownTemplateFile != "" {
		readmeTpl, err = os.ReadFile(opts.ReadmeMarkdownTemplateFile)
		if err != nil {
			return errors.Wrapf(err, "%s failed to read README template file %q",
				opts.IO.ColorScheme().FailureIcon(),
				opts.ReadmeMarkdownTemplateFile,
			)
		}
	}
	updatedTpl.ReadmeMarkdownTemplate = readmeTpl

	// read variable options file and parse hcl
	var variables []*models.HashicorpCloudWaypointTFModuleVariable
	if opts.VariableOptionsFile != "" {
		vars, err := internal.ParseVariableOptionsFile(opts.VariableOptionsFile)
		if err != nil {
			return errors.Wrapf(err, "%s failed to read Variable Options hcl file %q",
				opts.IO.ColorScheme().FailureIcon(),
				opts.VariableOptionsFile,
			)
		}
		variables = vars
		updatedTpl.VariableOptions = variables
	}

	// if a variable file is present but represents an empty list of variables,
	// we need to set the fieldmask for variables to clear them out

	_, err = opts.WS.WaypointServiceUpdateApplicationTemplate6(
		&waypoint_service.WaypointServiceUpdateApplicationTemplate6Params{
			NamespaceID:                     ns.ID,
			Context:                         opts.Ctx,
			ExistingApplicationTemplateName: opts.Name,
			ApplicationTemplate:             updatedTpl,
			// FieldMask:                       &fieldMasks,
		}, nil,
	)
	if err != nil {
		return errors.Wrapf(err, "%s failed to update template %q",
			opts.IO.ColorScheme().FailureIcon(),
			opts.Name,
		)
	}

	fmt.Fprintf(opts.IO.Err(), "%s Template %q updated.\n",
		opts.IO.ColorScheme().SuccessIcon(),
		opts.ID,
	)

	return nil
}
