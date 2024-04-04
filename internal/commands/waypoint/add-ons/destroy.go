package addons

import (
	"fmt"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2023-08-18/client/waypoint_service"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/flagvalue"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/pkg/errors"
)

func NewCmdDestroy(ctx *cmd.Context, opts *AddOnOpts) *cmd.Command {
	cmd := &cmd.Command{
		Name:      "destroy",
		ShortHelp: "Destroy an HCP Waypoint add-ons.",
		LongHelp: heredoc.New(ctx.IO).Must(`
The {{ template "mdCodeOrBold" "hcp waypoint add-ons destroy" }} command lets you
destroy an existing HCP Waypoint add-on.
`),
		Examples: []cmd.Example{
			{
				Preamble: "Destroy an HCP Waypoint add-on:",
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
$ hcp waypoint add-ons destroy -n=my-addon
`),
			},
		},
		RunF: func(c *cmd.Command, args []string) error {
			if opts.testFunc != nil {
				return opts.testFunc(c, args)
			}
			return addOnDestroy(opts)
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
					Description:  "The name of the add-on to destroy.",
					Value:        flagvalue.Simple("", &opts.Name),
					Required:     true,
				},
			},
		},
	}
	return cmd
}

func addOnDestroy(opts *AddOnOpts) error {
	ns, err := opts.Namespace()
	if err != nil {
		return err
	}

	_, err = opts.WS.WaypointServiceDestroyAddOn2(
		&waypoint_service.WaypointServiceDestroyAddOn2Params{
			NamespaceID: ns.ID,
			Context:     opts.Ctx,
			AddOnName:   opts.Name,
		}, nil,
	)
	if err != nil {
		return errors.Wrapf(err, "%s failed to destroy add-on %q",
			opts.IO.ColorScheme().FailureIcon(),
			opts.Name)
	}

	fmt.Fprintf(opts.IO.Out(), "Add-on %s destroyed\n", opts.Name)

	return nil
}
