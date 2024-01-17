package auth

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/hcp/internal/commands/auth/mocks"
	hcpAuth "github.com/hashicorp/hcp/internal/pkg/auth"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
	"github.com/stretchr/testify/require"

	"github.com/hashicorp/hcp-sdk-go/auth"
	"github.com/hashicorp/hcp-sdk-go/auth/workload"
	hcpconf "github.com/hashicorp/hcp-sdk-go/config"
)

func TestLoginOpts_Validate(t *testing.T) {
	t.Parallel()
	r := require.New(t)

	// Empty should error
	o := &LoginOpts{}
	r.ErrorContains(o.Validate(), "programmer error")

	// Fill out system set fields
	o = &LoginOpts{
		IO:            iostreams.Test(),
		Profile:       profile.TestProfile(t),
		ConfigFn:      hcpconf.NewHCPConfig,
		CredentialDir: hcpAuth.CredentialsDir,
	}
	r.NoError(o.Validate())

	// Set both the creds file and client ID
	o.CredentialFile = "test"
	o.ClientID = "bad"
	r.ErrorContains(o.Validate(), "both credential file and client id/secret may not be set")

	o.CredentialFile = ""
	r.ErrorContains(o.Validate(), "both client id and client secret must be set")

	o.ClientID = ""
	o.ClientSecret = "bad"
	r.ErrorContains(o.Validate(), "both client id and client secret must be set")
}

func TestLogin_Browser(t *testing.T) {
	t.Parallel()
	r := require.New(t)

	m := mocks.NewMockHCPConfig(t)
	newHCP := func(opts ...hcpconf.HCPConfigOption) (hcpconf.HCPConfig, error) {
		return m, nil
	}

	io := iostreams.Test()
	o := &LoginOpts{
		IO:            io,
		Profile:       profile.TestProfile(t),
		ConfigFn:      newHCP,
		CredentialDir: t.TempDir(),
	}

	// Expect the mock to be called
	m.EXPECT().Token().Return(nil, nil).Twice()

	// Run the command
	r.NoError(loginRun(o))
	r.Contains(io.Error.String(), "Successfully logged in!")

	// Run the command with quiet
	io.Error.Reset()
	o.Quiet = true
	r.NoError(loginRun(o))
	r.Empty(io.Error.Bytes())
}

func TestLogin_SP(t *testing.T) {
	t.Parallel()
	r := require.New(t)

	m := mocks.NewMockHCPConfig(t)
	newHCP := func(opts ...hcpconf.HCPConfigOption) (hcpconf.HCPConfig, error) {
		return m, nil
	}

	io := iostreams.Test()
	o := &LoginOpts{
		IO:            io,
		Profile:       profile.TestProfile(t),
		ConfigFn:      newHCP,
		CredentialDir: t.TempDir(),

		ClientID:     "123",
		ClientSecret: "456",
	}

	// Expect the mock to be called
	m.EXPECT().Token().Return(nil, nil)

	// Run the command
	r.NoError(loginRun(o))
	r.Contains(io.Error.String(), "Successfully logged in!")

	// Check that the credential file is written
	f, err := os.Open(filepath.Join(o.CredentialDir, hcpAuth.CredFileName))
	r.NoError(err)

	var credFile auth.CredentialFile
	r.NoError(json.NewDecoder(f).Decode(&credFile))
	r.Equal(auth.CredentialFileSchemeServicePrincipal, credFile.Scheme)
	r.Equal(o.ClientID, credFile.Oauth.ClientID)
	r.Equal(o.ClientSecret, credFile.Oauth.ClientSecret)
}

func TestLogin_CredFile(t *testing.T) {
	t.Parallel()
	r := require.New(t)

	m := mocks.NewMockHCPConfig(t)
	newHCP := func(opts ...hcpconf.HCPConfigOption) (hcpconf.HCPConfig, error) {
		return m, nil
	}

	// Write a credential file
	f, err := os.CreateTemp("", "example")
	r.NoError(err)
	defer func() {
		r.NoError(f.Close())
		r.NoError(os.Remove(f.Name()))
	}()

	credFile := auth.CredentialFile{
		ProjectID: "123",
		Scheme:    auth.CredentialFileSchemeWorkload,
		Workload: &workload.IdentityProviderConfig{
			ProviderResourceName: "iam/project/123/service-principal/foo/workload-identity-federation/bar",
			File: &workload.FileCredentialSource{
				Path: "hello",
			},
		},
	}
	r.NoError(json.NewEncoder(f).Encode(&credFile))

	io := iostreams.Test()
	o := &LoginOpts{
		IO:             io,
		Profile:        profile.TestProfile(t),
		ConfigFn:       newHCP,
		CredentialDir:  t.TempDir(),
		CredentialFile: f.Name(),
	}

	// Expect the mock to be called
	m.EXPECT().Token().Return(nil, nil)

	// Run the command
	r.NoError(loginRun(o))
	r.Contains(io.Error.String(), "Successfully logged in!")

	// Check that the credential file is written
	copiedCF, err := os.Open(filepath.Join(o.CredentialDir, hcpAuth.CredFileName))
	r.NoError(err)
	defer func() {
		r.NoError(copiedCF.Close())
	}()

	var copied auth.CredentialFile
	r.NoError(json.NewDecoder(copiedCF).Decode(&copied))
	r.EqualValues(credFile, copied)
}

func Test_getHCPConfig(t *testing.T) {
	t.Parallel()
	r := require.New(t)

	// Write a credential file
	tempDir := t.TempDir()
	f, err := os.Create(filepath.Join(tempDir, hcpAuth.CredFileName))
	r.NoError(err)
	defer func() {
		r.NoError(f.Close())
		r.NoError(os.Remove(f.Name()))
	}()

	credFile := auth.CredentialFile{
		ProjectID: "123",
		Scheme:    auth.CredentialFileSchemeWorkload,
		Workload: &workload.IdentityProviderConfig{
			ProviderResourceName: "iam/project/123/service-principal/foo/workload-identity-federation/bar",
			File: &workload.FileCredentialSource{
				Path: "hello",
			},
		},
	}
	r.NoError(json.NewEncoder(f).Encode(&credFile))

	// Get an HCP config using the temp directory containing the cred file
	conf, err := hcpAuth.GetHCPConfigFromDir(tempDir)
	r.NoError(err)
	_, err = conf.Token()
	r.ErrorContains(err, "failed to get new token: failed to open credential file \"hello\"")
}
