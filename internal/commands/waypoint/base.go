package waypoint

import (
	"context"

	"github.com/hashicorp/hcp-cli-sdk/hcpcli"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	wp02 "github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2022-02-03/client"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2022-02-03/client/action"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2022-02-03/client/waypoint_control_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2023-08-18/client"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2023-08-18/client/waypoint_service"
)

// BaseCommand is embedded in all commands to provide common logic and data.
//
// The unexported values are not available until after Init is called. Some
// values are only available in certain circumstances, read the documentation
// for the field to determine if that is the case.
type BaseCommand struct {
	*hcpcli.BaseCommand
}

// Close cleans up any resources that the command created. This should be
// deferred by any CLI command that embeds BaseCommand in the Run command.
func (c *BaseCommand) Close() error {
	return nil
}

func (c *BaseCommand) WPClient() (waypoint_service.ClientService, error) {
	cl, err := c.HCPClient()
	if err != nil {
		return nil, err
	}

	return client.New(cl, nil).WaypointService, nil
}

func (c *BaseCommand) ActionsClient() (action.ClientService, error) {
	cl, err := c.HCPClient()
	if err != nil {
		return nil, err
	}

	// Note(XX): clarification, wp02 is not waypoint 2.0 but waypoint 1.0. Will be deleted after 2.0 sdk goes public.
	return wp02.New(cl, nil).Action, nil
}

func (c *BaseCommand) GetNamespace() (string, error) {
	cl, err := c.HCPClient()
	if err != nil {
		return "", err
	}

	wp := wp02.New(cl, nil).WaypointControlService

	resp, err := wp.WaypointControlServiceGetNamespace(&waypoint_control_service.WaypointControlServiceGetNamespaceParams{
		LocationOrganizationID: c.Config.Organization,
		LocationProjectID:      c.Config.Project,
		Context:                context.Background(),
	}, nil)
	if err != nil {
		return "", err
	}

	return resp.Payload.Namespace.ID, nil
}

func (c *BaseCommand) GetNamespaceNext() (string, error) {
	client, err := c.WPClient()
	if err != nil {
		return "", err
	}

	// NOTE(XX): Add error if org ID & project ID are empty
	// NOTE(XX): Add a way for the user to give us the Org & Project Names since Org & Project ID are hard to find
	resp, err := client.WaypointServiceGetNamespace(
		&waypoint_service.WaypointServiceGetNamespaceParams{
			LocationOrganizationID: c.Config.Organization,
			LocationProjectID:      c.Config.Project,
			Context:                context.Background(),
		}, nil,
	)
	if err != nil {
		return "", err
	}

	return resp.Payload.Namespace.ID, nil
}

type ErrorCode interface {
	error
	Code() int
}

func (c *BaseCommand) StatusCode(err error) int {
	if v, ok := err.(ErrorCode); ok {
		return v.Code()
	}

	return -1
}

type ErrorPayload interface {
	GetPayload() *models.GrpcGatewayRuntimeError
}

func (c *BaseCommand) ShowError(err error) {
	ep, ok := err.(ErrorPayload)
	if !ok {
		return
	}

	grpcErr := ep.GetPayload()

	if grpcErr.Message != "" {
		if grpcErr.Error != "" {
			c.Ui.VerbosePrintf("Service reported: %s (%s) (%d)", grpcErr.Message, grpcErr.Error, grpcErr.Code)
		} else {
			c.Ui.VerbosePrintf("Service reported: %s (%d)", grpcErr.Message, grpcErr.Code)
		}
	} else {
		c.Ui.VerbosePrintf("Service reported an unknown error (%d)", grpcErr.Code)
	}
}
