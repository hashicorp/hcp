package templates

import (
	"fmt"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2023-08-18/client/waypoint_service"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/flagvalue"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/pkg/errors"
)

func NewCmdDelete(ctx *cmd.Context, opts *TemplateOpts) *cmd.Command {
	cmd := &cmd.Command{
		Name:      "delete",
		ShortHelp: "Delete an existing Waypoint template.",
		LongHelp: heredoc.New(ctx.IO).Must(`
The {{ template "mdCodeOrBold" "hcp waypoint templates delete" }} command lets you delete
existing HCP Waypoint templates.
		`),
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
					Name:         "name",
					Shorthand:    "n",
					DisplayValue: "NAME",
					Description:  "The name of the template to be deleted.",
					Value:        flagvalue.Simple("", &opts.Name),
					Required:     true,
				},
			},
		},
	}

	return cmd
}

func templateDelete(opts *TemplateOpts) error {
	ns, err := opts.Namespace()
	if err != nil {
		return err
	}

	_, err = opts.WS.WaypointServiceDeleteApplicationTemplate2(
		&waypoint_service.WaypointServiceDeleteApplicationTemplate2Params{
			NamespaceID:             ns.ID,
			Context:                 opts.Ctx,
			ApplicationTemplateName: opts.Name,
		}, nil,
	)
	if err != nil {
		return errors.Wrapf(err, "failed to delete template %q", opts.Name)
	}

	fmt.Fprintf(opts.IO.Err(), "%s Template %q deleted.\n",
		opts.IO.ColorScheme().SuccessIcon(),
		opts.Name,
	)

	return nil
}
