package actionconfig

import (
	"fmt"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2023-08-18/client/waypoint_service"
	"github.com/hashicorp/hcp/internal/commands/waypoint/opts"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/flagvalue"
)

type DeleteOpts struct {
	opts.WaypointOpts

	Name string
	// We intentionally don't support ID for delete yet.
	Id string
}

func NewCmdDelete(ctx *cmd.Context) *cmd.Command {
	opts := &DeleteOpts{
		WaypointOpts: opts.New(ctx),
	}

	cmd := &cmd.Command{
		Name:      "delete",
		ShortHelp: "Delete an existing action configuration.",
		LongHelp:  "Delete an existing action configuration. This will remove the config completely from HCP Waypoint.",
		RunF: func(c *cmd.Command, args []string) error {
			return deleteActionConfig(c, args, opts)
		},
		PersistentPreRun: func(c *cmd.Command, args []string) error {
			return cmd.RequireOrgAndProject(ctx)
		},
		Flags: cmd.Flags{
			Local: []*cmd.Flag{
				{
					Name:        "name",
					Shorthand:   "n",
					Description: "The name of the action configuration to delete.",
					Value:       flagvalue.Simple("", &opts.Name),
					Required:    true,
				},
			},
		},
	}

	return cmd
}

func deleteActionConfig(c *cmd.Command, args []string, opts *DeleteOpts) error {
	ns, err := opts.Namespace()
	if err != nil {
		return err
	}

	// Make action name a string pointer
	actionName := &opts.Name
	_, err = opts.WS.WaypointServiceDeleteActionConfig(&waypoint_service.WaypointServiceDeleteActionConfigParams{
		NamespaceID: ns.ID,
		Context:     opts.Ctx,
		ActionName:  actionName,
	}, nil)
	if err != nil {
		fmt.Fprintf(opts.IO.Err(), "Error deleting action config: %s", err)
		return err
	}

	fmt.Fprintf(opts.IO.Out(), "Action config %q deleted.", opts.Name)
	return nil
}
