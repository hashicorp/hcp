// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package gatewaypools

import (
	"github.com/hashicorp/hcp-sdk-go/auth"
	preview_models "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/preview/2023-11-28/models"
	"github.com/hashicorp/hcp/internal/pkg/format"
)

type gatewayPoolWithIntegrations struct {
	GatewayPool  *preview_models.Secrets20231128GatewayPool
	Oauth        *auth.OauthConfig
	Integrations []string
}

type displayer struct {
	gatewayPools []*gatewayPoolWithIntegrations

	extraFields []format.Field
	format      format.Format
	single      bool
}

func newDisplayer(single bool, gatewayPools ...*gatewayPoolWithIntegrations) *displayer {
	return &displayer{
		gatewayPools: gatewayPools,
		single:       single,
		format:       format.Table,
	}
}

func (d *displayer) AddExtraFields(fields ...format.Field) *displayer {
	d.extraFields = append(d.extraFields, fields...)
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
	if d.gatewayPools != nil {
		if d.single {
			if len(d.gatewayPools) != 1 {
				return nil
			}

			return d.gatewayPools[0]
		}

		return d.gatewayPools
	}

	if d.single {
		if len(d.gatewayPools) != 1 {
			return nil
		}

		return d.gatewayPools[0]
	}

	return d.gatewayPools
}

func (d *displayer) FieldTemplates() []format.Field {
	fields := []format.Field{
		{
			Name:        "GatewayPool Name",
			ValueFormat: "{{ .GatewayPool.Name }}",
		},
		{
			Name:        "Description",
			ValueFormat: "{{ .GatewayPool.Description }}",
		},
	}

	if d.extraFields != nil {
		fields = append(fields, d.extraFields...)
	}

	return fields
}
