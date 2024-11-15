// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package internal

import (
	"os"
	"sort"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2023-08-18/models"
)

func ParseInputVariablesFile(path string) ([]*models.HashicorpCloudWaypointInputVariable, error) {
	input, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return parseInputVariables(path, input)
}

func parseInputVariables(filename string, input []byte) ([]*models.HashicorpCloudWaypointInputVariable, error) {
	var hc hclInputVariablesFile
	var ctx hcl.EvalContext
	if err := hclsimple.Decode(filename, input, &ctx, &hc); err != nil {
		return nil, err
	}

	var variables []*models.HashicorpCloudWaypointInputVariable
	if len(hc.Variables) > 0 {
		for k, v := range hc.Variables {
			variables = append(variables, &models.HashicorpCloudWaypointInputVariable{
				Name:  k,
				Value: v,
			})
		}
	}

	sort.Slice(variables, func(i, j int) bool {
		return variables[i].Name < variables[j].Name
	})
	return variables, nil
}

type hclInputVariablesFile struct {
	Variables map[string]string `hcl:"variables,remain"`
}
