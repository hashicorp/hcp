// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package profile

import (
	"net/http"
	"testing"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/organization_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/project_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/models"
	mock_organization_service "github.com/hashicorp/hcp/internal/pkg/api/mocks/github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/organization_service"
	mock_project_service "github.com/hashicorp/hcp/internal/pkg/api/mocks/github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/project_service"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSet(t *testing.T) {
	t.Parallel()
	cases := []struct {
		Name         string
		Property     string
		Value        string
		Error        string
		CheckProfile func(p *profile.Profile, r *require.Assertions)
	}{
		{
			Name:     "can't set name",
			Property: "name",
			Value:    "test",
			Error:    "to update a profile name use hcp profile profiles rename",
		},
		{
			Name:     "invalid top-level key",
			Property: "unknown-top-level",
			Value:    "test",
			Error:    "property with name \"unknown-top-level\" does not exist",
		},
		{
			Name:     "invalid nested key",
			Property: "core/unknown-key",
			Value:    "test",
			Error:    "property with name \"core/unknown-key\" does not exist",
		},
		{
			Name:     "basic top-level property",
			Property: "organization_id",
			Value:    "123",
			CheckProfile: func(p *profile.Profile, r *require.Assertions) {
				r.Equal("123", p.OrganizationID)
			},
		},
		{
			Name:     "basic core property",
			Property: "core/no_color",
			Value:    "true",
			CheckProfile: func(p *profile.Profile, r *require.Assertions) {
				r.True(*p.Core.NoColor)
			},
		},
		{
			Name:     "basic core property - invalid type conversion",
			Property: "core/no_color",
			Value:    "bad-value",
			Error:    "cannot parse 'core.no_color' as bool",
		},
		{
			Name:     "basic core property - invalid value",
			Property: "core/output_format",
			Value:    "bad-value",
			Error:    "invalid output_format \"bad-value\". Must be one of:",
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			r := require.New(t)

			io := iostreams.Test()
			o := &SetOpts{
				IO:             io,
				Profile:        profile.TestProfile(t),
				ProjectService: nil,
				Property:       c.Property,
				Value:          c.Value,
				isAuthed:       func() (bool, error) { return true, nil },
			}

			err := setRun(o)
			if c.Error == "" {
				r.NoError(err)
				c.CheckProfile(o.Profile, r)
			} else {
				r.ErrorContains(err, c.Error)
			}
		})
	}
}

func TestSet_Project(t *testing.T) {
	t.Parallel()
	r := require.New(t)
	psvc := mock_project_service.NewMockClientService(t)
	io := iostreams.Test()
	l := profile.TestLoader(t)
	p := l.DefaultProfile()
	r.NoError(p.Write())
	o := &SetOpts{
		IO:             io,
		Profile:        p,
		ProjectService: psvc,
		Property:       "project_id",
	}

	setup := func(quiet, tty, authed bool, projectID string) {
		o.Value = projectID
		io.SetQuiet(quiet)
		io.InputTTY = tty
		io.ErrorTTY = tty
		io.Input.Reset()
		io.Error.Reset()
		io.Output.Reset()
		o.isAuthed = func() (bool, error) { return authed, nil }
	}

	checkProject := func(expected string) {
		loadedProfile, err := l.LoadProfile(p.Name)
		r.NoError(err)
		r.Equal(expected, loadedProfile.ProjectID)
	}

	// Run with quiet off, TTY's, authenticated, and return that the user has access to the project
	{
		setup(false, true, true, "123")
		psvc.EXPECT().ProjectServiceGet(mock.MatchedBy(func(getReq *project_service.ProjectServiceGetParams) bool {
			return getReq != nil && getReq.ID == o.Value
		}), mock.Anything).Once().Return(nil, nil)
		r.NoError(setRun(o))
		checkProject("123")
	}

	// Call again but return permission denied and accept the prompt
	{
		setup(false, true, true, "accept")
		psvc.EXPECT().ProjectServiceGet(mock.MatchedBy(func(getReq *project_service.ProjectServiceGetParams) bool {
			return getReq != nil && getReq.ID == o.Value
		}), mock.Anything).Once().Return(nil, project_service.NewProjectServiceGetDefault(http.StatusForbidden))

		// Since we are running with quiet disabled and tty's, expect a prompt. Answer yes
		r.NoError(io.Input.WriteByte('y'))

		// Check that we received a prompt request and save the profile
		r.NoError(setRun(o))
		r.Contains(io.Error.String(), "You do not appear to have access to project")
		r.Contains(io.Error.String(), "Are you sure you wish to set the")
		checkProject("accept")
	}

	// Call again but return permission denied and do not accept the prompt
	{
		setup(false, true, true, "no-accept")
		psvc.EXPECT().ProjectServiceGet(mock.MatchedBy(func(getReq *project_service.ProjectServiceGetParams) bool {
			return getReq != nil && getReq.ID == o.Value
		}), mock.Anything).Once().Return(nil, project_service.NewProjectServiceGetDefault(http.StatusForbidden))

		// Since we are running with quiet disabled and tty's, expect a prompt. Answer yes
		r.NoError(io.Input.WriteByte('n'))

		// Check that we received a prompt request and save the profile
		r.NoError(setRun(o))
		r.Contains(io.Error.String(), "You do not appear to have access to project")
		r.Contains(io.Error.String(), "Are you sure you wish to set the")
		checkProject("accept")
	}

	// Run again but with quiet
	{
		setup(true, true, true, "789")
		r.NoError(setRun(o))
		r.NotContains(io.Error.String(), "Are you sure you wish to set the")
		checkProject("789")
	}

	// Run again but with no quiet but no tty
	{
		setup(false, false, true, "012")
		r.NoError(setRun(o))
		r.NotContains(io.Error.String(), "Are you sure you wish to set the")
		checkProject("012")
	}

	// Run again but unauthenticated
	{
		setup(true, true, false, "789")
		r.NoError(setRun(o))
		r.NotContains(io.Error.String(), "Are you sure you wish to set the")
		checkProject("789")
	}
}

func TestSet_Organization(t *testing.T) {
	t.Parallel()
	r := require.New(t)
	osvc := mock_organization_service.NewMockClientService(t)
	io := iostreams.Test()
	l := profile.TestLoader(t)
	p := l.DefaultProfile()
	r.NoError(p.Write())
	o := &SetOpts{
		IO:                  io,
		Profile:             p,
		OrganizationService: osvc,
		Property:            "organization_id",
	}

	setup := func(quiet, tty, authed bool, orgID string) {
		o.Value = orgID
		io.SetQuiet(quiet)
		io.InputTTY = tty
		io.ErrorTTY = tty
		io.Input.Reset()
		io.Error.Reset()
		io.Output.Reset()
		o.isAuthed = func() (bool, error) { return authed, nil }
	}

	orgResp := func(ids ...string) *organization_service.OrganizationServiceListOK {
		resp := &organization_service.OrganizationServiceListOK{
			Payload: &models.HashicorpCloudResourcemanagerOrganizationListResponse{
				Organizations: []*models.HashicorpCloudResourcemanagerOrganization{},
			},
		}

		for _, id := range ids {
			resp.Payload.Organizations = append(resp.Payload.Organizations,
				&models.HashicorpCloudResourcemanagerOrganization{ID: id})
		}

		return resp
	}

	checkOrg := func(expected string) {
		loadedProfile, err := l.LoadProfile(p.Name)
		r.NoError(err)
		r.Equal(expected, loadedProfile.OrganizationID)
	}

	// Run with quiet off, TTY's, authenticated, and return the user is a member of the org
	{
		setup(false, true, true, "member")
		osvc.EXPECT().OrganizationServiceList(mock.Anything, mock.Anything).Once().Return(orgResp("1", "2", "member"), nil)
		r.NoError(setRun(o))
		checkOrg("member")
	}

	// Not be a member, expect prompting and respond no
	{
		setup(false, true, true, "not-a-member")
		osvc.EXPECT().OrganizationServiceList(mock.Anything, mock.Anything).Once().Return(orgResp("1", "2", "123"), nil)

		// Answer yes to prompt
		r.NoError(io.Input.WriteByte('n'))

		// Check that we received a prompt request and save the profile
		r.NoError(setRun(o))
		r.Contains(io.Error.String(), "You do not appear to be a member of organization")
		r.Contains(io.Error.String(), "Are you sure you wish to set the")
		checkOrg("member")
	}

	// Not be a member, expect prompting and respond yes
	{
		setup(false, true, true, "not-a-member")
		osvc.EXPECT().OrganizationServiceList(mock.Anything, mock.Anything).Once().Return(orgResp("1", "2", "123"), nil)

		// Answer yes to prompt
		r.NoError(io.Input.WriteByte('y'))

		// Check that we received a prompt request and save the profile
		r.NoError(setRun(o))
		r.Contains(io.Error.String(), "You do not appear to be a member of organization")
		r.Contains(io.Error.String(), "Are you sure you wish to set the")
		checkOrg("not-a-member")
	}

	// Do not be a member; but quiet
	{
		setup(true, true, true, "not-a-member-quiet")
		r.NoError(setRun(o))
		checkOrg("not-a-member-quiet")
	}

	// Do not be a member; but no tty
	{
		setup(false, false, true, "not-a-member-no-tty")
		r.NoError(setRun(o))
		checkOrg("not-a-member-no-tty")
	}

	// Do not be a member; but unauthenticated
	{
		setup(false, false, false, "not-a-member-no-tty")
		r.NoError(setRun(o))
		checkOrg("not-a-member-no-tty")
	}

	// Return error
	{
		setup(false, true, true, "error")
		osvc.EXPECT().OrganizationServiceList(mock.Anything, mock.Anything).Once().
			Return(nil, organization_service.NewOrganizationServiceListDefault(http.StatusInternalServerError))

		// Check that we received a prompt request and save the profile
		r.ErrorContains(setRun(o), "failed to list organizations for principal")
		checkOrg("not-a-member-no-tty")
	}
}
