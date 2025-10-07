// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: MPL-2.0

package internal

import (
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2024-11-22/models"
)

func ParseVariableOptionsFile(path string) ([]*models.HashicorpCloudWaypointV20241122TFModuleVariable, error) {
	input, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return parseVariableOptions(path, input)
}

// parseVariableOptions reads the input bytes and parses the HCL file to extract
// the variable options. Note that we intentionally do not provide much in terms
// of validation of the HCL file, as we expect the HCL to be validated by the
// server side.
//
// # Example contents of a vars.hcl file
//
//	variable_option "string_variable" {
//	  options = [
//	    "a string value",
//	  ]
//	  user_editable = false
//	}
//
//	variable_option "misc_variable" {
//	  options = [
//	    "another string value",
//	  ]
//	  user_editable = false
//	}
func parseVariableOptions(filename string, input []byte) ([]*models.HashicorpCloudWaypointV20241122TFModuleVariable, error) {
	var hc hclVariableOptionsFile
	var ctx hcl.EvalContext
	// the Decode method expects a filename to provide context to the error; it
	// does not actually load anything from the file system
	if err := hclsimple.Decode(filename, input, &ctx, &hc); err != nil {
		return nil, err
	}

	// var variables []*models.HashicorpCloudWaypointV20241122TFModuleVariable
	variables := make([]*models.HashicorpCloudWaypointV20241122TFModuleVariable, 0)
	if len(hc.VariableOptions) > 0 {
		for _, v := range hc.VariableOptions {
			variables = append(variables, &models.HashicorpCloudWaypointV20241122TFModuleVariable{
				Name:         v.Name,
				Options:      v.Options,
				UserEditable: v.UserEditable,
			})
		}
	}
	return variables, nil
}

type hclVariableOption struct {
	Name         string   `hcl:",label"`
	Options      []string `hcl:"options"`
	UserEditable bool     `hcl:"user_editable,optional"`
}

type hclVariableOptionsFile struct {
	VariableOptions []*hclVariableOption `hcl:"variable_option,block"`
}
