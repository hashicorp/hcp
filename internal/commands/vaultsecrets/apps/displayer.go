// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apps

import (
	preview_models "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/preview/2023-11-28/models"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-06-13/models"
	"github.com/hashicorp/hcp/internal/pkg/format"
)

type displayer struct {
	apps        []*models.Secrets20230613App
	previewApps []*preview_models.Secrets20231128App

	single bool
}

func newDisplayer(single bool, apps ...*models.Secrets20230613App) *displayer {
	return &displayer{
		apps:   apps,
		single: single,
	}
}

func newDisplayerPreview(single bool, apps ...*preview_models.Secrets20231128App) *displayer {
	return &displayer{
		previewApps: apps,
		single:      single,
	}
}

func (d *displayer) DefaultFormat() format.Format {
	return format.Table
}

func (d *displayer) Payload() any {
	if d.previewApps != nil {
		if d.single {
			if len(d.previewApps) != 1 {
				return nil
			}

			return d.previewApps[0]
		}

		return d.previewApps
	}

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
