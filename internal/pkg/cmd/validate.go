package cmd

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/spf13/pflag"
)

const (
	// ShortHelpMaxLength is the maximum length of the short help text.
	shortHelpMaxLength = 60
)

var (
	// commandNameRegex is used to validate the command names. It enforces that
	// the command name is lower case and contains only letters and hyphens.
	commandNameRegex = regexp.MustCompile(`^[a-z]+([-][a-z]+)*$`)

	// shortHelpRegex is used to validate the short help text. It enforces that
	// the short help text starts with a capital letter, ends with a period, and
	// contains only letters, apostrophes, hyphens, and spaces.
	shortHelpRegex = regexp.MustCompile(`^[A-Z][a-zA-Z-\s']+\.$`)

	// flagNameRegex is used to validate the flag names. It enforces that the
	// flag description starts with a capital letter and ends with a period.
	flagDescriptionRegex = regexp.MustCompile(`(?s)^[A-Z].+\.$`)

	// argsPreambleRegex is used to validate the preamble of the positional
	// arguments. It enforces that the preamble starts with a capital letter and
	// ends with a period.
	argsPreambleRegex = regexp.MustCompile(`(?s)^[A-Z].+\.$`)
)

func (c *Command) Validate() error {
	// Validate ourselves and then the children.
	if err := c.validate(); err != nil {
		return err
	}

	for _, child := range c.children {
		if err := child.Validate(); err != nil {
			return fmt.Errorf("error validating command %s: %w", child.Name, err)
		}
	}

	return nil
}

// TODO
// Deprecate Example Title/Postamble

func (c *Command) validate() error {
	// Validate the name
	if c.Name == "" {
		return fmt.Errorf("command name cannot be empty")
	} else if !commandNameRegex.MatchString(c.Name) {
		return fmt.Errorf("only lower case names with hyphens are allowed")
	}

	// Ensure the aliases are valid and there are no duplicates in the aliases
	aliases := make(map[string]struct{}, len(c.Aliases))
	for _, alias := range c.Aliases {
		// Ensure the alias is not the name
		if alias == c.Name {
			return fmt.Errorf("command name cannot be an alias")
		}

		// Check for duplicates
		if _, ok := aliases[alias]; ok {
			return fmt.Errorf("duplicate alias %q found", alias)
		}
		aliases[alias] = struct{}{}

		// Vaidate the alias
		if alias == "" {
			return fmt.Errorf("alias name is empty")
		} else if !commandNameRegex.MatchString(alias) {
			return fmt.Errorf("alias %q: only lower case names with hyphens are allowed", alias)
		}
	}

	// Validate that the help text is set
	if c.ShortHelp == "" || c.LongHelp == "" {
		return fmt.Errorf("short and long help text must be set")
	}

	// Validate the short help.
	if len(c.ShortHelp) > shortHelpMaxLength {
		return fmt.Errorf("short help text is too long. Max length is %d; got %q (%d)",
			shortHelpMaxLength, c.ShortHelp, len(c.ShortHelp))
	} else if !shortHelpRegex.MatchString(c.ShortHelp) {
		return fmt.Errorf("short help text must start with a capital letter, end with a period, and contain only letters, apostrophes, hyphens, and spaces; got %q", c.ShortHelp)
	}

	// Validate the additional documentation sections
	for i, d := range c.AdditionalDocs {
		if err := d.validate(); err != nil {
			return fmt.Errorf("error validating documentation section %d: %w", i, err)
		}
	}

	// Validate the examples
	for i, e := range c.Examples {
		if err := e.validate(); err != nil {
			return fmt.Errorf("error validating example %d: %w", i, err)
		}
	}

	// Validate IO is set
	if err := c.validateIO(); err != nil {
		return err
	}

	// Validate the Flags
	if err := c.validateFlags(); err != nil {
		return err
	}

	// validate the positional arguments
	if err := c.Args.validate(); err != nil {
		return fmt.Errorf("error validating positional arguments: %w", err)
	}

	// Validate that either RunF or Children are set, but not both.
	if c.RunF == nil && len(c.children) == 0 {
		return fmt.Errorf("either RunF or Children must be set")
	} else if c.RunF != nil && len(c.children) > 0 {
		return fmt.Errorf("both RunF and Children cannot be set")
	}

	return nil
}

// validateIO checks that the io is set on the command or any parent command.
func (c *Command) validateIO() error {
	for c := c; c != nil; c = c.parent {
		if c.io != nil {
			return nil
		}
	}

	return fmt.Errorf("io not set on command or any parent command")
}

func (c *Command) validateFlags() error {
	for _, flag := range c.Flags.Local {
		if err := flag.validate(); err != nil {
			return fmt.Errorf("error validating local flag %q: %w", flag.Name, err)
		}
	}

	for _, flag := range c.Flags.Persistent {
		if err := flag.validate(); err != nil {
			return fmt.Errorf("error validating persistent flag %q: %w", flag.Name, err)
		}
	}

	// Ensure local flags do not override parent persistent flags
	var flagErr error
	localFlags, inheritedFlags := c.localFlags(), c.parentPersistentFlags()
	localFlags.VisitAll(func(f *pflag.Flag) {
		if flagErr != nil {
			return
		}

		if inheritedFlags.Lookup(f.Name) != nil {
			flagErr = fmt.Errorf("local flag %q overrides inherited persistent flag", f.Name)
		}
	})
	if flagErr != nil {
		return flagErr
	}

	return nil
}

func (f *Flag) validate() error {
	if f.Name == "" {
		return fmt.Errorf("name cannot be empty")
	} else if f.Name != strings.ToLower(f.Name) {
		return fmt.Errorf("name is not lowercase")
	}
	if f.Shorthand != strings.ToLower(f.Shorthand) {
		return fmt.Errorf("shorthand %q is not lowercase", f.Shorthand)
	} else if len(f.Shorthand) > 1 {
		return fmt.Errorf("shorthand %q must be a single character", f.Shorthand)
	}
	if f.DisplayValue != strings.ToUpper(f.DisplayValue) {
		return fmt.Errorf("display value %q is not uppercase", f.DisplayValue)
	}
	if !flagDescriptionRegex.MatchString(f.Description) {
		return fmt.Errorf("description must start with a capital letter and end with a period; got %q", f.Description)
	}
	if f.Value == nil {
		return fmt.Errorf("value cannot be nil")
	}

	return nil
}

// validate validates the documentation section.
func (d *DocSection) validate() error {
	if d.Title == "" {
		return fmt.Errorf("title cannot be empty")
	} else if strings.HasSuffix(d.Title, ".") {
		return fmt.Errorf("title cannot end with a period")
	}
	if d.Documentation == "" {
		return fmt.Errorf("documentation cannot be empty")
	}

	return nil
}

// validate validates the positional arguments.
func (a *PositionalArguments) validate() error {
	// Start capital and end with a period if set.
	if a.Preamble != "" && !argsPreambleRegex.MatchString(a.Preamble) {
		return fmt.Errorf("preable must start with a capital letter and end with a period")
	}

	l := len(a.Args)
	for i, p := range a.Args {
		if err := p.validate(i == l-1); err != nil {
			return fmt.Errorf("error validating positional argument %d: %w", i, err)
		}
	}

	return nil
}

// validate validates the positional argument. isLast indicates if the positional
// argument is the last argument.
func (a *PositionalArgument) validate(isLast bool) error {
	if a.Name == "" {
		return fmt.Errorf("name cannot be empty")
	} else if a.Name != strings.ToUpper(a.Name) {
		return fmt.Errorf("name %q is not uppercase", a.Name)
	}

	if a.Documentation == "" {
		return fmt.Errorf("documentation cannot be empty")
	} else if !strings.HasSuffix(a.Documentation, ".") {
		return fmt.Errorf("documentation must end with a period")
	}

	if a.Optional && !isLast {
		return fmt.Errorf("optional positional argument %q must be the last argument", a.Name)
	}
	if a.Repeatable && !isLast {
		return fmt.Errorf("repeatable positional argument %q must be the last argument", a.Name)
	}

	return nil
}

// validate validates the example.
func (e *Example) validate() error {
	if !strings.HasPrefix(e.Command, "$ ") {
		return fmt.Errorf("example command must start with $")
	}

	return nil
}
