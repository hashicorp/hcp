package application

import (
	"fmt"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2023-08-18/client/waypoint_service"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/flagvalue"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/pkg/errors"
)

func NewCmdDestroyApplication(ctx *cmd.Context, opts *ApplicationOpts) *cmd.Command {
	cmd := &cmd.Command{
		Name:      "destroy",
		ShortHelp: "Destroy an HCP Waypoint application and its infrastructure.",
		LongHelp: heredoc.New(ctx.IO).Must(`
The {{ Bold "hcp waypoint applications destroy" }} command lets you destroy
an HCP Waypoint application and its infrastructure.
		`),
		Examples: []cmd.Example{
			{
				Preamble: "Destroy an HCP Waypoint application:",
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
$ hcp waypoint applications destroy -n my-application
`),
			},
		},
		RunF: func(c *cmd.Command, args []string) error {
			if opts.testFunc != nil {
				return opts.testFunc(c, args)
			}
			return applicationDestroy(opts)
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
					Description:  "The name of the application to destroy.",
					Value:        flagvalue.Simple("", &opts.Name),
					Required:     true,
				},
			},
		},
	}

	return cmd
}

func applicationDestroy(opts *ApplicationOpts) error {
	ns, err := opts.Namespace()
	if err != nil {
		return errors.Wrap(err, "failed to get namespace")
	}

	_, err = opts.WS.WaypointServiceDestroyApplication2(
		&waypoint_service.WaypointServiceDestroyApplication2Params{
			NamespaceID:     ns.ID,
			Context:         opts.Ctx,
			ApplicationName: opts.Name,
		}, nil,
	)
	if err != nil {
		return errors.Wrapf(err, "failed to destroy application %q", opts.Name)
	}

	fmt.Fprintf(opts.IO.Err(), "Application %q destroyed.", opts.Name)

	return nil
}
