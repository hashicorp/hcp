package internal

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Test_VariableOptionsFileParse tests the parsing of variable options from a
// file.
func Test_VariableOptionsFileParse(t *testing.T) {
	t.Parallel()

	t.Run("can parse variables", func(t *testing.T) {
		t.Parallel()

		r := require.New(t)

		hcl := `
          variable_option "string_variable" {
            options = [
              "a string value",
            ]
            user_editable = false
          }
`

		variableInputs, err := parseVariableOptions("blah.hcl", []byte(hcl))
		r.NoError(err)
		r.Equal(1, len(variableInputs))
		r.Equal("string_variable", variableInputs[0].Name)
		r.Equal(false, variableInputs[0].UserEditable)
	})

	t.Run("handles multiple options", func(t *testing.T) {
		t.Parallel()

		r := require.New(t)

		hcl := `
          variable_option "string_variable" {
            options = [
              "a string value",
        		"another string value",
            ]
            user_editable =true
          }
`

		variableInputs, err := parseVariableOptions("blah.hcl", []byte(hcl))
		r.NoError(err)
		r.Equal(1, len(variableInputs))
		r.Equal("string_variable", variableInputs[0].Name)
		r.Equal(2, len(variableInputs[0].Options))
		r.Equal(true, variableInputs[0].UserEditable)
	})

	t.Run("handles multiple variables", func(t *testing.T) {
		t.Parallel()

		r := require.New(t)

		hcl := `
variable_option "string_variable" {
  options = [
    "a string value",
		"another",
  ]
  user_editable = false
}

variable_option "misc_variable" {
  options = [
    8,
		2,
  ]
  user_editable = false
}
`

		variableInputs, err := parseVariableOptions("blah.hcl", []byte(hcl))
		r.NoError(err)
		r.Equal(2, len(variableInputs))
		r.Equal("string_variable", variableInputs[0].Name)
		r.Equal(2, len(variableInputs[0].Options))
		r.Equal(false, variableInputs[0].UserEditable)

		r.Equal("misc_variable", variableInputs[1].Name)
		r.Equal(2, len(variableInputs[1].Options))
		r.Equal(false, variableInputs[1].UserEditable)
	})

	t.Run("errors on missing attributes", func(t *testing.T) {
		t.Parallel()

		r := require.New(t)

		hcl := `
variable_option "" {}
`

		_, err := parseVariableOptions("blah.hcl", []byte(hcl))
		r.Error(err)
	})

	t.Run("returns nil empty", func(t *testing.T) {
		t.Parallel()

		r := require.New(t)

		hcl := ``

		variableInputs, err := parseVariableOptions("blah.hcl", []byte(hcl))
		r.NoError(err)
		r.Equal(0, len(variableInputs))
	})
}
