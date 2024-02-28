package format_test

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
)

func ExampleOutputter() {
	// Create the outputter. This is typically passed to the command.
	io := iostreams.Test()
	outputter := format.New(io)

	// Resource is an example resource that we want to display
	type Resource struct {
		Name        string
		ID          string
		Description string
		Bytes       int
	}

	// Build our mock resources. Typically this is the response payload from an API
	// request.
	payload := []Resource{
		{
			Name:        "hello",
			ID:          "123",
			Description: "world",
			Bytes:       100,
		},
		{
			Name:        "another",
			ID:          "456",
			Description: "example",
			Bytes:       1024 * 1024,
		},
	}

	// For displaying a table of the exact values, Show can be used:
	_ = outputter.Show(payload, format.Table)

	// Since the IO is a test io, manually print it.
	// We trim the lines to make examples testing pass correctly.
	lines := strings.Split(io.Output.String(), "\n")
	for i, l := range lines {
		lines[i] = strings.TrimSpace(l)
	}
	fmt.Println(strings.Join(lines, "\n"))

	// Output:
	// Name     ID   Description  Bytes
	// hello    123  world        100
	// another  456  example      1048576
	//

}

func ExampleDisplayer() {
	// Create the outputter. This is typically passed to the command.
	io := iostreams.Test()
	outputter := format.New(io)

	// Resource is an example resource that we want to display
	type Resource struct {
		Name        string
		ID          string
		Description string
		Bytes       int
	}

	// Build our mock resources. Typically this is the response payload from an API
	// request.
	payload := []Resource{
		{
			Name:        "hello",
			ID:          "123",
			Description: "world",
			Bytes:       100,
		},
		{
			Name:        "another",
			ID:          "456",
			Description: "example",
			Bytes:       1024 * 1024,
		},
	}

	// If you wish to format the values differently, use a Displayer:

	// Define the fields that we want to ExampleDisplayer
	var fields = []format.Field{
		format.NewField("Name", "{{ .Name }}"),
		format.NewField("ID", "{{ .ID }}"),
		format.NewField("Description", "{{ .Description }}"),
		format.NewField("Bytes", "{{ .Bytes }} bytes"),
	}

	// Build the displayer
	d := format.NewDisplayer(payload, format.Pretty, fields)

	// Run the displayer
	if err := outputter.Display(d); err != nil {
		fmt.Printf("error displaying resources: %s\n", err)
	}

	// Since the IO is a test io, manually print it
	fmt.Println(io.Output.String())

	// Output:
	// Name:        hello
	// ID:          123
	// Description: world
	// Bytes:       100 bytes
	// ---
	// Name:        another
	// ID:          456
	// Description: example
	// Bytes:       1048576 bytes
}
