// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package profile

import (
	"fmt"
	"math/rand"
	"net/http"
	"testing"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/client/iam_service"
	iam_models "github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/models"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/organization_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/project_service"
	rm_models "github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/models"
	preview_secret_service "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/preview/2023-11-28/client/secret_service"
	preview_secret_models "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/preview/2023-11-28/models"
	mock_iam_service "github.com/hashicorp/hcp/internal/pkg/api/mocks/github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/client/iam_service"
	mock_organization_service "github.com/hashicorp/hcp/internal/pkg/api/mocks/github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/organization_service"
	mock_project_service "github.com/hashicorp/hcp/internal/pkg/api/mocks/github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/project_service"
	mock_preview_secret_secvice "github.com/hashicorp/hcp/internal/pkg/api/mocks/github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/preview/2023-11-28/client/secret_service"

	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
	"github.com/hashicorp/hcp/internal/pkg/testing/promptio"
	"github.com/manifoldco/promptui"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type initMocks struct {
	IAMClient           *mock_iam_service.MockClientService
	OrganizationService *mock_organization_service.MockClientService
	ProjectService      *mock_project_service.MockClientService
	SecretSevice        *mock_preview_secret_secvice.MockClientService
}

func getInitMocks(t *testing.T, opts *InitOpts) initMocks {
	m := initMocks{
		IAMClient:           mock_iam_service.NewMockClientService(t),
		OrganizationService: mock_organization_service.NewMockClientService(t),
		ProjectService:      mock_project_service.NewMockClientService(t),
		SecretSevice:        mock_preview_secret_secvice.NewMockClientService(t),
	}

	if opts != nil {
		opts.IAMClient = m.IAMClient
		opts.OrganizationService = m.OrganizationService
		opts.ProjectService = m.ProjectService
		opts.SecretService = m.SecretSevice
	}

	return m
}

func (m *initMocks) callerIdentitySP(orgID, projID string) {
	resp := iam_service.IamServiceGetCallerIdentityOK{
		Payload: &iam_models.HashicorpCloudIamGetCallerIdentityResponse{
			Principal: &iam_models.HashicorpCloudIamPrincipal{
				Service: &iam_models.HashicorpCloudIamServicePrincipal{
					OrganizationID: orgID,
					ProjectID:      projID,
				},
			},
		},
	}

	m.IAMClient.EXPECT().IamServiceGetCallerIdentity(mock.Anything, mock.Anything).Return(&resp, nil)
}

func (m *initMocks) callerIdentityUser() {
	resp := iam_service.IamServiceGetCallerIdentityOK{
		Payload: &iam_models.HashicorpCloudIamGetCallerIdentityResponse{
			Principal: &iam_models.HashicorpCloudIamPrincipal{
				User: &iam_models.HashicorpCloudIamUserPrincipal{
					Email:    "test@test.com",
					FullName: "Unit Test",
					ID:       "test",
					Subject:  "test",
				},
			},
		},
	}

	m.IAMClient.EXPECT().IamServiceGetCallerIdentity(mock.Anything, mock.Anything).Return(&resp, nil)
}

func (m *initMocks) projectListErr(code int) {
	err := project_service.NewProjectServiceListDefault(code)
	m.ProjectService.EXPECT().ProjectServiceList(mock.Anything, mock.Anything).Return(nil, err)
}

func (m *initMocks) projectList(count int) []*rm_models.HashicorpCloudResourcemanagerProject {
	ok := project_service.NewProjectServiceListOK()
	ok.Payload = &rm_models.HashicorpCloudResourcemanagerProjectListResponse{
		Projects: []*rm_models.HashicorpCloudResourcemanagerProject{},
	}

	for i := 0; i < count; i++ {
		ok.Payload.Projects = append(ok.Payload.Projects,
			&rm_models.HashicorpCloudResourcemanagerProject{
				Description: fmt.Sprintf("description %d", i),
				ID:          fmt.Sprintf("id_%d", i),
				Name:        fmt.Sprintf("name_%d", i),
			})
	}

	m.ProjectService.EXPECT().ProjectServiceList(mock.Anything, mock.Anything).Return(ok, nil)
	return ok.Payload.Projects
}

func (m *initMocks) orgList(count int) []*rm_models.HashicorpCloudResourcemanagerOrganization {
	ok := organization_service.NewOrganizationServiceListOK()
	ok.Payload = &rm_models.HashicorpCloudResourcemanagerOrganizationListResponse{
		Organizations: []*rm_models.HashicorpCloudResourcemanagerOrganization{},
	}

	for i := 0; i < count; i++ {
		ok.Payload.Organizations = append(ok.Payload.Organizations,
			&rm_models.HashicorpCloudResourcemanagerOrganization{
				ID:   fmt.Sprintf("id_%d", i),
				Name: fmt.Sprintf("name_%d", i),
			})
	}

	m.OrganizationService.EXPECT().OrganizationServiceList(mock.Anything, mock.Anything).Return(ok, nil)
	return ok.Payload.Organizations
}

func (m *initMocks) vaultSecretsAppsList(count int) []*preview_secret_models.Secrets20231128App {
	ok := preview_secret_service.NewListAppsOK()
	ok.Payload = &preview_secret_models.Secrets20231128ListAppsResponse{
		Apps: []*preview_secret_models.Secrets20231128App{},
	}

	for i := 0; i < count; i++ {
		ok.Payload.Apps = append(ok.Payload.Apps,
			&preview_secret_models.Secrets20231128App{
				Name: fmt.Sprintf("app_name_%d", i),
			})
	}

	m.SecretSevice.EXPECT().ListApps(mock.Anything, mock.Anything).Return(ok, nil)
	return ok.Payload.Apps
}

func TestInit_OrgAndProject_SP_NoList(t *testing.T) {
	t.Parallel()
	r := require.New(t)

	io := iostreams.Test()
	io.ErrorTTY = true
	io.InputTTY = true

	opts := InitOpts{
		IO:      io,
		Profile: profile.TestProfile(t),
	}
	mocks := getInitMocks(t, &opts)

	// Expect a call to GetCallerIdentity
	orgID, projID := "123", "456"
	mocks.callerIdentitySP(orgID, projID)

	// Fail the list projects call with permission denied
	mocks.projectListErr(http.StatusForbidden)

	// Say no to service config prompt
	_, err := io.Input.WriteRune('n')
	r.NoError(err)

	_, err = io.Input.WriteRune(promptui.KeyEnter)
	r.NoError(err)

	r.NoError(opts.run())
	r.Equal(orgID, opts.Profile.OrganizationID)
	r.Equal(projID, opts.Profile.ProjectID)
}

func TestInit_OrgAndProject_SP_List(t *testing.T) {
	t.Parallel()
	r := require.New(t)

	ioTest := iostreams.Test()
	ioTest.ErrorTTY = true
	ioTest.InputTTY = true

	io := promptio.Wrap(ioTest)
	opts := InitOpts{
		IO:      io,
		Profile: profile.TestProfile(t),
	}
	mocks := getInitMocks(t, &opts)

	// Expect a call to GetCallerIdentity
	orgID, projID := "123", "456"
	mocks.callerIdentitySP(orgID, projID)

	// Fail the list projects call with permission denied
	projects := mocks.projectList(5)

	// Send a down character and enter
	_, err := io.Input.WriteRune(promptui.KeyNext)
	r.NoError(err)
	_, err = io.Input.WriteRune(promptui.KeyEnter)
	r.NoError(err)

	// Say no to service config prompt
	_, err = io.Input.WriteRune('n')
	r.NoError(err)

	_, err = io.Input.WriteRune(promptui.KeyEnter)
	r.NoError(err)

	r.NoError(opts.run())
	r.Equal(orgID, opts.Profile.OrganizationID)
	r.Equal(projects[1].ID, opts.Profile.ProjectID)
}

func TestInit_OrgAndProject_User(t *testing.T) {
	t.Parallel()
	cases := []struct {
		Name                  string
		NumOrgs               int
		NumProjects           int
		NumVaultSecretsApps   int
		ConfigureVaultSecrets bool

		Error string
	}{
		{
			Name:    "No org",
			NumOrgs: 0,
			Error:   "there are no valid organizations for your principal.",
		},
		{
			Name:        "One org / No Project",
			NumOrgs:     1,
			NumProjects: 0,
			Error:       "there are no valid projects for your principal",
		},
		{
			Name:        "One org / One Project",
			NumOrgs:     1,
			NumProjects: 1,
		},
		{
			Name:        "One org / Many Project",
			NumOrgs:     1,
			NumProjects: 100,
		},
		{
			Name:        "Many org / Many Project",
			NumOrgs:     10,
			NumProjects: 10,
		},
		{
			Name:                  "Success: Many org / Many Project / No Vault Secets App/ Do not configure service config",
			NumOrgs:               10,
			NumProjects:           10,
			NumVaultSecretsApps:   0,
			ConfigureVaultSecrets: false,
		},
		{
			Name:                  "Success: Many org / Many Project / No Vault Secrets App / Configure Vault Secrets service",
			NumOrgs:               10,
			NumProjects:           10,
			NumVaultSecretsApps:   0,
			ConfigureVaultSecrets: true,
		},
		{
			Name:                  "Success: Many org / Many Project / 1 Vault Secrets App / Configure Vault Secrets service",
			NumOrgs:               10,
			NumProjects:           10,
			NumVaultSecretsApps:   1,
			ConfigureVaultSecrets: true,
		},
		{
			Name:                  "Success: Many org / Many Project / Many Vault Secrets App / Configure Vault Secrets service",
			NumOrgs:               10,
			NumProjects:           10,
			NumVaultSecretsApps:   100,
			ConfigureVaultSecrets: true,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			r := require.New(t)

			ioTest := iostreams.Test()
			ioTest.ErrorTTY = true
			ioTest.InputTTY = true

			io := promptio.Wrap(ioTest)
			opts := InitOpts{
				IO:      io,
				Profile: profile.TestProfile(t),
			}
			mocks := getInitMocks(t, &opts)

			// Expect a call to GetCallerIdentity
			mocks.callerIdentityUser()

			// Capture the selected IDs
			selectedOrgID, selectedProjID, selectedVaultSecretsAppName := "", "", ""

			// Return the expected number of orgs
			orgs := mocks.orgList(c.NumOrgs)

			// We have to interact with the prompter
			if c.NumOrgs > 1 {
				selection := rand.Intn(c.NumOrgs)
				selectedOrgID = orgs[selection].ID

				// Send a down character and enter
				for i := 0; i < selection; i++ {
					_, err := io.Input.WriteRune(promptui.KeyNext)
					r.NoError(err)
				}

				// Select
				_, err := io.Input.WriteRune(promptui.KeyEnter)
				r.NoError(err)

			} else if c.NumOrgs == 1 {
				selectedOrgID = orgs[0].ID
			}

			// Return the expected number of projects
			var projects []*rm_models.HashicorpCloudResourcemanagerProject
			if c.NumOrgs != 0 {
				projects = mocks.projectList(c.NumProjects)
			}

			// We have to interact with the prompter
			if c.NumProjects > 1 {
				selection := rand.Intn(c.NumProjects)
				selectedProjID = projects[selection].ID

				// Send a down character and enter
				for i := 0; i < selection; i++ {
					_, err := io.Input.WriteRune(promptui.KeyNext)
					r.NoError(err)
				}

				// Select
				_, err := io.Input.WriteRune(promptui.KeyEnter)
				r.NoError(err)
			} else if c.NumProjects == 1 {
				selectedProjID = projects[0].ID
			}

			if c.ConfigureVaultSecrets {
				// Say yes to configuring service config
				_, err := io.Input.WriteRune('y')
				r.NoError(err)

				// Send a down character and enter
				_, err = io.Input.WriteRune(promptui.KeyNext)
				r.NoError(err)

				// Select Vault-Secrets
				_, err = io.Input.WriteRune(promptui.KeyEnter)
				r.NoError(err)
			} else {
				// Say no to configuring service config
				_, err := io.Input.WriteRune('n')
				r.NoError(err)

				_, err = io.Input.WriteRune(promptui.KeyEnter)
				r.NoError(err)
			}

			var vaultSecretsApps []*preview_secret_models.Secrets20231128App
			if c.ConfigureVaultSecrets {
				vaultSecretsApps = mocks.vaultSecretsAppsList(c.NumVaultSecretsApps)
			}

			if c.ConfigureVaultSecrets {
				if c.NumVaultSecretsApps > 1 {
					selection := rand.Intn(c.NumVaultSecretsApps)
					selectedVaultSecretsAppName = vaultSecretsApps[selection].Name

					// Send a down character and enter
					for i := 0; i < selection; i++ {
						_, err := io.Input.WriteRune(promptui.KeyNext)
						r.NoError(err)
					}

					// Select
					_, err := io.Input.WriteRune(promptui.KeyEnter)
					r.NoError(err)

				} else if c.NumVaultSecretsApps == 1 {
					selectedVaultSecretsAppName = vaultSecretsApps[0].Name
				}
			}

			err := opts.run()
			if c.Error != "" {
				r.ErrorContains(err, c.Error)
			} else {
				r.NoError(err)
				r.Equal(selectedOrgID, opts.Profile.OrganizationID)
				r.Equal(selectedProjID, opts.Profile.ProjectID)
				if opts.Profile.VaultSecrets != nil {
					r.Equal(selectedVaultSecretsAppName, opts.Profile.VaultSecrets.AppName)
				}
			}
		})
	}
}
