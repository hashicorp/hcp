// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: MPL-2.0

package keys

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/hashicorp/hcp-sdk-go/auth"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/client/service_principals_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/models"
	mock_service_principals_service "github.com/hashicorp/hcp/internal/pkg/api/mocks/github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/client/service_principals_service"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewCmdCreate(t *testing.T) {
	t.Parallel()

	goodCredFilePath := "/path/good.json"
	cases := []struct {
		Name    string
		Args    []string
		Profile func(t *testing.T) *profile.Profile
		Error   string
		Expect  *CreateOpts
	}{
		{
			Name:    "No Org",
			Profile: profile.TestProfile,
			Args:    []string{},
			Error:   "Organization ID must be configured before running the command.",
		},
		{
			Name: "Too many args",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123")
			},
			Args:  []string{"foo", "bar"},
			Error: "accepts 1 arg(s), received 2",
		},
		{
			Name: "Bad output-cred-file",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123")
			},
			Args:  []string{"foo", "--output-cred-file", "non-json"},
			Error: "credential file must be a json file",
		},
		{
			Name: "Good",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123")
			},
			Args: []string{"foo"},
			Expect: &CreateOpts{
				Name: "foo",
			},
		},
		{
			Name: "Good output-cred-file",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123")
			},
			Args: []string{"foo", "--output-cred-file", goodCredFilePath},
			Expect: &CreateOpts{
				Name:               "foo",
				CredentialFilePath: &goodCredFilePath,
			},
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			r := require.New(t)

			// Create a context.
			io := iostreams.Test()
			ctx := &cmd.Context{
				IO:          io,
				Profile:     c.Profile(t),
				Output:      format.New(io),
				HCP:         &client.Runtime{},
				ShutdownCtx: context.Background(),
			}

			var gotOpts *CreateOpts
			createCmd := NewCmdCreate(ctx, func(o *CreateOpts) error {
				gotOpts = o
				return nil
			})
			createCmd.SetIO(io)

			code := createCmd.Run(c.Args)
			if c.Error != "" {
				r.NotZero(code)
				r.Contains(io.Error.String(), c.Error)
				return
			}

			r.Zero(code, io.Error.String())
			r.NotNil(gotOpts)
			r.Equal(c.Expect.Name, gotOpts.Name)
			r.EqualValues(c.Expect.CredentialFilePath, gotOpts.CredentialFilePath)
		})
	}
}

func TestCreateRun_NonCredFile(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Name               string
		Profile            func(t *testing.T) *profile.Profile
		RespErr            bool
		SPName             string
		ParentResourceName string
		Error              string
	}{
		{
			Name: "Server error",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123")
			},
			SPName:             "test-sp",
			ParentResourceName: "iam/organization/123/service-principal/test-sp",
			RespErr:            true,
			Error:              "failed to create service principal key: [POST /2019-12-10/{parent_resource_name}/keys][403]",
		},
		{
			Name: "Good org",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123")
			},
			SPName:             "test-sp",
			ParentResourceName: "iam/organization/123/service-principal/test-sp",
		},
		{
			Name: "Good project",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123").SetProjectID("456")
			},
			SPName:             "test-sp",
			ParentResourceName: "iam/project/456/service-principal/test-sp",
		},
		{
			Name: "Good full",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123").SetProjectID("456")
			},
			SPName:             "iam/project/789/service-principal/test-sp",
			ParentResourceName: "iam/project/789/service-principal/test-sp",
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			r := require.New(t)

			io := iostreams.Test()
			spService := mock_service_principals_service.NewMockClientService(t)
			opts := &CreateOpts{
				Ctx:     context.Background(),
				Profile: c.Profile(t),
				IO:      io,
				Output:  format.New(io),
				Client:  spService,
				Name:    c.SPName,
			}

			// Expect a request to get the user.
			call := spService.EXPECT().ServicePrincipalsServiceCreateServicePrincipalKey(mock.MatchedBy(func(req *service_principals_service.ServicePrincipalsServiceCreateServicePrincipalKeyParams) bool {
				return req.ParentResourceName == c.ParentResourceName
			}), nil).Once()

			secret := "superSecret12312232"
			id := "ASDKJLAD1412412"
			now := time.Now()
			rn := fmt.Sprintf("iam/%s/service-principal/%s/key/%s", c.ParentResourceName, c.SPName, id)
			if c.RespErr {
				call.Return(nil, service_principals_service.NewServicePrincipalsServiceCreateServicePrincipalKeyDefault(http.StatusForbidden))
			} else {
				ok := service_principals_service.NewServicePrincipalsServiceCreateServicePrincipalKeyOK()
				ok.Payload = &models.HashicorpCloudIamCreateServicePrincipalKeyResponse{
					ClientSecret: secret,
					Key: &models.HashicorpCloudIamServicePrincipalKey{
						ClientID:     id,
						CreatedAt:    strfmt.DateTime(now),
						ResourceName: rn,
						State:        models.HashicorpCloudIamServicePrincipalKeyStateACTIVE.Pointer(),
					},
				}

				call.Return(ok, nil)
			}

			// Run the command
			err := createRun(opts)
			if c.Error != "" {
				r.ErrorContains(err, c.Error)
				return
			}

			r.NoError(err)
			r.Contains(io.Output.String(), id)
			r.Contains(io.Output.String(), rn)
			r.Contains(io.Output.String(), c.SPName)
			r.Contains(io.Output.String(), strfmt.DateTime(now).String())
		})
	}
}

func TestCreateRun_CredFile(t *testing.T) {
	t.Parallel()
	r := require.New(t)

	// Determine the test file path.
	p := filepath.Join(t.TempDir(), "cred.json")

	io := iostreams.Test()
	spService := mock_service_principals_service.NewMockClientService(t)
	opts := &CreateOpts{
		Ctx:                context.Background(),
		Profile:            profile.TestProfile(t).SetOrgID("123").SetProjectID("456"),
		IO:                 io,
		Output:             format.New(io),
		Client:             spService,
		Name:               "test-sp",
		CredentialFilePath: &p,
	}

	// Expect a request to get the user.
	call := spService.EXPECT().ServicePrincipalsServiceCreateServicePrincipalKey(mock.Anything, nil).Once()

	secret := "superSecret12312232"
	id := "ASDKJLAD1412412"
	now := time.Now()
	rn := fmt.Sprintf("iam/project/456/service-principal/test-sp/key/%s", id)
	ok := service_principals_service.NewServicePrincipalsServiceCreateServicePrincipalKeyOK()
	ok.Payload = &models.HashicorpCloudIamCreateServicePrincipalKeyResponse{
		ClientSecret: secret,
		Key: &models.HashicorpCloudIamServicePrincipalKey{
			ClientID:     id,
			CreatedAt:    strfmt.DateTime(now),
			ResourceName: rn,
			State:        models.HashicorpCloudIamServicePrincipalKeyStateACTIVE.Pointer(),
		},
	}

	call.Return(ok, nil)

	// Run the command
	err := createRun(opts)
	r.NoError(err)
	r.Empty(io.Output.String())
	r.Contains(io.Error.String(), "credential file written to")

	// Ensure we can read the file.
	f, err := auth.ReadCredentialFile(p)
	r.NoError(err)
	r.Equal(f.ProjectID, "456")
	r.Equal(f.Scheme, auth.CredentialFileSchemeServicePrincipal)
	r.Equal(f.Oauth.ClientID, id)
	r.Equal(f.Oauth.ClientSecret, secret)
	r.Nil(f.Workload)
}
