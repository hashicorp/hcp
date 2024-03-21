package cmd

import (
	"encoding/json"
	"io"
	"path/filepath"
	"slices"
	"strings"
)

// navItem is a single item in the navigation JSON.
type navItem struct {
	Title  string     `json:"title"`
	Path   string     `json:"path,omitempty"`
	Routes []*navItem `json:"routes,omitempty"`
}

// GenNavJSON generates the navigation JSON in the format that hcp-docs expects,
// for the command structure.
func GenNavJSON(c *Command, w io.Writer) error {

	root := &navItem{}
	genNavJSON(c, root, "cli/commands")

	// Create the top level nav item
	nav := &navItem{
		Title:  "Commands (CLI)",
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
