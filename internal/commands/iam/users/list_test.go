// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: MPL-2.0

package users

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/go-openapi/runtime/client"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/client/iam_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/models"
	cloud "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	mock_iam_service "github.com/hashicorp/hcp/internal/pkg/api/mocks/github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/client/iam_service"
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
		Resp    [][]*models.HashicorpCloudIamUserPrincipal
		RespErr bool
		Error   string
	}{
		{
			Name:    "Server error",
			RespErr: true,
			Error:   "failed to list users: [GET /iam/2019-12-10/organizations/{organization_id}/user-principals][403]",
		},
		{
			Name: "Good no pagination",
			Resp: [][]*models.HashicorpCloudIamUserPrincipal{
				{
					{
						Email:    "alex@hashicorp.com",
						FullName: "Alex Dadgar",
						ID:       "5b68791a-944f-4efd-b4d0-22648ffe0e41",
					},
				},
			},
		},
		{
			Name: "Good pagination",
			Resp: [][]*models.HashicorpCloudIamUserPrincipal{
				{
					{
						Email:    "bob@hashicorp.com",
						FullName: "Bob Bobert",
						ID:       "ff4dc0bf-951c-46f3-a646-f113132a51e9",
					},
					{
						Email:    "charlie@hashicorp.com",
						FullName: "Charlie Charleton",
						ID:       "85066601-e47d-46dc-92f8-9732f1763718",
					},
				},
				{
					{
						Email:    "david@hashicorp.com",
						FullName: "David Davey",
						ID:       "695e3a49-996d-4db7-a9bf-43d873aed9e0",
					},
					{
						Email:    "edward@hashicorp.com",
						FullName: "Edward Edgerton",
						ID:       "3f630e37-4b99-41a3-b6f9-c2e085e145bb",
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
			iam := mock_iam_service.NewMockClientService(t)
			opts := &ListOpts{
				Ctx:     context.Background(),
				Profile: profile.TestProfile(t).SetOrgID("123"),
				Output:  format.New(io),
				Client:  iam,
			}

			e := iam.EXPECT()
			for i := 0; i < len(c.Resp) || c.RespErr; i++ {
				i := i
				call := e.IamServiceListUserPrincipalsByOrganization(mock.MatchedBy(func(req *iam_service.IamServiceListUserPrincipalsByOrganizationParams) bool {
					// Expect an org
					if req.OrganizationID != "123" {
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
					call.Return(nil, iam_service.NewIamServiceListUserPrincipalsByOrganizationDefault(http.StatusForbidden))
					break
				} else {
					ok := iam_service.NewIamServiceListUserPrincipalsByOrganizationOK()
					ok.Payload = &models.HashicorpCloudIamListUserPrincipalsByOrganizationResponse{
						UserPrincipals: c.Resp[i],
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
					r.Contains(io.Output.String(), p.FullName)
					r.Contains(io.Output.String(), p.Email)
				}
			}
		})
	}
}
