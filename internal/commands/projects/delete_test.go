// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package projects

import (
	"context"
	"net/http"
	"testing"

	"github.com/go-openapi/runtime/client"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/project_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/models"
	cloud "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	mock_project_service "github.com/hashicorp/hcp/internal/pkg/api/mocks/github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/project_service"
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
	}{
		{
			Name:    "No Org",
			Profile: profile.TestProfile,
			Args:    []string{},
			Error:   "Organization ID and Project ID must be configured before running the command.",
		},
		{
			Name: "Too many args",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123").SetProjectID("456")
			},
			Args:  []string{"foo", "bar"},
			Error: "no arguments allowed, but received 2",
		},
		{
			Name: "Good",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123").SetProjectID("456")
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
			r.NotNil(gotOpts.Profiles)
		})
	}
}

func TestDeleteRun(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Name    string
		RespErr bool
		Error   string
	}{
		{
			Name:    "Server error",
			RespErr: true,
			Error:   "failed to delete project: [DELETE /resource-manager/2019-12-10/projects/{id}][403]",
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
			project := mock_project_service.NewMockClientService(t)
			opts := &DeleteOpts{
				Ctx:      context.Background(),
				Profile:  profile.TestProfile(t).SetOrgID("123").SetProjectID("456"),
				Profiles: profile.TestLoader(t),
				IO:       io,
				Client:   project,
			}

			// Expect a request to create the project.
			call := project.EXPECT().ProjectServiceDelete(mock.MatchedBy(func(req *project_service.ProjectServiceDeleteParams) bool {
				return req.ID == "456"
			}), nil).Once()

			if c.RespErr {
				call.Return(nil, project_service.NewProjectServiceDeleteDefault(http.StatusForbidden))
			} else {
				ok := project_service.NewProjectServiceDeleteOK()
				ok.Payload = &models.HashicorpCloudResourcemanagerProjectDeleteResponse{
					Operation: &cloud.HashicorpCloudOperationOperation{
						ID:    "op-456",
						State: cloud.HashicorpCloudOperationOperationStateRUNNING.Pointer(),
					},
				}

				call.Return(ok, nil)
			}

			// Run the command
			err := deleteRun(opts)
			if c.Error != "" {
				r.ErrorContains(err, c.Error)
				return
			}

			r.NoError(err)
			r.Contains(io.Error.String(), "Project \"456\" deleted")
		})
	}
}

func TestDeleteRun_ProjectInProfile(t *testing.T) {
	t.Parallel()
	r := require.New(t)

	projectID := "proj-to-delete"

	io := iostreams.Test()
	project := mock_project_service.NewMockClientService(t)
	l := profile.TestLoader(t)

	opts := &DeleteOpts{
		Ctx:      context.Background(),
		Profile:  l.DefaultProfile().SetOrgID("123").SetProjectID(projectID),
		Profiles: l,
		IO:       io,
		Client:   project,
	}

	// Create two profiles with the project ID set
	expectedProfiles := []string{"profile-a", "profile-b"}
	for _, name := range expectedProfiles {
		p, err := l.NewProfile(name)
		r.NoError(err)
		p.ProjectID = projectID
		r.NoError(p.Write())
	}

	// Expect a request to create the project.
	ok := project_service.NewProjectServiceDeleteOK()
	ok.Payload = &models.HashicorpCloudResourcemanagerProjectDeleteResponse{
		Operation: &cloud.HashicorpCloudOperationOperation{
			ID:    "op-456",
			State: cloud.HashicorpCloudOperationOperationStateRUNNING.Pointer(),
		},
	}
	project.EXPECT().ProjectServiceDelete(mock.Anything, nil).Return(ok, nil).Once()

	// Run the command
	err := deleteRun(opts)
	r.NoError(err)

	// Expect to be warned
	r.Contains(io.Error.String(), "The following profiles have")
	for _, name := range expectedProfiles {
		r.Contains(io.Error.String(), name)
	}
}

func TestDeleteRun_RejectPrompt(t *testing.T) {
	t.Parallel()
	r := require.New(t)

	io := iostreams.Test()
	io.ErrorTTY = true
	io.InputTTY = true

	project := mock_project_service.NewMockClientService(t)
	l := profile.TestLoader(t)
	opts := &DeleteOpts{
		Ctx:      context.Background(),
		Profile:  l.DefaultProfile().SetOrgID("123").SetProjectID("456"),
		Profiles: l,
		IO:       io,
		Client:   project,
	}

	// Reject the deletion
	_, err := io.Input.WriteRune('n')
	r.NoError(err)

	// Run the command
	err = deleteRun(opts)
	r.NoError(err)

	// Expect to be warned
	r.Contains(io.Error.String(), "Your project will be deleted.")

	// We did not mock a call to delete the project, so if successful, we
	// exited.
}
