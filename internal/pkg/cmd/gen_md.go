package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/spf13/pflag"
)

const markdownExtension = ".mdx"

// LinkHandler is a function that can be used to modify the links in the
// generated markdown. The path string is the unmodified path to the file.
type LinkHandler func(path string) string

func GenMarkdownTree(c *Command, dir string, link LinkHandler) error {
	// Create the directory if it doesn't exist
	if err := os.MkdirAll(dir, 0766); err != nil {
		return err
	}

	// Determine the filename
	filename := "index" + markdownExtension
	if c.RunF != nil {
		filename = c.Name + markdownExtension
	}

	// Create the file
	f, err := os.Create(filepath.Join(dir, filename))
	if err != nil {
		return err
	}
	defer f.Close()

	// Generate the markdown
	if err := GenMarkdown(c, f, link); err != nil {
		return err
	}

	for _, c := range c.children {
		dir := dir
		if len(c.children) > 0 {
			dir = filepath.Join(dir, c.Name)
		}

		if err := GenMarkdownTree(c, dir, link); err != nil {
			return err
		}
	}

	return nil
}

// GenMarkdown creates custom markdown output.
func GenMarkdown(c *Command, w io.Writer, link LinkHandler) error {
	cs := c.getIO().ColorScheme()

	buf := new(bytes.Buffer)
	name := c.commandPath()

	buf.WriteString("---\n")
	buf.WriteString(fmt.Sprintf("page_title: %s\n", name))
	buf.WriteString(fmt.Sprintf("description: |-\n  %s\n", c.ShortHelp))
	buf.WriteString("---\n\n")

	_, err := buf.WriteTo(w)
	if err != nil {
		return fmt.Errorf("error writing header: %w", err)
	}

	buf.WriteString("# " + name + "\n\n")
	buf.WriteString(fmt.Sprintf("%s\n\n", c.ShortHelp))

	// Disable markdown escaping and then re-enable. This is needed because
	// the flags / args would otherwise generate markdown that will not render
	// because we are using a code block.
	io := c.getIO()
	c.SetIO(iostreams.Test())
	buf.WriteString("## Usage\n\n")
	buf.WriteString(fmt.Sprintf("```shell-session\n$ %s\n```\n\n", c.useLine()))
	c.SetIO(io)

	// Description
	if len(c.LongHelp) > 0 {
		buf.WriteString("## Description\n\n")
		buf.WriteString(c.LongHelp + "\n\n")
	}

	// Aliases
	if len(c.Aliases) > 0 {
		buf.WriteString("## Aliases\n\n")
		for a, u := range c.aliasUsages() {
			buf.WriteString(fmt.Sprintf("%s - `%s`\n", a, u))
		}
		buf.WriteString("\n")
	}

	// Examples
	if len(c.Examples) > 0 {
		buf.WriteString("## Examples\n\n")

		for _, e := range c.Examples {
			buf.WriteString(fmt.Sprintf("%s\n\n", e.Preamble))
			buf.WriteString(fmt.Sprintf("```shell-session\n%s\n```\n\n", e.Command))
		}
	}

	// Children commands
	if len(c.children) > 0 {
		var commands, groups []string
		for _, c := range c.children {
			path := strings.ReplaceAll(c.commandPath(), " ", "/")
			entry := fmt.Sprintf("- [%s](%s) - %s", c.Name, link(path), c.ShortHelp)

			if c.RunF != nil {
				commands = append(commands, entry)
			} else {
				groups = append(groups, entry)
			}
		}
		if len(groups) > 0 {
			buf.WriteString("## Command Groups\n\n")
			buf.WriteString(strings.Join(groups, "\n") + "\n\n")
		}

		if len(commands) > 0 {
			buf.WriteString("## Commands\n\n")
			buf.WriteString(strings.Join(commands, "\n") + "\n\n")
		}
	}

	// Positional arguments
	genMarkdownPositionalArgs(c, cs, buf)

	// Print flags
	genMarkdownFlags(c, cs, buf)

	// Additional docs
	for _, d := range c.AdditionalDocs {
		buf.WriteString(fmt.Sprintf("## %s\n", d.Title))
		buf.WriteString(d.Documentation + "\n")
	}

	_, err = buf.WriteTo(w)
	return err
}

func genMarkdownPositionalArgs(c *Command, cs *iostreams.ColorScheme, buf *bytes.Buffer) {
	if len(c.Args.Args) == 0 {
		return
	}

	buf.WriteString("## Positional Arguments\n\n")
	p := c.Args
	if p.Preamble != "" {
		fmt.Fprintln(buf, p.Preamble)
	}

	for _, a := range p.Args {
		nameUpper := strings.ToUpper(a.Name)
		repeatable := ""
		if a.Repeatable {
			repeatable = fmt.Sprintf(" [%s ...]", nameUpper)
		}
		fmt.Fprintf(buf, "`%s%s`\n\n", nameUpper, repeatable)

		if a.Optional {
			fmt.Fprintln(buf, cs.String("Optional Argument\n").Italic().String())
		}
		fmt.Fprintln(buf, a.Documentation)
		fmt.Fprintln(buf)
	}
}

func genMarkdownFlags(c *Command, cs *iostreams.ColorScheme, buf *bytes.Buffer) {
	// If we are the root command, just print global flags.
	if c.parent == nil && c.RunF == nil {
		buf.WriteString("## Global Flags\n\n")
		genMarkdownFlagsetUsage(c.globalFlags(), buf)
	}

	// Print flags only if the command is runnable
	if c.RunF == nil {
		return
	}

	flagSets := []struct {
		flags *pflag.FlagSet
		name  string
	}{
		{
			flags: c.localFlags(),
			name:  "",
		},
		{
			flags: c.inheritedFlags(),
			name:  "Inherited ",
		},
	}

	for _, set := range flagSets {
		required, optional := splitRequiredFlags(set.flags)
		if required.HasFlags() {
			buf.WriteString(fmt.Sprintf("## Required %sFlags\n\n", set.name))
			genMarkdownFlagsetUsage(required, buf)

			if optional.HasFlags() {
				buf.WriteString(fmt.Sprintf("## Optional %sFlags\n\n", set.name))
				genMarkdownFlagsetUsage(optional, buf)
			}
		} else if optional.HasFlags() {
			buf.WriteString(fmt.Sprintf("## %sFlags\n\n", set.name))
			genMarkdownFlagsetUsage(optional, buf)
		}
	}
}

func genMarkdownFlagsetUsage(flags *pflag.FlagSet, buf *bytes.Buffer) {
	flags.VisitAll(func(flag *pflag.Flag) {
		if flag.Hidden {
			return
		}

		longDisplay := flagString(flag)
		if flag.Shorthand != "" && flag.ShorthandDeprecated == "" {
			fmt.Fprintf(buf, "`-%s, %s`\n\n", flag.Shorthand, longDisplay)
		} else {
			fmt.Fprintf(buf, "`%s`\n\n", longDisplay)
		}

		// Add the usage
		fmt.Fprintf(buf, "%s\n\n", flag.Usage)
	})
}

// navItem is a single item in the navigation JSON.
type navItem struct {
	Title  string     `json:"title"`
	Path   string     `json:"path,omitempty"`
	Routes []*navItem `json:"routes,omitempty"`
}

// GenNavJSON generates the navigation JSON for the command structure.
func GenNavJSON(c *Command, w io.Writer) error {

	root := &navItem{}
	genNavJSON(c, root, "cli/commands")

	// Create the top level nav item
	nav := &navItem{
		Title:  "Command Reference",
		Routes: root.Routes[0].Routes,
	}

	// Serialize the JSON
	e := json.NewEncoder(w)
	e.SetIndent("", "  ")
	if err := e.Encode(nav); err != nil {
		return err
	}

	return nil
}

// genNavJSON is a recursive function that generates the navigation JSON for
// the command structure.
func genNavJSON(c *Command, nav *navItem, path string) {
	// Generate a new nav item for this command
	var self *navItem

	if c.parent != nil {
		path = filepath.Join(path, c.Name)
	}

	// Handle being a command group
	if len(c.children) > 0 {
		self = &navItem{
			Title: c.Name,
			Routes: []*navItem{
				{
					Title: "Overview",
					Path:  path,
				},
			},
		}
	} else {
		self = &navItem{
			Title: c.Name,
			Path:  path,
		}
	}

	// Sort the children by name
	slices.SortFunc(c.children, func(i, j *Command) int {
		return strings.Compare(i.Name, j.Name)
	})

	// If we have children, create a new nav item for each child
	for _, child := range c.children {
		genNavJSON(child, self, path)
	}

	nav.Routes = append(nav.Routes, self)
}
