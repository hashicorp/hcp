package agent

import (
	"context"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2023-08-18/client/waypoint_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2023-08-18/models"
)

type StatusOperation struct {
	wp waypoint_service.ClientService

	namespace string
	cfgID     string
	runID     string

	Message string
	Values  map[string]string
	Status  string
}

func (s *StatusOperation) Run(ctx context.Context, log hclog.Logger) (OperationStatus, error) {
	ret, err := s.wp.WaypointServiceSendStatusLog(&waypoint_service.WaypointServiceSendStatusLogParams{
		ActionConfigID: s.cfgID,
		ActionRunSeq:   s.runID,
		NamespaceID:    s.namespace,

		Body: &models.HashicorpCloudWaypointWaypointServiceSendStatusLogBody{
			StatusLog: &models.HashicorpCloudWaypointStatusLog{
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
