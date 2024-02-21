package actionconfig

import (
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2023-08-18/client/waypoint_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2023-08-18/models"

	"github.com/hashicorp/hcp/internal/commands/waypoint/opts"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/flagvalue"
)

type CreateOpts struct {
	opts.WaypointOpts

	Name        string
	Description string
	// Request Types. We only support setting a Oneof
	Request *models.HashicorpCloudWaypointActionConfigRequest
	// Workarounds due to not being able to set these values directly in cmd.Flag
	RequestCustomMethod string
	// Must be hacked to work, there's no supporting string map flag type
	RequestHeaders    map[string]string
	RequestHeadersRaw []string // Expected to be "key=value" and will return an err if not
}

func NewCmdCreate(ctx *cmd.Context) *cmd.Command {
	opts := &CreateOpts{
		WaypointOpts: opts.New(ctx),
		Request: &models.HashicorpCloudWaypointActionConfigRequest{
			Custom: &models.HashicorpCloudWaypointActionConfigFlavorCustom{},
			Github: &models.HashicorpCloudWaypointActionConfigFlavorGitHub{},
		},
	}

	cmd := &cmd.Command{
		Name:      "create",
		ShortHelp: "Create a new action configuration.",
		LongHelp:  "Create a new action configuration to be used to launch an action with.",
		RunF: func(c *cmd.Command, args []string) error {
			return createActionConfig(c, args, opts)
		},
		PersistentPreRun: func(c *cmd.Command, args []string) error {
			return cmd.RequireOrgAndProject(ctx)
		},
		Flags: cmd.Flags{
			Local: []*cmd.Flag{
				{
					Name:        "name",
					Shorthand:   "n",
					Description: "The name of the action configuration.",
					Value:       flagvalue.Simple("", &opts.Name),
				},
				{
					Name:        "description",
					Shorthand:   "d",
					Description: "The description of the action configuration.",
					Value:       flagvalue.Simple("", &opts.Description),
				},
				// Custom Requests
				{
					Name:        "url",
					Description: "The URL of the action configuration.",
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
					Value:       flagvalue.SimpleSlice([]string{}, &opts.RequestHeadersRaw),
					Repeatable:  true,
				},
				// GitHub Requests
			},
		},
	}

	return cmd
}

func createActionConfig(c *cmd.Command, args []string, opts *CreateOpts) error {
	// Validate Request Type is not set to all options
	if opts.Request != nil {
		if opts.Request.Custom != nil && opts.Request.Github != nil {
			return errors.New("only one request type can be set")
		}
		if opts.Request.Custom != nil {
			return errors.New("gitHub request types are not yet supported")
		}
	}

	// Work arounds for missing flag support
	if opts.Request.Custom != nil {
		// Parse the headers. This is a workaround for not having a map string flag type
		if len(opts.RequestHeadersRaw) > 0 {
			opts.RequestHeaders = make(map[string]string)
			for _, header := range opts.RequestHeadersRaw {
				// Split the header into key and value with the string package
				kv := strings.Split(header, "=")
				if len(kv) != 2 {
					return errors.New("invalid header format. Must be in the format 'key=value'")
				}
				opts.Request.Custom.Headers = append(opts.Request.Custom.Headers, &models.HashicorpCloudWaypointActionConfigFlavorCustomHeader{
					Key:   kv[0],
					Value: kv[1],
				})
			}
		}

		// Cast the method to the correct type
		// TODO(briancain): this isn't working
		// opts.Request.Custom.Method = models.HashicorpCloudWaypointActionConfigFlavorCustomMethod(opts.RequestCustomMethod)
	}

	// Ok, run the command!!
	ns, err := opts.Namespace()
	if err != nil {
		return err
	}
	_, err = opts.WS.WaypointServiceCreateActionConfig(&waypoint_service.WaypointServiceCreateActionConfigParams{
		NamespaceID: ns.ID,
		Context:     opts.Ctx,
		Body: &models.HashicorpCloudWaypointWaypointServiceCreateActionConfigBody{
			ActionConfig: &models.HashicorpCloudWaypointActionConfig{
				Name:        opts.Name,
				Description: opts.Description,
				Request:     opts.Request,
			},
		},
	}, nil)

	if err != nil {
		fmt.Fprintf(opts.IO.Err(), "Error creating action config: %s", err)
		return err
	}

	fmt.Fprintf(opts.IO.Out(), "Action config %q created.", opts.Name)

	return nil
}
