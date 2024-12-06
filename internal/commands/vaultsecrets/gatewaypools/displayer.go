// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package gatewaypools

import (
	"github.com/hashicorp/hcp-sdk-go/auth"
	preview_models "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-11-28/models"
	"github.com/hashicorp/hcp/internal/pkg/format"
)

type gatewayPoolWithIntegrations struct {
	GatewayPool  *preview_models.Secrets20231128GatewayPool `json:"gateway_pool"`
	Integrations []string                                   `json:"integrations,omitempty"`
}

type gatewayPoolWithOauth struct {
	GatewayPool *preview_models.Secrets20231128GatewayPool `json:"gateway_pool"`
	Oauth       *auth.OauthConfig                          `json:"oauth,omitempty"`
}

type displayer struct {
	gatewayPools                 []*preview_models.Secrets20231128GatewayPool
	gatewayPoolsWithIntegrations *gatewayPoolWithIntegrations
	gatewayPoolWithOauth         *gatewayPoolWithOauth
	gateways                     []*preview_models.Secrets20231128Gateway

	fields []format.Field
	format format.Format
	single bool
}

func newDisplayer(single bool) *displayer {
	return &displayer{
		single: single,
		format: format.Table,
	}
}

func (d *displayer) GatewayPools(gatewayPools ...*preview_models.Secrets20231128GatewayPool) *displayer {
	d.gatewayPools = gatewayPools
	return d
}

func (d *displayer) GatewayPoolWithIntegrations(gatewayPool *preview_models.Secrets20231128GatewayPool, integrations ...string) *displayer {
	d.gatewayPoolsWithIntegrations = &gatewayPoolWithIntegrations{
		GatewayPool:  gatewayPool,
		Integrations: integrations,
	}

	return d
}

func (d *displayer) Gateways(gateways ...*preview_models.Secrets20231128Gateway) *displayer {
	d.gateways = gateways
	return d
}

func (d *displayer) GatewayPoolsWithOauth(gpo *gatewayPoolWithOauth) *displayer {
	d.gatewayPoolWithOauth = gpo
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

func (d *displayer) DefaultFormat() format.Format {
	return d.format
}

func (d *displayer) previewGatewayPools() any {
	if d.single {
		if len(d.gatewayPools) != 1 {
			return nil
		}
		return d.gatewayPools[0]
	}

	return d.gatewayPools
}

func (d *displayer) previewGateways() any {
	if d.single {
		if len(d.gateways) != 1 {
			return nil
		}
		return d.gateways[0]
	}

	return d.gateways
}

func (d *displayer) Payload() any {
	if d.gatewayPools != nil {
		return d.previewGatewayPools()
	}
	if d.gatewayPoolsWithIntegrations != nil {
		return d.gatewayPoolsWithIntegrations
	}
	if d.gateways != nil {
		return d.previewGateways()
	}
	if d.gatewayPoolWithOauth != nil {
		return d.gatewayPoolWithOauth
	}

	return nil
}

func (d *displayer) FieldTemplates() []format.Field {
	return d.fields
}
