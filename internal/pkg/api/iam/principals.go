// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"

	iam "github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/client/iam_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/models"
)

const (
	// maxBatchGetPrincipalsSize is the maximum number of principals that
	// can be retrieved in a given batch.
	maxBatchGetPrincipalsSize = 1000
)

// BatchGetPrincipals retrieves the requested principals in a batch. If the
// number of principals exceeds the batch limit, multiple requests will be made.
func BatchGetPrincipals(ctx context.Context, organizationID string, client iam.ClientService, principals []string, view *models.HashicorpCloudIamPrincipalView) ([]*models.HashicorpCloudIamPrincipal, error) {
	var allPrincipals []*models.HashicorpCloudIamPrincipal

	n := len(principals)
	for i := 0; i < n; i += maxBatchGetPrincipalsSize {
		params := iam.NewIamServiceBatchGetPrincipalsParams()
		params.OrganizationID = organizationID
		params.View = (*string)(view.Pointer())
		params.PrincipalIds = principals[i:min(i+maxBatchGetPrincipalsSize, n)]

		resp, err := client.IamServiceBatchGetPrincipals(params, nil)
		if err != nil {
			return nil, err
		}

		allPrincipals = append(allPrincipals, resp.Payload.Principals...)
	}

	return allPrincipals, nil
}
