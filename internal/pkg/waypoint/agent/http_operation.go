// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package agent

import (
	"context"
	"net/http"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2024-11-22/client/waypoint_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2024-11-22/models"
	"github.com/hashicorp/hcp/internal/pkg/profile"
)

type HTTPOperation struct {
	URL string
}

func (h *HTTPOperation) Run(
	ctx context.Context,
	log hclog.Logger,
	api waypoint_service.ClientService,
	profile *profile.Profile,
	opInfo *models.HashicorpCloudWaypointV20241122AgentOperation,
) (OperationStatus, error) {
	resp, err := http.Get(h.URL)
	if err != nil {
		return errStatus, err
	}

	defer resp.Body.Close()

	return cleanStatus, nil
}
