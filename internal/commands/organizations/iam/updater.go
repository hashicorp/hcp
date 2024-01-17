package iam

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/organization_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/models"
	"github.com/hashicorp/hcp/internal/pkg/api/iampolicy"
)

// iamUpdater meets the iampolicy.ResourceUpdater interface. It is used to
// manage IAM bindings.
type iamUpdater struct {
	orgID  string
	client organization_service.ClientService
}

func (u *iamUpdater) GetIamPolicy(ctx context.Context) (*models.HashicorpCloudResourcemanagerPolicy, error) {
	params := organization_service.NewOrganizationServiceGetIamPolicyParams()
	params.ID = u.orgID
	res, err := u.client.OrganizationServiceGetIamPolicy(params, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve organization IAM policy: %w", err)
	}

	return res.GetPayload().Policy, nil
}

func (u *iamUpdater) SetIamPolicy(ctx context.Context, policy *models.HashicorpCloudResourcemanagerPolicy) (*models.HashicorpCloudResourcemanagerPolicy, error) {
	params := organization_service.NewOrganizationServiceSetIamPolicyParams()
	params.ID = u.orgID
	params.Body = organization_service.OrganizationServiceSetIamPolicyBody{
		Policy: policy,
	}

	res, err := u.client.OrganizationServiceSetIamPolicy(params, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to set organization IAM policy: %w", err)
	}

	return res.GetPayload().Policy, nil
}

// Ensure we meet the interface.
var _ iampolicy.ResourceUpdater = &iamUpdater{}
