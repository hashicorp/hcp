package profile

import (
	"testing"

	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
	"github.com/stretchr/testify/require"
)

func TestDisplay(t *testing.T) {
	t.Parallel()

	io := iostreams.Test()
	p := profile.TestProfile(t)
	p.OrganizationID = "123"
	p.ProjectID = "456"
	p.Core = &profile.Core{
		NoColor: new(bool),
	}

	t.Run("default", func(t *testing.T) {
		t.Parallel()
		r := require.New(t)

		opts := &DisplayOpts{
			IO:      io,
			Profile: p,
		}
		r.NoError(displayRun(opts))
		r.Contains(io.Output.String(), "project_id")
		r.Contains(io.Output.String(), "core {")
	})

	t.Run("json", func(t *testing.T) {
		t.Parallel()
		r := require.New(t)

		opts := &DisplayOpts{
			IO:      io,
			Profile: p,
			Format:  format.JSON,
		}
		r.NoError(displayRun(opts))
		r.Contains(io.Output.String(), "ProjectID")
	})
}
