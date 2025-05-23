// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package agent

import (
	"context"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2024-11-22/client/waypoint_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2024-11-22/models"
	"github.com/hashicorp/hcp/internal/pkg/profile"
)

type NoopWrapper struct {
	Operation Operation
}

func (n *NoopWrapper) Run(
	ctx context.Context,
	log hclog.Logger,
	api waypoint_service.ClientService,
	profile *profile.Profile,
	opInfo *models.HashicorpCloudWaypointV20241122AgentOperation,
) (OperationStatus, error) {
	return cleanStatus, nil
}

func NoopOperations(o Operation) Operation {
	return &NoopWrapper{Operation: o}
}
