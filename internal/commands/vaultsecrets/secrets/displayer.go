// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package secrets

import (
	"fmt"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-06-13/models"
	preview_models "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-11-28/models"
	"github.com/hashicorp/hcp/internal/pkg/format"
)

const multiValueSecretFmt = "{{ range $key, $value := %s }}{{printf \"%%s: %%s\\n\" $key $value}}{{ end }}"

type displayer struct {
	secrets              []*models.Secrets20230613Secret
	previewSecrets       []*preview_models.Secrets20231128Secret
	openAppSecrets       []*preview_models.Secrets20231128OpenSecret
	secretType           string
	secretTypeFormatters map[string]secretTypeFormatter
	fields               []format.Field
	format               format.Format
}

type secretTypeFormatter interface {
	fields() []format.Field
}

type kvSecretFormatter struct{}

func (k kvSecretFormatter) fields() []format.Field {
	return []format.Field{
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
	}
}

type dynamicSecretFormatter struct{}

func (k dynamicSecretFormatter) fields() []format.Field {
	return []format.Field{
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
	}
}

type rotatingSecretFormatter struct{}

func (k rotatingSecretFormatter) fields() []format.Field {
	return []format.Field{
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
	}
}

func newDisplayer() *displayer {
	return &displayer{
		format: format.Table,
		secretTypeFormatters: map[string]secretTypeFormatter{
			secretTypeKV:       kvSecretFormatter{},
			secretTypeDynamic:  dynamicSecretFormatter{},
			secretTypeRotating: rotatingSecretFormatter{},
		},
	}
}

func (d *displayer) Secrets(secrets ...*models.Secrets20230613Secret) *displayer {
	d.secrets = secrets
	if len(secrets) == 1 {
		d.secretType = secretTypeKV
	}
	return d
}

func (d *displayer) PreviewSecrets(secrets ...*preview_models.Secrets20231128Secret) *displayer {
	d.previewSecrets = secrets
	if len(secrets) == 1 {
		d.secretType = secrets[0].Type
	}
	return d
}

func (d *displayer) OpenAppSecrets(secrets ...*preview_models.Secrets20231128OpenSecret) *displayer {
	d.openAppSecrets = secrets
	if len(secrets) == 1 {
		d.secretType = secrets[0].Type
	}
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

	secretTypeFormatter := d.secretTypeFormatters[d.secretType]
	return append(fields, secretTypeFormatter.fields()...)
}

func (d *displayer) secretsPayload() any {
	if len(d.secrets) == 1 {
		return d.secrets[0]
	}
	return d.secrets
}

func (d *displayer) previewSecretsPayload() any {
	if len(d.previewSecrets) == 1 {
		return d.previewSecrets[0]
	}
	return d.previewSecrets
}

func (d *displayer) openAppSecretsPayload() any {
	if len(d.openAppSecrets) == 1 {
		return d.openAppSecrets[0]
	}
	return d.openAppSecrets
}

type rotatingSecretsDisplayer struct {
	previewRotatingSecrets []*preview_models.Secrets20231128RotatingSecretConfig
	single                 bool

	format format.Format
}

func newRotatingSecretsDisplayer(single bool) *rotatingSecretsDisplayer {
	return &rotatingSecretsDisplayer{
		single: single,
		format: format.Table,
	}
}

func (r *rotatingSecretsDisplayer) PreviewRotatingSecrets(secrets ...*preview_models.Secrets20231128RotatingSecretConfig) *rotatingSecretsDisplayer {
	r.previewRotatingSecrets = secrets
	return r
}

func (r *rotatingSecretsDisplayer) DefaultFormat() format.Format {
	return r.format
}

func (r *rotatingSecretsDisplayer) Payload() any {
	if r.single {
		if len(r.previewRotatingSecrets) != 1 {
			return nil
		}
		return r.previewRotatingSecrets[0]
	}
	return r.previewRotatingSecrets
}

func (r *rotatingSecretsDisplayer) FieldTemplates() []format.Field {
	return []format.Field{
		{
			Name:        "Secret Name",
			ValueFormat: "{{ .Name }}",
		},
		{
			Name:        " Integration Name",
			ValueFormat: "{{ .IntegrationName }}",
		},
		{
			Name:        " Rotation Policy",
			ValueFormat: "{{ .RotationPolicyName }}",
		},
	}
}
