// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package secrets

import (
	preview_models "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/preview/2023-11-28/models"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-06-13/models"
	"github.com/hashicorp/hcp/internal/pkg/format"
)

type displayer struct {
	secrets                    []*models.Secrets20230613Secret
	previewSecrets             []*preview_models.Secrets20231128Secret
	openAppSecrets             []*models.Secrets20230613OpenSecret
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

func (d *displayer) Secrets(secrets ...*models.Secrets20230613Secret) *displayer {
	d.secrets = secrets
	return d
}

func (d *displayer) PreviewSecrets(secrets ...*preview_models.Secrets20231128Secret) *displayer {
	d.previewSecrets = secrets
	return d
}

func (d *displayer) OpenAppSecrets(secrets ...*models.Secrets20230613OpenSecret) *displayer {
	d.openAppSecrets = secrets
	return d
}

func (d *displayer) StaticVersions(secrets ...*preview_models.Secrets20231128SecretStaticVersion) *displayer {
	d.previewStaticSecretVerions = secrets
	return d
}

func (d *displayer) AddFields(fields []format.Field) []format.Field {
	d.fields = append(d.fields, fields...)
	return d.fields
}

func (d *displayer) SetDefaultFormat(f format.Format) *displayer {
	d.format = f
	return d
}

func (d *displayer) DefaultFormat() format.Format {
	return d.format
}

func (d *displayer) Payload() any {
	if d.previewSecrets != nil {
		return d.previewSecretsPayload()
	}

	if d.openAppSecrets != nil {
		return d.openAppSecretsPayload()
	}

	if d.previewStaticSecretVerions != nil {
		return d.previewStaticSecretVerionsPayload()
	}

	if d.secrets == nil {
		return nil
	}
	return d.secretsPayload()
}

func (d *displayer) FieldTemplates() []format.Field {
	if d.openAppSecrets != nil {
		return d.openAppSecretsFieldTemplate()
	}

	if d.previewStaticSecretVerions != nil {
		return d.previewStaticSecretVerionsFieldTemplate()
	}

	return d.secretsFieldTemplate()
}

func (d *displayer) secretsFieldTemplate() []format.Field {
	return []format.Field{
		{
			Name:        "Secret Name",
			ValueFormat: "{{ .Name }}",
		},
		{
			Name:        "Latest Version",
			ValueFormat: "{{ .LatestVersion }}",
		},
		{
			Name:        "Created At",
			ValueFormat: "{{ .CreatedAt }}",
		},
	}
}

func (d *displayer) openAppSecretsFieldTemplate() []format.Field {
	fields := d.secretsFieldTemplate()
	fields = append(fields, format.Field{
		Name:        "Value",
		ValueFormat: "{{ .Version.Value }}",
	})
	return fields
}

func (displayer) previewStaticSecretVerionsFieldTemplate() []format.Field {
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
func (d *displayer) secretsPayload() any {
	if d.single {
		if len(d.secrets) != 1 {
			return nil
		}
		return d.secrets[0]
	}
	return d.secrets
}

func (d *displayer) previewSecretsPayload() any {
	if d.single {
		if len(d.previewSecrets) != 1 {
			return nil
		}
		return d.previewSecrets[0]
	}
	return d.previewSecrets
}

func (d *displayer) openAppSecretsPayload() any {
	if d.single {
		if len(d.openAppSecrets) != 1 {
			return nil
		}
		return d.openAppSecrets[0]
	}
	return d.openAppSecrets
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
