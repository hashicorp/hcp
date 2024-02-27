package format_test

import (
	"fmt"

	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
)

func ExampleDisplayer() {
	// Create the outputter. This is typically passed to the command.
	io := iostreams.Test()
	outputter := format.New(io)

	// Resource is an example resource that we want to display
	type Resource struct {
		Name        string
		ID          string
		Description string
	}

	// Define the fields that we want to ExampleDisplayer
	var fields = []format.Field{
		format.NewField("Name", "{{ .Name }}"),
		format.NewField("ID", "{{ .ID }}"),
		format.NewField("Description", "{{ .Description }}"),
	}

	// Build our mock resources. Typically this is the response payload from an API
	// request.
	payload := []Resource{
		{
			Name:        "hello",
			ID:          "123",
			Description: "world",
		},
		{
			Name:        "another",
			ID:          "456",
			Description: "example",
		},
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
	// ---
	// Name:        another
	// ID:          456
	// Description: example
}
