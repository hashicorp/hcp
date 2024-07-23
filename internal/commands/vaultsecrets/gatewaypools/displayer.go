// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package gatewaypools

import (
	preview_models "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/preview/2023-11-28/models"
	"github.com/hashicorp/hcp/internal/pkg/format"
)

type gatewayPoolWithIntegrations struct {
	GatewayPool  *preview_models.Secrets20231128GatewayPool
	Gateways     []*preview_models.Secrets20231128Gateway
	Integrations []string
}

type displayer struct {
	gatewayPools []*gatewayPoolWithIntegrations

	single bool
}

func newDisplayer(single bool, gatewayPools ...*gatewayPoolWithIntegrations) *displayer {
	return &displayer{
		gatewayPools: gatewayPools,
		single:       single,
	}
}

func (d *displayer) DefaultFormat() format.Format {
	return format.Table
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
	return []format.Field{
		{
			Name:        "GatewayPool Name",
			ValueFormat: "{{ .Name }}",
		},
		{
			Name:        "Description",
			ValueFormat: "{{ .Description }}",
		},
	}
}
