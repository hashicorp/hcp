// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package agent

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2023-08-18/models"
	"github.com/pkg/errors"
	"github.com/zclconf/go-cty/cty"
)

type OperationStatus struct {
	Status string
	Code   int
}

var (
	errStatus   = OperationStatus{Code: -1}
	cleanStatus = OperationStatus{Code: 0}
)

type Operation interface {
	Run(ctx context.Context, log hclog.Logger) (OperationStatus, error)
}

type Executor struct {
	Log    hclog.Logger
	Config *Config
}

var ErrUnknownOperation = errors.New("unknown operation")

func (e *Executor) IsAvailable(opInfo *models.HashicorpCloudWaypointAgentOperation) (bool, error) {
	// TODO this can also validate the operation body against what the operation can support and
	// return an error here.
	return e.Config.IsAvailable(opInfo.Group, opInfo.ID)
}

func (e *Executor) Execute(ctx context.Context, opInfo *models.HashicorpCloudWaypointAgentOperation) (OperationStatus, error) {
	var hctx hcl.EvalContext

	input := make(map[string]cty.Value)
	// New ones
	varInputs := make(map[string]cty.Value)
	actionInputs := make(map[string]cty.Value)
	appInputs := make(map[string]cty.Value)
	appOutputKeys := make(map[string]cty.Value)
	addOnInputs := make(map[string]cty.Value)
	addOnOutputKeys := make(map[string]cty.Value)

	if len(opInfo.Body) != 0 {

		var rawInput map[string]any

		err := json.Unmarshal(opInfo.Body, &rawInput)
		if err != nil {
			return errStatus, err
		}

		for k, v := range rawInput {
			switch sv := v.(type) {
			case float64:
				input[k] = cty.NumberFloatVal(sv)
			case string:
				input[k] = cty.StringVal(sv)
			case bool:
				input[k] = cty.BoolVal(sv)
			default:
				// TODO how should we deal with these?
			}
		}

		for k, v := range input {
			parts := strings.Split(k, ".")
			// Join the rest of the parts back together
			rest := strings.Join(parts[1:], ".")
			switch parts[0] {
			case "var":
				// - var.<key>
				varInputs[rest] = v
			case "action":
				// - action.<key>
				actionInputs[rest] = v
			case "application":
				if parts[1] == "outputs" {
					// - application.outputs.<key>
					// Application outputs
					outputKey := parts[2]
					appOutputKeys[outputKey] = v
				} else {
					// - application.<key>
					// Static variables
					appInputs[rest] = v
				}
			case "addon":
				// - addon.<instance-name>.outputs.<key>
				if parts[2] == "outputs" {
					addOnInstName := parts[1]
					addOnOutputKey := parts[3]
					addOnOutputKeys[addOnInstName] = cty.ObjectVal(map[string]cty.Value{
						addOnOutputKey: v,
					})
				}
				// - addon.<instance-name>.<key>
				// Currently not supported in the server
				// addOnInputs[rest] = v
			default:
				// CLI encountered an unknown input from the server. We ignore it,
				// because when we parse it in a second it will fail to parse with
				// as a missing or unknown var.
			}
		}

		// Merge all outputs keys
		appInputs["outputs"] = cty.ObjectVal(appOutputKeys)
		for k, v := range addOnOutputKeys {
			addOnInputs[k] = cty.ObjectVal(map[string]cty.Value{
				"outputs": v,
			})
		}
	}

	hctx.Variables = map[string]cty.Value{
		"waypoint": cty.ObjectVal(map[string]cty.Value{
			"run_id": cty.StringVal(opInfo.ActionRunID),
		}),
		"var":         cty.ObjectVal(varInputs),
		"action":      cty.ObjectVal(actionInputs),
		"addon":       cty.ObjectVal(addOnInputs),
		"application": cty.ObjectVal(appInputs),
	}

	op, err := e.Config.Action(opInfo.Group, opInfo.ID, &hctx)
	if err != nil {
		return errStatus, err
	}

	return op.Run(ctx, e.Log)
}
