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
	cloud "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
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
		Name    string
		Args    []string
		Profile func(t *testing.T) *profile.Profile
		Error   string
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
			Error: "no arguments allowed, but received 2",
		},
		{
			Name: "Good",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123")
			},
			Args: []string{},
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

			cmd := NewCmdList(ctx, func(o *ListOpts) error {
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
		})
	}
}

func TestListRun(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Name    string
		Resp    [][]*models.HashicorpCloudIamServicePrincipal
		RespErr bool
		Error   string
	}{
		{
			Name:    "Server error",
			RespErr: true,
			Error:   "failed to list service principals: [GET /2019-12-10/iam/{parent_resource_name}/service-principals][403]",
		},
		{
			Name: "Good no pagination",
			Resp: [][]*models.HashicorpCloudIamServicePrincipal{
				{
					{
						CreatedAt:      strfmt.DateTime(time.Date(2021, 1, 1, 1, 2, 3, 4, time.UTC)),
						ID:             "iam.service-principal:123",
						Name:           "test-sp",
						OrganizationID: "123",
						ProjectID:      "456",
						ResourceName:   "iam/project/456/service-principal/test-sp",
					},
				},
			},
		},
		{
			Name: "Good pagination",
			Resp: [][]*models.HashicorpCloudIamServicePrincipal{
				{
					{
						CreatedAt:      strfmt.DateTime(time.Date(2021, 1, 1, 1, 2, 3, 4, time.UTC)),
						ID:             "iam.service-principal:123",
						Name:           "test-sp",
						OrganizationID: "123",
						ProjectID:      "456",
						ResourceName:   "iam/project/456/service-principal/test-sp",
					},
					{
						CreatedAt:      strfmt.DateTime(time.Date(2021, 1, 1, 1, 2, 3, 4, time.UTC)),
						ID:             "iam.service-principal:12426354",
						Name:           "test-sp2",
						OrganizationID: "123",
						ProjectID:      "456",
						ResourceName:   "iam/project/456/service-principal/test-sp2",
					},
				},
				{
					{
						CreatedAt:      strfmt.DateTime(time.Date(2021, 1, 1, 1, 2, 3, 4, time.UTC)),
						ID:             "iam.service-principal:907122741",
						Name:           "test-sp-3",
						OrganizationID: "123",
						ProjectID:      "456",
						ResourceName:   "iam/project/456/service-principal/test-sp-3",
					},
					{
						CreatedAt:      strfmt.DateTime(time.Date(2021, 1, 1, 1, 2, 3, 4, time.UTC)),
						ID:             "iam.service-principal:578192",
						Name:           "test-sp-4",
						OrganizationID: "123",
						ProjectID:      "456",
						ResourceName:   "iam/project/456/service-principal/test-sp-4",
					},
				},
			},
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
				Profile: profile.TestProfile(t).SetOrgID("123"),
				Output:  format.New(io),
				Client:  spService,
			}

			e := spService.EXPECT()
			for i := 0; i < len(c.Resp) || c.RespErr; i++ {
				i := i
				call := e.ServicePrincipalsServiceListServicePrincipals(mock.MatchedBy(func(req *service_principals_service.ServicePrincipalsServiceListServicePrincipalsParams) bool {
					// Expect an org
					if req.ParentResourceName != "organization/123" {
						return false
					}

					// No initial pagination
					if i == 0 && req.PaginationNextPageToken != nil {
						return false
					} else if i >= 1 && *req.PaginationNextPageToken != fmt.Sprintf("next-page-%d", i) {
						// Expect a page token
						return false
					}

					return true
				}), nil)

				if c.RespErr {
					call.Return(nil, service_principals_service.NewServicePrincipalsServiceListServicePrincipalsDefault(http.StatusForbidden))
					break
				} else {
					ok := service_principals_service.NewServicePrincipalsServiceListServicePrincipalsOK()
					ok.Payload = &models.HashicorpCloudIamListServicePrincipalsResponse{
						ServicePrincipals: c.Resp[i],
					}

					if i < len(c.Resp)-1 {
						ok.Payload.Pagination = &cloud.HashicorpCloudCommonPaginationResponse{
							NextPageToken: fmt.Sprintf("next-page-%d", i+1),
						}
					}

					call.Return(ok, nil)
				}
			}

			// Run the command
			err := listRun(opts)
			if c.Error != "" {
				r.ErrorContains(err, c.Error)
				return
			}

			r.NoError(err)

			// Check we outputted the project
			for _, page := range c.Resp {
				for _, p := range page {
					r.Contains(io.Output.String(), p.ID)
					r.Contains(io.Output.String(), p.Name)
					r.Contains(io.Output.String(), p.ResourceName)
					r.Contains(io.Output.String(), p.CreatedAt.String())
				}
			}
		})
	}
}
