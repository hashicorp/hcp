// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package workloadidentityproviders

import (
	"context"
	"fmt"
	"net/http"
	"slices"
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

func TestNewCmdCreateOIDC(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Name    string
		Args    []string
		Profile func(t *testing.T) *profile.Profile
		Error   string
		Expect  *CreateOIDCOpts
	}{
		{
			Name:    "No Org",
			Profile: profile.TestProfile,
			Args: []string{
				"test",
				"--service-principal", "iam/project/PROJECT/service-principal/example-sp",
				"--issuer", "https://example.com/",
				"--conditional-access", "aws.arn matches \"arn:aws:iam::123456789012:role/example-role/*\"",
			},
			Error: "Organization ID and Project ID must be configured before running the command.",
		},
		{
			Name: "No Project",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123")
			},
			Args: []string{
				"test",
				"--service-principal", "iam/project/PROJECT/service-principal/example-sp",
				"--issuer", "https://example.com/",
				"--conditional-access", "aws.arn matches \"arn:aws:iam::123456789012:role/example-role/*\"",
			},
			Error: "Organization ID and Project ID must be configured before running the command.",
		},
		{
			Name: "Too many args",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123").SetProjectID("456")
			},
			Args: []string{
				"test", "extra",
				"--service-principal", "iam/project/PROJECT/service-principal/example-sp",
				"--issuer", "https://example.com/",
				"--conditional-access", "aws.arn matches \"arn:aws:iam::123456789012:role/example-role/*\"",
			},
			Error: "accepts 1 arg(s), received 2",
		},
		{
			Name: "Missing flag",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123").SetProjectID("456")
			},
			Args: []string{
				"test",
				"--service-principal", "iam/project/PROJECT/service-principal/example-sp",
				"--conditional-access", "aws.arn matches \"arn:aws:iam::123456789012:role/example-role/*\"",
			},
			Error: "missing required flag: --issuer=URI",
		},
		{
			Name: "Good",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123").SetProjectID("456")
			},
			Args: []string{
				"test",
				"--service-principal", "iam/project/PROJECT/service-principal/example-sp",
				"--issuer", "https://example.com/",
				"--allowed-audience", "example-audience",
				"--conditional-access", "aws.arn matches \"arn:aws:iam::123456789012:role/example-role/*\"",
				"--description", "example",
			},
			Expect: &CreateOIDCOpts{
				Name:              "test",
				SP:                "iam/project/PROJECT/service-principal/example-sp",
				IssuerURI:         "https://example.com/",
				AllowedAudiences:  []string{"example-audience"},
				ConditionalAccess: "aws.arn matches \"arn:aws:iam::123456789012:role/example-role/*\"",
				Description:       "example",
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

			var gotOpts *CreateOIDCOpts
			createCmd := NewCmdCreateOIDC(ctx, func(o *CreateOIDCOpts) error {
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
			r.Equal(c.Expect.SP, gotOpts.SP)
			r.Equal(c.Expect.IssuerURI, gotOpts.IssuerURI)
			r.EqualValues(c.Expect.AllowedAudiences, gotOpts.AllowedAudiences)
			r.Equal(c.Expect.ConditionalAccess, gotOpts.ConditionalAccess)
			r.Equal(c.Expect.Description, gotOpts.Description)
		})
	}
}

func TestCreateOIDC_Validation(t *testing.T) {
	t.Parallel()
	r := require.New(t)
	io := iostreams.Test()
	opts := &CreateOIDCOpts{
		Ctx:               context.Background(),
		Profile:           profile.TestProfile(t).SetOrgID("123").SetProjectID("456"),
		Output:            format.New(io),
		IO:                io,
		Name:              "test-sp",
		SP:                "iam/project/456/service-principals/test-sp",
		IssuerURI:         "https://example.com/",
		ConditionalAccess: "jwt.sub == \"test\"",
		Description:       "example",
	}

	err := createOIDCRun(opts)
	r.ErrorContains(err, "invalid service principal resource name: iam/project/456/service-principals/test-sp")

}

func TestCreateOIDCRun(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Name              string
		Profile           func(t *testing.T) *profile.Profile
		RespErr           bool
		WIPName           string
		SPResouceName     string
		Issuer            string
		AllowedAudiences  []string
		ConditionalAccess string
		Description       string
		Error             string
	}{
		{
			Name: "Server error",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123").SetProjectID("456")
			},
			WIPName:           "test-sp",
			SPResouceName:     "iam/project/456/service-principal/test-sp",
			Issuer:            "https://example.com/",
			AllowedAudiences:  []string{"example-audience", "test"},
			ConditionalAccess: "jwt.sub == \"test\"",
			Description:       "example",
			RespErr:           true,
			Error:             "failed to create workload identity provider: [POST /2019-12-10/{parent_resource_name}/workload-identity-providers][403]",
		},
		{
			Name: "Good",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123").SetProjectID("456")
			},
			WIPName:           "test-sp",
			SPResouceName:     "iam/project/456/service-principal/test-sp",
			Issuer:            "https://example.com/",
			AllowedAudiences:  []string{"example-audience", "test"},
			ConditionalAccess: "aws.arn matches \"arn:aws:iam::123456789012:role/example-role/*\"",
			Description:       "example",
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			r := require.New(t)

			io := iostreams.Test()
			spService := mock_service_principals_service.NewMockClientService(t)
			opts := &CreateOIDCOpts{
				Ctx:               context.Background(),
				Profile:           c.Profile(t),
				Output:            format.New(io),
				IO:                io,
				Name:              c.WIPName,
				SP:                c.SPResouceName,
				IssuerURI:         c.Issuer,
				AllowedAudiences:  c.AllowedAudiences,
				ConditionalAccess: c.ConditionalAccess,
				Description:       c.Description,
				Client:            spService,
			}

			// Expect a request to get the user.
			call := spService.EXPECT().ServicePrincipalsServiceCreateWorkloadIdentityProvider(mock.MatchedBy(func(req *service_principals_service.ServicePrincipalsServiceCreateWorkloadIdentityProviderParams) bool {
				return req.ParentResourceName == c.SPResouceName &&
					req.Body.Name == c.WIPName &&
					req.Body.Provider.OidcConfig.IssuerURI == c.Issuer &&
					slices.Equal(req.Body.Provider.OidcConfig.AllowedAudiences, c.AllowedAudiences) &&
					req.Body.Provider.ConditionalAccess == c.ConditionalAccess &&
					req.Body.Provider.Description == c.Description
			}), nil).Once()

			id := "iam.workload-identity-provider:124124"
			now := strfmt.DateTime(time.Now())
			rn := fmt.Sprintf("%s/workload-identity-provider/%s", c.SPResouceName, c.WIPName)

			if c.RespErr {
				call.Return(nil, service_principals_service.NewServicePrincipalsServiceCreateWorkloadIdentityProviderDefault(http.StatusForbidden))
			} else {
				ok := service_principals_service.NewServicePrincipalsServiceCreateWorkloadIdentityProviderOK()
				ok.Payload = &models.HashicorpCloudIamCreateWorkloadIdentityProviderResponse{
					Provider: &models.HashicorpCloudIamWorkloadIdentityProvider{
						CreatedAt:         &now,
						ResourceID:        id,
						ResourceName:      rn,
						ConditionalAccess: c.ConditionalAccess,
						Description:       c.Description,
						OidcConfig: &models.HashicorpCloudIamOIDCWorkloadIdentityProviderConfig{
							IssuerURI:        c.Issuer,
							AllowedAudiences: c.AllowedAudiences,
						},
					},
				}

				call.Return(ok, nil)
			}

			// Run the command
			err := createOIDCRun(opts)
			if c.Error != "" {
				r.ErrorContains(err, c.Error)
				return
			}

			r.NoError(err)
			r.Contains(io.Output.String(), id)
			r.Contains(io.Output.String(), rn)
			r.Contains(io.Output.String(), c.Description)
			r.Contains(io.Output.String(), c.Issuer)
			r.Contains(io.Output.String(), c.AllowedAudiences[0])
			r.Contains(io.Output.String(), c.ConditionalAccess)
			r.Contains(io.Output.String(), now.String())
		})
	}
}
