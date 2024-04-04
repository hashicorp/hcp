package addons

import (
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2023-08-18/client/waypoint_service"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/flagvalue"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/pkg/errors"
)

func NewCmdList(ctx *cmd.Context, opts *AddOnOpts) *cmd.Command {
	cmd := &cmd.Command{
		Name:      "list",
		ShortHelp: "List HCP Waypoint add-ons.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp waypoint add-ons list" }} command lists
		HCP Waypoint add-ons. By supplying the "name" flag, you can list add-ons
		for a specific application.
`),
		Examples: []cmd.Example{
			{
				Preamble: "List all HCP Waypoint add-ons:",
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
$ hcp waypoint add-ons list
`),
			},
			{
				Preamble: "List HCP Waypoint add-ons for a specific application:",
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
$ hcp waypoint add-ons list --application-name my-application
`),
			},
		},
		RunF: func(cmd *cmd.Command, args []string) error {
			if opts.testFunc != nil {
				return opts.testFunc(cmd, args)
			}
			return addOnsList(opts)
		},
		PersistentPreRun: func(c *cmd.Command, args []string) error {
			return cmd.RequireOrgAndProject(ctx)
		},
		Flags: cmd.Flags{
			Local: []*cmd.Flag{
				{
					Name:         "application-name",
					DisplayValue: "APPLICATION_NAME",
					Description:  "The name of the application to list add-ons for.",
					Value:        flagvalue.Simple("", &opts.ApplicationName),
					Required:     false,
				},
			},
		},
	}
	return cmd
}

func addOnsList(opts *AddOnOpts) error {
	ns, err := opts.Namespace()
	if err != nil {
		return err
	}

	resp, err := opts.WS.WaypointServiceListAddOns(
		&waypoint_service.WaypointServiceListAddOnsParams{
			NamespaceID:     ns.ID,
			Context:         opts.Ctx,
			ApplicationName: &opts.ApplicationName,
		}, nil,
	)
	if err != nil {
		return errors.Wrapf(err, "%s failed to list add-ons",
			opts.IO.ColorScheme().FailureIcon())
	}

	return opts.Output.Show(resp.GetPayload().AddOns, format.Pretty)
}
