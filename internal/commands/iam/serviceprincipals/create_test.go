// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: MPL-2.0

package serviceprincipals

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

func TestNewCmdCreate(t *testing.T) {
	t.Parallel()

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
			Name: "Good",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123")
			},
			Args: []string{"foo"},
			Expect: &CreateOpts{
				Name: "foo",
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
		})
	}
}

func TestCreateRun(t *testing.T) {
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
			ParentResourceName: "organization/123",
			RespErr:            true,
			Error:              "failed to create service principal: [POST /2019-12-10/iam/{parent_resource_name}/service-principals][403]",
		},
		{
			Name: "Good org",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123")
			},
			SPName:             "test-sp",
			ParentResourceName: "organization/123",
		},
		{
			Name: "Good project",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123").SetProjectID("456")
			},
			SPName:             "test-sp",
			ParentResourceName: "project/456",
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
				Output:  format.New(io),
				Client:  spService,
				Name:    c.SPName,
			}

			// Expect a request to get the user.
			call := spService.EXPECT().ServicePrincipalsServiceCreateServicePrincipal(mock.MatchedBy(func(req *service_principals_service.ServicePrincipalsServiceCreateServicePrincipalParams) bool {
				return req.ParentResourceName == c.ParentResourceName && req.Body.Name == c.SPName
			}), nil).Once()

			id := "iam.service-principal:124124"
			now := time.Now()
			rn := fmt.Sprintf("iam/%s/service-principal/%s", c.ParentResourceName, c.SPName)
			if c.RespErr {
				call.Return(nil, service_principals_service.NewServicePrincipalsServiceCreateServicePrincipalDefault(http.StatusForbidden))
			} else {
				ok := service_principals_service.NewServicePrincipalsServiceCreateServicePrincipalOK()
				ok.Payload = &models.HashicorpCloudIamCreateServicePrincipalResponse{
					ServicePrincipal: &models.HashicorpCloudIamServicePrincipal{
						CreatedAt:      strfmt.DateTime(now),
						ID:             id,
						Name:           c.SPName,
						OrganizationID: opts.Profile.OrganizationID,
						ProjectID:      opts.Profile.ProjectID,
						ResourceName:   rn,
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
