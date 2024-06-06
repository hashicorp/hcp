// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/resource_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/models"
	"github.com/hashicorp/hcp/internal/pkg/api/iampolicy"
)

// iamUpdater meets the iampolicy.ResourceUpdater interface. It is used to
// manage IAM bindings.
type iamUpdater struct {
	resourceName string
	client       resource_service.ClientService
}

func (u *iamUpdater) GetIamPolicy(ctx context.Context) (*models.HashicorpCloudResourcemanagerPolicy, error) {
	params := resource_service.NewResourceServiceGetIamPolicyParams()
	params.ResourceName = &u.resourceName

	res, err := u.client.ResourceServiceGetIamPolicy(params, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve group IAM policy: %w", err)
	}

	return res.GetPayload().Policy, nil
}

func (u *iamUpdater) SetIamPolicy(ctx context.Context, policy *models.HashicorpCloudResourcemanagerPolicy) (*models.HashicorpCloudResourcemanagerPolicy, error) {
	params := resource_service.NewResourceServiceSetIamPolicyParams()
	params.Body = &models.HashicorpCloudResourcemanagerResourceSetIamPolicyRequest{
		ResourceName: u.resourceName,
		Policy:       policy,
	}

	res, err := u.client.ResourceServiceSetIamPolicy(params, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to set group IAM policy: %w", err)
	}

	return res.GetPayload().Policy, nil
}

// Ensure we meet the interface.
var _ iampolicy.ResourceUpdater = &iamUpdater{}
