package applications

import (
	"fmt"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2023-08-18/client/waypoint_service"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/flagvalue"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/pkg/errors"
)

func NewCmdApplicationsDestroy(ctx *cmd.Context, opts *ApplicationOpts) *cmd.Command {
	cmd := &cmd.Command{
		Name:      "destroy",
		ShortHelp: "Destroy an HCP Waypoint applications and its infrastructure.",
		LongHelp: heredoc.New(ctx.IO).Must(`
The {{ template "mdCodeOrBold" "hcp waypoint applications destroy" }} command lets you destroy
an HCP Waypoint applications and its infrastructure.
		`),
		Examples: []cmd.Example{
			{
				Preamble: "Destroy an HCP Waypoint applications:",
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
$ hcp waypoint applications destroy -n my-applications
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
					Description:  "The name of the applications to destroy.",
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

	if opts.IO.CanPrompt() {
		ok, err := opts.IO.PromptConfirm(
			"The HCP Waypoint application will be deleted.\n\n" +
				"Do you want to continue")
		if err != nil {
			return fmt.Errorf("failed to retrieve confirmation: %w", err)
		}

		if !ok {
			return nil
		}
	}

	_, err = opts.WS.WaypointServiceDestroyApplication2(
		&waypoint_service.WaypointServiceDestroyApplication2Params{
			NamespaceID:     ns.ID,
			Context:         opts.Ctx,
			ApplicationName: opts.Name,
		}, nil,
	)
	if err != nil {
		return errors.Wrapf(err, "failed to destroy applications %q", opts.Name)
	}

	fmt.Fprintf(opts.IO.Err(), "Application %q destroyed.", opts.Name)

	return nil
}
