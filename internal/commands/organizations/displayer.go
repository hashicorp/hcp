// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: MPL-2.0

package organizations

import (
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/models"
	"github.com/hashicorp/hcp/internal/pkg/format"
)

type displayer struct {
	orgs          []*models.HashicorpCloudResourcemanagerOrganization
	defaultFormat format.Format
	single        bool
}

func newDisplayer(defaultFormat format.Format, single bool, orgs ...*models.HashicorpCloudResourcemanagerOrganization) *displayer {
	return &displayer{
		orgs:          orgs,
		defaultFormat: defaultFormat,
		single:        single,
	}
}

func (d *displayer) DefaultFormat() format.Format {
	return d.defaultFormat
}

func (d *displayer) Payload() any {
	if d.single {
		if len(d.orgs) != 1 {
			return nil
		}

		return d.orgs[0]
	}

	return d.orgs
}

func (d *displayer) FieldTemplates() []format.Field {
	fields := []format.Field{
		{
			Name:        "Name",
			ValueFormat: "{{ .Name }}",
		},
		{
			Name:        "ID",
			ValueFormat: "{{ .ID }}",
		},
	}

	if d.single {
		fields = append(fields,
			format.Field{
				Name:        "Owner Principal ID",
				ValueFormat: "{{ .Owner.User }}",
			},
			format.Field{
				Name:        "State",
				ValueFormat: "{{ .State }}",
			},
		)
	}

	fields = append(fields, format.Field{
		Name:        "Created At",
		ValueFormat: "{{ .CreatedAt }}",
	})

	return fields
}
