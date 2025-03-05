// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opts

import (
	"context"

	service20241122 "github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2024-11-22/client/waypoint_service"

	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
)

func New(ctx *cmd.Context) WaypointOpts {
	return WaypointOpts{
		Ctx:          ctx.ShutdownCtx,
		Profile:      ctx.Profile,
		IO:           ctx.IO,
		Output:       ctx.Output,
		WS2024Client: service20241122.New(ctx.HCP, nil),
	}
}

type WaypointOpts struct {
	WS2024Client service20241122.ClientService

	Ctx     context.Context
	Profile *profile.Profile
	IO      iostreams.IOStreams
	Output  *format.Outputter
}
