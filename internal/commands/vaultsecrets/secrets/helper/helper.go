// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package helper

import (
	"context"
	"fmt"

	preview_secret_service "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/preview/2023-11-28/client/secret_service"
	preview_models "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/preview/2023-11-28/models"
	"github.com/hashicorp/hcp/internal/commands/vaultsecrets/secrets/appname"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/posener/complete"
)

// PredictAppName returns a predict function for application names.
func PredictSecretName(ctx *cmd.Context, c *cmd.Command, client preview_secret_service.ClientService) complete.PredictFunc {
	return func(args complete.Args) []string {
		// Parse the args
		remainingArgs, err := ctx.ParseFlags(c, args.All)
		if err != nil {
			return nil
		}

		if len(remainingArgs) > 1 {
			return nil
		}

		// Parse the app name from the flags or profile
		if err := appname.Require(ctx); err != nil {
			return nil
		}

		appName := appname.Get()
		secrets, err := getSecrets(ctx.ShutdownCtx, ctx.Profile.OrganizationID, ctx.Profile.ProjectID, appName, client)
		if err != nil {
			return nil
		}

		names := make([]string, len(secrets))
		for i, d := range secrets {
			names[i] = d.Name
		}
		return names
	}
}

func getSecrets(ctx context.Context, orgID, projectID, appName string, client preview_secret_service.ClientService) ([]*preview_models.Secrets20231128Secret, error) {
	req := preview_secret_service.NewListAppSecretsParamsWithContext(ctx)
	req.OrganizationID = orgID
	req.ProjectID = projectID
	req.AppName = appName

	var secrets []*preview_models.Secrets20231128Secret
	for {

		resp, err := client.ListAppSecrets(req, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to list groups: %w", err)
		}
		secrets = append(secrets, resp.Payload.Secrets...)
		if resp.Payload.Pagination == nil || resp.Payload.Pagination.NextPageToken == "" {
			break
		}

		next := resp.Payload.Pagination.NextPageToken
		req.PaginationNextPageToken = &next
	}

	return secrets, nil
}
