// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package gateway_pools

import (
	preview_models "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/preview/2023-11-28/models"
	"github.com/hashicorp/hcp/internal/pkg/format"
)

type displayer struct {
	gatewayPools []*preview_models.Secrets20231128GatewayPool

	single bool
}

func newDisplayer(single bool, gatewayPools ...*preview_models.Secrets20231128GatewayPool) *displayer {
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
			ValueFormat: "{{ .GatewayPoolName }}",
		},
		{
			Name:        "Description",
			ValueFormat: "{{ .Description }}",
		},
	}
}
