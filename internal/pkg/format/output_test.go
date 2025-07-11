// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package format_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	example "github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2024-11-22/models"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
)

func TestOutputter_SetFormat(t *testing.T) {
	t.Parallel()
	r := require.New(t)
	io := iostreams.Test()
	out := format.New(io)

	// Create our displayer default to pretty printing
	d := &KVDisplayer{
		KVs: []*KV{
			{
				Key:   "Hello",
				Value: "World!",
			},
		},
		Default: format.Pretty,
	}

	// Force the format to JSON
	out.SetFormat(format.JSON)

	// Display the table
	r.NoError(out.Display(d))

	// Ensure we can unmarshal the output as JSON
	var parsed *KV
	r.NoError(json.Unmarshal(io.Output.Bytes(), &parsed))
	r.Equal(d.KVs[0], parsed)
}

type InnerL2Struct struct {
	Name string
}

type InnerL1Struct struct {
	Name  string
	Inner *InnerL2Struct
}

type OuterStruct struct {
	Name  string
	Inner *InnerL1Struct
}

func TestNilInnerStruct(t *testing.T) {
	t.Parallel()
	r := require.New(t)

	kv := &OuterStruct{
		Name: "OuterStruct",
		// we leave inner nil on purpose
	}

	io := iostreams.Test()
	out := format.New(io)
	err := out.Show(kv, format.Pretty)

	r.NoError(err)
	fmt.Println(io.Output.String())
	r.Equal("Name:             OuterStruct\nInner Name:       \nInner Inner Name: \n", io.Output.String())
}

func TestNilInnerL2Struct(t *testing.T) {
	t.Parallel()
	r := require.New(t)

	kv := &OuterStruct{
		Name: "OuterStruct",
		Inner: &InnerL1Struct{
			Name: "InnerL1Struct",
			// we leave inner nil on purpose
		},
	}

	io := iostreams.Test()
	out := format.New(io)
	err := out.Show(kv, format.Pretty)

	r.NoError(err)
	fmt.Println(io.Output.String())
	r.Equal("Name:             OuterStruct\nInner Name:       InnerL1Struct\nInner Inner Name: \n", io.Output.String())
}

func TestNonNilInnerStruct(t *testing.T) {
	t.Parallel()
	r := require.New(t)

	kv := &OuterStruct{
		Name: "OuterStruct",
		Inner: &InnerL1Struct{
			Name: "InnerL1Struct",
			Inner: &InnerL2Struct{
				Name: "InnerStruct",
			},
		},
	}

	io := iostreams.Test()
	out := format.New(io)
	err := out.Show(kv, format.Pretty)

	r.NoError(err)
	fmt.Println(io.Output.String())
	r.Equal("Name:             OuterStruct\nInner Name:       InnerL1Struct\nInner Inner Name: InnerStruct\n", io.Output.String())
}

func TestWithSlice(t *testing.T) {
	t.Parallel()
	r := require.New(t)
	j := `{
  "action_configs": [
    {
      "id": "00000000-0000-0000-0000-000000000000",
      "action_url": "",
      "name": "Agent Smith",
      "request": {
        "agent": {
          "op": {
            "id": "Agent Smith",
            "body": "",
            "action_run_id": "",
            "group": "Enforcements"
          }
        }
      },
      "description": "test description",
      "created_at": "2024-08-16T18:11:19.777071Z"
    },
    {
      "id": "11111111-1111-1111-1111-111111111111",
      "action_url": "",
      "name": "Example",
      "request": {
        "custom": {
          "method": "GET",
          "headers": [],
          "url": "https://hashicorp.com",
          "body": ""
        }
      },
      "description": "Runs an action against https://hashicorp.com",
      "created_at": "2024-06-13T17:31:17.436255Z"
    },
    {
      "id": "22222222-2222-2222-2222-222222222222",
      "action_url": "",
      "name": "Variables",
      "request": {
        "custom": {
          "method": "GET",
          "headers": [],
          "url": "https://${var.company}.com",
          "body": ""
        }
      },
      "description": "An action to test the variables feature.",
      "created_at": "2024-08-07T21:56:00.043627Z"
    }
  ],
  "pagination": {
    "next_page_token": "",
    "previous_page_token": ""
  }
}`
	thing := example.HashicorpCloudWaypointV20241122ListActionConfigResponse{}
	err := json.Unmarshal([]byte(j), &thing)
	r.NoError(err)
	io := iostreams.Test()
	out := format.New(io)
	err = out.Show(thing.ActionConfigs, format.Pretty)

	r.NoError(err)
	fmt.Println(io.Output.String())

	expected := `Action UR L:                              
Created At:                               2024-08-16T18:11:19.777Z
Description:                              test description
ID:                                       00000000-0000-0000-0000-000000000000
Name:                                     Agent Smith
Request Agent Op Action Run ID:           
Request Agent Op Body:                    
Request Agent Op Group:                   Enforcements
Request Agent Op ID:                      Agent Smith
Request Custom Body:                      
Request Custom Headers:                   
Request Custom Method:                    
Request Custom UR L:                      
Request Github Enable Debug Log:          
Request Github Gh Enabled Workflow Param: 
Request Github Git Ref:                   
Request Github Inputs:                    
Request Github Install Name:              
Request Github Repository:                
Request Github Workflow ID:               
---
Action UR L:                              
Created At:                               2024-06-13T17:31:17.436Z
Description:                              Runs an action against https://hashicorp.com
ID:                                       11111111-1111-1111-1111-111111111111
Name:                                     Example
Request Agent Op Action Run ID:           
Request Agent Op Body:                    
Request Agent Op Group:                   
Request Agent Op ID:                      
Request Custom Body:                      
Request Custom Headers:                   []
Request Custom Method:                    GET
Request Custom UR L:                      https://hashicorp.com
Request Github Enable Debug Log:          
Request Github Gh Enabled Workflow Param: 
Request Github Git Ref:                   
Request Github Inputs:                    
Request Github Install Name:              
Request Github Repository:                
Request Github Workflow ID:               
---
Action UR L:                              
Created At:                               2024-08-07T21:56:00.043Z
Description:                              An action to test the variables feature.
ID:                                       22222222-2222-2222-2222-222222222222
Name:                                     Variables
Request Agent Op Action Run ID:           
Request Agent Op Body:                    
Request Agent Op Group:                   
Request Agent Op ID:                      
Request Custom Body:                      
Request Custom Headers:                   []
Request Custom Method:                    GET
Request Custom UR L:                      https://${var.company}.com
Request Github Enable Debug Log:          
Request Github Gh Enabled Workflow Param: 
Request Github Git Ref:                   
Request Github Inputs:                    
Request Github Install Name:              
Request Github Repository:                
Request Github Workflow ID:               
`
	r.Equal(expected, io.Output.String())
}
