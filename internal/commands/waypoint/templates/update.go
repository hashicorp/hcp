package templates

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
$ hcp waypoint templates update -n my-template \
  -s "My Template Summary" \
  -d "My Template Description" \
  -readme-markdown-template-file "README.tpl" \
  --tfc-no-code-module-source "app.terraform.io/hashicorp/dir/template" \
  --tfc-no-code-module-version "1.0.2" \
  --tfc-project-name "my-tfc-project" \
  --tfc-project-id "prj-123456 \
  -l "label1" \
  -l "label2" 
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
					Name:         "tfc-no-code-module-source",
					DisplayValue: "TFC_NO_CODE_MODULE_SOURCE",
					Description: heredoc.New(ctx.IO).Must(`
			The source of the Terraform no-code module. 
			The expected format is "NAMESPACE/NAME/PROVIDER". An
			optional "HOSTNAME/" can be added at the beginning for
			a private registry.
					`),
					Value: flagvalue.Simple("", &opts.TerraformNoCodeModuleSource),
				},
				{
					Name:         "tfc-no-code-module-version",
					DisplayValue: "TFC_NO_CODE_MODULE_VERSION",
					Description:  "The version of the Terraform no-code module.",
					Value:        flagvalue.Simple("", &opts.TerraformNoCodeModuleVersion),
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
			return errors.Wrapf(err, "%s failed to read README template file %q",
				opts.IO.ColorScheme().FailureIcon(),
				opts.ReadmeMarkdownTemplateFile,
			)
		}
	}

	updatedTpl := &models.HashicorpCloudWaypointApplicationTemplate{
		Name:                   opts.UpdatedName,
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
	}

	_, err = opts.WS.WaypointServiceUpdateApplicationTemplate2(
		&waypoint_service.WaypointServiceUpdateApplicationTemplate2Params{
			NamespaceID:                     ns.ID,
			Context:                         opts.Ctx,
			ExistingApplicationTemplateName: opts.Name,
			Body: &models.HashicorpCloudWaypointWaypointServiceUpdateApplicationTemplateBody{
				ApplicationTemplate: updatedTpl,
			},
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
