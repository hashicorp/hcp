// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package actions

import (
	"errors"
	"fmt"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2024-11-22/client/waypoint_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2024-11-22/models"

	"github.com/hashicorp/hcp/internal/commands/waypoint/opts"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/flagvalue"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
)

type CreateOpts struct {
	opts.WaypointOpts

	Name        string
	Description string
	// Request Types. We only support setting a Oneof
	Request *models.HashicorpCloudWaypointActionConfigRequest
	// Workarounds due to not being able to set these values directly in cmd.Flag
	RequestCustomMethod string
	RequestHeaders      map[string]string
}

func NewCmdCreate(ctx *cmd.Context) *cmd.Command {
	opts := &CreateOpts{
		WaypointOpts: opts.New(ctx),
		Request: &models.HashicorpCloudWaypointActionConfigRequest{
			Custom: &models.HashicorpCloudWaypointActionConfigFlavorCustom{},
		},
	}

	cmd := &cmd.Command{
		Name:      "create",
		ShortHelp: "Create a new action configuration.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp waypoint actions create" }} command
		creates a new action to be used to launch an action with.
		`),
		RunF: func(c *cmd.Command, args []string) error {
			return createAction(c, args, opts)
		},
		PersistentPreRun: func(c *cmd.Command, args []string) error {
			return cmd.RequireOrgAndProject(ctx)
		},
		Flags: cmd.Flags{
			Local: []*cmd.Flag{
				{
					Name:        "name",
					Shorthand:   "n",
					Description: "The name of the action.",
					Value:       flagvalue.Simple("", &opts.Name),
				},
				{
					Name:        "description",
					Shorthand:   "d",
					Description: "The description of the action.",
					Value:       flagvalue.Simple("", &opts.Description),
				},
				// Custom Requests
				{
					Name:        "url",
					Description: "The URL of the action.",
					Value:       flagvalue.Simple("", &opts.Request.Custom.URL),
				},
				{
					Name:        "body",
					Description: "The request body to submit when running the action.",
					Value:       flagvalue.Simple("", &opts.Request.Custom.Body),
				},
				{
					Name:        "method",
					Description: "The HTTP method to use when making the request.",
					Value:       flagvalue.Simple("GET", &opts.RequestCustomMethod),
				},
				{
					Name:        "header",
					Description: "The headers to include in the request. This flag can be specified multiple times.",
					Value:       flagvalue.SimpleMap(map[string]string{}, &opts.RequestHeaders),
					Repeatable:  true,
				},
				// GitHub Requests
			},
		},
	}

	return cmd
}

func createAction(c *cmd.Command, args []string, opts *CreateOpts) error {
	// Validate Request Type is not set to all options
	if opts.Request != nil {
		if opts.Request.Custom != nil && opts.Request.Custom.URL != "" && opts.Request.Github != nil {
			return errors.New("only one request type can be set")
		}
		if opts.Request.Github != nil {
			return errors.New("gitHub request types are not yet supported")
		}
	}

	// Mutate flag values to Custom options
	if opts.Request.Custom != nil {
		// Parse the headers
		for k, v := range opts.RequestHeaders {
			opts.Request.Custom.Headers = append(opts.Request.Custom.Headers, &models.HashicorpCloudWaypointActionConfigFlavorCustomHeader{
				Key:   k,
				Value: v,
			})
		}

		// Cast the string to a const for the sdk API
		customMethod := models.HashicorpCloudWaypointActionConfigFlavorCustomMethod(opts.RequestCustomMethod)
		opts.Request.Custom.Method = &customMethod
	}

	// Ok, run the command!!
	_, err := opts.WS2024Client.WaypointServiceCreateActionConfig(&waypoint_service.WaypointServiceCreateActionConfigParams{
		NamespaceLocationOrganizationID: opts.Profile.OrganizationID,
		NamespaceLocationProjectID:      opts.Profile.ProjectID,
		Context:                         opts.Ctx,
		Body: &models.HashicorpCloudWaypointV20241122WaypointServiceCreateActionConfigBody{
			ActionConfig: &models.HashicorpCloudWaypointActionConfig{
				Name:        opts.Name,
				Description: opts.Description,
				Request:     opts.Request,
			},
		},
	}, nil)

	if err != nil {
		return fmt.Errorf("failed to create action %q: %w", opts.Name, err)
	}

	fmt.Fprintf(opts.IO.Err(), "Action %q created.", opts.Name)

	return nil
}
