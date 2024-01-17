package profile

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoader_New(t *testing.T) {
	t.Parallel()
	r := require.New(t)

	// Test that we create the directory if it doesn't yet exist.
	dir := filepath.Join(t.TempDir(), "hcp")
	l, err := newLoader(dir)
	r.NoError(err)
	r.NotNil(l)

	// Check the directory and the profiles sub-dir was created.
	r.DirExists(dir)
	r.DirExists(filepath.Join(dir, ProfileDir))
}

func TestLoader_GetActiveProfile(t *testing.T) {
	t.Parallel()

	t.Run("no active profile", func(t *testing.T) {
		t.Parallel()
		r := require.New(t)
		l, err := newLoader(t.TempDir())
		r.NoError(err)
		active, err := l.GetActiveProfile()
		r.Nil(active)
		r.ErrorIs(err, ErrNoActiveProfileFilePresent)
	})

	t.Run("empty active profile", func(t *testing.T) {
		t.Parallel()
		r := require.New(t)
		l := TestLoader(t)
		active := l.DefaultActiveProfile()
		active.Name = ""
		r.NoError(active.Write())

		p, err := l.GetActiveProfile()
		r.Nil(p)
		r.ErrorIs(err, ErrActiveProfileFileEmpty)
	})

	t.Run("malformed active profile", func(t *testing.T) {
		t.Parallel()
		r := require.New(t)
		l := TestLoader(t)

		// Write a bad active profile
		r.NoError(os.WriteFile(filepath.Join(l.configDir, ActiveProfileFileName), []byte("invalid!"), 0x777))

		// Read the malformed profile
		p, err := l.GetActiveProfile()
		r.Nil(p)
		r.Error(err)
	})

	t.Run("valid active profile", func(t *testing.T) {
		t.Parallel()
		r := require.New(t)
		l := TestLoader(t)

		active := l.DefaultActiveProfile()
		active.Name = t.Name()
		r.NoError(active.Write())

		p, err := l.GetActiveProfile()
		r.NoError(err)
		r.Equal(t.Name(), p.Name)
	})
}

func TestLoader_ListProfiles(t *testing.T) {
	t.Parallel()

	validProfileNames := []string{"bar", "baz", "foo"}
	slices.Sort(validProfileNames)
	t.Run("empty profiles directory", func(t *testing.T) {
		t.Parallel()
		r := require.New(t)
		l := TestLoader(t)
		profiles, err := l.ListProfiles()
		r.Empty(profiles)
		r.NoError(err)
	})

	t.Run("valid profiles", func(t *testing.T) {
		t.Parallel()
		r := require.New(t)
		l := TestLoader(t)

		// Create some profiles
		for _, n := range validProfileNames {
			p, err := l.NewProfile(n)
			r.NoError(err)
			r.NoError(p.Write())
		}

		profiles, err := l.ListProfiles()
		slices.Sort(profiles)
		r.Equal(profiles, validProfileNames)
		r.NoError(err)
	})

	t.Run("one invalid profile", func(t *testing.T) {
		t.Parallel()
		r := require.New(t)
		l := TestLoader(t)

		// Create some profiles
		for _, n := range validProfileNames {
			p, err := l.NewProfile(n)
			r.NoError(err)
			r.NoError(p.Write())
		}

		// Write an invalid file
		r.NoError(os.WriteFile(filepath.Join(l.configDir, ProfileDir, "not_a_profile.json"), []byte("invalid!"), 0x777))

		profiles, err := l.ListProfiles()
		r.Empty(profiles)
		r.ErrorContains(err, "unexpected non-hcl file")
	})
}

func TestLoader_LoadProfile(t *testing.T) {
	t.Parallel()

	t.Run("no profile", func(t *testing.T) {
		t.Parallel()
		r := require.New(t)
		l := TestLoader(t)

		p, err := l.LoadProfile("test")
		r.Nil(p)
		r.ErrorIs(err, ErrNoProfileFilePresent)
	})

	t.Run("invalid profile", func(t *testing.T) {
		t.Parallel()
		r := require.New(t)
		l := TestLoader(t)

		// Write an invalid profile to disk
		name := "test"
		path := filepath.Join(l.configDir, ProfileDir, fmt.Sprintf("%s.hcl", name))
		r.NoError(os.WriteFile(path, []byte("invalid!"), 0x777))

		p, err := l.LoadProfile(name)
		r.Nil(p)
		r.ErrorContains(err, "failed to decode profile")
	})

	t.Run("mismatched profile name", func(t *testing.T) {
		t.Parallel()
		r := require.New(t)
		l := TestLoader(t)

		// Write an invalid profile to disk
		name := "test"
		path := filepath.Join(l.configDir, ProfileDir, fmt.Sprintf("%s.hcl", name))
		r.NoError(os.WriteFile(path, []byte(`name = "other"
organization_id = "123"
project_id = "456"`,
		), 0x777))

		p, err := l.LoadProfile(name)
		r.Nil(p)
		r.ErrorContains(err, "profile path name does not match name in file")
	})

	t.Run("valid profile", func(t *testing.T) {
		t.Parallel()
		r := require.New(t)
		l := TestLoader(t)

		p, err := l.NewProfile("test")
		r.NoError(err)
		p.OrganizationID = "123"
		p.ProjectID = "456"
		r.NoError(p.Write())

		out, err := l.LoadProfile(p.Name)
		r.NotNil(out)
		r.Equal(p.Name, out.Name)
		r.Equal(p.OrganizationID, out.OrganizationID)
		r.Equal(p.ProjectID, out.ProjectID)
		r.NoError(err)
	})

	t.Run("invalid profile name", func(t *testing.T) {
		t.Parallel()
		r := require.New(t)
		l := TestLoader(t)

		_, err := l.NewProfile("test!@#$")
		r.ErrorContains(err, "profile name may only include")
	})
}

func TestLoader_LoadProfiles(t *testing.T) {
	t.Parallel()

	t.Run("no profile", func(t *testing.T) {
		t.Parallel()
		r := require.New(t)
		l := TestLoader(t)

		profiles, err := l.LoadProfiles()
		r.Nil(profiles)
		r.NoError(err)
	})

	t.Run("valid profile", func(t *testing.T) {
		t.Parallel()
		r := require.New(t)
		l := TestLoader(t)

		p, err := l.NewProfile("test")
		r.NoError(err)
		p.OrganizationID = "123"
		p.ProjectID = "456"
		r.NoError(p.Write())

		out, err := l.LoadProfiles()
		r.NoError(err)
		r.Len(out, 1)
		r.Equal(p.Name, out[0].Name)
		r.Equal(p.OrganizationID, out[0].OrganizationID)
		r.Equal(p.ProjectID, out[0].ProjectID)
		r.NoError(err)
	})

	t.Run("valid profiles", func(t *testing.T) {
		t.Parallel()
		r := require.New(t)
		l := TestLoader(t)

		p, err := l.NewProfile("test")
		r.NoError(err)
		p.OrganizationID = "123"
		p.ProjectID = "456"
		r.NoError(p.Write())

		p2, err := l.NewProfile("test2")
		r.NoError(err)
		p2.OrganizationID = "456"
		p2.ProjectID = "789"
		r.NoError(p2.Write())

		out, err := l.LoadProfiles()
		r.NoError(err)
		r.NotNil(out)
		r.Equal(p.Name, out[0].Name)
		r.Equal(p.OrganizationID, out[0].OrganizationID)
		r.Equal(p.ProjectID, out[0].ProjectID)
		r.Equal(p2.Name, out[1].Name)
		r.Equal(p2.OrganizationID, out[1].OrganizationID)
		r.Equal(p2.ProjectID, out[1].ProjectID)
	})
}

func TestLoader_DeleteProfile(t *testing.T) {
	t.Parallel()

	t.Run("no profile", func(t *testing.T) {
		t.Parallel()
		r := require.New(t)
		l := TestLoader(t)

		p, err := l.NewProfile("test")
		r.NoError(err)
		p.OrganizationID = "123"
		p.ProjectID = "456"
		r.NoError(p.Write())

		r.NoError(l.DeleteProfile("test"))
	})

	t.Run("existing profile", func(t *testing.T) {
		t.Parallel()
		r := require.New(t)
		l := TestLoader(t)

		// Write an invalid profile to disk
		name := "test"
		path := filepath.Join(l.configDir, ProfileDir, fmt.Sprintf("%s.hcl", name))
		r.NoError(os.WriteFile(path, []byte("invalid!"), 0x777))

		p, err := l.LoadProfile(name)
		r.Nil(p)
		r.ErrorContains(err, "failed to decode profile")
	})
}
