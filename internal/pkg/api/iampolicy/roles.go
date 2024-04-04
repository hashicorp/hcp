// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iampolicy

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/organization_service"
	"github.com/posener/complete"
)

type rolesAutcomplete struct {
	ctx      context.Context
	client   organization_service.ClientService
	orgID    string
	allRoles []string
}

func AutocompleteRoles(ctx context.Context, orgID string, client organization_service.ClientService) complete.Predictor {
	return &rolesAutcomplete{
		ctx:      ctx,
		client:   client,
		orgID:    orgID,
		allRoles: nil,
	}
}

func (r *rolesAutcomplete) Predict(complete.Args) []string {
	if err := r.getRoles(); err != nil {
		return nil
	}

	return r.allRoles
}

func (r *rolesAutcomplete) getRoles() error {
	req := organization_service.NewOrganizationServiceListRolesParamsWithContext(r.ctx)
	req.ID = r.orgID

	var roles []string
	for {
		resp, err := r.client.OrganizationServiceListRoles(req, nil)
		if err != nil {
			return fmt.Errorf("failed to list organization roles: %w", err)
		}

		for _, r := range resp.Payload.Roles {
			roles = append(roles, r.ID)
		}

		if resp.Payload.Pagination == nil || resp.Payload.Pagination.NextPageToken == "" {
			break
		}

		next := resp.Payload.Pagination.NextPageToken
		req.PaginationNextPageToken = &next
	}

	r.allRoles = roles
	return nil
}
