// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package secrets

import (
	"fmt"

	preview_models "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/preview/2023-11-28/models"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-06-13/models"
	"github.com/hashicorp/hcp/internal/pkg/format"
)

const multiValueSecretFmt = "{{ range $key, $value := %s }}{{printf \"%%s: %%s\\n\" $key $value}}{{ end }}"

type displayer struct {
	secrets        []*models.Secrets20230613Secret
	previewSecrets []*preview_models.Secrets20231128Secret
	openAppSecrets []*preview_models.Secrets20231128OpenSecret
	secretType     string
	fields         []format.Field
	format         format.Format
}

func newDisplayer() *displayer {
	return &displayer{
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

func (d *displayer) OpenAppSecrets(secrets ...*preview_models.Secrets20231128OpenSecret) *displayer {
	d.openAppSecrets = secrets
	return d
}

func (d *displayer) AddFields(fields ...format.Field) *displayer {
	d.fields = append(d.fields, fields...)
	return d
}

func (d *displayer) SetDefaultFormat(f format.Format) *displayer {
	d.format = f
	return d
}

func (d *displayer) SetSecretType(s string) *displayer {
	d.secretType = s
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

	if d.secrets == nil {
		return nil
	}
	return d.secretsPayload()
}

func (d *displayer) FieldTemplates() []format.Field {
	fields := d.secretsFieldTemplate()
	if d.openAppSecrets != nil {
		fields = d.openAppSecretsFieldTemplate()
	}

	return append(fields, d.fields...)
}

func (d *displayer) secretsFieldTemplate() []format.Field {
	fields := []format.Field{
		{
			Name:        "Secret Name",
			ValueFormat: "{{ .Name }}",
		},
	}

	if len(d.previewSecrets) > 0 {
		fields = append(fields, format.Field{
			Name:        "Type",
			ValueFormat: "{{ .Type }}",
		})
	}

	fields = append(fields, []format.Field{
		{
			Name:        "Created At",
			ValueFormat: "{{ .CreatedAt }}",
		},
		{
			Name:        "Latest Version",
			ValueFormat: "{{ if eq (printf \"%v\" .LatestVersion) \"0\" }}-{{ else }}{{ .LatestVersion }}{{ end }}",
		},
	}...)

	return fields
}

func (d *displayer) openAppSecretsFieldTemplate() []format.Field {
	fields := []format.Field{
		{
			Name:        "Secret Name",
			ValueFormat: "{{ .Name }}",
		},
		{
			Name:        "Type",
			ValueFormat: "{{ .Type }}",
		},
	}

	// Secret type specific fields
	switch d.secretType {
	case secretTypeKV:
		fields = append(fields, []format.Field{
			{
				Name:        "Created At",
				ValueFormat: "{{ .CreatedAt }}",
			},
			{
				Name:        "Latest Version",
				ValueFormat: "{{ .LatestVersion }}",
			},
			{
				Name:        "Value",
				ValueFormat: "{{ .StaticVersion.Value }}",
			},
		}...)
	case secretTypeDynamic:
		fields = append(fields, []format.Field{
			{
				Name:        "Created At",
				ValueFormat: "{{ .DynamicInstance.CreatedAt }}",
			},
			{
				Name:        "Expires At",
				ValueFormat: "{{ .DynamicInstance.ExpiresAt }}",
			},
			{
				Name:        "Time-to-Live",
				ValueFormat: "{{ .DynamicInstance.TTL }}",
			},
			{
				Name:        "Values",
				ValueFormat: fmt.Sprintf(multiValueSecretFmt, ".DynamicInstance.Values"),
			},
		}...)
	case secretTypeRotating:
		fields = append(fields, []format.Field{
			{
				Name:        "Created At",
				ValueFormat: "{{ .CreatedAt }}",
			},
			{
				Name:        "Expires At",
				ValueFormat: "{{ .RotatingVersion.ExpiresAt }}",
			},
			{
				Name:        "Latest Version",
				ValueFormat: "{{ .LatestVersion }}",
			},
			{
				Name:        "Values",
				ValueFormat: fmt.Sprintf(multiValueSecretFmt, ".RotatingVersion.Values"),
			},
		}...)
	}

	return fields
}

func (d *displayer) secretsPayload() any {
	if len(d.secrets) > 1 {
		return d.secrets
	}
	if len(d.secrets) == 1 {
		return d.secrets[0]
	}
	return nil
}

func (d *displayer) previewSecretsPayload() any {
	if len(d.previewSecrets) > 1 {
		return d.previewSecrets
	}
	if len(d.previewSecrets) == 1 {
		return d.previewSecrets[0]
	}
	return nil
}

func (d *displayer) openAppSecretsPayload() any {
	if len(d.openAppSecrets) > 1 {
		return d.openAppSecrets
	}
	if len(d.openAppSecrets) == 1 {
		return d.openAppSecrets[0]
	}
	return nil
}
