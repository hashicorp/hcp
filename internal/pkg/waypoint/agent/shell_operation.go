// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package agent

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"slices"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2024-11-22/client/waypoint_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2024-11-22/models"
	"github.com/hashicorp/hcp/internal/pkg/profile"
)

type ShellOperation struct {
	Arguments     []string
	Environment   map[string]string
	DockerOptions *DockerOptions
}

type DockerOptions struct {
	Image string `hcl:"image"`
}

func (s *ShellOperation) Run(
	ctx context.Context,
	log hclog.Logger,
	api waypoint_service.ClientService,
	profile *profile.Profile,
	opInfo *models.HashicorpCloudWaypointV20241122AgentOperation,
) (OperationStatus, error) {
	if s.DockerOptions != nil {
		return s.runUnderDocker(ctx, log)
	}

	return s.exec(ctx, log, s.Arguments)
}

func (s *ShellOperation) exec(ctx context.Context, log hclog.Logger, cmd []string) (OperationStatus, error) {
	c := exec.CommandContext(ctx, s.Arguments[0], s.Arguments[1:]...)

	c.Env = slices.Clone(os.Environ())

	for k, v := range s.Environment {
		c.Env = append(c.Env, k+"="+v)
	}

	var out bytes.Buffer

	c.Stdout = &out
	c.Stderr = &out

	err := c.Run()
	log.Debug("output from shell operation", "command", s.Arguments[0], "output", out.String(), "error", err)

	status := OperationStatus{
		Code: c.ProcessState.ExitCode(),
	}

	data := bytes.TrimSpace(out.Bytes())

	// Only return last line of output for security reasons.
	// Users can use the API to send more output as a status log, if desired.
	if idx := bytes.LastIndexByte(data, '\n'); idx != -1 {
		status.Status = "output: " + string(data[idx:])
	} else if len(data) > 0 {
		status.Status = "output: " + string(data)
	}

	return status, nil
}

func (s *ShellOperation) runUnderDocker(ctx context.Context, log hclog.Logger) (OperationStatus, error) {
	args := append([]string{
		"run", "--rm", s.DockerOptions.Image,
	}, s.Arguments...)

	return s.exec(ctx, log, args)
}
