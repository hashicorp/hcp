package template

import (
	"fmt"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2023-08-18/client/waypoint_service"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/flagvalue"
	"github.com/pkg/errors"
)

func NewCmdDelete(ctx *cmd.Context, opts *TemplateOpts) *cmd.Command {
	cmd := &cmd.Command{
		Name:      "delete",
		ShortHelp: "Delete an existing Waypoint template.",
		LongHelp: "Delete an existing Waypoint template. This will remove" +
			" the template completely from HCP Waypoint.",
		RunF: func(c *cmd.Command, args []string) error {
			if opts.testFunc != nil {
				return opts.testFunc(c, args)
			}
			return templateDelete(opts)
		},
		PersistentPreRun: func(c *cmd.Command, args []string) error {
			return cmd.RequireOrgAndProject(ctx)
		},
		Flags: cmd.Flags{
			Local: []*cmd.Flag{
				{
					Name:         "id",
					DisplayValue: "ID",
					Description:  "The ID of the template to be deleted.",
					Value:        flagvalue.Simple("", &opts.ID),
					// When the API supports deleting by name, instead of only ID,
					// this should no longer be required, and a name flag should
					// be added.
					Required: true,
				},
				{
					Name:         "name",
					Shorthand:    "n",
					DisplayValue: "NAME",
					Description:  "The name of the template to be deleted.",
					Value:        flagvalue.Simple("", &opts.Name),
					Required:     false,
					// When the API supports deleting by name, instead of only ID,
					// this should no longer be hidden.
					Hidden: true,
				},
			},
		},
	}

	return cmd
}

func templateDelete(opts *TemplateOpts) error {
	ns, err := opts.Namespace()
	if err != nil {
		return errors.Wrapf(err, "Unable to access HCP project")
	}

	_, err = opts.WS.WaypointServiceDeleteApplicationTemplate(
		&waypoint_service.WaypointServiceDeleteApplicationTemplateParams{
			NamespaceID:             ns.ID,
			Context:                 opts.Ctx,
			ApplicationTemplateID:   opts.ID,
			ApplicationTemplateName: &opts.Name,
		}, nil,
	)
	if err != nil {
		return fmt.Errorf("failed to delete template %q: %w", opts.ID, err)
	}

	fmt.Fprintf(opts.IO.Err(), "Template %q deleted.", opts.ID)

	return nil
}
