// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package groups

import (
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/models"
	"github.com/hashicorp/hcp/internal/pkg/format"
)

type displayer struct {
	groups        []*models.HashicorpCloudIamGroup
	defaultFormat format.Format
	single        bool
}

func newDisplayer(defaultFormat format.Format, single bool, groups ...*models.HashicorpCloudIamGroup) *displayer {
	return &displayer{
		groups:        groups,
		defaultFormat: defaultFormat,
		single:        single,
	}
}

func (d *displayer) DefaultFormat() format.Format {
	return d.defaultFormat
}

func (d *displayer) Payload() any {
	if d.single {
		if len(d.groups) != 1 {
			return nil
		}

		return d.groups[0]
	}

	return d.groups
}

func (d *displayer) FieldTemplates() []format.Field {
	return []format.Field{
		{
			Name:        "Resource Name",
			ValueFormat: "{{ .ResourceName }}",
		},
		{
			Name:        "Resource ID",
			ValueFormat: "{{ .ResourceID }}",
		},
		{
			Name:        "Display Name",
			ValueFormat: "{{ .DisplayName }}",
		},
		{
			Name:        "Member Count",
			ValueFormat: "{{ .MemberCount }}",
		},
		{
			Name:        "Description",
			ValueFormat: "{{ .Description }}",
		},
	}
}
