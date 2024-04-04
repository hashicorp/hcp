// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package format

import (
	"bytes"
	"fmt"
	"reflect"
	"text/template"

	"github.com/rodaine/table"
)

// TableFormatter is an optional interface to implement to customize how a table
// is outputted.
type TableFormatter interface {
	HeaderFormatter(input string) string
	FirstColumnFormatter(input string) string
}

// outputTable outputs the payload as a table.
func (o *Outputter) outputTable(d Displayer) error {
	// Gather the headers and the row template
	fields := d.FieldTemplates()
	headers := make([]interface{}, len(fields))
	for i, f := range fields {
		headers[i] = f.Name
	}

	// Instantiate the header and apply default formatting.
	tbl := table.New(headers...)
	tbl.WithWriter(o.io.Out())
	tbl.WithFirstColumnFormatter(o.defaultFirstColumnFormatter())
	tbl.WithHeaderFormatter(o.defaultHeaderFormatter())

	// If the displayer has implemented the table formatter, then use it.
	formatter, ok := d.(TableFormatter)
	if ok {
		tbl.WithFirstColumnFormatter(wrapFormatter(formatter.FirstColumnFormatter))
		tbl.WithHeaderFormatter(wrapFormatter(formatter.HeaderFormatter))
	}

	// Get the payload
	var p any
	if tp, ok := d.(TemplatedPayload); ok {
		p = tp.TemplatedPayload()
	} else {
		p = d.Payload()
	}

	// Build the rows
	var rows [][]string
	rv := reflect.ValueOf(p)

	// If the payload is a slice, render each row and add it to the table.
	if rv.Kind() == reflect.Slice {
		for i := 0; i < rv.Len(); i++ {
			vf := rv.Index(i)
			row, err := renderRow(vf.Interface(), fields)
			if err != nil {
				return err
			}

			rows = append(rows, row)
		}
	} else {
		// Render the payload as a row and add it to the table.
		row, err := renderRow(p, fields)
		if err != nil {
			return err
		}

		rows = append(rows, row)
	}

	// Output the table
	tbl.SetRows(rows)
	tbl.Print()
	return nil
}

// defaultHeaderFormatter is the default header formatter which prints the
// header in green.
func (o *Outputter) defaultHeaderFormatter() func(string, ...interface{}) string {
	return func(format string, vals ...interface{}) string {
		cs := o.io.ColorScheme()
		formatted := fmt.Sprintf(format, vals...)
		return cs.String(formatted).Color(cs.Green()).Underline().String()
	}
}

// defaultFirstColumnFormatter is the default first column formatter which
// prints the column in yellow.
func (o *Outputter) defaultFirstColumnFormatter() func(string, ...interface{}) string {
	return func(format string, vals ...interface{}) string {
		cs := o.io.ColorScheme()
		formatted := fmt.Sprintf(format, vals...)
		return cs.String(formatted).Color(cs.Yellow()).String()
	}
}

// wrapFormatter wraps a TableFormatter function to be the type expected by our
// table implementation.
func wrapFormatter(formatter func(string) string) func(string, ...interface{}) string {
	return func(format string, vals ...interface{}) string {
		formatted := fmt.Sprintf(format, vals...)
		return formatter(formatted)
	}
}

// renderRow renders each field by executing the text/template given the
// payload.
func renderRow(p any, fields []Field) ([]string, error) {
	renderedFields := make([]string, len(fields))
	for i, f := range fields {
		tmpl, err := template.New("hcp").Parse(f.ValueFormat)
		if err != nil {
			return nil, err
		}

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, p); err != nil {
			return nil, err
		}

		renderedFields[i] = buf.String()
	}

	return renderedFields, nil
}
