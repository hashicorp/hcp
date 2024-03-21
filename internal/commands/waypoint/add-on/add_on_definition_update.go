package addon

import (
	"fmt"
	"os"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2023-08-18/client/waypoint_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2023-08-18/models"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/flagvalue"
	"github.com/pkg/errors"
)

func NewCmdAddOnDefinitionUpdate(ctx *cmd.Context, opts *AddOnDefinitionOpts) *cmd.Command {
	cmd := &cmd.Command{
		Name:      "update",
		ShortHelp: "Update a HCP Waypoint add-on definition.",
		LongHelp:  "Update a HCP Waypoint add-on definition.",
		Examples: []cmd.Example{
			{
				Preamble: "Update a HCP Waypoint add-on definition:",
				Command:  "hcp waypoint add-ons definitions update -n=cli-test",
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
					Name:         "tfc-no-code-module-source",
					DisplayValue: "TFC_NO_CODE_MODULE_SOURCE",
					Description:  "The Terraform Cloud no-code module source.",
					Value:        flagvalue.Simple("", &opts.TerraformNoCodeModuleSource),
				},
				{
					Name:         "tfc-no-code-module-version",
					DisplayValue: "TFC_NO_CODE_MODULE_VERSION",
					Description:  "The Terraform Cloud no-code module version.",
					Value:        flagvalue.Simple("", &opts.TerraformNoCodeModuleVersion),
				},
				{
					Name:         "tfc-project-id",
					DisplayValue: "TFC_PROJECT_ID",
					Description:  "The Terraform Cloud project ID.",
					Value:        flagvalue.Simple("", &opts.TerraformCloudProjectID),
				},
				{
					Name:         "tfc-project-name",
					DisplayValue: "TFC_PROJECT_NAME",
					Description:  "The Terraform Cloud project name.",
					Value:        flagvalue.Simple("", &opts.TerraformCloudProjectName),
				},
			},
		},
	}

	return cmd
}

func addOnDefinitionUpdate(opts *AddOnDefinitionOpts) error {
	ns, err := opts.Namespace()
	if err != nil {
		return errors.Wrap(err, "unable to access HCP project")
	}

	var readmeTpl []byte
	if opts.ReadmeMarkdownTemplateFile != "" {
		readmeTpl, err = os.ReadFile(opts.ReadmeMarkdownTemplateFile)
		if err != nil {
			return errors.Wrapf(err, "failed to read README markdown template file %q", opts.ReadmeMarkdownTemplateFile)
		}
	}

	_, err = opts.WS.WaypointServiceUpdateAddOnDefinition2(
		&waypoint_service.WaypointServiceUpdateAddOnDefinition2Params{
			NamespaceID:                 ns.ID,
			Context:                     opts.Ctx,
			ExistingAddOnDefinitionName: opts.Name,
			Body: &models.HashicorpCloudWaypointWaypointServiceUpdateAddOnDefinitionBody{
				Summary:                opts.Summary,
				Description:            opts.Description,
				ReadmeMarkdownTemplate: readmeTpl,
				Labels:                 opts.Labels,
				TerraformNocodeModule: &models.HashicorpCloudWaypointTerraformNocodeModule{
					Source:  opts.TerraformNoCodeModuleSource,
					Version: opts.TerraformNoCodeModuleVersion,
				},
				TerraformCloudWorkspaceDetails: &models.HashicorpCloudWaypointTerraformCloudWorkspaceDetails{
					ProjectID: opts.TerraformCloudProjectID,
					Name:      opts.TerraformCloudProjectName,
				},
			},
		}, nil,
	)
	if err != nil {
		return errors.Wrapf(err, "failed to update add-on definition %q", opts.Name)
	}

	fmt.Fprintf(opts.IO.Err(), "Add-on definition %q updated.", opts.Name)

	return nil
}
