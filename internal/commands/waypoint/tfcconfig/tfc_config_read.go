// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tfcconfig

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2024-11-22/client/waypoint_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2024-11-22/models"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
	"github.com/pkg/errors"
)

func NewCmdRead(ctx *cmd.Context, runF func(opts *ReadOpts) error) *cmd.Command {
	opts := &ReadOpts{
		Ctx:            ctx.ShutdownCtx,
		Profile:        ctx.Profile,
		Output:         ctx.Output,
		IO:             ctx.IO,
		WaypointClient: waypoint_service.New(ctx.HCP, nil),
	}

	c := &cmd.Command{
		Name:      "read",
		ShortHelp: "Read TFC Config properties.",
		LongHelp: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
			The {{template "mdCodeOrBold" "hcp waypoint tfc-config read"}} command returns
			the TFC Organization name and a redacted form of the TFC Team token that is set
			for this HCP Project. There can only be one TFC Config set for each HCP Project.
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
	return c
}

func readRun(opts *ReadOpts) error {
	resp, err := opts.WaypointClient.WaypointServiceGetTFCConfig(
		&waypoint_service.WaypointServiceGetTFCConfigParams{
			NamespaceLocationOrganizationID: opts.Profile.OrganizationID,
			NamespaceLocationProjectID:      opts.Profile.ProjectID,
			Context:                         opts.Ctx,
		}, nil,
	)
	if err != nil {
		return errors.Wrapf(err, "%s failed to get TFC Config", opts.IO.ColorScheme().FailureIcon())
	}
	if resp.Payload.TfcConfig == nil {
		fmt.Fprintf(opts.IO.Out(), "%s No TFC Config found for this project\n",
			opts.IO.ColorScheme().FailureIcon())
		return nil
	}
	d := newDisplayer(format.Pretty, resp.Payload.TfcConfig)
	return opts.Output.Display(d)
}

type configDisplayer struct {
	config        *models.HashicorpCloudWaypointV20241122TFCConfig
	defaultFormat format.Format
}

func newDisplayer(defaultFormat format.Format, config *models.HashicorpCloudWaypointV20241122TFCConfig) *configDisplayer {
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

type ReadOpts struct {
	Ctx            context.Context
	Profile        *profile.Profile
	Output         *format.Outputter
	IO             iostreams.IOStreams
	WaypointClient waypoint_service.ClientService
}
