package heredoc_test

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
)

func Example() {
	io, _ := iostreams.System(context.Background())
	out := heredoc.New(io).Mustf(`
	This is an example of documenting a command. You can format in values %s you want.

	Really long lines will automatically get wrapped for you. So you don't need to worry about being super rigorous about where you wrap a line.

	However if you have a block of text that you want to preserve the formatting of, you can use the PreserveNewLines function as shown below.

	{{ PreserveNewLines }}
	{
	  "description": "JSON block to preserve the formatting of",
	  "cool": true
	}
	{{ PreserveNewLines }}

	You can also colorize and stylize text. This is useful if you want to highlight a command.

	Such as, run {{ Bold "hcp your command" }} to do awesome things.

	The available style functions are: Bold, Faint, Italic, Underline, Blink, CrossOut.

	You can color output with the Color function. It is invoked as Color <text color> <optional foreground> "text".

	The valdid colors are: red, green, yellow, orange, gray, white, black, or #<hex>. The case doesn't matter.

	For example, {{ Color "Red" "this could be an error" }}.

	You may have noticed that all these lines have an indent. heredoc will automatically dedent for you.

	  But you can still further indent and it will be maintained for you.

	Lastly, blank spaces at the start and end will be stripped so that you can start your text on a new line and end it like this.
	`, "wherever")

	fmt.Fprintln(io.Out(), out)
	// Output:
	// This is an example of documenting a command. You can format in values wherever
	// you want.
	//
	// Really long lines will automatically get wrapped for you. So you don't need to
	// worry about being super rigorous about where you wrap a line.
	//
	// However if you have a block of text that you want to preserve the formatting of,
	// you can use the PreserveNewLines function as shown below.
	//
	// {
	//   "description": "JSON block to preserve the formatting of",
	//   "cool": true
	// }
	//
	// You can also colorize and stylize text. This is useful if you want to highlight
	// a command.
	//
	// Such as, run hcp your command to do awesome things.
	//
	// The available style functions are: Bold, Faint, Italic, Underline, Blink,
	// CrossOut.
	//
	// You can color output with the Color function. It is invoked as Color <text
	// color> <optional foreground> "text".
	//
	// The valdid colors are: red, green, yellow, orange, gray, white, black, or
	// #<hex>. The case doesn't matter.
	//
	// For example, this could be an error.
	//
	// You may have noticed that all these lines have an indent. heredoc will
	// automatically dedent for you.
	//
	//   But you can still further indent and it will be maintained for you.
	//
	// Lastly, blank spaces at the start and end will be stripped so that you can start
	// your text on a new line and end it like this.
}
