package apps

import (
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-06-13/models"
	"github.com/hashicorp/hcp/internal/pkg/format"
)

type displayer struct {
	apps   []*models.Secrets20230613App
	single bool
}

func newDisplayer(single bool, apps ...*models.Secrets20230613App) *displayer {
	return &displayer{
		apps:   apps,
		single: single,
	}
}

func (d *displayer) DefaultFormat() format.Format {
	return format.Pretty
}

func (d *displayer) Payload() any {
	if d.single {
		if len(d.apps) != 1 {
			return nil
		}

		return d.apps[0]
	}

	return d.apps
}

func (d *displayer) FieldTemplates() []format.Field {
	return []format.Field{
		{
			Name:        "App Name",
			ValueFormat: "{{ .Name }}",
		},
		{
			Name:        "Description",
			ValueFormat: "{{ .Description }}",
		},
	}
}
