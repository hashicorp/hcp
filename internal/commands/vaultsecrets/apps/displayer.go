// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apps

import (
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-11-28/models"
	"github.com/hashicorp/hcp/internal/pkg/format"
)

type displayer struct {
	apps []*models.Secrets20231128App

	single bool
}

func newDisplayer(single bool, apps ...*models.Secrets20231128App) *displayer {
	return &displayer{
		apps:   apps,
		single: single,
	}
}

func (d *displayer) DefaultFormat() format.Format {
	return format.Table
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
