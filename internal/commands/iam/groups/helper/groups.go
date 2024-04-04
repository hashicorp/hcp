// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package helper

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/client/groups_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/models"
	"github.com/hashicorp/hcp/internal/pkg/api/resourcename"
	"github.com/posener/complete"
)

var (
	// GroupResourceName is a regex that matches a group resource name
	GroupResourceName = regexp.MustCompile(`^iam/organization/.+/group/.+$`)
)

const (
	// GroupNameArgDoc is the documentation for accepting a group name as an
	// argument.
	GroupNameArgDoc = `
	The name of the group to %s. The name may be specified as either:

	{{ PreserveNewLines }}
	* The group's resource name. Formatted as {{ template "mdCodeOrBold" "iam/organization/ORG_ID/group/GROUP_NAME" }}
	* The resource name suffix, {{ template "mdCodeOrBold" "GROUP_NAME" }}.
	{{ PreserveNewLines }}
	`
)

func ResourceName(groupName, orgID string) string {
	rn := groupName
	if !GroupResourceName.MatchString(rn) {
		rn = fmt.Sprintf("iam/organization/%s/group/%s", orgID, groupName)
	}

	return rn
}

// PredictGroupResourceNameSuffix is an argument prediction function that predicts a group
// resource name suffix.
func PredictGroupResourceNameSuffix(ctx context.Context, orgID string, client groups_service.ClientService) complete.PredictFunc {
	return func(args complete.Args) []string {
		if len(args.Completed) > 1 {
			return nil
		}

		groups, err := GetGroups(ctx, orgID, client)
		if err != nil {
			return nil
		}

		names := make([]string, len(groups))
		for i, g := range groups {
			_, parts, err := resourcename.Parse(g.ResourceName)
			if err != nil {
				return nil
			}

			names[i] = parts[len(parts)-1].Name
		}

		return names
	}
}

// GetGroups retrieves the groups in the organization.
func GetGroups(ctx context.Context, orgID string, client groups_service.ClientService) ([]*models.HashicorpCloudIamGroup, error) {
	req := groups_service.NewGroupsServiceListGroupsParamsWithContext(ctx)
	req.ParentResourceName = fmt.Sprintf("organization/%s", orgID)

	var groups []*models.HashicorpCloudIamGroup
	for {

		resp, err := client.GroupsServiceListGroups(req, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to list groups: %w", err)
		}
		groups = append(groups, resp.Payload.Groups...)
		if resp.Payload.Pagination == nil || resp.Payload.Pagination.NextPageToken == "" {
			break
		}

		next := resp.Payload.Pagination.NextPageToken
		req.PaginationNextPageToken = &next
	}

	return groups, nil
}
