// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package keys

import (
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/models"
	"github.com/hashicorp/hcp/internal/pkg/format"
)

type displayer struct {
	keys          []*models.HashicorpCloudIamServicePrincipalKey
	defaultFormat format.Format
	single        bool
	secret        string
}

func newDisplayer(defaultFormat format.Format, keys ...*models.HashicorpCloudIamServicePrincipalKey) *displayer {
	return &displayer{
		keys:          keys,
		defaultFormat: defaultFormat,
		single:        false,
	}
}

func newSecretDisplayer(defaultFormat format.Format, key *models.HashicorpCloudIamServicePrincipalKey, secret string) *displayer {
	return &displayer{
		keys:          []*models.HashicorpCloudIamServicePrincipalKey{key},
		defaultFormat: defaultFormat,
		single:        true,
		secret:        secret,
	}
}

func (d *displayer) DefaultFormat() format.Format {
	return d.defaultFormat
}

func (d *displayer) Payload() any {
	if d.single {
		if len(d.keys) != 1 {
			return nil
		}

		return d.keys[0]
	}

	return d.keys
}

func (d *displayer) FieldTemplates() []format.Field {
	fields := []format.Field{
		{
			Name:        "Resource Name",
			ValueFormat: "{{ .ResourceName }}",
		},
		{
			Name:        "Client ID",
			ValueFormat: "{{ .ClientID }}",
		},
		{
			Name:        "Created At",
			ValueFormat: "{{ .CreatedAt }}",
		},
	}

	if d.secret != "" {
		secret := format.Field{Name: "Client Secret", ValueFormat: d.secret}
		fields = append(fields[:2], append([]format.Field{secret}, fields[2:]...)...)
	}

	return fields
}
