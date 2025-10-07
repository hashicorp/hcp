// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: MPL-2.0

package keys

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
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

func TestNewCmdList(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Name     string
		Args     []string
		Profile  func(t *testing.T) *profile.Profile
		Error    string
		ExpectSP string
	}{
		{
			Name:    "No Org",
			Profile: profile.TestProfile,
			Args:    []string{},
			Error:   "Organization ID must be configured",
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
			Name: "Good",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123").SetProjectID("456")
			},
			Args:     []string{"test-sp"},
			ExpectSP: "test-sp",
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

			var gotOpts *ListOpts
			cmd := NewCmdList(ctx, func(o *ListOpts) error {
				gotOpts = o
				return nil
			})
			cmd.SetIO(io)

			code := cmd.Run(c.Args)
			if c.Error != "" {
				r.NotZero(code)
				r.Contains(io.Error.String(), c.Error)
				return
			}

			r.Zero(code, io.Error.String())
			r.Equal(c.ExpectSP, gotOpts.Name)
		})
	}
}

func TestListRun(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Name         string
		Profile      func(t *testing.T) *profile.Profile
		RespErr      bool
		SPName       string
		ResourceName string
		Error        string
	}{
		{
			Name: "Server error",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123")
			},
			SPName:       "test-sp",
			ResourceName: "iam/organization/123/service-principal/test-sp",
			RespErr:      true,
			Error:        "failed to read service principal to retrieve keys: [GET /2019-12-10/{resource_name}][403]",
		},
		{
			Name: "Good org prefix",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123")
			},
			SPName:       "test-sp",
			ResourceName: "iam/organization/123/service-principal/test-sp",
		},
		{
			Name: "Good org resource name",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123")
			},
			SPName:       "iam/organization/123/service-principal/test-sp",
			ResourceName: "iam/organization/123/service-principal/test-sp",
		},
		{
			Name: "Good project prefix",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123").SetProjectID("456")
			},
			SPName:       "test-sp",
			ResourceName: "iam/project/456/service-principal/test-sp",
		},
		{
			Name: "Good project resource name",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123").SetProjectID("456")
			},
			SPName:       "iam/project/789/service-principal/test-sp",
			ResourceName: "iam/project/789/service-principal/test-sp",
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			r := require.New(t)

			io := iostreams.Test()
			spService := mock_service_principals_service.NewMockClientService(t)
			opts := &ListOpts{
				Ctx:     context.Background(),
				Profile: c.Profile(t),
				Output:  format.New(io),
				Client:  spService,
				Name:    c.SPName,
			}

			// Expect a request to get the user.
			call := spService.EXPECT().ServicePrincipalsServiceGetServicePrincipal(mock.MatchedBy(func(req *service_principals_service.ServicePrincipalsServiceGetServicePrincipalParams) bool {
				return req.ResourceName == c.ResourceName
			}), nil).Once()

			k1 := &models.HashicorpCloudIamServicePrincipalKey{
				ClientID:     "ADSKLAFHDOQEQR",
				CreatedAt:    strfmt.DateTime(time.Now().Add(-1 * time.Hour)),
				ResourceName: fmt.Sprintf("%s/key/%s", c.ResourceName, "ADSKLAFHDOQEQR"),
			}
			k2 := &models.HashicorpCloudIamServicePrincipalKey{
				ClientID:     "QWERTYIOP",
				CreatedAt:    strfmt.DateTime(time.Now().Add(-3 * time.Hour)),
				ResourceName: fmt.Sprintf("%s/key/%s", c.ResourceName, "QWERTYIOP"),
			}

			if c.RespErr {
				call.Return(nil, service_principals_service.NewServicePrincipalsServiceGetServicePrincipalDefault(http.StatusForbidden))
			} else {
				ok := service_principals_service.NewServicePrincipalsServiceGetServicePrincipalOK()
				ok.Payload = &models.HashicorpCloudIamGetServicePrincipalResponse{
					ServicePrincipal: &models.HashicorpCloudIamServicePrincipal{
						CreatedAt:      strfmt.DateTime(time.Now()),
						ID:             "iam.service-principal:124124",
						Name:           c.SPName,
						OrganizationID: opts.Profile.OrganizationID,
						ProjectID:      opts.Profile.ProjectID,
						ResourceName:   c.ResourceName,
					},
					Keys: []*models.HashicorpCloudIamServicePrincipalKey{k1, k2},
				}
				call.Return(ok, nil)
			}

			// Run the command
			err := listRun(opts)
			if c.Error != "" {
				r.ErrorContains(err, c.Error)
				return
			}

			r.NoError(err)
			r.Contains(io.Output.String(), k1.ResourceName)
			r.Contains(io.Output.String(), k1.ClientID)
			r.Contains(io.Output.String(), k1.CreatedAt.String())
			r.Contains(io.Output.String(), k2.ResourceName)
			r.Contains(io.Output.String(), k2.ClientID)
			r.Contains(io.Output.String(), k2.CreatedAt.String())
		})
	}
}
