package profile

import (
	"bytes"
	"fmt"
	"slices"

	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/muesli/reflow/indent"
	"golang.org/x/exp/maps"
)

// availableProperties returns a document section describing all the available
// properties to be set on the profile.
func availablePropertiesDoc(io iostreams.IOStreams) cmd.DocSection {
	return cmd.DocSection{
		Title:         "Available Properties",
		Documentation: availableProperties(io).build(),
	}
}

func availableProperties(io iostreams.IOStreams) *availablePropertiesBuilder {
	b := newAvailablePropertiesBuilder(io)
	addCoreProperties(b)
	return b
}

func addCoreProperties(b *availablePropertiesBuilder) {
	b.AddProperty("", "organization_id", "Organization ID of the HCP organization to operate on.")
	b.AddProperty("", "project_id", `Project ID of the HCP project to operate on by default. This can be overridden by using the global {{ Bold "--project" }} flag.`)

	b.AddProperty("core", "no_color", "If True, color will not be used when printing messages in the terminal.")
	b.AddProperty("core", "verbosity", `
		Default logging verbosity for {{ Bold "hcp" }} commands. This is the
		equivalent of using the global --verbosity flag. Supported log levels:
		{{ Bold "trace" }}, {{ Bold "debug" }}, {{ Bold "info" }}, {{ Bold "warn" }}, and
		{{ Bold "error" }}. `)
	b.AddProperty("core", "output_format", `
		Default output format for {{ Bold "hcp" }} commands. This is the
		equivalent of using the global --format flag. Supported output foramts:
		{{ Bold "pretty" }}, {{ Bold "table" }}, and {{ Bold "json" }}.`)
}

type availablePropertiesBuilder struct {
	io         iostreams.IOStreams
	properties map[string]map[string]string
}

func newAvailablePropertiesBuilder(io iostreams.IOStreams) *availablePropertiesBuilder {
	return &availablePropertiesBuilder{
		io:         io,
		properties: make(map[string]map[string]string),
	}
}

func (b availablePropertiesBuilder) AddProperty(component, property, description string, args ...any) {
	c, ok := b.properties[component]
	if !ok {
		b.properties[component] = make(map[string]string)
		c = b.properties[component]
	}

	c[property] = heredoc.New(b.io).Mustf(description, args...)
}

func (b availablePropertiesBuilder) build() string {
	var buf bytes.Buffer
	cs := b.io.ColorScheme()

	// Start with the core section first
	topLevel, ok := b.properties[""]
	if ok {
		keys := maps.Keys(topLevel)
		slices.Sort(keys)
		for _, k := range keys {
			fmt.Fprintln(&buf, cs.String(k).Bold())
			fmt.Fprintln(&buf, indent.String(topLevel[k], 2))
			fmt.Fprintln(&buf)
		}
	}

	allComponents := maps.Keys(b.properties)
	slices.Sort(allComponents)
	for _, c := range allComponents {
		if c == "" {
			continue
		}

		// Print the component
		fmt.Fprintln(&buf, cs.String(c).Underline().String())

		keys := maps.Keys(b.properties[c])
		slices.Sort(keys)
		for _, k := range keys {
			fmt.Fprintln(&buf, indent.String(cs.String(k).Bold().String(), 2))
			fmt.Fprintln(&buf, indent.String(b.properties[c][k], 4))
			fmt.Fprintln(&buf)
		}
	}

	return buf.String()
}
