package iam

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/project_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/models"
	"github.com/hashicorp/hcp/internal/pkg/api/iampolicy"
)

// iamUpdater meets the iampolicy.ResourceUpdater interface. It is used to
// manage IAM bindings.
type iamUpdater struct {
	projectID string
	client    project_service.ClientService
}

func (u *iamUpdater) GetIamPolicy(ctx context.Context) (*models.HashicorpCloudResourcemanagerPolicy, error) {
	params := project_service.NewProjectServiceGetIamPolicyParams()
	params.ID = u.projectID
	res, err := u.client.ProjectServiceGetIamPolicy(params, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve project IAM policy: %w", err)
	}

	return res.GetPayload().Policy, nil
}

func (u *iamUpdater) SetIamPolicy(ctx context.Context, policy *models.HashicorpCloudResourcemanagerPolicy) (*models.HashicorpCloudResourcemanagerPolicy, error) {
	params := project_service.NewProjectServiceSetIamPolicyParams()
	params.ID = u.projectID
	params.Body = project_service.ProjectServiceSetIamPolicyBody{
		Policy: policy,
	}

	res, err := u.client.ProjectServiceSetIamPolicy(params, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to set project IAM policy: %w", err)
	}

	return res.GetPayload().Policy, nil
}

// Ensure we meet the interface.
var _ iampolicy.ResourceUpdater = &iamUpdater{}
