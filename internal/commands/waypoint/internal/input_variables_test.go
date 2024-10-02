package internal

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_InputVariablesFile(t *testing.T) {
	t.Parallel()

	t.Run("can parse variables", func(t *testing.T) {
		t.Parallel()

		r := require.New(t)

		hcl := `
key="value"
key2="value2"
`

		inputVars, err := parseInputVariables("blah.hcl", []byte(hcl))
		r.NoError(err)
		r.Equal(2, len(inputVars))
		r.Equal("key", inputVars[0].Name)
		r.Equal("value", inputVars[0].Value)
		r.Equal("key2", inputVars[1].Name)
		r.Equal("value2", inputVars[1].Value)
	})

	t.Run("handles empty file", func(t *testing.T) {
		t.Parallel()

		r := require.New(t)

		hcl := ``

		inputVars, err := parseInputVariables("blah.hcl", []byte(hcl))
		r.NoError(err)
		r.Equal(0, len(inputVars))
	})

	t.Run("handles empty values", func(t *testing.T) {
		t.Parallel()

		r := require.New(t)

		hcl := `
key=""
key2=""
`

		inputVars, err := parseInputVariables("blah.hcl", []byte(hcl))
		r.NoError(err)
		r.Equal(2, len(inputVars))
		r.Equal("key", inputVars[0].Name)
		r.Equal("", inputVars[0].Value)
		r.Equal("key2", inputVars[1].Name)
		r.Equal("", inputVars[1].Value)
	})

	t.Run("fail to parse", func(t *testing.T) {
		t.Parallel()

		r := require.New(t)

		hcl := `
key======
`

		_, err := parseInputVariables("blah.hcl", []byte(hcl))
		r.Error(err)
	})
}
