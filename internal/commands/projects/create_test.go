// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package projects

import (
	"context"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-billing/preview/2020-11-05/client/billing_account_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/project_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/models"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	billingModels "github.com/hashicorp/hcp-sdk-go/clients/cloud-billing/preview/2020-11-05/models"
	mock_billing_account_service "github.com/hashicorp/hcp/internal/pkg/api/mocks/github.com/hashicorp/hcp-sdk-go/clients/cloud-billing/preview/2020-11-05/client/billing_account_service"
	mock_project_service "github.com/hashicorp/hcp/internal/pkg/api/mocks/github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/project_service"
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
			Error:   "Organization ID must be configured",
		},
		{
			Name: "No name",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123")
			},
			Args:  []string{},
			Error: "accepts 1 arg(s), received 0",
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
			Args: []string{"foo", "--description=test", "--set-as-default"},
			Expect: &CreateOpts{
				Name:        "foo",
				Description: "test",
				Default:     true,
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
			cmd := NewCmdCreate(ctx, func(o *CreateOpts) error {
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
			r.NotNil(gotOpts)
			r.Equal(c.Expect.Name, gotOpts.Name)
			r.Equal(c.Expect.Description, gotOpts.Description)
			r.Equal(c.Expect.Default, gotOpts.Default)
		})
	}
}

func TestCreateRun(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Name    string
		Resp    *models.HashicorpCloudResourcemanagerProject
		RespErr bool
		Error   string
		Default bool
	}{
		{
			Name:    "Server error",
			Default: false,
			RespErr: true,
			Error:   "failed to create project: [POST /resource-manager/2019-12-10/projects][403]",
		},
		{
			Name: "Good no default",
			Resp: &models.HashicorpCloudResourcemanagerProject{
				CreatedAt:   strfmt.DateTime(time.Now()),
				Description: "Good test",
				ID:          "456",
				Name:        "Good",
			},
			Default: false,
		},
		{
			Name: "Good default",
			Resp: &models.HashicorpCloudResourcemanagerProject{
				CreatedAt:   strfmt.DateTime(time.Now()),
				Description: "Good test",
				ID:          "456",
				Name:        "Good",
			},
			Default: true,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			r := require.New(t)

			io := iostreams.Test()
			project := mock_project_service.NewMockClientService(t)
			billing := mock_billing_account_service.NewMockClientService(t)
			opts := &CreateOpts{
				Ctx:           context.Background(),
				Profile:       profile.TestProfile(t).SetOrgID("123"),
				IO:            io,
				Output:        format.New(io),
				Name:          "test",
				Description:   "testing",
				Default:       c.Default,
				ProjectClient: project,
				BillingClient: billing,
			}

			// Expect a request to create the project.
			call := project.EXPECT().ProjectServiceCreate(mock.MatchedBy(func(getReq *project_service.ProjectServiceCreateParams) bool {
				if getReq.Body == nil {
					return false
				}
				if getReq.Body.Parent == nil || getReq.Body.Parent.ID != "123" {
					return false
				}

				return getReq.Body.Name == opts.Name && getReq.Body.Description == opts.Description
			}), nil).Once()

			if c.RespErr {
				call.Return(nil, project_service.NewProjectServiceCreateDefault(http.StatusForbidden))
			} else {
				ok := project_service.NewProjectServiceCreateOK()
				ok.Payload = &models.HashicorpCloudResourcemanagerProjectCreateResponse{
					Project: c.Resp,
				}

				call.Return(ok, nil)

				// Expect the billing get request
				country := "test-country"
				projectIDs := []string{"p1", "p2"}
				getResp := &billing_account_service.BillingAccountServiceGetOK{
					Payload: &billingModels.Billing20201105GetBillingAccountResponse{
						BillingAccount: &billingModels.Billing20201105BillingAccount{
							Country:        billingModels.NewBilling20201105Country(billingModels.Billing20201105Country(country)),
							ID:             defaultBillingAccountID,
							Name:           "Billing Account",
							OrganizationID: opts.Profile.OrganizationID,
							ProjectIds:     projectIDs,
						},
					},
				}

				billing.EXPECT().BillingAccountServiceGet(mock.MatchedBy(func(getReq *billing_account_service.BillingAccountServiceGetParams) bool {
					return getReq.OrganizationID == opts.Profile.OrganizationID && getReq.ID == defaultBillingAccountID
				}), nil).Once().Return(getResp, nil)

				// Expect the update request
				projectIDs = append(projectIDs, c.Resp.ID)
				billing.EXPECT().BillingAccountServiceUpdate(mock.MatchedBy(func(updateReq *billing_account_service.BillingAccountServiceUpdateParams) bool {
					return updateReq.OrganizationID == opts.Profile.OrganizationID &&
						updateReq.ID == getResp.Payload.BillingAccount.ID &&
						updateReq.Body.Name == getResp.Payload.BillingAccount.Name &&
						string(*updateReq.Body.Country) == string(*getResp.Payload.BillingAccount.Country) &&
						reflect.DeepEqual(updateReq.Body.ProjectIds, projectIDs)
				}), nil).Once().Return(&billing_account_service.BillingAccountServiceUpdateOK{}, nil)
			}

			// Run the command
			err := createRun(opts)
			if c.Error != "" {
				r.ErrorContains(err, c.Error)
				return
			}

			r.NoError(err)
			if c.Default {
				r.Equal(c.Resp.ID, opts.Profile.ProjectID)
				r.Contains(io.Error.String(), "set as default project in active profile")
			} else {
				r.NotEqual(c.Resp.ID, opts.Profile.ProjectID)
				r.Empty(io.Error.String())
			}

			// Check we outputted the project
			r.Contains(io.Output.String(), c.Resp.ID)
			r.Contains(io.Output.String(), c.Resp.Name)
			r.Contains(io.Output.String(), c.Resp.Description)
		})
	}
}
