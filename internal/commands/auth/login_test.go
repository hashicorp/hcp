// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package auth

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/hcp/internal/commands/auth/mocks"
	mock_iam_service "github.com/hashicorp/hcp/internal/pkg/api/mocks/github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/client/iam_service"
	hcpAuth "github.com/hashicorp/hcp/internal/pkg/auth"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/hashicorp/hcp-sdk-go/auth"
	"github.com/hashicorp/hcp-sdk-go/auth/workload"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/client/iam_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/models"
	hcpconf "github.com/hashicorp/hcp-sdk-go/config"
)

// getIAMClientFunc returns a getIAMClientFunc and a mock IAM client and returns
// the mock IAM client for use in tests.
func getIAMClientFunc(t *testing.T) (GetIAMClientFunc, *mock_iam_service.MockClientService) {
	client := mock_iam_service.NewMockClientService(t)
	return func(_ hcpconf.HCPConfig) (iam_service.ClientService, error) {
		return client, nil
	}, client
}

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

	getter, iamClient := getIAMClientFunc(t)
	io := iostreams.Test()
	o := &LoginOpts{
		IO:            io,
		Profile:       profile.TestProfile(t),
		GetIAM:        getter,
		ConfigFn:      newHCP,
		CredentialDir: t.TempDir(),
	}

	// Expect the mock to be called
	m.EXPECT().Token().Return(nil, nil).Twice()

	// Expect the IAM client to be called
	iamClient.EXPECT().IamServiceGetCallerIdentity(mock.Anything, mock.Anything).
		Return(&iam_service.IamServiceGetCallerIdentityOK{
			Payload: &models.HashicorpCloudIamGetCallerIdentityResponse{
				Principal: &models.HashicorpCloudIamPrincipal{
					ID:   "123",
					Type: models.HashicorpCloudIamPrincipalTypePRINCIPALTYPEUSER.Pointer(),
					User: &models.HashicorpCloudIamUserPrincipal{
						Email:    "foo@bar.com",
						FullName: "Foo",
						ID:       "123",
					},
				},
			},
		}, nil)

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
	t.Run("profile has org/project", func(t *testing.T) {
		t.Parallel()
		testLoginSP(t, true)
	})

	t.Run("profile does not have org/project", func(t *testing.T) {
		t.Parallel()
		testLoginSP(t, false)
	})

}

func testLoginSP(t *testing.T, profilePreconfigured bool) {
	r := require.New(t)

	m := mocks.NewMockHCPConfig(t)
	newHCP := func(opts ...hcpconf.HCPConfigOption) (hcpconf.HCPConfig, error) {
		return m, nil
	}

	p := profile.TestProfile(t)
	if profilePreconfigured {
		p.OrganizationID = "preconfigured-org"
		p.ProjectID = "preconfigured-project"
	}

	getter, iamClient := getIAMClientFunc(t)
	io := iostreams.Test()
	o := &LoginOpts{
		IO:            io,
		Profile:       p,
		GetIAM:        getter,
		ConfigFn:      newHCP,
		CredentialDir: t.TempDir(),

		ClientID:     "123",
		ClientSecret: "456",
	}

	// Expect the mock to be called
	m.EXPECT().Token().Return(nil, nil)

	// Expect the IAM client to be called
	orgID, projectID := "org-123", "project-456"
	if !profilePreconfigured {
		iamClient.EXPECT().IamServiceGetCallerIdentity(mock.Anything, mock.Anything).
			Return(&iam_service.IamServiceGetCallerIdentityOK{
				Payload: &models.HashicorpCloudIamGetCallerIdentityResponse{
					Principal: &models.HashicorpCloudIamPrincipal{
						ID:   "123",
						Type: models.HashicorpCloudIamPrincipalTypePRINCIPALTYPESERVICE.Pointer(),
						Service: &models.HashicorpCloudIamServicePrincipal{
							ID:             "123",
							OrganizationID: orgID,
							ProjectID:      projectID,
						},
					},
				},
			}, nil)
	}

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

	// Check the profile
	if profilePreconfigured {
		r.NotEqual(orgID, o.Profile.OrganizationID)
		r.NotEqual(projectID, o.Profile.ProjectID)
	} else {
		r.Equal(orgID, o.Profile.OrganizationID)
		r.Equal(projectID, o.Profile.ProjectID)
	}
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

	getter, iamClient := getIAMClientFunc(t)
	io := iostreams.Test()
	o := &LoginOpts{
		IO:             io,
		Profile:        profile.TestProfile(t),
		GetIAM:         getter,
		ConfigFn:       newHCP,
		CredentialDir:  t.TempDir(),
		CredentialFile: f.Name(),
	}

	// Expect the mock to be called
	m.EXPECT().Token().Return(nil, nil)

	// Expect the IAM client to be called
	orgID, projectID := "org-123", "project-456"
	iamClient.EXPECT().IamServiceGetCallerIdentity(mock.Anything, mock.Anything).
		Return(&iam_service.IamServiceGetCallerIdentityOK{
			Payload: &models.HashicorpCloudIamGetCallerIdentityResponse{
				Principal: &models.HashicorpCloudIamPrincipal{
					ID:   "123",
					Type: models.HashicorpCloudIamPrincipalTypePRINCIPALTYPESERVICE.Pointer(),
					Service: &models.HashicorpCloudIamServicePrincipal{
						ID:             "123",
						OrganizationID: orgID,
						ProjectID:      projectID,
					},
				},
			},
		}, nil)

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

	// Check the profile
	r.Equal(orgID, o.Profile.OrganizationID)
	r.Equal(projectID, o.Profile.ProjectID)

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
