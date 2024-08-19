// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package agent

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2023-08-18/models"
	"github.com/stretchr/testify/require"
)

func TestExecutor(t *testing.T) {
	t.Parallel()

	log := hclog.New(&hclog.LoggerOptions{
		Name:  "agent-exec",
		Level: hclog.Trace,
	})
	t.Run("can lookup the availability of an operation", func(t *testing.T) {
		t.Parallel()

		r := require.New(t)

		var (
			e Executor
		)

		e.Log = log

		hcl := `
		group "test" {
			action "launch" {
				run {
					command = "./launch.sh"
				}
			}
		}
`

		cfg, err := ParseConfig(hcl)
		r.NoError(err)

		e.Config = cfg

		ok, err := e.IsAvailable(&models.HashicorpCloudWaypointAgentOperation{
			Group: "test",
			ID:    "launch",
		})

		r.NoError(err)

		r.True(ok)
	})

	t.Run("executes the operation based on the id", func(t *testing.T) {
		t.Parallel()

		r := require.New(t)

		var (
			e Executor
		)

		e.Log = log
		hcl := `
		group "test" {
			action "launch" {
				run {
					command = "true"
				}
			}
		}
`

		cfg, err := ParseConfig(hcl)
		r.NoError(err)

		e.Config = cfg

		_, err = e.Execute(context.TODO(), &models.HashicorpCloudWaypointAgentOperation{
			Group: "test",
			ID:    "launch",
		})

		r.NoError(err)
	})

	t.Run("variables can be passed through", func(t *testing.T) {
		t.Parallel()

		r := require.New(t)

		var (
			e Executor
		)

		e.Log = log
		hcl := `
		group "test" {
			action "launch" {
				run {
					command = "echo '${var.type}'"
		    }
			}
		}
`

		cfg, err := ParseConfig(hcl)
		r.NoError(err)

		e.Config = cfg

		data, err := json.Marshal(map[string]any{
			"var.type":                              "nerf",
			"action.name":                           "launch",
			"application.templateName":              "cool-test-template", // pretend its in an app
			"application.name":                      "test",
			"application.outputs.run_id":            "1234",
			"application.inputs.region":             "us-west-1",
			"addon.abc123.outputs.database_url":     "http://localhost:5432",
			"addon.xyz098.outputs.load_balancer_ip": "http://localhost:8080",
			"var.local_variable":                    "local-value",
			"var.token":                             "token",
		})
		r.NoError(err)

		_, err = e.Execute(context.TODO(), &models.HashicorpCloudWaypointAgentOperation{
			Group: "test",
			ID:    "launch",
			Body:  data,
		})
		r.NoError(err)
	})

}
