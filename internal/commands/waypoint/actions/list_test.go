package actions

import (
	"context"
	"net/http"
	"testing"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2024-11-22/client/waypoint_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2024-11-22/models"
	"github.com/hashicorp/hcp/internal/commands/waypoint/opts"
	mock_waypoint_service "github.com/hashicorp/hcp/internal/pkg/api/mocks/github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2024-11-22/client/waypoint_service"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestListActionsRun(t *testing.T) {
	t.Parallel()
	cases := []struct {
		Name    string
		Resp    []*models.HashicorpCloudWaypointActionConfig
		RespErr bool
		Error   string
	}{
		{
			Name:    "Server error",
			RespErr: true,
			Error:   "error listing actions: [GET /waypoint/2024-11-22/organizations/{namespace.location.organization_id}/projects/{namespace.location.project_id}/actionconfigs][403]",
		},
		{
			Name: "Good empty",
			Resp: []*models.HashicorpCloudWaypointActionConfig{},
		},
		{
			Name: "Good",
			Resp: []*models.HashicorpCloudWaypointActionConfig{
				{
					Name:        "test-name",
					ActionURL:   "https://example.com",
					Description: "test-description",
				},
			},
		},
		{
			Name: "Good multiple",
			Resp: []*models.HashicorpCloudWaypointActionConfig{
				{
					Name:        "test-name",
					ActionURL:   "https://example.com",
					Description: "test-description",
				},
				{
					Name:        "test-name-2",
					ActionURL:   "https://example.com/2",
					Description: "test-description-2",
				},
			},
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			r := require.New(t)

			io := iostreams.Test()
			ws := mock_waypoint_service.NewMockClientService(t)
			opts := &ListOpts{
				WaypointOpts: opts.WaypointOpts{
					Ctx:          context.Background(),
					Profile:      profile.TestProfile(t).SetOrgID("123").SetProjectID("456"),
					Output:       format.New(io),
					WS2024Client: ws,
				},
			}

			call := ws.EXPECT().WaypointServiceListActionConfigs(mock.Anything, mock.Anything)

			if c.RespErr {
				call.Return(nil, waypoint_service.NewWaypointServiceListActionConfigsDefault(http.StatusForbidden))
			} else {
				ok := waypoint_service.NewWaypointServiceListActionConfigsOK()
				ok.Payload = &models.HashicorpCloudWaypointListActionConfigResponse{
					ActionConfigs: c.Resp,
				}
				call.Return(ok, nil)
			}

			err := listActions(nil, nil, opts)
			if c.Error != "" {
				r.ErrorContains(err, c.Error)
				return
			}

			r.NoError(err)

			for _, action := range c.Resp {
				r.Contains(io.Output.String(), action.Name)
				r.Contains(io.Output.String(), action.Description)
				r.Contains(io.Output.String(), action.ActionURL)
			}
		})
	}
}
