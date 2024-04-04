// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package organizations

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/organization_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/models"
	mock_organization_service "github.com/hashicorp/hcp/internal/pkg/api/mocks/github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/organization_service"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewCmdRead(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Name     string
		Args     []string
		Profile  func(t *testing.T) *profile.Profile
		ExpectID string
		Error    string
	}{
		{
			Name:    "Too many args",
			Profile: profile.TestProfile,
			Args:    []string{"foo", "bar"},
			Error:   "accepts 1 arg(s), received 2",
		},
		{
			Name:     "Good",
			Args:     []string{"123"},
			ExpectID: "123",
			Profile:  profile.TestProfile,
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

			var gotOpts *ReadOpts
			readCmd := NewCmdRead(ctx, func(o *ReadOpts) error {
				gotOpts = o
				return nil
			})
			readCmd.SetIO(io)

			code := readCmd.Run(c.Args)
			if c.Error != "" {
				r.NotZero(code)
				r.Contains(io.Error.String(), c.Error)
				return
			}

			r.Zero(code, io.Error.String())
			r.NotNil(gotOpts)
			r.Equal(c.ExpectID, gotOpts.ID)
		})
	}
}

func TestReadRun(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Name    string
		RespErr bool
		Error   string
	}{
		{
			Name:    "Server error",
			RespErr: true,
			Error:   "failed to read organization: [GET /resource-manager/2019-12-10/organizations/{id}][403]",
		},
		{
			Name: "Good",
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			r := require.New(t)

			io := iostreams.Test()
			org := mock_organization_service.NewMockClientService(t)
			opts := &ReadOpts{
				Ctx:     context.Background(),
				Profile: profile.TestProfile(t),
				IO:      io,
				Output:  format.New(io),
				Client:  org,
				ID:      "123",
			}

			// Expect a request to create the project.
			call := org.EXPECT().OrganizationServiceGet(mock.MatchedBy(func(req *organization_service.OrganizationServiceGetParams) bool {
				return req.ID == "123"
			}), nil).Once()

			if c.RespErr {
				call.Return(nil, organization_service.NewOrganizationServiceGetDefault(http.StatusForbidden))
			} else {
				ok := organization_service.NewOrganizationServiceGetOK()
				ok.Payload = &models.HashicorpCloudResourcemanagerOrganizationGetResponse{
					Organization: &models.HashicorpCloudResourcemanagerOrganization{
						CreatedAt: strfmt.DateTime(time.Now()),
						ID:        "123",
						Name:      "Hello",
						Owner: &models.HashicorpCloudResourcemanagerOrganizationOwner{
							User: "user-123",
						},
						State: models.HashicorpCloudResourcemanagerOrganizationOrganizationStateACTIVE.Pointer(),
					},
				}

				call.Return(ok, nil)
			}

			// Run the command
			err := readRun(opts)
			if c.Error != "" {
				r.ErrorContains(err, c.Error)
				return
			}

			r.NoError(err)
			r.Contains(io.Output.String(), "Hello")
			r.Contains(io.Output.String(), "123")
			r.Contains(io.Output.String(), "user-123")
			r.Contains(io.Output.String(), "ACTIVE")
		})
	}
}
