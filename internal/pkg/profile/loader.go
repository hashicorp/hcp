package profile

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/mitchellh/go-homedir"
)

const (
	// ConfigDir is the directory that contains HCP CLI configuration.
	ConfigDir = "~/.config/hcp/"

	// ProfileDir is the directory that contains HCP CLI profiles.
	ProfileDir = "profiles/"
)

var (
	// ErrNoActiveProfileFilePresent is returned if no active profile file
	// exists.
	ErrNoActiveProfileFilePresent = errors.New("active profile file doesn't exist")

	// ErrActiveProfileFileEmpty is returned if the active profile file is
	// empty.
	ErrActiveProfileFileEmpty = errors.New("active profile is unset")
)

// Loader is used to load and interact with profiles on disk.
type Loader struct {
	// configDir is the configuration directory.
	configDir string

	// profilesDir is the directory containing profiles.
	profilesDir string
}

// NewLoader returns a new loader or an error if the loader can't be
// instantiated.
func NewLoader() (*Loader, error) {
	return newLoader(ConfigDir)
}

// newLoader returns a new loader for the given config directory.
func newLoader(dir string) (*Loader, error) {
	path, err := homedir.Expand(dir)
	if err != nil {
		return nil, fmt.Errorf("error expanding HCP config directory path %q: %w", dir, err)
	}

	// Ensure the config directory exists.
	_, err = os.Stat(path)
	if err != nil {
		// If the directory doesn't exist, create it.
		if errors.Is(err, fs.ErrNotExist) {
			if err := os.MkdirAll(path, 0766); err != nil {
				return nil, fmt.Errorf("failed to created HCP config directory %q: %w", path, err)
			}
		} else {
			return nil, fmt.Errorf("failed to check if HCP config directory exists: %w", err)
		}
	}

	// Ensure the profiles directory exists.
	profilesDir := filepath.Join(path, ProfileDir)
	_, err = os.Stat(profilesDir)
	if err != nil {
		// If the directory doesn't exist, create it.
		if errors.Is(err, fs.ErrNotExist) {
			if err := os.MkdirAll(profilesDir, 0766); err != nil {
				return nil, fmt.Errorf("failed to created HCP profiles directory %q: %w", profilesDir, err)
			}
		} else {
			return nil, fmt.Errorf("failed to check if HCP profiles directory exists: %w", err)
		}
	}

	return &Loader{
		configDir:   path,
		profilesDir: profilesDir,
	}, nil
}

// GetActiveProfile returns the current profile
func (l *Loader) GetActiveProfile() (*ActiveProfile, error) {
	// Expand the active profile path.
	path := filepath.Join(l.configDir, ActiveProfileFileName)

	// Check if the file exists.
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNoActiveProfileFilePresent
		}

		return nil, err
	}

	// Decode the file
	var c ActiveProfile
	if err := hclsimple.DecodeFile(path, nil, &c); err != nil {
		return nil, err
	}

	// Check if no profile has been set.
	if c.Name == "" {
		return nil, ErrActiveProfileFileEmpty
	}

	c.dir = l.configDir
	return &c, nil
}

// DefaultActiveProfile returns an active profile set to default.
func (l *Loader) DefaultActiveProfile() *ActiveProfile {
	return &ActiveProfile{
		Name: "default",
		dir:  l.configDir,
	}
}

// ListProfiles returns the available profile names
func (l *Loader) ListProfiles() ([]string, error) {
	files, err := os.ReadDir(l.profilesDir)
	if err != nil {
		return nil, fmt.Errorf("unable to list profiles: %w", err)
	}

	profiles := make([]string, 0, len(files))
	for _, file := range files {
		n := file.Name()
		if file.IsDir() {
			return nil, fmt.Errorf("unexpected directory %q in profile %q directory. Please delete to recover", n, l.configDir)
		}

		if !strings.HasSuffix(n, ".hcl") {
			return nil, fmt.Errorf("unexpected non-hcl file %q in profile %q directory. Please delete to recover", n, l.configDir)
		}

		profiles = append(profiles, strings.TrimSuffix(n, ".hcl"))
	}

	return profiles, nil
}

// LoadProfile loads a profile given its name. If the profile can not be found,
// ErrNoProfileFilePresent will be returned. Otherwise, an error will be
// returned if the profile is invalid.
func (l *Loader) LoadProfile(name string) (*Profile, error) {
	// Expand the directory.
	path := filepath.Join(l.profilesDir, fmt.Sprintf("%s.hcl", name))

	// Check that the profile exists.
	_, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, ErrNoProfileFilePresent
		}

		return nil, err
	}

	// Decode the profile.
	var c Profile
	if err := hclsimple.DecodeFile(path, nil, &c); err != nil {
		return nil, fmt.Errorf("failed to decode profile: %w", err)
	}

	// Validate the name matches in the path and file.
	if name != c.Name {
		return nil, fmt.Errorf("profile path name does not match name in file. %q versus %q. Please rename file or name within the profile file to reconcile", name, c.Name)
	}

	c.dir = l.profilesDir
	return &c, nil
}

// LoadProfiles loads all the available profiles
func (l *Loader) LoadProfiles() ([]*Profile, error) {
	profileNames, err := l.ListProfiles()
	if err != nil {
		return nil, err
	}

	var profiles []*Profile
	for _, n := range profileNames {
		p, err := l.LoadProfile(n)
		if err != nil {
			return nil, fmt.Errorf("failed to load profile %q: %w", n, err)
		}
		profiles = append(profiles, p)
	}

	return profiles, nil
}

// DeleteProfile deletes the profile with the given name. If the profile can not be found,
// ErrNoProfileFilePresent will be returned. Otherwise, an error will be
// returned if the profile can not be deleted for any other reason..
func (l *Loader) DeleteProfile(name string) error {
	// Expand the directory.
	path := filepath.Join(l.profilesDir, fmt.Sprintf("%s.hcl", name))

	// Try to delete the file
	err := os.Remove(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return ErrNoProfileFilePresent
		}

		return err
	}

	return nil
}

// DefaultProfile returns the minimal default profile.
func (l *Loader) DefaultProfile() *Profile {
	return &Profile{
		Name: "default",
		dir:  l.profilesDir,
	}
}

// NewProfile returns an empty profile with the given name.
func (l *Loader) NewProfile(name string) (*Profile, error) {
	p := &Profile{
		Name: name,
		dir:  l.profilesDir,
	}

	return p, p.Validate()
}
