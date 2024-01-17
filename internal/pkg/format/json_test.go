package format_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/stretchr/testify/require"
)

func TestJSON_KV_Slice(t *testing.T) {
	t.Parallel()
	r := require.New(t)
	io := iostreams.Test()
	out := format.New(io)

	// Create our displayer
	d := &KVDisplayer{
		KVs: []*KV{
			{
				Key:   "Hello",
				Value: "World!",
			},
			{
				Key:   "Another",
				Value: "Test",
			},
		},
		Default: format.JSON,
	}

	// Display the table
	r.NoError(out.Display(d))

	var parsed []*KV
	r.NoError(json.Unmarshal(io.Output.Bytes(), &parsed))
	r.Equal(d.KVs, parsed)
}

func TestJSON_KV_Struct(t *testing.T) {
	t.Parallel()
	r := require.New(t)
	io := iostreams.Test()
	out := format.New(io)

	// Create our displayer
	d := &KVDisplayer{
		KVs: []*KV{
			{
				Key:   "Hello",
				Value: "World!",
			},
		},
		Default: format.JSON,
	}

	// Display the table
	r.NoError(out.Display(d))

	var parsed *KV
	r.NoError(json.Unmarshal(io.Output.Bytes(), &parsed))
	r.Equal(d.KVs[0], parsed)
}

func TestJSON_Complex_Slice(t *testing.T) {
	t.Parallel()
	r := require.New(t)
	io := iostreams.Test()
	out := format.New(io)

	// Create our displayer
	d := &ComplexDisplayer{
		Data: []*Complex{
			{
				Name:        "Test",
				Description: "Test description",
				Version:     12,
				CreatedAt:   time.Now().Add(-5 * time.Second),
				UpdatedAt:   time.Now().Add(-1 * time.Second),
			},
			{
				Name:        "Other",
				Description: "Other description",
				Version:     15,
				CreatedAt:   time.Now().Add(-10 * time.Minute),
				UpdatedAt:   time.Now().Add(-3 * time.Second),
			},
		},
		Default: format.JSON,
	}

	// Display the table
	r.NoError(out.Display(d))

	var parsed []*Complex
	r.NoError(json.Unmarshal(io.Output.Bytes(), &parsed))

	// Check the timestamps are equal and then clear
	for i, d := range d.Data {
		r.True(d.CreatedAt.Equal(parsed[i].CreatedAt))
		r.True(d.UpdatedAt.Equal(parsed[i].UpdatedAt))
		d.CreatedAt = time.Time{}
		d.UpdatedAt = time.Time{}
		parsed[i].CreatedAt = time.Time{}
		parsed[i].UpdatedAt = time.Time{}
	}

	r.Equal(d.Data, parsed)
}

func TestJSON_Complex_Struct(t *testing.T) {
	t.Parallel()
	r := require.New(t)
	io := iostreams.Test()
	out := format.New(io)

	// Create our displayer
	d := &ComplexDisplayer{
		Data: []*Complex{
			{
				Name:        "Test",
				Description: "Test description",
				Version:     12,
				CreatedAt:   time.Now().Add(-5 * time.Second),
				UpdatedAt:   time.Now().Add(-1 * time.Second),
			},
		},
		Default: format.JSON,
	}

	// Display the table
	r.NoError(out.Display(d))

	var parsed *Complex
	r.NoError(json.Unmarshal(io.Output.Bytes(), &parsed))

	// Check the timestamps are equal and then clear
	r.True(d.Data[0].CreatedAt.Equal(parsed.CreatedAt))
	r.True(d.Data[0].UpdatedAt.Equal(parsed.UpdatedAt))
	d.Data[0].CreatedAt = time.Time{}
	d.Data[0].UpdatedAt = time.Time{}
	parsed.CreatedAt = time.Time{}
	parsed.UpdatedAt = time.Time{}

	r.Equal(d.Data[0], parsed)
}
