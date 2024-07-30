// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package integrations

import (
	preview_models "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/preview/2023-11-28/models"
	"github.com/hashicorp/hcp/internal/pkg/format"
)

type twilioDisplayer struct {
	previewTwilioIntegrations []*preview_models.Secrets20231128TwilioIntegration

	single bool
}

func newTwilioDisplayer(single bool, integrations ...*preview_models.Secrets20231128TwilioIntegration) *twilioDisplayer {
	return &twilioDisplayer{
		previewTwilioIntegrations: integrations,
		single:                    single,
	}
}

func (t *twilioDisplayer) DefaultFormat() format.Format {
	return format.Table
}

func (t *twilioDisplayer) Payload() any {
	if t.previewTwilioIntegrations != nil {
		return t.previewTwilioIntegrationsPayload()
	}

	return nil
}

func (t *twilioDisplayer) previewTwilioIntegrationsPayload() any {
	if t.single {
		if len(t.previewTwilioIntegrations) != 1 {
			return nil
		}
		return t.previewTwilioIntegrations[0]
	}
	return t.previewTwilioIntegrations
}

func (t *twilioDisplayer) FieldTemplates() []format.Field {
	fields := []format.Field{
		{
			Name:        "Integration Name",
			ValueFormat: "{{ .IntegrationName }}",
		},
	}

	if t.single {
		return append(fields, []format.Field{
			{
				Name:        "Account SID",
				ValueFormat: "{{ .TwilioAccountSid }}",
			},
		}...)
	} else {
		return fields
	}
}

type mongodbDisplayer struct {
	previewMongoDBIntegrations []*preview_models.Secrets20231128MongoDBAtlasIntegration

	single bool
}

func newMongoDBDisplayer(single bool, integrations ...*preview_models.Secrets20231128MongoDBAtlasIntegration) *mongodbDisplayer {
	return &mongodbDisplayer{
		previewMongoDBIntegrations: integrations,
		single:                     single,
	}
}

func (m *mongodbDisplayer) DefaultFormat() format.Format {
	return format.Table
}

func (m *mongodbDisplayer) Payload() any {

	if m.previewMongoDBIntegrations != nil {
		return m.previewMongoDBIntegrationsPayload()
	}

	return nil
}

func (m *mongodbDisplayer) previewMongoDBIntegrationsPayload() any {
	if m.single {
		if len(m.previewMongoDBIntegrations) != 1 {
			return nil
		}
		return m.previewMongoDBIntegrations[0]
	}
	return m.previewMongoDBIntegrations
}

func (m *mongodbDisplayer) FieldTemplates() []format.Field {
	fields := []format.Field{
		{
			Name:        "Integration Name",
			ValueFormat: "{{ .IntegrationName }}",
		},
	}

	if m.single {
		return append(fields, []format.Field{
			{
				Name:        "API Public Key",
				ValueFormat: "{{ .MongodbAPIPublicKey }}",
			},
		}...)
	} else {
		return fields
	}
}

type genericDisplayer struct {
	integrations []*preview_models.Secrets20231128Integration

	single bool
}

func newGenericDisplayer(single bool, integrations ...*preview_models.Secrets20231128Integration) *genericDisplayer {
	return &genericDisplayer{
		integrations: integrations,
		single:       single,
	}
}

func (g *genericDisplayer) DefaultFormat() format.Format {
	return format.Table
}

func (g *genericDisplayer) Payload() any {
	if g.integrations != nil {
		return g.integrationsPayload()
	}

	return nil
}

func (g *genericDisplayer) integrationsPayload() any {
	if g.single {
		if len(g.integrations) != 1 {
			return nil
		}
		return g.integrations[0]
	}
	return g.integrations
}

func (g *genericDisplayer) FieldTemplates() []format.Field {
	return []format.Field{
		{
			Name:        "Integration Name",
			ValueFormat: "{{ .IntegrationName }}",
		},
	}
}
