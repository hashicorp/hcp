// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package agent

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/go-openapi/strfmt"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2023-08-18/client/waypoint_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2023-08-18/models"
	"github.com/hashicorp/hcp/internal/commands/waypoint/opts"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/flagvalue"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/pkg/errors"
)

type QueueOpts struct {
	opts.WaypointOpts

	Group string
	ID    string
	Body  string

	ActionRunID string
}

func NewCmdQueue(ctx *cmd.Context) *cmd.Command {
	opts := &QueueOpts{
		WaypointOpts: opts.New(ctx),
	}

	cmd := &cmd.Command{
		Name:      "queue",
		ShortHelp: "Queue an operation for an agent to execute.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp waypoint agent queue" }} command queues an operation for an agent to run.
		`),
		Flags: cmd.Flags{
			Local: []*cmd.Flag{
				{
					Name:         "id",
					Shorthand:    "i",
					DisplayValue: "ID",
					Description:  "Id of the operation to run.",
					Value:        flagvalue.Simple("", &opts.ID),
					Required:     true,
				},
				{
					Name:         "body",
					Shorthand:    "d",
					DisplayValue: "JSON",
					Description:  "JSON to pass to operation. Use @filename to read json from a file.",
					Value:        flagvalue.Simple("", &opts.Body),
				},
				{
					Name:         "action-run",
					DisplayValue: "ID",
					Description:  "Action run to associate operation with.",
					Value:        flagvalue.Simple("", &opts.ActionRunID),
				},
				{
					Name:         "group",
					Shorthand:    "g",
					DisplayValue: "NAME",
					Description:  "Agent group to run for operations on.",
					Value:        flagvalue.Simple("", &opts.Group),
					Required:     true,
				},
			},
		},
		RunF: func(c *cmd.Command, args []string) error {
			return agentQueue(c.Logger(), opts)
		},
	}

	return cmd
}

func agentQueue(log hclog.Logger, opts *QueueOpts) error {
	ns, err := opts.Namespace()
	if err != nil {
		return errors.Wrapf(err, "Unable to access HCP project")
	}

	var body strfmt.Base64

	if strings.HasPrefix(opts.Body, "@") {
		path := opts.Body[1:]
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("unable to read json file %s: %w", path, err)
		}

		if !json.Valid(data) {
			return fmt.Errorf("invalid json in file '%s'", path)
		}

		body = data
	} else if len(opts.Body) > 0 {
		if !json.Valid([]byte(opts.Body)) {
			return fmt.Errorf("invalid json specified on command line")
		}

		body = strfmt.Base64(opts.Body)
	}

	ctx := opts.Ctx

	_, err = opts.WS.WaypointServiceQueueAgentOperation(&waypoint_service.WaypointServiceQueueAgentOperationParams{
		NamespaceID: ns.ID,
		Body: &models.HashicorpCloudWaypointWaypointServiceQueueAgentOperationBody{
			Operation: &models.HashicorpCloudWaypointAgentOperation{
				ID:          opts.ID,
				ActionRunID: opts.ActionRunID,
				Body:        body,
				Group:       opts.Group,
			},
		},
		Context: ctx,
	}, nil)

	if err != nil {
		return fmt.Errorf("error queuing operation: %w", err)
	}

	fmt.Fprintf(opts.IO.Err(), "Operation '%s' queued.\n", opts.ID)
	return nil
}
