// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package agent

import (
	"context"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2024-11-22/client/waypoint_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2024-11-22/models"
)

type StatusOperation struct {
	wp waypoint_service.ClientService

	orgID  string
	projID string
	cfgID  string
	runID  string

	Message string
	Values  map[string]string
	Status  string
}

func (s *StatusOperation) Run(ctx context.Context, log hclog.Logger) (OperationStatus, error) {
	ret, err := s.wp.WaypointServiceSendStatusLog(&waypoint_service.WaypointServiceSendStatusLogParams{
		ActionRunSpecifierActionID:      s.cfgID,
		ActionRunSpecifierSequence:      s.runID,
		NamespaceLocationOrganizationID: s.orgID,
		NamespaceLocationProjectID:      s.projID,

		Body: &models.HashicorpCloudWaypointV20241122WaypointServiceSendStatusLogBody{
			StatusLog: &models.HashicorpCloudWaypointV20241122StatusLog{
				EmittedAt: strfmt.DateTime(time.Now()),
				Log:       s.Message,
				Metadata:  s.Values,
			},
		},
	}, nil)
	if err != nil {
		return OperationStatus{}, err
	}

	if ret.IsClientError() {
		log.Error("error sending status log (client side)", "error", ret.Error())
	}

	if ret.IsServerError() {
		log.Error("error sending status log (server side)", "error", ret.Error())
	}

	return cleanStatus, nil
}
