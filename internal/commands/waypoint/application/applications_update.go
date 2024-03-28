package application

import (
	"fmt"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2023-08-18/client/waypoint_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2023-08-18/models"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/flagvalue"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/pkg/errors"
)

func NewCmdApplicationsUpdate(ctx *cmd.Context, opts *ApplicationOpts) *cmd.Command {
	cmd := &cmd.Command{
		Name:      "update",
		ShortHelp: "Update an existing HCP Waypoint application.",
		LongHelp: heredoc.New(ctx.IO).Must(`
The {{ Bold "hcp waypoint applications update" }} command lets you update
an existing HCP Waypoint application.
		`),
		Examples: []cmd.Example{
			{
				Preamble: "Update an existing HCP Waypoint application:",
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
$ hcp waypoint applications update -n my-application -action-config-name my-action-config
`),
			},
		},
		RunF: func(c *cmd.Command, args []string) error {
			if opts.testFunc != nil {
				return opts.testFunc(c, args)
			}
			return applicationUpdate(opts)
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
					Description:  "The name of the HCP Waypoint application to update.",
					Value:        flagvalue.Simple("", &opts.Name),
					Required:     true,
				},
				{
					Name:         "action-config-name",
					DisplayValue: "ACTION_CONFIG_NAME",
					Description:  "The name of the action configuration to be added to the application.",
					Value:        flagvalue.SimpleSlice(nil, &opts.ActionConfigNames),
					Required:     false,
					Repeatable:   true,
				},
			},
		},
	}

	return cmd
}

func applicationUpdate(opts *ApplicationOpts) error {
	ns, err := opts.Namespace()
	if err != nil {
		return errors.Wrap(err, "failed to get namespace")
	}

	_, err = opts.WS.WaypointServiceUpdateApplication(
		&waypoint_service.WaypointServiceUpdateApplicationParams{
			NamespaceID: ns.ID,
			Context:     opts.Ctx,
			Body: &models.HashicorpCloudWaypointWaypointServiceUpdateApplicationBody{
				Name: opts.Name,
			},
		}, nil,
	)
	if err != nil {
		return errors.Wrapf(err, "failed to update application %q", opts.Name)
	}

	fmt.Fprintf(opts.IO.Err(), "Application %q updated.", opts.Name)

	return nil
}
