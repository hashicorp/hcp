// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package serviceprincipals

import (
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/models"
	"github.com/hashicorp/hcp/internal/pkg/format"
)

type displayer struct {
	sps           []*models.HashicorpCloudIamServicePrincipal
	defaultFormat format.Format
	single        bool
}

func newDisplayer(defaultFormat format.Format, single bool, sps ...*models.HashicorpCloudIamServicePrincipal) *displayer {
	return &displayer{
		sps:           sps,
		defaultFormat: defaultFormat,
		single:        single,
	}
}

func (d *displayer) DefaultFormat() format.Format {
	return d.defaultFormat
}

func (d *displayer) Payload() any {
	if d.single {
		if len(d.sps) != 1 {
			return nil
		}

		return d.sps[0]
	}

	return d.sps
}

func (d *displayer) FieldTemplates() []format.Field {
	return []format.Field{
		{
			Name:        "Resource Name",
			ValueFormat: "{{ .ResourceName }}",
		},
		{
			Name:        "Resource ID",
			ValueFormat: "{{ .ID }}",
		},
		{
			Name:        "Display Name",
			ValueFormat: "{{ .Name }}",
		},
		{
			Name:        "Created At",
			ValueFormat: "{{ .CreatedAt }}",
		},
	}
}
