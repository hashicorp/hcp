// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package workloadidentityproviders

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
		Expect  *ListOpts
	}{
		{
			Name:    "No Org",
			Profile: profile.TestProfile,
			Args: []string{
				"iam/project/my-project/service-principal/my-sp",
			},
			Error: "Organization ID and Project ID must be configured before running the command.",
		},
		{
			Name: "No Project",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123")
			},
			Args: []string{
				"iam/project/my-project/service-principal/my-sp",
			},
			Error: "Organization ID and Project ID must be configured before running the command.",
		},
		{
			Name: "Too many args",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123").SetProjectID("456")
			},
			Args: []string{
				"iam/project/my-project/service-principal/my-sp",
				"extra",
			},
			Error: "accepts 1 arg(s), received 2",
		},
		{
			Name: "Good",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123").SetProjectID("456")
			},
			Args: []string{
				"iam/project/my-project/service-principal/my-sp",
			},
			Expect: &ListOpts{
				SP: "iam/project/my-project/service-principal/my-sp",
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

			var gotOpts *ListOpts
			listCmd := NewCmdList(ctx, func(o *ListOpts) error {
				gotOpts = o
				return nil
			})
			listCmd.SetIO(io)

			code := listCmd.Run(c.Args)
			if c.Error != "" {
				r.NotZero(code)
				r.Contains(io.Error.String(), c.Error)
				return
			}

			r.Zero(code, io.Error.String())
			r.NotNil(gotOpts)
			r.Equal(c.Expect.SP, gotOpts.SP)
		})
	}
}

func TestListRun(t *testing.T) {
	t.Parallel()

	time1 := strfmt.DateTime(time.Now())
	time2 := strfmt.DateTime(time.Now().Add(time.Hour * 24))
	time3 := strfmt.DateTime(time.Now().Add(time.Hour * 48))

	cases := []struct {
		Name    string
		Resp    [][]*models.HashicorpCloudIamWorkloadIdentityProvider
		RespErr bool
		Error   string
	}{
		{
			Name:    "Server error",
			RespErr: true,
			Error:   "failed to list workload identity providers: [GET /2019-12-10/{parent_resource_name}/workload-identity-providers][403]",
		},
		{
			Name: "Good no pagination",
			Resp: [][]*models.HashicorpCloudIamWorkloadIdentityProvider{
				{
					{
						ConditionalAccess: "jwt_claims.sub == \"1\"",
						CreatedAt:         &time1,
						Description:       "1",
						OidcConfig: &models.HashicorpCloudIamOIDCWorkloadIdentityProviderConfig{
							AllowedAudiences: []string{"aud1"},
							IssuerURI:        "https://example-1.com",
						},
						ResourceID:   "iam.workload-identity-provider:1",
						ResourceName: "iam/project/my-project/service-principal/my-sp/workload-identity-provider/1",
					},
				},
			},
		},
		{
			Name: "Good pagination",
			Resp: [][]*models.HashicorpCloudIamWorkloadIdentityProvider{
				{
					{
						ConditionalAccess: "jwt_claims.sub == \"1\"",
						CreatedAt:         &time1,
						Description:       "1",
						OidcConfig: &models.HashicorpCloudIamOIDCWorkloadIdentityProviderConfig{
							AllowedAudiences: []string{"aud1"},
							IssuerURI:        "https://example-1.com",
						},
						ResourceID:   "iam.workload-identity-provider:1",
						ResourceName: "iam/project/my-project/service-principal/my-sp/workload-identity-provider/1",
					},
					{
						ConditionalAccess: "jwt_claims.sub == \"2\"",
						CreatedAt:         &time2,
						Description:       "2",
						OidcConfig: &models.HashicorpCloudIamOIDCWorkloadIdentityProviderConfig{
							AllowedAudiences: []string{"aud2"},
							IssuerURI:        "https://example-2.com",
						},
						ResourceID:   "iam.workload-identity-provider:2",
						ResourceName: "iam/project/my-project/service-principal/my-sp/workload-identity-provider/2",
					},
				},
				{
					{
						ConditionalAccess: "jwt_claims.sub == \"3\"",
						CreatedAt:         &time3,
						Description:       "3",
						OidcConfig: &models.HashicorpCloudIamOIDCWorkloadIdentityProviderConfig{
							AllowedAudiences: []string{"aud3"},
							IssuerURI:        "https://example-3.com",
						},
						ResourceID:   "iam.workload-identity-provider:3",
						ResourceName: "iam/project/my-project/service-principal/my-sp/workload-identity-provider/3",
					},
					{
						ConditionalAccess: "jwt_claims.sub == \"4\"",
						CreatedAt:         &time1,
						Description:       "4",
						OidcConfig: &models.HashicorpCloudIamOIDCWorkloadIdentityProviderConfig{
							AllowedAudiences: []string{"aud4"},
							IssuerURI:        "https://example-4.com",
						},
						ResourceID:   "iam.workload-identity-provider:4",
						ResourceName: "iam/project/my-project/service-principal/my-sp/workload-identity-provider/4",
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
				SP:      "iam/project/my-project/service-principal/my-sp",
			}

			e := spService.EXPECT()
			for i := 0; i < len(c.Resp) || c.RespErr; i++ {
				i := i
				call := e.ServicePrincipalsServiceListWorkloadIdentityProvider(mock.MatchedBy(func(req *service_principals_service.ServicePrincipalsServiceListWorkloadIdentityProviderParams) bool {
					// Expect the SP
					if req.ParentResourceName != opts.SP {
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
					call.Return(nil, service_principals_service.NewServicePrincipalsServiceListWorkloadIdentityProviderDefault(http.StatusForbidden))
					break
				} else {
					ok := service_principals_service.NewServicePrincipalsServiceListWorkloadIdentityProviderOK()
					ok.Payload = &models.HashicorpCloudIamListWorkloadIdentityProviderResponse{
						Providers: c.Resp[i],
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
					r.Contains(io.Output.String(), p.ResourceID)
					r.Contains(io.Output.String(), p.ResourceName)
					r.Contains(io.Output.String(), p.Description)
					r.Contains(io.Output.String(), p.OidcConfig.IssuerURI)
					r.Contains(io.Output.String(), p.OidcConfig.AllowedAudiences[0])
					r.Contains(io.Output.String(), p.ConditionalAccess)
					r.Contains(io.Output.String(), p.CreatedAt.String())
				}
			}
		})
	}
}
