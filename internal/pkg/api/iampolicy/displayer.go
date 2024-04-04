// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iampolicy

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/client/iam_service"
	iamModels "github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/models"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/models"
	"github.com/hashicorp/hcp/internal/pkg/api/iam"
	"github.com/hashicorp/hcp/internal/pkg/format"

	"golang.org/x/exp/maps"
)

type Displayer struct {
	policy         *models.HashicorpCloudResourcemanagerPolicy
	client         iam_service.ClientService
	principalNames map[string]string
}

func NewDisplayer(ctx context.Context, orgID string, policy *models.HashicorpCloudResourcemanagerPolicy, client iam_service.ClientService) (*Displayer, error) {
	d := &Displayer{
		policy:         policy,
		client:         client,
		principalNames: make(map[string]string, 32),
	}

	return d, d.resolvePrincipals(ctx, orgID)
}

func (d *Displayer) resolvePrincipals(ctx context.Context, orgID string) error {
	// Determine all the principal IDs
	principalIDs := make(map[string]struct{})
	for _, b := range d.policy.Bindings {
		for _, p := range b.Members {
			principalIDs[p.MemberID] = struct{}{}
		}
	}

	principals, err := iam.BatchGetPrincipals(ctx, orgID, d.client, maps.Keys(principalIDs), iamModels.HashicorpCloudIamPrincipalViewPRINCIPALVIEWFULL.Pointer())
	if err != nil {
		return fmt.Errorf("failed to resolve principals in IAM policy: %w", err)
	}

	for _, p := range principals {
		switch *p.Type {
		case iamModels.HashicorpCloudIamPrincipalTypePRINCIPALTYPEGROUP:
			d.principalNames[p.ID] = p.Group.DisplayName
		case iamModels.HashicorpCloudIamPrincipalTypePRINCIPALTYPESERVICE:
			d.principalNames[p.ID] = p.Service.Name
		case iamModels.HashicorpCloudIamPrincipalTypePRINCIPALTYPEUSER:
			d.principalNames[p.ID] = p.User.FullName
		case iamModels.HashicorpCloudIamPrincipalTypePRINCIPALTYPEUNSPECIFIED:
			return fmt.Errorf("invalid principal type for principal %q: %s", p.ID, *p.Type)
		}
	}

	return nil
}

func (d *Displayer) DefaultFormat() format.Format {
	return format.Table
}

// Payload is the object to display. Payload may return a single object or a
// slice of objects.
func (d *Displayer) Payload() any {
	return d.policy
}

// If we are displaying templated data, return the underlying bindings.
func (d *Displayer) TemplatedPayload() any {
	var bindings []*flattenedBinding
	for _, binding := range d.policy.Bindings {
		for _, member := range binding.Members {
			bindings = append(bindings, &flattenedBinding{
				RoleID:        binding.RoleID,
				PrincipalName: d.principalNames[member.MemberID],
				PrincipalID:   member.MemberID,
				PrincipalType: string(*member.MemberType),
			})
		}
	}
	return bindings
}

// FieldTemplates returns a slice of Fields. Each Field represents an field
// based on the payload to display to the user. It is common that the Field
// is simply a specific field of the payload struct being outputted.
func (d *Displayer) FieldTemplates() []format.Field {
	return []format.Field{
		{
			Name:        "Role ID",
			ValueFormat: "{{ .RoleID }}",
		},
		{
			Name:        "Principal Name",
			ValueFormat: "{{ .PrincipalName }}",
		},
		{
			Name:        "Principal ID",
			ValueFormat: "{{ .PrincipalID }}",
		},
		{
			Name:        "Principal Type",
			ValueFormat: "{{ .PrincipalType }}",
		},
	}
}

type flattenedBinding struct {
	RoleID        string
	PrincipalName string
	PrincipalID   string
	PrincipalType string
}

// Ensure we meet the interfaces
var _ format.Displayer = &Displayer{}
var _ format.TemplatedPayload = &Displayer{}
