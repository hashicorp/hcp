// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tfcconfig

import (
	"context"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2023-08-18/client/waypoint_service"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
)

func NewCmdTFCConfig(ctx *cmd.Context) *cmd.Command {
	cmd := &cmd.Command{
		Name:      "tfc-config",
		ShortHelp: "Manage Terraform Cloud Configurations.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp waypoint tfc-config" }} command group manages 
		the set of TFC Configs. New TFC Configs can be created using
		{{ template "mdCodeOrBold" "hcp waypoint tfc-config create" }} and existing
		profiles can be viewed using {{ template "mdCodeOrBold" "hcp waypoint tfc-config read" }}.
		`),
	}

	cmd.AddChild(NewCmdCreate(ctx, nil))
	cmd.AddChild(NewCmdDelete(ctx, nil))
	cmd.AddChild(NewCmdRead(ctx, nil))
	return cmd
}

func GetNamespace(ctx context.Context, client waypoint_service.ClientService, orgID, projectID string) (string, error) {

	resp, err := client.WaypointServiceGetNamespace(
		&waypoint_service.WaypointServiceGetNamespaceParams{
			LocationOrganizationID: orgID,
			LocationProjectID:      projectID,
			Context:                ctx,
		}, nil,
	)
	if err != nil {
		return "", err
	}

	return resp.Payload.Namespace.ID, nil

}
