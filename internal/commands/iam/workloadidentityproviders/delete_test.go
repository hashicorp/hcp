// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package workloadidentityproviders

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/go-openapi/runtime/client"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/client/service_principals_service"
	mock_service_principals_service "github.com/hashicorp/hcp/internal/pkg/api/mocks/github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/client/service_principals_service"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewCmdDelete(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Name    string
		Args    []string
		Profile func(t *testing.T) *profile.Profile
		Error   string
		Expect  *DeleteOpts
	}{
		{
			Name:    "No Org",
			Profile: profile.TestProfile,
			Args: []string{
				"iam/project/my-project/service-principal/my-sp/workload-identity-provider/example-wip",
			},
			Error: "Organization ID and Project ID must be configured before running the command.",
		},
		{
			Name: "No Project",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123")
			},
			Args: []string{
				"iam/project/my-project/service-principal/my-sp/workload-identity-provider/example-wip",
			},
			Error: "Organization ID and Project ID must be configured before running the command.",
		},
		{
			Name: "Too many args",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123").SetProjectID("456")
			},
			Args: []string{
				"iam/project/my-project/service-principal/my-sp/workload-identity-provider/example-wip",
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
				"iam/project/my-project/service-principal/my-sp/workload-identity-provider/example-wip",
			},
			Expect: &DeleteOpts{
				Name: "iam/project/my-project/service-principal/my-sp/workload-identity-provider/example-wip",
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

			var gotOpts *DeleteOpts
			deleteCmd := NewCmdDelete(ctx, func(o *DeleteOpts) error {
				gotOpts = o
				return nil
			})
			deleteCmd.SetIO(io)

			code := deleteCmd.Run(c.Args)
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

func TestDeleteRun(t *testing.T) {
	t.Parallel()

	wipRN := "iam/project/my-project/service-principal/my-sp/workload-identity-provider/example-wip"
	cases := []struct {
		Name    string
		Profile func(t *testing.T) *profile.Profile
		RespErr bool
		WIP     string
		Error   string
	}{
		{
			Name: "Server error",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123").SetProjectID("456")
			},
			WIP:     wipRN,
			RespErr: true,
			Error:   "failed to delete workload identity provider: [DELETE /2019-12-10/{resource_name_4}][403]",
		},
		{
			Name: "Good",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123").SetProjectID("456")
			},
			WIP: wipRN,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			r := require.New(t)

			io := iostreams.Test()
			spService := mock_service_principals_service.NewMockClientService(t)
			opts := &DeleteOpts{
				Ctx:     context.Background(),
				Profile: c.Profile(t),
				IO:      io,
				Name:    c.WIP,
				Client:  spService,
			}

			// Expect a request to get the user.
			call := spService.EXPECT().ServicePrincipalsServiceDeleteWorkloadIdentityProvider(mock.MatchedBy(func(req *service_principals_service.ServicePrincipalsServiceDeleteWorkloadIdentityProviderParams) bool {
				return req.ResourceName4 == c.WIP
			}), nil).Once()

			if c.RespErr {
				call.Return(nil, service_principals_service.NewServicePrincipalsServiceDeleteWorkloadIdentityProviderDefault(http.StatusForbidden))
			} else {
				ok := service_principals_service.NewServicePrincipalsServiceDeleteWorkloadIdentityProviderOK()
				call.Return(ok, nil)
			}

			// Run the command
			err := deleteRun(opts)
			if c.Error != "" {
				r.ErrorContains(err, c.Error)
				return
			}

			r.NoError(err)
			r.Contains(io.Error.String(), fmt.Sprintf("Workload identity provider %q deleted", c.WIP))
		})
	}
}
