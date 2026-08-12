// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: MPL-2.0

package iostreams_test

import (
	"fmt"

	"github.com/hashicorp/hcp/internal/pkg/iostreams"
)

// Example_output shows how text can be outputted with color and style
func Example() {
	// Use iostreams.System for real usage.
	io := iostreams.Test()

	cs := io.ColorScheme()
	_, _ = fmt.Fprintln(io.Out(), cs.String("Applying Style").Bold())
	_, _ = fmt.Fprintln(io.Out(), cs.String("Chaining Styles").Bold().Italic())

	_, _ = fmt.Fprintln(io.Out(), cs.String("Applying Color").Color(cs.Orange()))
	_, _ = fmt.Fprintln(io.Out(), cs.String("Applying Color and Style").Bold().Color(cs.Orange()))

	// Changing the background
	_, _ = fmt.Fprintln(io.Out(), cs.String("WARNING").Bold().Background(cs.Orange()).Color(cs.Black()))

	// Print the test output
	fmt.Print(io.Output.String())

	// Output:
	// Applying Style
	// Chaining Styles
	// Applying Color
	// Applying Color and Style
	// WARNING
}

// Example_secrets shows how a secret can be retrieved.
func ExampleIOStreams_ReadSecret() {
	// Use iostreams.System for real usage.
	io := iostreams.Test()
	io.InputTTY = true
	io.ErrorTTY = true

	// Mock stdin to demonstrate reading from stdin.
	io.Input.WriteString("pa$$w0rd")

	_, _ = fmt.Fprintln(io.Err(), "Whats your password?")
	data, err := io.ReadSecret()
	if err != nil {
		panic(err)
	}
	_, _ = fmt.Fprintf(io.Out(), "%q is a terrible password now!", string(data))

	// Print the test output
	fmt.Print(io.Error.String())
	fmt.Print(io.Output.String())

	// Output:
	// Whats your password?
	// "pa$$w0rd" is a terrible password now!
}
