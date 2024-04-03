package tfcconfig

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2023-08-18/client/waypoint_service"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
	"github.com/pkg/errors"
)

func NewCmdDelete(ctx *cmd.Context, runF func(opts *TFCConfigDeleteOpts) error) *cmd.Command {
	opts := &TFCConfigDeleteOpts{
		Ctx:            ctx.ShutdownCtx,
		Profile:        ctx.Profile,
		Output:         ctx.Output,
		IO:             ctx.IO,
		WaypointClient: waypoint_service.New(ctx.HCP, nil),
	}
	cmd := &cmd.Command{
		Name:      "delete",
		ShortHelp: "Delete TFC Configuration.",
		LongHelp: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
			The {{ template "mdCodeOrBold" "hcp waypoint tfc-config delete" }} command deletes
			the TFC Organization name and team token that is set for this HCP
			Project. Only one TFC Config is allowed for each HCP Project.
		`),
		Examples: []cmd.Example{
			{
				Preamble: `Delete the saved TFC Config from Waypoint for this HCP Project ID:`,
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				$ hcp waypoint tfc-config delete example-org
`),
			},
		},
		RunF: func(c *cmd.Command, args []string) error {
			if runF != nil {
				return runF(opts)
			}
			return deleteRun(opts)
		},
		PersistentPreRun: func(c *cmd.Command, args []string) error {
			return cmd.RequireOrgAndProject(ctx)
		},
	}
	return cmd

}

func deleteRun(opts *TFCConfigDeleteOpts) error {
	nsID, err := GetNamespace(opts.Ctx, opts.WaypointClient, opts.Profile.OrganizationID, opts.Profile.ProjectID)
	if err != nil {
		return err
	}

	resp, err := opts.WaypointClient.WaypointServiceDeleteTFCConfig(
		&waypoint_service.WaypointServiceDeleteTFCConfigParams{
			NamespaceID: nsID,
			Context:     opts.Ctx,
		}, nil,
	)
	if err != nil {
		return errors.Wrapf(err, "%s error deleting TFC config", opts.IO.ColorScheme().FailureIcon())
	}

	if resp.IsSuccess() {
		fmt.Fprintf(opts.IO.Err(), "%s TFC Config successfully deleted!\n", opts.IO.ColorScheme().SuccessIcon())
	}

	return nil
}

type TFCConfigDeleteOpts struct {
	Ctx            context.Context
	Profile        *profile.Profile
	Output         *format.Outputter
	IO             iostreams.IOStreams
	WaypointClient waypoint_service.ClientService
}
