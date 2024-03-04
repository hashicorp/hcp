package actionconfig

import (
	"fmt"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2023-08-18/client/waypoint_service"
	"github.com/hashicorp/hcp/internal/commands/waypoint/opts"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/format"
)

type ListOpts struct {
	opts.WaypointOpts
}

func NewCmdList(ctx *cmd.Context) *cmd.Command {
	opts := &ListOpts{
		WaypointOpts: opts.New(ctx),
	}

	cmd := &cmd.Command{
		Name:      "list",
		ShortHelp: "List all known action configurations.",
		LongHelp:  "List all known action configurations from HCP Waypoint.",
		RunF: func(c *cmd.Command, args []string) error {
			return listActionConfig(c, args, opts)
		},
		PersistentPreRun: func(c *cmd.Command, args []string) error {
			return cmd.RequireOrgAndProject(ctx)
		},
	}

	return cmd
}

func listActionConfig(c *cmd.Command, args []string, opts *ListOpts) error {
	ns, err := opts.Namespace()
	if err != nil {
		return err
	}

	resp, err := opts.WS.WaypointServiceListActionConfigs(&waypoint_service.WaypointServiceListActionConfigsParams{
		NamespaceID: ns.ID,
		Context:     opts.Ctx,
	}, nil)
	if err != nil {
		return fmt.Errorf("error listing action configurations: %w", err)
	}

	respPayload := resp.GetPayload()
	d := format.NewDisplayer(respPayload.ActionConfigs, format.Pretty, actionConfigFields)

	return opts.Output.Display(d)
}
