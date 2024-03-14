package template

import (
	"fmt"
	"os"

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
		LongHelp: "Create a new HCP Waypoint template, from which applications " +
			"can be created.",
		Examples: []cmd.Example{
			{
				Preamble: "Create a new HCP Waypoint template:",
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
$ hcp waypoint templates create -n my-template \
-s "My Template Summary" \
-d "My Template Description" \
-readme-markdown-template-file "README.tpl" \
-tfc-no-code-module-source "app.terraform.io/hashicorp/dir/template" \
-tfc-no-code-module-version "1.0.2" \
-tfc-project-name "my-tfc-project" \
-tfc-project-id "prj-123456 \
-l "label1" \
-l "label2" \
-t "key1=value1" \
-t "key2=value2"
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
					Name:         "tfc-no-code-module-version",
					DisplayValue: "TFC_NO_CODE_MODULE_VERSION",
					Description:  "The version of the Terraform no-code module.",
					Value:        flagvalue.Simple("", &opts.TerraformNoCodeModuleVersion),
					Required:     true,
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
					Description: "The ID of the Terraform Cloud project where" +
						" applications using this template will be created.",
					Value:    flagvalue.Simple("", &opts.TerraformCloudProjectID),
					Required: true,
				},
			},
		},
	}

	return cmd
}

func templateCreate(opts *TemplateOpts) error {
	ns, err := opts.Namespace()
	if err != nil {
		return errors.Wrapf(err, "Unable to access HCP project")
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
			return fmt.Errorf("failed to read README markdown template file: %w", err)
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
					TerraformNocodeModule: &models.HashicorpCloudWaypointTerraformNocodeModule{
						Source:  opts.TerraformNoCodeModuleSource,
						Version: opts.TerraformNoCodeModuleVersion,
					},
					TerraformCloudWorkspaceDetails: &models.HashicorpCloudWaypointTerraformCloudWorkspaceDetails{
						Name:      opts.TerraformCloudProjectName,
						ProjectID: opts.TerraformCloudProjectID,
					},
				},
			},
			Context: opts.Ctx,
		}, nil)
	if err != nil {
		return fmt.Errorf("failed to create template %q: %w", opts.Name, err)
	}

	fmt.Fprintf(opts.IO.Err(), "Template %q created.", opts.Name)

	return nil
}
