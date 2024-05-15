// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package secrets

import (
	preview_secret_models "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/preview/2023-11-28/models"
	models "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-06-13/models"
	"github.com/hashicorp/hcp/internal/pkg/format"
)

type displayer struct {
	secrets        []*models.Secrets20230613Secret
	previewSecrets []*preview_secret_models.Secrets20231128Secret
	single         bool
}

func newDisplayer(single bool, secrets ...*models.Secrets20230613Secret) *displayer {
	return &displayer{
		secrets: secrets,
		single:  single,
	}
}

func newDisplayerPreview(single bool, secrets ...*preview_secret_models.Secrets20231128Secret) *displayer {
	return &displayer{
		previewSecrets: secrets,
		single:         single,
	}
}

func (d *displayer) DefaultFormat() format.Format {
	return format.Table
}

func (d *displayer) Payload() any {
	if d.previewSecrets != nil {
		if d.single {
			if len(d.previewSecrets) != 1 {
				return nil
			}

			return d.previewSecrets[0]
		}

		return d.previewSecrets
	}

	if d.single {
		if len(d.secrets) != 1 {
			return nil
		}

		return d.secrets[0]
	}

	return d.secrets
}

func (d *displayer) FieldTemplates() []format.Field {
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
