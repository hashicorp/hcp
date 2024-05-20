// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package versions

import (
	preview_models "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/preview/2023-11-28/models"
	"github.com/hashicorp/hcp/internal/pkg/format"
)

type displayer struct {
	previewStaticSecretVerions []*preview_models.Secrets20231128SecretStaticVersion
	single                     bool
	fields                     []format.Field
	format                     format.Format
}

func newDisplayer(single bool) *displayer {
	return &displayer{
		single: single,
		format: format.Table,
	}
}

func (d *displayer) StaticVersions(secrets ...*preview_models.Secrets20231128SecretStaticVersion) *displayer {
	d.previewStaticSecretVerions = secrets
	return d
}

func (d *displayer) SetDefaultFormat(f format.Format) *displayer {
	d.format = f
	return d
}

func (d *displayer) DefaultFormat() format.Format {
	return d.format
}

func (d *displayer) Payload() any {
	return d.previewStaticSecretVerionsPayload()

}

func (d *displayer) FieldTemplates() []format.Field {
	return d.previewStaticSecretVersionsFieldTemplate()
}

func (displayer) previewStaticSecretVersionsFieldTemplate() []format.Field {
	return []format.Field{
		{
			Name:        "Version",
			ValueFormat: "{{ .Version }}",
		},
		{
			Name:        "Created At",
			ValueFormat: "{{ .CreatedAt }}",
		},
		{
			Name:        "Created By",
			ValueFormat: "{{ .CreatedBy.Email }}",
		},
	}
}

func (d *displayer) previewStaticSecretVerionsPayload() any {
	if d.single {
		if len(d.previewStaticSecretVerions) != 1 {
			return nil
		}
		return d.previewStaticSecretVerions[0]
	}
	return d.previewStaticSecretVerions
}
