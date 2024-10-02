package internal

import (
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2023-08-18/models"
)

func ParseVariableOptionsFile(path string) ([]*models.HashicorpCloudWaypointTFModuleVariable, error) {
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
//	  type = "string"
//	  options = [
//	    "a string value",
//	  ]
//	  user_editable = false
//	}
//
//	variable_option "misc_variable" {
//	  type = "string"
//	  options = [
//	    "another string value",
//	  ]
//	  user_editable = false
//	}
func parseVariableOptions(filename string, input []byte) ([]*models.HashicorpCloudWaypointTFModuleVariable, error) {
	var hc hclVariableOptionsFile
	var ctx hcl.EvalContext
	// the Decode method expects a filename to provide context to the error; it
	// does not actually load anything from the file system
	if err := hclsimple.Decode(filename, input, &ctx, &hc); err != nil {
		return nil, err
	}

	// var variables []*models.HashicorpCloudWaypointTFModuleVariable
	variables := make([]*models.HashicorpCloudWaypointTFModuleVariable, 0)
	if len(hc.VariableOptions) > 0 {
		for _, v := range hc.VariableOptions {
			variables = append(variables, &models.HashicorpCloudWaypointTFModuleVariable{
				Name:         v.Name,
				VariableType: v.Type,
				Options:      v.Options,
				UserEditable: v.UserEditable,
			})
		}
	}
	return variables, nil
}

type hclVariableOption struct {
	Name         string   `hcl:",label"`
	Type         string   `hcl:"type"`
	Options      []string `hcl:"options"`
	UserEditable bool     `hcl:"user_editable,optional"`
}

type hclVariableOptionsFile struct {
	VariableOptions []*hclVariableOption `hcl:"variable_option,block"`
}
