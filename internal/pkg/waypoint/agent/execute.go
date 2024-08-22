// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package agent

import (
	"context"
	"encoding/json"
	"fmt"
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
	var (
		hctx   hcl.EvalContext
		varMap map[string]any
	)

	if len(opInfo.Body) != 0 {
		var rawInput map[string]any

		err := json.Unmarshal(opInfo.Body, &rawInput)
		if err != nil {
			return errStatus, err
		}

		varMap, err = buildVariableMap(rawInput)
		if err != nil {
			return errStatus, err
		}
	} else {
		varMap = make(map[string]any)
	}

	// NOTE(briancain): This might overwrite any waypoint varibles if we ever
	// decided to set those on the server. Additionally, it might make more sense
	// to include this `waypoint.run_id` as a variable set by the server in the
	// future.
	varMap["waypoint"] = map[string]any{
		"run_id": opInfo.ActionRunID,
	}
	_, ctyMap := anyToCty(varMap)

	// Set all variables from the server on the HCL context so we can parse the
	// agent config if any interpolated variables are defined
	hctx.Variables = ctyMap

	op, err := e.Config.Action(opInfo.Group, opInfo.ID, &hctx)
	if err != nil {
		return errStatus, err
	}

	return op.Run(ctx, e.Log)
}

// buildVariableMap takes a map of string any values where the keys are expected
// to be dot separated and builds a nested map of the values. This format can
// then be used by anyToCty to walk the map structure and build a map of cty
// values that HCL understands and can use to parse a config.
func buildVariableMap(rawInput map[string]any) (map[string]any, error) {
	ret := make(map[string]any)

	for k, v := range rawInput {
		parts := strings.Split(k, ".")
		cur := ret

		for _, p := range parts[:len(parts)-1] {
			if curVal, ok := cur[p]; ok {
				curMap, ok := curVal.(map[string]any)
				if ok {
					cur = curMap
				} else {
					return nil, errors.Errorf("invalid input key %s %v", k, curVal)
				}
			} else {
				curMap := make(map[string]any)
				cur[p] = curMap
				cur = curMap
			}
		}

		cur[parts[len(parts)-1]] = v
	}

	return ret, nil
}

// anyToCty takes a map of string any values and converts them to cty values
// that HCL understands. This function will walk the map and convert the values
// to cty values that HCL can use to parse a config.
func anyToCty(objMap map[string]any) (cty.Value, map[string]cty.Value) {
	obj := make(map[string]cty.Value)

	for k, v := range objMap {
		switch sv := v.(type) {
		case map[string]any:
			// Recuse and walk the map for its children
			obj[k], _ = anyToCty(sv)
		case float64:
			obj[k] = cty.NumberFloatVal(sv)
		case bool:
			obj[k] = cty.BoolVal(sv)
		case string:
			obj[k] = cty.StringVal(sv)
		default:
			// Unhandled var type
			obj[k] = cty.StringVal(fmt.Sprintf("%v", v))
		}
	}

	return cty.ObjectVal(obj), obj
}
