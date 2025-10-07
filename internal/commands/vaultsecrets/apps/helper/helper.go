// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: MPL-2.0

package helper

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-11-28/client/secret_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-11-28/models"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/posener/complete"
)

// PredictAppName returns a predict function for application names.
func PredictAppName(ctx *cmd.Context, c *cmd.Command, client secret_service.ClientService) complete.PredictFunc {
	return func(args complete.Args) []string {
		// Parse the args
		remainingArgs, err := ctx.ParseFlags(c, args.All)
		if err != nil {
			return nil
		}

		if len(remainingArgs) > 1 {
			return nil
		}

		apps, err := getApps(ctx.ShutdownCtx, ctx.Profile.OrganizationID, ctx.Profile.ProjectID, client)
		if err != nil {
			return nil
		}

		names := make([]string, len(apps))
		for i, d := range apps {
			names[i] = d.Name
		}

		return names
	}
}

func getApps(ctx context.Context, orgID, projectID string, client secret_service.ClientService) ([]*models.Secrets20231128App, error) {
	req := secret_service.NewListAppsParamsWithContext(ctx)
	req.OrganizationID = orgID
	req.ProjectID = projectID
	var apps []*models.Secrets20231128App
	for {

		resp, err := client.ListApps(req, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to list groups: %w", err)
		}
		apps = append(apps, resp.Payload.Apps...)
		if resp.Payload.Pagination == nil || resp.Payload.Pagination.NextPageToken == "" {
			break
		}

		next := resp.Payload.Pagination.NextPageToken
		req.PaginationNextPageToken = &next
	}

	return apps, nil
}
