// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package workloadidentityproviders

import (
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/models"
	"github.com/hashicorp/hcp/internal/pkg/format"
)

type displayer struct {
	wips          []*models.HashicorpCloudIamWorkloadIdentityProvider
	defaultFormat format.Format
	single        bool
}

func newDisplayer(defaultFormat format.Format, single bool, wips ...*models.HashicorpCloudIamWorkloadIdentityProvider) *displayer {
	return &displayer{
		wips:          wips,
		defaultFormat: defaultFormat,
		single:        single,
	}
}

func (d *displayer) DefaultFormat() format.Format {
	return d.defaultFormat
}

func (d *displayer) Payload() any {
	if d.single {
		if len(d.wips) != 1 {
			return nil
		}

		return d.wips[0]
	}

	return d.wips
}

func (d *displayer) FieldTemplates() []format.Field {
	return []format.Field{
		{
			Name:        "Resource Name",
			ValueFormat: "{{ .ResourceName }}",
		},
		{
			Name:        "Resource ID",
			ValueFormat: "{{ .ResourceID }}",
		},
		{
			Name:        "AWS Account ID",
			ValueFormat: "{{ if .AwsConfig -}} {{ .AwsConfig.AccountID }} {{- end }}",
		},
		{
			Name:        "OIDC Issuer",
			ValueFormat: "{{ if .OidcConfig -}} {{ .OidcConfig.IssuerURI }} {{- end }}",
		},
		{
			Name:        "OIDC Allowed Audiences",
			ValueFormat: "{{ if .OidcConfig -}} {{ .OidcConfig.AllowedAudiences }} {{- end }}",
		},
		{
			Name:        "Conditional Access",
			ValueFormat: "{{ .ConditionalAccess }}",
		},
		{
			Name:        "Description",
			ValueFormat: "{{ .Description }}",
		},
		{
			Name:        "Created At",
			ValueFormat: "{{ .CreatedAt }}",
		},
	}
}
