// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package integrations

import (
	preview_models "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/preview/2023-11-28/models"
	"github.com/hashicorp/hcp/internal/pkg/format"
)

type displayer struct {
	previewTwilioIntegrations  []*preview_models.Secrets20231128TwilioIntegration
	previewMongoDBIntegrations []*preview_models.Secrets20231128MongoDBAtlasIntegration

	single bool
}

func newTwilioDisplayer(single bool, integrations ...*preview_models.Secrets20231128TwilioIntegration) *displayer {
	return &displayer{
		previewTwilioIntegrations: integrations,
		single:                    single,
	}
}

func newMongoDBDisplayer(single bool, integrations ...*preview_models.Secrets20231128MongoDBAtlasIntegration) *displayer {
	return &displayer{
		previewMongoDBIntegrations: integrations,
		single:                     single,
	}
}

func (d *displayer) DefaultFormat() format.Format {
	return format.Table
}

func (d *displayer) Payload() any {
	if d.previewTwilioIntegrations != nil {
		return d.previewTwilioIntegrationsPayload()
	}

	if d.previewMongoDBIntegrations != nil {
		return d.previewMongoDBIntegrationsPayload()
	}

	return nil
}

func (d *displayer) previewTwilioIntegrationsPayload() any {
	if d.single {
		if len(d.previewTwilioIntegrations) != 1 {
			return nil
		}
		return d.previewTwilioIntegrations[0]
	}
	return d.previewTwilioIntegrations
}

func (d *displayer) previewMongoDBIntegrationsPayload() any {
	if d.single {
		if len(d.previewMongoDBIntegrations) != 1 {
			return nil
		}
		return d.previewMongoDBIntegrations[0]
	}
	return d.previewMongoDBIntegrations
}

func (d *displayer) FieldTemplates() []format.Field {
	return []format.Field{
		{
			Name:        "Integration Name",
			ValueFormat: "{{ .IntegrationName }}",
		},
		{
			Name:        "Account SID",
			ValueFormat: "{{ .TwilioAccountSid }}",
		},
	}
}
