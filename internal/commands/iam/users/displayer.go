package users

import (
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/models"
	"github.com/hashicorp/hcp/internal/pkg/format"
)

type displayer struct {
	users         []*models.HashicorpCloudIamUserPrincipal
	defaultFormat format.Format
	single        bool
}

func newDisplayer(defaultFormat format.Format, single bool, users ...*models.HashicorpCloudIamUserPrincipal) *displayer {
	return &displayer{
		users:         users,
		defaultFormat: defaultFormat,
		single:        single,
	}
}

func (d *displayer) DefaultFormat() format.Format {
	return d.defaultFormat
}

func (d *displayer) Payload() any {
	if d.single {
		if len(d.users) != 1 {
			return nil
		}

		return d.users[0]
	}

	return d.users
}

func (d *displayer) FieldTemplates() []format.Field {
	return []format.Field{
		{
			Name:        "Name",
			ValueFormat: "{{ .FullName }}",
		},
		{
			Name:        "ID",
			ValueFormat: "{{ .ID }}",
		},
		{
			Name:        "Email",
			ValueFormat: "{{ .Email }}",
		},
	}
}
