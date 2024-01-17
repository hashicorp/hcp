package projects

import (
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/models"
	"github.com/hashicorp/hcp/internal/pkg/format"
)

type displayer struct {
	projects      []*models.HashicorpCloudResourcemanagerProject
	defaultFormat format.Format
	single        bool
}

func newDisplayer(defaultFormat format.Format, single bool, projects ...*models.HashicorpCloudResourcemanagerProject) *displayer {
	return &displayer{
		projects:      projects,
		defaultFormat: defaultFormat,
		single:        single,
	}
}

func (d *displayer) DefaultFormat() format.Format {
	return d.defaultFormat
}

func (d *displayer) Payload() any {
	if d.single {
		if len(d.projects) != 1 {
			return nil
		}

		return d.projects[0]
	}

	return d.projects
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
		{
			Name:        "Created At",
			ValueFormat: "{{ .CreatedAt }}",
		},
	}
}
