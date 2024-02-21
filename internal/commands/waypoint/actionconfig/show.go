package actionconfig

import (
	"fmt"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2023-08-18/client/waypoint_service"
	"github.com/hashicorp/hcp/internal/commands/waypoint/opts"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/flagvalue"
)

type ShowOpts struct {
	opts.WaypointOpts

	Name string
}

func NewCmdShow(ctx *cmd.Context) *cmd.Command {
	opts := &ShowOpts{
		WaypointOpts: opts.New(ctx),
	}

	cmd := &cmd.Command{
		Name:      "show",
		ShortHelp: "Show more details about an action configurations.",
		LongHelp:  "Show more details about an action configurations.",
		RunF: func(c *cmd.Command, args []string) error {
			return showActionConfig(c, args, opts)
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
					Required:    true,
				},
			},
		},
	}

	return cmd
}

func showActionConfig(c *cmd.Command, args []string, opts *ShowOpts) error {
	ns, err := opts.Namespace()
	if err != nil {
		return err
	}

	// Make action name a string pointer
	actionName := &opts.Name
	_, err = opts.WS.WaypointServiceGetActionConfig(&waypoint_service.WaypointServiceGetActionConfigParams{
		NamespaceID: ns.ID,
		Context:     opts.Ctx,
		ActionName:  actionName,
	}, nil)
	if err != nil {
		fmt.Fprintf(opts.IO.Err(), "Error getting action config: %s", err)
		return err
	}

	// respPayload := resp.GetPayload()
	// actionCfg := respPayload.ActionConfig
	// latestRun := respPayload.LatestRun
	// totalRuns := respPayload.TotalRuns

	// TODO(briancain): https://github.com/hashicorp/hcp/issues/16

	return nil
}
