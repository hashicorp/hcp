// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package actions

import (
	"errors"
	"fmt"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2024-11-22/client/waypoint_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2024-11-22/models"

	wpopts "github.com/hashicorp/hcp/internal/commands/waypoint/opts"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/flagvalue"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
)

type CreateOpts struct {
	wpopts.WaypointOpts

	Name        string
	Description string
	// Request Types. We only support setting a Oneof
	Request *models.HashicorpCloudWaypointV20241122ActionConfigRequest
	// Workarounds due to not being able to set these values directly in cmd.Flag
	RequestCustomMethod string
	RequestHeaders      map[string]string

	// Agent flavor fields
	AgentGroup     string
	AgentOperation string
}

func NewCmdCreate(ctx *cmd.Context, opts *CreateOpts) *cmd.Command {
	opts.WaypointOpts = wpopts.New(ctx)
	if opts.Request == nil {
		opts.Request = &models.HashicorpCloudWaypointV20241122ActionConfigRequest{
			Custom: &models.HashicorpCloudWaypointV20241122ActionConfigFlavorCustom{},
		}
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
					Name:         "name",
					Shorthand:    "n",
					DisplayValue: "NAME",
					Description:  "The name of the action.",
					Value:        flagvalue.Simple("", &opts.Name),
					Required:     true,
				},
				{
					Name:         "description",
					Shorthand:    "d",
					DisplayValue: "DESCRIPTION",
					Description:  "The description of the action.",
					Value:        flagvalue.Simple("", &opts.Description),
				},
				// Custom Requests
				{
					Name:         "url",
					DisplayValue: "URL",
					Description:  "The URL of the action.",
					Value:        flagvalue.Simple("", &opts.Request.Custom.URL),
				},
				{
					Name:         "body",
					DisplayValue: "BODY",
					Description:  "The request body to submit when running the action.",
					Value:        flagvalue.Simple("", &opts.Request.Custom.Body),
				},
				{
					Name:         "method",
					DisplayValue: "METHOD",
					Description:  "The HTTP method to use when making the request.",
					Value:        flagvalue.Simple("GET", &opts.RequestCustomMethod),
				},
				{
					Name:         "header",
					DisplayValue: "KEY=VALUE",
					Description:  "The headers to include in the request. This flag can be specified multiple times.",
					Value:        flagvalue.SimpleMap(map[string]string{}, &opts.RequestHeaders),
					Repeatable:   true,
				},
				// GitHub Requests

				// Agent Requests
				{
					Name:         "agent-group",
					DisplayValue: "GROUP",
					Description:  "The agent group to use for the action.",
					Value:        flagvalue.Simple("", &opts.AgentGroup),
				},
				{
					Name:         "agent-operation",
					DisplayValue: "OPERATION",
					Description:  "The operation ID to run in the agent group.",
					Value:        flagvalue.Simple("", &opts.AgentOperation),
				},
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

	// Determine action flavor
	hasAgent := opts.AgentGroup != "" || opts.AgentOperation != ""
	hasCustom := opts.Request.Custom != nil && opts.Request.Custom.URL != ""

	// Validate agent flags
	if hasAgent {
		if hasCustom {
			return errors.New("cannot specify both custom action and agent action flags")
		}

		if opts.AgentGroup == "" {
			return errors.New("agent-group must be specified when using agent action")
		}

		if opts.AgentOperation == "" {
			return errors.New("agent-operation must be specified when using agent action")
		}

		// Create agent flavor request
		opts.Request.Agent = &models.HashicorpCloudWaypointV20241122ActionConfigFlavorAgent{
			Op: &models.HashicorpCloudWaypointV20241122AgentOperation{
				Group: opts.AgentGroup,
				ID:    opts.AgentOperation,
			},
		}
		// Clear custom request since we're using agent
		opts.Request.Custom = nil
	} else if hasCustom {
		// Parse the headers
		for k, v := range opts.RequestHeaders {
			opts.Request.Custom.Headers = append(opts.Request.Custom.Headers, &models.HashicorpCloudWaypointV20241122ActionConfigFlavorCustomHeader{
				Key:   k,
				Value: v,
			})
		}

		// Cast the string to a const for the sdk API
		customMethod := models.HashicorpCloudWaypointV20241122ActionConfigFlavorCustomMethod(opts.RequestCustomMethod)
		opts.Request.Custom.Method = &customMethod
	} else {
		return errors.New("must specify either custom action flags (--url) or agent action flags (--agent-group, --agent-operation)")
	}

	// Ok, run the command!!
	resp, err := opts.WS2024Client.WaypointServiceCreateActionConfig(&waypoint_service.WaypointServiceCreateActionConfigParams{
		NamespaceLocationOrganizationID: opts.Profile.OrganizationID,
		NamespaceLocationProjectID:      opts.Profile.ProjectID,
		Context:                         opts.Ctx,
		Body: &models.HashicorpCloudWaypointV20241122WaypointServiceCreateActionConfigBody{
			ActionConfig: &models.HashicorpCloudWaypointV20241122ActionConfig{
				Name:        opts.Name,
				Description: opts.Description,
				Request:     opts.Request,
			},
		},
	}, nil)

	if err != nil {
		return fmt.Errorf("failed to create action %q: %w", opts.Name, err)
	}

	if resp == nil || resp.GetPayload() == nil {
		return fmt.Errorf("no response returned from API")
	}
	actionCfg := resp.GetPayload().ActionConfig
	if actionCfg == nil {
		return fmt.Errorf("no action config returned from API")
	}

	// Choose fields based on action flavor
	var fields []format.Field
	if actionCfg.Request != nil && actionCfg.Request.Agent != nil {
		fields = []format.Field{
			format.NewField("Name", "{{ .Name }}"),
			format.NewField("Description", "{{ .Description }}"),
			format.NewField("Agent Group", "{{ .Request.Agent.Op.Group }}"),
			format.NewField("Agent Operation", "{{ .Request.Agent.Op.ID }}"),
		}
	} else if actionCfg.Request != nil && actionCfg.Request.Custom != nil {
		fields = []format.Field{
			format.NewField("Name", "{{ .Name }}"),
			format.NewField("Description", "{{ .Description }}"),
			format.NewField("URL", "{{ .Request.Custom.URL }}"),
			format.NewField("Method", "{{ .Request.Custom.Method }}"),
		}
	}

	d := format.NewDisplayer(&models.HashicorpCloudWaypointV20241122ActionConfig{
		Name:        actionCfg.Name,
		Description: actionCfg.Description,
		Request:     actionCfg.Request,
	}, format.Pretty, fields)
	return opts.Output.Display(d)
}
