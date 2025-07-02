// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2024-11-22/client/waypoint_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2024-11-22/models"
	"github.com/hashicorp/hcp/internal/pkg/profile"
)

type StatusOperation struct {
	Message string
	Values  map[string]string
	Status  string
}

func (s *StatusOperation) Run(
	ctx context.Context,
	log hclog.Logger,
	api waypoint_service.ClientService,
	profile *profile.Profile,
	opInfo *models.HashicorpCloudWaypointV20241122AgentOperation,
) (OperationStatus, error) {
	if api == nil {
		return OperationStatus{}, fmt.Errorf("the status operation requires an API client and none was provided")
	}
	if profile == nil {
		return OperationStatus{}, fmt.Errorf("the status operation requires a profile and none was provided")
	}
	if opInfo == nil || opInfo.ActionRunID == "" {
		return OperationStatus{}, fmt.Errorf("the status operation requires a run ID and none was provided")
	}

	ret, err := api.WaypointServiceSendStatusLog2(&waypoint_service.WaypointServiceSendStatusLog2Params{
		NamespaceLocationOrganizationID: profile.OrganizationID,
		NamespaceLocationProjectID:      profile.ProjectID,

		ActionRunID: opInfo.ActionRunID,

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
