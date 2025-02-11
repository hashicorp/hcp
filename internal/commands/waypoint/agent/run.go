// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2024-11-22/client/waypoint_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2024-11-22/models"
	"github.com/pkg/errors"

	"github.com/hashicorp/hcp/internal/commands/waypoint/opts"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/flagvalue"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/hashicorp/hcp/internal/pkg/waypoint/agent"
)

type RunOpts struct {
	opts.WaypointOpts
	Groups []string

	ConfigPath string
	Config     *agent.Config
}

func NewCmdRun(ctx *cmd.Context) *cmd.Command {
	opts := &RunOpts{
		WaypointOpts: opts.New(ctx),
	}

	cmd := &cmd.Command{
		Name:      "run",
		ShortHelp: "Start the Waypoint Agent.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp waypoint agent run" }} command executes a local Waypoint Agent.
		`),
		Flags: cmd.Flags{
			Local: []*cmd.Flag{
				{
					Name:         "config",
					Shorthand:    "c",
					DisplayValue: "PATH",
					Description:  "Path to configuration file for agent.",
					Value:        flagvalue.Simple("agent.hcl", &opts.ConfigPath),
				},
			},
		},
		PersistentPreRun: func(c *cmd.Command, args []string) error {
			return cmd.RequireOrgAndProject(ctx)
		},
		RunF: func(c *cmd.Command, args []string) error {
			return agentRun(c.Logger(), opts)
		},
	}

	return cmd
}

var agentRunDuration = 60 * time.Second

func agentRun(log hclog.Logger, opts *RunOpts) error {
	// Only set level to info if not in debug mode
	if !log.IsDebug() {
		// Give log.Info level feedback to user when they run the agent CLI
		log.SetLevel(hclog.Info)
	}

	cfg, err := agent.ParseConfigFile(opts.ConfigPath)
	if err != nil {
		return err
	}

	opts.Groups = cfg.Groups()

	ns, err := opts.Namespace()
	if err != nil {
		return errors.Wrapf(err, "Unable to access HCP project")
	}

	ctx := opts.Ctx

	// check the groups!
	resp2, err := opts.WS2024Client.WaypointServiceValidateAgentGroups(&waypoint_service.WaypointServiceValidateAgentGroupsParams{
		NamespaceLocationOrganizationID: opts.Profile.OrganizationID,
		NamespaceLocationProjectID:      opts.Profile.ProjectID,
		Body: &models.HashicorpCloudWaypointV20241122WaypointServiceValidateAgentGroupsBody{
			Groups: opts.Groups,
		},
		Context: opts.Ctx,
	}, nil)
	if err != nil {
		return errors.Wrapf(err, "Error validating agent group names")
	}

	if len(resp2.Payload.UnknownGroups) > 0 {
		fmt.Fprintf(opts.IO.Err(), "Unknown agent groups detected:\n")

		for _, g := range resp2.Payload.UnknownGroups {
			fmt.Fprintf(opts.IO.Err(), "  %s\n ", g)
		}
		return nil
	}

	retry := time.NewTimer(agentRunDuration)
	defer retry.Stop()

	log.Info("Waypoint agent initialized",
		"hcp-org", opts.Profile.OrganizationID,
		"hcp-project", opts.Profile.ProjectID,
		"waypoint-namespace", ns.ID,
		"groups", opts.Groups,
	)

	exec := &agent.Executor{
		Log:    log,
		Config: cfg,
	}

	for {
		opCfg, err := opts.WS2024Client.WaypointServiceRetrieveAgentOperation(&waypoint_service.WaypointServiceRetrieveAgentOperationParams{
			Body: &models.HashicorpCloudWaypointV20241122WaypointServiceRetrieveAgentOperationBody{
				Groups: opts.Groups,
			},
			NamespaceLocationOrganizationID: opts.Profile.OrganizationID,
			NamespaceLocationProjectID:      opts.Profile.ProjectID,
			Context:                         ctx,
		}, nil)

		if err != nil {
			log.Error("error reading agent operation", "error", err)
		} else if ao := opCfg.Payload.Operation; ao != nil {
			runOp(log, ctx, opts, ao, exec, ns.ID)
		}

		retry.Reset(agentRunDuration)

		select {
		case <-opts.Ctx.Done():
			return nil
		case <-retry.C:
			// ok
		}
	}
}

func runOp(
	log hclog.Logger,
	ctx context.Context,
	opts *RunOpts,
	ao *models.HashicorpCloudWaypointAgentOperation,
	exec *agent.Executor,
	ns string,
) {
	var (
		status      string
		statusCode  int
		sequenceNum string
	)

	log = log.With("group", ao.Group, "operation", ao.ID, "action-run-id", ao.ActionRunID)

	if ao.ActionRunID != "" {
		log.Info("reporting action run starting")

		resp, err := opts.WS2024Client.WaypointServiceStartingAction(&waypoint_service.WaypointServiceStartingActionParams{
			Body: &models.HashicorpCloudWaypointV20241122WaypointServiceStartingActionBody{
				ActionRunID: ao.ActionRunID,
				GroupName:   ao.Group,
			},
			Context:                         ctx,
			NamespaceLocationOrganizationID: opts.Profile.OrganizationID,
			NamespaceLocationProjectID:      opts.Profile.ProjectID,
		}, nil)

		if err != nil {
			log.Error("unable to register action as starting", "error", err)
		} else {
			if resp != nil && resp.Payload != nil {
				sequenceNum = resp.Payload.Sequence

			}
			defer func() {
				log.Info("reporting action run ended", "status", status, "status-code", statusCode, "action-run-sequence", sequenceNum)

				_, err = opts.WS2024Client.WaypointServiceEndingAction(&waypoint_service.WaypointServiceEndingActionParams{
					Body: &models.HashicorpCloudWaypointV20241122WaypointServiceEndingActionBody{
						ActionRunID: resp.Payload.ActionRunID,
						FinalStatus: status,
						StatusCode:  int32(statusCode),
					},
					Context:                         ctx,
					NamespaceLocationOrganizationID: opts.Profile.OrganizationID,
					NamespaceLocationProjectID:      opts.Profile.ProjectID,
				}, nil)

				if err != nil {
					log.Error("unable to send ending action", "error", err)
				}
			}()
		}
	}

	ok, err := exec.IsAvailable(ao)
	if err != nil {
		status = "internal error: " + err.Error()
		statusCode = 1

		log.Error("error resolving operation", "error", err)
		return
	}

	if !ok {
		status = "unknown operation: " + ao.ID
		statusCode = 127

		log.Error("requested unknown operation", "status", status, "status-code", statusCode)
		return
	}

	opStat, err := exec.Execute(ctx, ao)
	if err != nil {
		status = "error execution operation: " + err.Error()

		log.Error("error executing operation", "error", err)
		return
	}

	status = opStat.Status
	statusCode = opStat.Code

	log.Info("finished operation")
}
