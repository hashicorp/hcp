package format

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInferFields(t *testing.T) {
	t.Parallel()

	r := require.New(t)

	s1 := struct {
		Name string
	}{
		Name: "s1",
	}

	r.Equal([]Field{
		{Name: "Name", ValueFormat: "{{ .Name }}"},
	}, inferFields(s1, nil))

	s2 := struct {
		Name string `json:"owner,omitempty"`
	}{
		Name: "s2",
	}

	r.Equal([]Field{
		{Name: "Owner", ValueFormat: "{{ .Name }}"},
	}, inferFields(s2, nil))

	s3 := struct {
		Name string `json:"owner,omitempty"`
		Age  int
	}{
		Name: "s3",
	}

	r.Equal([]Field{
		{Name: "Owner", ValueFormat: "{{ .Name }}"},
	}, inferFields(s3, []string{"owner"}))

	// Shows that the json tag wins as the title even if it's specified as
	// by struct field name
	r.Equal([]Field{
		{Name: "Owner", ValueFormat: "{{ .Name }}"},
	}, inferFields(s3, []string{"Name"}))

	s4 := []struct {
		Name string `json:"owner,omitempty"`
		Age  int
	}{
		{
			Name: "s3",
		},
	}

	r.Equal([]Field{
		{Name: "Owner", ValueFormat: "{{ .Name }}"},
	}, inferFields(s4, []string{"owner"}))

	r.Equal([]Field{
		{Name: "Value", ValueFormat: "{{ . }}"},
	}, inferFields(1, nil))

	s5 := struct {
		Name string
		max  int
	}{
		Name: "s2",
		max:  10,
	}

	r.Equal([]Field{
		{Name: "Name", ValueFormat: "{{ .Name }}"},
	}, inferFields(s5, nil))

}
