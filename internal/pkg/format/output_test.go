// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package format_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

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
