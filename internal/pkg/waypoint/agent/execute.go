// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package agent

import (
	"context"
	"encoding/json"

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
	}

	hctx.Variables = map[string]cty.Value{
		"waypoint": cty.ObjectVal(map[string]cty.Value{
			"run_id": cty.StringVal(opInfo.ActionRunID),
		}),
		"var": cty.ObjectVal(input),
	}

	op, err := e.Config.Action(opInfo.Group, opInfo.ID, &hctx)
	if err != nil {
		return errStatus, err
	}

	return op.Run(ctx, e.Log)
}
