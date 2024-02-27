package tfcconfig

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2023-08-18/client/waypoint_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2023-08-18/models"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
)

func NewCmdRead(ctx *cmd.Context, runF func(opts *TFCConfigReadOpts) error) *cmd.Command {
	opts := &TFCConfigReadOpts{
		Ctx:            ctx.ShutdownCtx,
		Profile:        ctx.Profile,
		Output:         ctx.Output,
		IO:             ctx.IO,
		WaypointClient: waypoint_service.New(ctx.HCP, nil),
	}

	cmd := &cmd.Command{
		Name:      "read",
		ShortHelp: "Read TFC Config properties.",
		LongHelp: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
			The {{Bold "hcp waypoint tfc-config read"}} command returns the TFC Organization name and a redacted form
		of the TFC Team token that is set for this HCP Project. There can only be one TFC Config set for each HCP Project.
		`),
		Examples: []cmd.Example{
			{
				Preamble: `Retrieve the saved TFC Config from Waypoint for this HCP Project ID:`,
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				$ hcp waypoint tfc-config read`),
			},
		},
		RunF: func(c *cmd.Command, args []string) error {
			if runF != nil {
				return runF(opts)
			}
			return readRun(opts)
		},
		PersistentPreRun: func(c *cmd.Command, args []string) error {
			return cmd.RequireOrgAndProject(ctx)
		},
	}
	return cmd
}

func readRun(opts *TFCConfigReadOpts) error {
	nsID, err := GetNamespace(opts.Ctx, opts.WaypointClient, opts.Profile.OrganizationID, opts.Profile.ProjectID)
	if err != nil {
		return fmt.Errorf("error getting namespace: %w", err)
	}
	resp, err := opts.WaypointClient.WaypointServiceGetTFCConfig(
		&waypoint_service.WaypointServiceGetTFCConfigParams{
			NamespaceID: nsID,
			Context:     opts.Ctx,
		}, nil,
	)
	if err != nil {
		return fmt.Errorf("error retrieving TFC config: %w", err)

	}
	d := newDisplayer(format.Pretty, resp.Payload.TfcConfig)
	return opts.Output.Display(d)
}

type configDisplayer struct {
	config        *models.HashicorpCloudWaypointTFCConfig
	defaultFormat format.Format
}

func newDisplayer(defaultFormat format.Format, config *models.HashicorpCloudWaypointTFCConfig) *configDisplayer {
	return &configDisplayer{
		config:        config,
		defaultFormat: defaultFormat,
	}
}

func (d configDisplayer) DefaultFormat() format.Format { return d.defaultFormat }
func (d configDisplayer) Payload() any                 { return d.config }

func (d configDisplayer) FieldTemplates() []format.Field {
	return []format.Field{
		{
			Name:        "Organization Name",
			ValueFormat: "{{ .OrganizationName }}",
		},
		{
			Name:        "Token",
			ValueFormat: "{{ .Token }}",
		},
	}
}

type TFCConfigReadOpts struct {
	Ctx            context.Context
	Profile        *profile.Profile
	Output         *format.Outputter
	IO             iostreams.IOStreams
	WaypointClient waypoint_service.ClientService
}
