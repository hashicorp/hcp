package definitions

import (
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2023-08-18/client/waypoint_service"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/flagvalue"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/pkg/errors"
)

func NewCmdAddOnDefinitionRead(ctx *cmd.Context, opts *AddOnDefinitionOpts) *cmd.Command {
	cmd := &cmd.Command{
		Name:      "read",
		ShortHelp: "Read an HCP Waypoint add-on definition.",
		LongHelp: heredoc.New(ctx.IO).Must(`
The {{ template "mdCodeOrBold" "hcp waypoint add-ons definitions read" }}
command lets you read an existing HCP Waypoint add-on definition.
`),
		Examples: []cmd.Example{
			{
				Preamble: "Read an HCP Waypoint add-on definition:",
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
$ hcp waypoint add-ons definitions read -n my-addon-definition
`),
			},
		},
		RunF: func(c *cmd.Command, args []string) error {
			if opts.testFunc != nil {
				return opts.testFunc(c, args)
			}
			return addOnDefinitionRead(opts)
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
					Description:  "The name of the add-on definition.",
					Value:        flagvalue.Simple("", &opts.Name),
					Required:     true,
				},
			},
		},
	}
	return cmd
}

func addOnDefinitionRead(opts *AddOnDefinitionOpts) error {
	ns, err := opts.Namespace()
	if err != nil {
		return err
	}

	getResp, err := opts.WS.WaypointServiceGetAddOnDefinition2(
		&waypoint_service.WaypointServiceGetAddOnDefinition2Params{
			NamespaceID:         ns.ID,
			Context:             opts.Ctx,
			AddOnDefinitionName: opts.Name,
		}, nil,
	)
	if err != nil {
		return errors.Wrapf(err, "%s failed to get add-on definition %q",
			opts.IO.ColorScheme().FailureIcon(),
			opts.Name,
		)
	}

	getRespPayload := getResp.GetPayload()

	return opts.Output.Show(getRespPayload.AddOnDefinition, format.Pretty)
}
