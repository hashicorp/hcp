// table outputs data in a table format. The package is heavily inspired by
// github.com/gosuri/uitable.
package table

import (
	"fmt"
	"strings"
	"sync"

	"github.com/muesli/ansi"
	"github.com/muesli/reflow/padding"
	"github.com/muesli/reflow/truncate"
	"github.com/muesli/reflow/wordwrap"
)

// Separator is the default column seperator
var Separator = "\t"

// Table represents a decorator that renders the data in formatted in a table
type Table struct {
	// Rows is the collection of rows in the table
	Rows []*Row

	// MaxColWidth is the maximum allowed width for cells in the table
	MaxColWidth uint

	// Wrap when set to true wraps the contents of the columns when the length exceeds the MaxColWidth
	Wrap bool

	// Separator is the seperator for columns in the table. Default is "\t"
	Separator string

	// HeaderFormatter is a function that formats the header of the table
	HeaderFormatter func(input string) string

	// FirstColumnFormatter is a function that formats the first column of the table
	FirstColumnFormatter func(input string) string

	mtx *sync.RWMutex
}

// New returns a new Table with default values
func New() *Table {
	return &Table{
		Separator: Separator,
		mtx:       new(sync.RWMutex),
	}
}

// AddRow adds a new row to the table
func (t *Table) AddRow(data ...interface{}) *Table {
	t.mtx.Lock()
	defer t.mtx.Unlock()
	r := NewRow(data...)
	t.Rows = append(t.Rows, r)
	return t
}

// Bytes returns the []byte value of table
func (t *Table) Bytes() []byte {
	return []byte(t.String())
}

// String returns the string value of table
func (t *Table) String() string {
	t.mtx.RLock()
	defer t.mtx.RUnlock()

	if len(t.Rows) == 0 {
		return ""
	}

	// determine the width for each column (cell in a row)
	var colwidths []uint
	for _, row := range t.Rows {
		for i, cell := range row.Cells {
			// resize colwidth array
			if i+1 > len(colwidths) {
				colwidths = append(colwidths, 0)
			}
			cellwidth := cell.LineWidth()
			if t.MaxColWidth != 0 && cellwidth > t.MaxColWidth {
				cellwidth = t.MaxColWidth
			}

			if cellwidth > colwidths[i] {
				colwidths[i] = cellwidth
			}
		}
	}

	var lines []string
	for i, row := range t.Rows {
		row.Separator = t.Separator
		if i == 0 {
			row.HeaderFormatter = t.HeaderFormatter
		} else {
			row.FirstColumnFormatter = t.FirstColumnFormatter
		}
		for i, cell := range row.Cells {
			cell.Width = colwidths[i]
			cell.Wrap = t.Wrap
		}
		lines = append(lines, row.String())
	}
	return strings.Join(lines, "\n")
}

// Row represents a row in a table
type Row struct {
	// Cells is the group of cell for the row
	Cells []*Cell

	// Separator for tabular columns
	Separator string

	// HeaderFormatter is a function that formats the header of the table
	HeaderFormatter func(input string) string

	// FirstColumnFormatter is a function that formats the first column of the table
	FirstColumnFormatter func(input string) string
}

// NewRow returns a new Row and adds the data to the row
func NewRow(data ...interface{}) *Row {
	r := &Row{Cells: make([]*Cell, len(data))}
	for i, d := range data {
		r.Cells[i] = &Cell{Data: d}
	}
	return r
}

// String returns the string representation of the row
func (r *Row) String() string {
	// get the max number of lines for each cell
	var lc int // line count
	for _, cell := range r.Cells {
		if clc := len(strings.Split(cell.String(), "\n")); clc > lc {
			lc = clc
		}
	}

	// allocate a two-dimentional array of cells for each line and add size them
	cells := make([][]*Cell, lc)
	for x := 0; x < lc; x++ {
		cells[x] = make([]*Cell, len(r.Cells))
		for y := 0; y < len(r.Cells); y++ {
			cells[x][y] = &Cell{Width: r.Cells[y].Width, Wrap: r.Cells[y].Wrap}
		}
	}

	// insert each line in a cell as new cell in the cells array
	for y, cell := range r.Cells {
		lines := strings.Split(cell.String(), "\n")
		for x, line := range lines {
			cells[x][y].Data = line
		}
	}

	// format each line
	lines := make([]string, lc)
	for x := range lines {
		line := make([]string, len(cells[x]))
		for y := range cells[x] {
			val := cells[x][y].String()
			if r.HeaderFormatter != nil {
				val = r.HeaderFormatter(val)
			}
			if y == 0 && r.FirstColumnFormatter != nil {
				val = r.FirstColumnFormatter(val)
			}

			line[y] = val
		}
		lines[x] = strings.Join(line, r.Separator)
	}
	return strings.Join(lines, "\n")
}

// Cell represents a column in a row
type Cell struct {
	// Width is the width of the cell
	Width uint

	// Wrap when true wraps the contents of the cell when the lenght exceeds the width
	Wrap bool

	// Data is the cell data
	Data interface{}
}

// LineWidth returns the max width of all the lines in a cell
func (c *Cell) LineWidth() uint {
	width := 0
	for _, s := range strings.Split(c.String(), "\n") {
		w := ansi.PrintableRuneWidth(s)
		if w > width {
			width = w
		}
	}
	return uint(width)
}

// String returns the string formated representation of the cell
func (c *Cell) String() string {
	if c.Data == nil {
		return padding.String(" ", c.Width)
	}
	s := fmt.Sprint(c.Data)
	if c.Width > 0 {
		if c.Wrap && uint(ansi.PrintableRuneWidth(s)) > c.Width {
			return wordwrap.String(s, int(c.Width)) // <- breaks things
		} else if !c.Wrap && uint(ansi.PrintableRuneWidth(s)) > c.Width {
			return truncate.StringWithTail(s, c.Width, "...")
		} else if len(s) != 0 {
			return padding.String(s, c.Width)
		} else {
			return strings.Repeat(" ", int(c.Width))
		}
	}
	return s
}
