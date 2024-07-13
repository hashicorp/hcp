// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package gateway_pools

import (
	"context"
	"fmt"

	preview_secret_service "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/preview/2023-11-28/client/secret_service"
	preview_models "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/preview/2023-11-28/models"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/posener/complete"
)

// PredictGatewayPoolName returns a predict function for gateway pools names.
func PredictGatewayPoolName(ctx *cmd.Context, c *cmd.Command, client preview_secret_service.ClientService) complete.PredictFunc {
	return func(args complete.Args) []string {
		// Parse the args
		remainingArgs, err := ctx.ParseFlags(c, args.All)
		if err != nil {
			return nil
		}

		if len(remainingArgs) > 1 {
			return nil
		}

		gwps, err := listGatewayPools(ctx.ShutdownCtx, ctx.Profile.OrganizationID, ctx.Profile.ProjectID, client)
		if err != nil {
			return nil
		}

		names := make([]string, len(gwps))
		for i, d := range gwps {
			names[i] = d.Name
		}

		return names
	}
}

func listGatewayPools(ctx context.Context, orgID, projectID string, client preview_secret_service.ClientService) ([]*preview_models.Secrets20231128GatewayPool, error) {
	req := preview_secret_service.NewListGatewayPoolsParamsWithContext(ctx)
	req.OrganizationID = orgID
	req.ProjectID = projectID

	resp, err := client.ListGatewayPools(req, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list gateway pools: %w", err)
	}
	if resp.Payload == nil || resp.Payload.GatewayPools == nil {
		return nil, fmt.Errorf("failed to list gateway pools: empty response")
	}

	return resp.Payload.GatewayPools, nil
}
