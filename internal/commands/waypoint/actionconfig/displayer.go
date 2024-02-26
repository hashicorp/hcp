package actionconfig

import (
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2023-08-18/models"
	"github.com/hashicorp/hcp/internal/pkg/format"
)

// TODO(briancain): https://github.com/hashicorp/hcp/issues/16
type displayer struct {
	actionConfigs []*models.HashicorpCloudWaypointActionConfig
	defaultFormat format.Format
	single        bool
}

func newDisplayer(defaultFormat format.Format, single bool, actionConfigs ...*models.HashicorpCloudWaypointActionConfig) *displayer {
	return &displayer{
		actionConfigs: actionConfigs,
		defaultFormat: defaultFormat,
		single:        single,
	}
}

func (d *displayer) DefaultFormat() format.Format {
	return d.defaultFormat
}

func (d *displayer) Payload() any {
	if d.single {
		if len(d.actionConfigs) != 1 {
			return nil
		}

		return d.actionConfigs[0]
	}

	return d.actionConfigs
}

func (d *displayer) FieldTemplates() []format.Field {
	return []format.Field{
		{
			Name:        "Name",
			ValueFormat: "{{ .Name }}",
		},
		{
			Name:        "ID",
			ValueFormat: "{{ .ID }}",
		},
		{
			Name:        "Description",
			ValueFormat: "{{ .Description }}",
		},
		// TODO(briancain): Show Request field nested structs
	}
}
