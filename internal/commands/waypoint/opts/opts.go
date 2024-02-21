package opts

import (
	"context"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2023-08-18/client/waypoint_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2023-08-18/models"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
	"github.com/pkg/errors"
)

func New(ctx *cmd.Context) WaypointOpts {
	return WaypointOpts{
		Ctx:     ctx.ShutdownCtx,
		Profile: ctx.Profile,
		IO:      ctx.IO,
		Output:  ctx.Output,
		WS:      waypoint_service.New(ctx.HCP, nil),
	}
}

type WaypointOpts struct {
	WS waypoint_service.ClientService

	Ctx     context.Context
	Profile *profile.Profile
	IO      iostreams.IOStreams
	Output  *format.Outputter
}

func (w *WaypointOpts) Namespace() (*models.HashicorpCloudWaypointNamespace, error) {
	resp, err := w.WS.WaypointServiceGetNamespace(&waypoint_service.WaypointServiceGetNamespaceParams{
		LocationOrganizationID: w.Profile.OrganizationID,
		LocationProjectID:      w.Profile.ProjectID,
		Context:                w.Ctx,
	}, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to access HCP project")
	}

	return resp.Payload.Namespace, nil
}
