package helper

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/client/service_principals_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/models"
	"github.com/hashicorp/hcp/internal/pkg/api/resourcename"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/posener/complete"
)

var (
	// SPResourceName is a regex that matches a service principal resource name
	SPResourceName = regexp.MustCompile(`^iam/(organization|project)/.+/service-principal/.+$`)
)

const (
	// SPNameArgDoc is the documentation for accepting a service principal name as an
	// argument.
	SPNameArgDoc = `
	The name of the service principal to %s. The name may be specified as either:

	  * The service principal's resource name. Formatted as:
		{{ Italic "iam/project/PROJECT_ID/service-principal/SP_NAME" }} or
		{{ Italic "iam/organization/ORG_ID/service-principal/SP_NAME" }}
	  * The resource name suffix, SP_NAME.
	`
)

func ResourceName(spName, orgID, projectID string) string {
	if SPResourceName.MatchString(spName) {
		return spName
	}

	if projectID == "" || projectID == "-" {
		return fmt.Sprintf("iam/organization/%s/service-principal/%s", orgID, spName)
	}

	return fmt.Sprintf("iam/project/%s/service-principal/%s", projectID, spName)
}

// PredictSPResourceNameSuffix is an argument prediction function that predicts
// a service-principal resource name suffix.
func PredictSPResourceNameSuffix(ctx *cmd.Context, c *cmd.Command, client service_principals_service.ClientService) complete.PredictFunc {
	return func(args complete.Args) []string {
		// Parse the args
		remainingArgs, err := ctx.ParseFlags(c, args.All)
		if err != nil {
			return nil
		}

		if len(remainingArgs) > 1 || len(args.Completed) > 0 {
			return nil
		}

		sps, err := GetSPs(ctx.ShutdownCtx, ctx.Profile.OrganizationID, ctx.Profile.ProjectID, client)
		if err != nil {
			return nil
		}

		names := make([]string, len(sps))
		for i, g := range sps {
			_, parts, err := resourcename.Parse(g.ResourceName)
			if err != nil {
				return nil
			}

			names[i] = parts[len(parts)-1].Name
		}

		return names
	}
}

// GetSPs retrieves the service principals in the organization or project. If
// project is unset or set to "-", the organization service principals will be
// retrieved.
func GetSPs(ctx context.Context, orgID, projectID string, client service_principals_service.ClientService) ([]*models.HashicorpCloudIamServicePrincipal, error) {
	req := service_principals_service.NewServicePrincipalsServiceListServicePrincipalsParamsWithContext(ctx)
	req.ParentResourceName = fmt.Sprintf("organization/%s", orgID)
	if projectID != "" && projectID != "-" {
		req.ParentResourceName = fmt.Sprintf("project/%s", projectID)
	}

	var sps []*models.HashicorpCloudIamServicePrincipal
	for {

		resp, err := client.ServicePrincipalsServiceListServicePrincipals(req, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to list service principals: %w", err)
		}
		sps = append(sps, resp.Payload.ServicePrincipals...)
		if resp.Payload.Pagination == nil || resp.Payload.Pagination.NextPageToken == "" {
			break
		}

		next := resp.Payload.Pagination.NextPageToken
		req.PaginationNextPageToken = &next
	}

	return sps, nil
}
