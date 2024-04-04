// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package promptio_test

import (
	"io"

	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/testing/promptio"
	"github.com/manifoldco/promptui"
)

func Example() {
	// Write to the stream input buffer to simulate interaction with the prompt
	stream := iostreams.Test()

	// Write down and then enter
	_, _ = stream.Input.WriteRune(promptui.KeyNext)
	_, _ = stream.Input.WriteRune(promptui.KeyEnter)

	promptIO := promptio.Wrap(stream)
	prompt := promptui.Select{
		Label: "Your prompt",
		Items: []string{"test", "other"},
		Stdin: io.NopCloser(promptIO.In()),
	}

	// Run the prompt
	_, _, _ = prompt.Run() //nolint:dogsled
}
