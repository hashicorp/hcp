The `heredoc` package is a wrapper around `text/template` that provides helper
functions suited for the CLI, such as outputting stylized and colored text,
creating hyperlinks, etc. Further, the heredoc package automatically wraps long
lines and removes any shared indentation from the text.

Refer to the package examples and documentation for all possible functions and
features.

In this document, we will cover some of the most common use cases.

# Command LongHelp

Command's LongHelp should always begin as follows:
* Command: `The {{ template "mdCodeOrBold" "hcp <command path>" }} command ...`
* Command Group : `The {{ template "mdCodeOrBold" "hcp <command group path>" }} command group...`

As an example, consider the command definition for project list:

```go
cmd := &cmd.Command{
    Name:      "list",
    ShortHelp: "List HCP projects.",
    LongHelp: heredoc.New(ctx.IO).Must(`
    The {{ template "mdCodeOrBold" "hcp projects list" }} command lists HCP projects.
    `),
    ...
}
```

The `mdCodeOrBold` template will render bold text when outputting to the terminal and will
render a code block when outputting to markdown.

# Bold Text

To output bold text, use `{{ Bold "text" }}`:

```go
heredoc.New(ctx.IO).Must(`
    {{ Bold "This is bold text." }}
`)
```

# Italic Text

To output italic text, use `{{ Italic "text" }}`:

```go
heredoc.New(ctx.IO).Must(`
    {{ Italic "This is italic text." }}
`)
```

# Color Text

To output colored text, use `{{ Color "color_name" "text" }}`:

```go
heredoc.New(ctx.IO).Must(`
    {{ Color "green" "This is green text." }}
`)
```

To output colored text with a different foreground and background color,
use `{{ Color "foreground_color" "background_color" "text" }}`:

```go
heredoc.New(ctx.IO).Must(`
    {{ Color "white" "red" "This is red background text." }}
`)
```

The commands can be pipelined. For example, to output bold and green text:

```go
heredoc.New(ctx.IO).Must(`
    {{ Bold (Color "green" "This is bold and green text.") }}
`)
```

Valid Color values are: "red", "green", "yellow", "orange", "gray", white",
"black" (case insensitive), or "#<hex>".

# Preserve New Lines

If you are outputting text that should have its new lines preserved, such as
JSON, use paired `{{ PreserveNewLines }}`:

```go
heredoc.New(ctx.IO).Must(`
    {{ PreserveNewLines }}
    {
      "key": "value"
    }
    {{ PreserveNewLines }}
`)
```

# Emit Markdown specific output

The template can be used to emit markdown specific output using the `{{ IsMD }}`
function:

```go
heredoc.New(ctx.IO).Must(`
    {{ if IsMD }}
    This is markdown output.
    {{ else }}
    This is not markdown output.
    {{ end }}
`)
```
