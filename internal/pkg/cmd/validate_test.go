package cmd

import (
	"testing"

	"github.com/hashicorp/hcp/internal/pkg/flagvalue"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/stretchr/testify/require"
)

func getGoodCommand() *Command {
	var projectID string
	parent := &Command{
		Name:      "parent-cmd",
		Aliases:   []string{"parent"},
		ShortHelp: "This is a short help message.",
		LongHelp:  "This is a long help message.",
		Flags: Flags{
			Persistent: []*Flag{
				{
					Name:         "project",
					Shorthand:    "p",
					Description:  "Project ID.",
					DisplayValue: "ID",
					Value:        flagvalue.Simple("", &projectID),
				},
			},
		},
		io: iostreams.Test(),
	}

	// Add a child command
	var count int
	child := &Command{
		Name:      "child-cmd",
		Aliases:   []string{"child"},
		ShortHelp: "This is a short help message.",
		LongHelp:  "This is a long help message.",
		Flags: Flags{
			Local: []*Flag{
				{
					Name:         "count",
					Description:  "Count of things to print.",
					DisplayValue: "N",
					Value:        flagvalue.Simple(0, &count),
				},
			},
		},
		Args: PositionalArguments{
			Preamble: "This is a preamble.",
			Args: []PositionalArgument{
				{
					Name:          "TEXT",
					Documentation: "Text to repeatedly print.",
					Optional:      false,
					Repeatable:    false,
				},
				{
					Name:          "PREFIX",
					Documentation: "Prefix to prepend to the text.",
					Optional:      false,
					Repeatable:    false,
				},
			},
		},
		Examples: []Example{
			{
				Preamble: "This is an example invocation:",
				Command:  "$ hcp parent child --count 5",
			},
		},
		AdditionalDocs: []DocSection{
			{
				Title:         "More Details",
				Documentation: "This will explain everything.",
			},
		},
		RunF: func(cmd *Command, args []string) error {
			return nil
		},
	}
	parent.AddChild(child)
	return parent
}

func TestCommand_Validate(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		command func() *Command
		error   string
	}{
		{
			name:    "good",
			command: getGoodCommand,
			error:   "",
		},
		{
			name: "no io",
			command: func() *Command {
				c := getGoodCommand()
				c.io = nil
				return c
			},
			error: "io not set on command or any parent command",
		},
		{
			name: "no runf or children",
			command: func() *Command {
				c := getGoodCommand()
				c.children[0].RunF = nil
				return c
			},
			error: "either RunF or Children must be set",
		},
		{
			name: "both runf and children",
			command: func() *Command {
				c := getGoodCommand()
				c.RunF = func(cmd *Command, args []string) error { return nil }
				return c
			},
			error: "both RunF and Children cannot be set",
		},
		{
			name: "no name",
			command: func() *Command {
				c := getGoodCommand()
				c.Name = ""
				return c
			},
			error: "command name cannot be empty",
		},
		{
			name: "bad name characters",
			command: func() *Command {
				c := getGoodCommand()
				c.Name = "ThisIsBad"
				return c
			},
			error: "only lower case names with hyphens are allowed",
		},
		{
			name: "bad alias characters",
			command: func() *Command {
				c := getGoodCommand()
				c.Aliases = append(c.Aliases, "ThisIsBad")
				return c
			},
			error: "only lower case names with hyphens are allowed",
		},
		{
			name: "duplicate aliases",
			command: func() *Command {
				c := getGoodCommand()
				c.Aliases = append(c.Aliases, "good", "good")
				return c
			},
			error: "duplicate alias \"good\" found",
		},
		{
			name: "duplicate name and alias",
			command: func() *Command {
				c := getGoodCommand()
				c.Aliases = append(c.Aliases, c.Name)
				return c
			},
			error: "command name cannot be an alias",
		},
		{
			name: "no short help",
			command: func() *Command {
				c := getGoodCommand()
				c.ShortHelp = ""
				return c
			},
			error: "short and long help text must be set",
		},
		{
			name: "no long help",
			command: func() *Command {
				c := getGoodCommand()
				c.LongHelp = ""
				return c
			},
			error: "short and long help text must be set",
		},
		{
			name: "short help is too long",
			command: func() *Command {
				c := getGoodCommand()
				c.ShortHelp = "This is a very long help message that is too long to be valid."
				return c
			},
			error: "short help text is too long. Max length is 60; got",
		},
		{
			name: "short help doesn't start with capital",
			command: func() *Command {
				c := getGoodCommand()
				c.ShortHelp = "bad short."
				return c
			},
			error: "short help text must start with a capital letter, end with a period, and contain only letters, apostrophes, hyphens, and spaces",
		},
		{
			name: "short help doesn't end with a period",
			command: func() *Command {
				c := getGoodCommand()
				c.ShortHelp = "Bad short"
				return c
			},
			error: "short help text must start with a capital letter, end with a period, and contain only letters, apostrophes, hyphens, and spaces",
		},
		{
			name: "short help has bad char",
			command: func() *Command {
				c := getGoodCommand()
				c.ShortHelp = "Bad $hort."
				return c
			},
			error: "short help text must start with a capital letter, end with a period, and contain only letters, apostrophes, hyphens, and spaces",
		},
		{
			name: "additional docs has title",
			command: func() *Command {
				c := getGoodCommand()
				c.children[0].AdditionalDocs[0].Title = ""
				return c
			},
			error: "error validating documentation section 0: title cannot be empty",
		},
		{
			name: "additional docs has no period",
			command: func() *Command {
				c := getGoodCommand()
				c.children[0].AdditionalDocs[0].Title = "test."
				return c
			},
			error: "error validating documentation section 0: title cannot end with a period",
		},
		{
			name: "additional docs has no docs",
			command: func() *Command {
				c := getGoodCommand()
				c.children[0].AdditionalDocs[0].Documentation = ""
				return c
			},
			error: "error validating documentation section 0: documentation cannot be empty",
		},
		{
			name: "example preamble set",
			command: func() *Command {
				c := getGoodCommand()
				c.children[0].Examples[0].Preamble = ""
				return c
			},
			error: "error validating example 0: preamble cannot be empty",
		},
		{
			name: "example preamble start with capital",
			command: func() *Command {
				c := getGoodCommand()
				c.children[0].Examples[0].Preamble = "bad preamble:"
				return c
			},
			error: "error validating example 0: preamble must start with a capital letter and end with a colon",
		},
		{
			name: "example preamble end with colon",
			command: func() *Command {
				c := getGoodCommand()
				c.children[0].Examples[0].Preamble = "Bad preamble"
				return c
			},
			error: "error validating example 0: preamble must start with a capital letter and end with a colon",
		},
		{
			name: "examples start with a $",
			command: func() *Command {
				c := getGoodCommand()
				c.children[0].Examples[0].Command = "hcp parent child --count 5"
				return c
			},
			error: "error validating example 0: example command must start with $",
		},
		{
			name: "flag name is set",
			command: func() *Command {
				c := getGoodCommand()
				c.Flags.Persistent[0].Name = ""
				return c
			},
			error: "error validating persistent flag \"\": name cannot be empty",
		},
		{
			name: "flag name must be lower",
			command: func() *Command {
				c := getGoodCommand()
				c.Flags.Persistent[0].Name = "BAD"
				return c
			},
			error: "error validating persistent flag \"BAD\": name is not lowercase",
		},
		{
			name: "flag shorthand must be lower",
			command: func() *Command {
				c := getGoodCommand()
				c.Flags.Persistent[0].Shorthand = "B"
				return c
			},
			error: "error validating persistent flag \"project\": shorthand \"B\" is not lowercase",
		},
		{
			name: "flag shorthand too long",
			command: func() *Command {
				c := getGoodCommand()
				c.Flags.Persistent[0].Shorthand = "bbb"
				return c
			},
			error: "error validating persistent flag \"project\": shorthand \"bbb\" must be a single character",
		},
		{
			name: "flag display value must be upper case",
			command: func() *Command {
				c := getGoodCommand()
				c.Flags.Persistent[0].DisplayValue = "id"
				return c
			},
			error: "error validating persistent flag \"project\": display value \"id\" is not uppercase",
		},
		{
			name: "flag description lowercase start",
			command: func() *Command {
				c := getGoodCommand()
				c.Flags.Persistent[0].Description = "this is a description."
				return c
			},
			error: "error validating persistent flag \"project\": description must start with a capital letter and end with a period",
		},
		{
			name: "flag description end with period",
			command: func() *Command {
				c := getGoodCommand()
				c.Flags.Persistent[0].Description = "This is a description"
				return c
			},
			error: "error validating persistent flag \"project\": description must start with a capital letter and end with a period",
		},
		{
			name: "flag description no value",
			command: func() *Command {
				c := getGoodCommand()
				c.Flags.Persistent[0].Value = nil
				return c
			},
			error: "error validating persistent flag \"project\": value cannot be nil",
		},
		{
			name: "flags don't override parent persistent",
			command: func() *Command {
				c := getGoodCommand()
				c.children[0].Flags.Local[0].Name = "project"
				return c
			},
			error: "local flag \"project\" overrides inherited persistent flag",
		},
		{
			name: "PositionalArgs preamble is valid",
			command: func() *Command {
				c := getGoodCommand()
				c.children[0].Args.Preamble = "bad preamble."
				return c
			},
			error: "error validating positional arguments: preable must start with a capital letter and end with a period",
		},
		{
			name: "PositionalArg name is set",
			command: func() *Command {
				c := getGoodCommand()
				c.children[0].Args.Args[0].Name = ""
				return c
			},
			error: "error validating positional argument 0: name cannot be empty",
		},
		{
			name: "PositionalArg name must be uppercase",
			command: func() *Command {
				c := getGoodCommand()
				c.children[0].Args.Args[0].Name = "bad"
				return c
			},
			error: "error validating positional argument 0: name \"bad\" is not uppercase",
		},
		{
			name: "PositionalArg documentation must be set",
			command: func() *Command {
				c := getGoodCommand()
				c.children[0].Args.Args[0].Documentation = ""
				return c
			},
			error: "error validating positional argument 0: documentation cannot be empty",
		},
		{
			name: "PositionalArg documentation must end with a period",
			command: func() *Command {
				c := getGoodCommand()
				c.children[0].Args.Args[0].Documentation = "bad docs"
				return c
			},
			error: "error validating positional argument 0: documentation must end with a period",
		},
		{
			name: "PositionalArg optional must be last",
			command: func() *Command {
				c := getGoodCommand()
				c.children[0].Args.Args[0].Optional = true
				return c
			},
			error: "error validating positional argument 0: optional positional argument \"TEXT\" must be the last argument",
		},
		{
			name: "PositionalArg repeated must be last",
			command: func() *Command {
				c := getGoodCommand()
				c.children[0].Args.Args[0].Repeatable = true
				return c
			},
			error: "error validating positional argument 0: repeatable positional argument \"TEXT\" must be the last argument",
		},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			r := require.New(t)
			err := c.command().Validate()
			if c.error != "" {
				r.ErrorContains(err, c.error)
			} else {
				r.NoError(err)
			}
		})
	}
}
