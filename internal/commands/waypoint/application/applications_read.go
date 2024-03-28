package application

import (
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2023-08-18/client/waypoint_service"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/flagvalue"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/pkg/errors"
)

func NewCmdReadApplication(ctx *cmd.Context, opts *ApplicationOpts) *cmd.Command {
	cmd := &cmd.Command{
		Name:      "read",
		ShortHelp: "Read an HCP Waypoint application.",
		LongHelp: heredoc.New(ctx.IO).Must(`,
			The {{ Bold "hcp waypoint applications read" }}" command lets you read
			details about an HCP Waypoint application.
`),
		Examples: []cmd.Example{
			{
				Preamble: "Read an HCP Waypoint application:",
				Command:  "$ hcp waypoint applications read -n my-application",
			},
		},
		RunF: func(c *cmd.Command, args []string) error {
			if opts.testFunc != nil {
				return opts.testFunc(c, args)
			}
			return applicationRead(opts)
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
					Description:  "The name of the HCP Waypoint application.",
					Value:        flagvalue.Simple("", &opts.Name),
					Required:     true,
				},
			},
		},
	}

	return cmd
}

func applicationRead(opts *ApplicationOpts) error {
	ns, err := opts.Namespace()
	if err != nil {
		return errors.Wrap(err, "failed to get namespace")
	}

	getResp, err := opts.WS.WaypointServiceGetApplication2(
		&waypoint_service.WaypointServiceGetApplication2Params{
			NamespaceID:     ns.ID,
			Context:         opts.Ctx,
			ApplicationName: opts.Name,
		}, nil,
	)
	if err != nil {
		return errors.Wrapf(err, "failed to get application %q", opts.Name)
	}

	return opts.Output.Show(getResp.GetPayload().Application, format.Pretty)
}
