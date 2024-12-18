// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package profile

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/client/iam_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/organization_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/project_service"
	resources "github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/models"
	preview_secret_service "github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-11-28/client/secret_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-11-28/models"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/flagvalue"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
	"github.com/manifoldco/promptui"
)

const (
	serviceNameVaultSecrets = "Vault Secrets"
)

func NewCmdInit(ctx *cmd.Context) *cmd.Command {
	opts := &InitOpts{
		Ctx:                 ctx.ShutdownCtx,
		IO:                  ctx.IO,
		Profile:             ctx.Profile,
		IAMClient:           iam_service.New(ctx.HCP, nil),
		OrganizationService: organization_service.New(ctx.HCP, nil),
		ProjectService:      project_service.New(ctx.HCP, nil),
		SecretService:       preview_secret_service.New(ctx.HCP, nil),
	}
	cmd := &cmd.Command{
		Name:      "init",
		ShortHelp: "Initialize the current profile.",
		LongHelp: heredoc.New(ctx.IO).Mustf(`
		The {{ template "mdCodeOrBold" "hcp profile init" }} command configures
		the HCP CLI to run commands against the correct context; namely against
		the desired organization and project ID. This command is interactive. To
		set configuration using non-interactively prefer using {{ template "mdCodeOrBold" "hcp profile set" }}.

		For a list of all available options, run {{ template "mdCodeOrBold" "hcp config --help" }}.
		`),
		RunF: func(c *cmd.Command, args []string) error {
			if !ctx.IO.CanPrompt() {
				return fmt.Errorf("command cannot run with disabled prompts")
			}

			return opts.run()
		},
		Flags: cmd.Flags{
			Local: []*cmd.Flag{
				{
					Name:          "vault-secrets",
					Description:   "Initializes Vault Secrets configuration.",
					IsBooleanFlag: true,
					Value:         flagvalue.Simple(false, &opts.VaultSecrets),
				},
			},
		},
	}
	return cmd
}

type InitOpts struct {
	Ctx     context.Context
	IO      iostreams.IOStreams
	Profile *profile.Profile

	// Clients
	IAMClient           iam_service.ClientService
	OrganizationService organization_service.ClientService
	ProjectService      project_service.ClientService
	SecretService       preview_secret_service.ClientService

	// Flags
	VaultSecrets bool
}

func (i *InitOpts) run() error {
	if err := i.configureOrgAndProject(); err != nil {
		return fmt.Errorf("failed configuring organization and project: %w", err)
	}

	// If the user hasn't explicitly specified a service to configure, prompt to see if they would like to.
	if err := i.serviceConfigPrompt(); err != nil {
		return err
	}

	if i.VaultSecrets {
		if err := i.configureVaultSecrets(); err != nil {
			return fmt.Errorf("failed configuring profile for Vault Secrets: %w", err)
		}
	}

	return i.Profile.Write()
}

// serviceConfigPrompt prompts the user to configure a service.
func (i *InitOpts) serviceConfigPrompt() error {
	if i.VaultSecrets {
		return nil
	}

	ok, err := i.IO.PromptConfirm("\nWould you like to configure any service related config")
	if err != nil {
		return fmt.Errorf("failed to retrieve confirmation: %w", err)
	}

	if !ok {
		return nil
	}

	prompt := promptui.Select{
		Label:  "Please select the service you would like to configure",
		Items:  []string{serviceNameVaultSecrets},
		Stdin:  io.NopCloser(i.IO.In()),
		Stdout: iostreams.NopWriteCloser(i.IO.Err()),
	}

	_, result, err := prompt.Run()
	if err != nil {
		return fmt.Errorf("service selection prompt failed: %w", err)
	}

	if result == serviceNameVaultSecrets {
		i.VaultSecrets = true
	}
	return nil
}

func (i *InitOpts) configureOrgAndProject() error {
	// Retrieve whether the authenticated principal is bound to a single
	// organization or has a default project binding.
	org, proj, err := i.getCallerIdentity()
	if err != nil {
		return err
	}

	// If the principal is a service principal, it will only ever have access to
	// a single organization. If we detect this do not prompt for organization.
	if org != "" {
		i.Profile.OrganizationID = org
	} else {
		org, err := i.gatherOrganizationID()
		if err != nil {
			return err
		}

		i.Profile.OrganizationID = org
	}

	projectID, err := i.gatherProjectID(proj)
	if err != nil {
		return err
	}

	i.Profile.ProjectID = projectID

	return nil
}

func (i *InitOpts) configureVaultSecrets() error {
	params := &preview_secret_service.ListAppsParams{
		Context:        i.Ctx,
		ProjectID:      i.Profile.ProjectID,
		OrganizationID: i.Profile.OrganizationID,
	}

	var apps []*models.Secrets20231128App
	for {
		resp, err := i.SecretService.ListApps(params, nil)

		if err != nil {
			return fmt.Errorf("failed to list apps: %w", err)
		}

		apps = append(apps, resp.Payload.Apps...)
		if resp.Payload.Pagination == nil || resp.Payload.Pagination.NextPageToken == "" {
			break
		}

		next := resp.Payload.Pagination.NextPageToken
		params.PaginationNextPageToken = &next
	}

	appCount := len(apps)
	if appCount <= 0 {
		appsCreateDoc := heredoc.New(i.IO, heredoc.WithPreserveNewlines()).Must(`
No Vault Secrets application found. Create one and set on the active profile by issuing:

	$ hcp vault-secrets apps create test-app --description="Test app"
	$ hcp profile set vault-secrets/app_name test-app
`)
		fmt.Fprintf(i.IO.Err(), "\n%s\n\n", appsCreateDoc)
		return nil
	}

	appName := apps[0].Name
	if appCount > 1 {
		prompt := promptui.Select{
			Label: "Multiple apps found. Please select the one you would like to configure.",
			Items: apps,
			Templates: &promptui.SelectTemplates{
				Active:   `> {{ .Name }}`,
				Inactive: `{{ .Name }}`,
				Details: `
----- Apps -----
{{ "Name:" | faint }}   {{ .Name }}
{{ "Description" | faint }} {{ .Description }}
`,
			},
			HideSelected: true,
			Size:         15,
			Stdin:        io.NopCloser(i.IO.In()),
			Stdout:       iostreams.NopWriteCloser(i.IO.Err()),
			Searcher: func(term string, index int) bool {
				term = strings.ToLower(term)
				name := strings.ToLower(apps[index].Name)
				return strings.Contains(name, term)
			},
		}

		i, _, err := prompt.Run()
		if err != nil {
			return fmt.Errorf("prompt failed: %w", err)
		}
		appName = apps[i].Name
	}

	cs := i.IO.ColorScheme()
	fmt.Fprintf(i.IO.Err(), "%s App with name %q selected\n", cs.SuccessIcon(), appName)
	i.Profile.VaultSecrets = &profile.VaultSecretsConf{
		AppName: appName,
	}
	return nil
}

func (i *InitOpts) gatherOrganizationID() (string, error) {
	req := organization_service.NewOrganizationServiceListParamsWithContext(i.Ctx)
	orgs, err := i.OrganizationService.OrganizationServiceList(req, nil)
	if err != nil {
		return "", err
	}

	orgsCount := len(orgs.Payload.Organizations)
	if orgsCount <= 0 {
		return "", fmt.Errorf("there are no valid organizations for your principal. Create one by visiting the HCP Portal (https://portal.cloud.hashicorp.com)")
	}

	orgID := orgs.Payload.Organizations[0].ID
	orgName := orgs.Payload.Organizations[0].Name
	if orgsCount > 1 {
		prompt := promptui.Select{
			Label: "Multiple organizations found. Please select the one you would like to configure.",
			Items: orgs.Payload.Organizations,
			Templates: &promptui.SelectTemplates{
				Active:   `> {{ .Name }}`,
				Inactive: `{{ .Name }}`,
				Details: `
----- Organization -----
{{ "Name:" | faint }}   {{ .Name }}
{{ "ID:" | faint }}     {{ .ID }}
`,
			},
			HideSelected: true,
			Stdin:        io.NopCloser(i.IO.In()),
			Stdout:       iostreams.NopWriteCloser(i.IO.Err()),
			Searcher: func(term string, index int) bool {
				term = strings.ToLower(term)
				id := strings.ToLower(orgs.Payload.Organizations[index].ID)
				name := strings.ToLower(orgs.Payload.Organizations[index].Name)
				if strings.Contains(id, term) {
					return true
				} else if strings.Contains(name, term) {
					return true
				}

				return false
			},
		}

		i, _, err := prompt.Run()
		if err != nil {
			return "", err
		}

		orgID = orgs.Payload.Organizations[i].ID
		orgName = orgs.Payload.Organizations[i].Name
	}

	cs := i.IO.ColorScheme()
	fmt.Fprintf(i.IO.Err(), "%s Organization with name %q and ID %q selected\n", cs.SuccessIcon(), orgName, orgID)
	return orgID, nil
}

func (i *InitOpts) gatherProjectID(detectedProject string) (string, error) {
	scopeType := string(resources.HashicorpCloudResourcemanagerResourceIDResourceTypeORGANIZATION)
	params := &project_service.ProjectServiceListParams{
		ScopeID:   &i.Profile.OrganizationID,
		ScopeType: &scopeType,
		Context:   i.Ctx,
	}

	// TODO Switch to LookupAccessibleResources and use pagination. When
	// available, we can autocomplete set project_id as well.
	projects, err := i.ProjectService.ProjectServiceList(params, nil)
	if err != nil {
		var listErr *project_service.ProjectServiceListDefault
		if errors.As(err, &listErr) && listErr.IsCode(http.StatusForbidden) && detectedProject != "" {
			fmt.Fprintln(i.IO.Err(), heredoc.New(i.IO).Mustf(`
{{ Color "yellow" "Principal does not have permission to list projects." }}

Using the project the principal was created in:
%s`, detectedProject))
			return detectedProject, nil
		}

		return "", fmt.Errorf("unable to list projects the current principal has access to: \n\n%w", err)
	}

	projCount := len(projects.Payload.Projects)
	if projCount <= 0 {
		return "", fmt.Errorf("there are no valid projects for your principal")
	}

	projectID := projects.Payload.Projects[0].ID
	projectName := projects.Payload.Projects[0].Name
	if projCount > 1 {
		cursor := 0
		if detectedProject != "" {
			for i, p := range projects.Payload.Projects {
				if p.ID == detectedProject {
					cursor = i
					break
				}
			}
		}

		prompt := promptui.Select{
			Label: "Multiple projects found. Please select the one you would like to configure.",
			Items: projects.Payload.Projects,
			Templates: &promptui.SelectTemplates{
				Active:   `> {{ .Name }}`,
				Inactive: `{{ .Name }}`,
				Details: `
----- Project -----
{{ "Name:" | faint }}        {{ .Name }}
{{ "ID:" | faint }}          {{ .ID }}
{{ "Description:" | faint }} {{ .Description }}
`,
			},
			HideSelected: true,
			CursorPos:    cursor,
			Size:         15,
			Stdin:        io.NopCloser(i.IO.In()),
			Stdout:       iostreams.NopWriteCloser(i.IO.Err()),
			Searcher: func(term string, index int) bool {
				term = strings.ToLower(term)
				id := strings.ToLower(projects.Payload.Projects[index].ID)
				name := strings.ToLower(projects.Payload.Projects[index].Name)
				if strings.Contains(id, term) {
					return true
				} else if strings.Contains(name, term) {
					return true
				}

				return false
			},
		}

		i, _, err := prompt.Run()
		if err != nil {
			return "", fmt.Errorf("interactive selection failed: %w", err)
		}

		projectID = projects.Payload.Projects[i].ID
		projectName = projects.Payload.Projects[i].Name
	}

	cs := i.IO.ColorScheme()
	fmt.Fprintf(i.IO.Err(), "%s Project with name %q and ID %q selected\n", cs.SuccessIcon(), projectName, projectID)
	return projectID, nil
}

func (i *InitOpts) getCallerIdentity() (string, string, error) {
	callerIdentityParams := iam_service.NewIamServiceGetCallerIdentityParamsWithContext(i.Ctx)
	ident, err := i.IAMClient.IamServiceGetCallerIdentity(callerIdentityParams, nil)
	if err != nil {
		return "", "", err
	}

	// basically everything in the response could be nil
	if ident.Payload == nil || ident.Payload.Principal == nil {
		return "", "", nil
	}

	if ident.Payload.Principal.Service != nil {
		return ident.Payload.Principal.Service.OrganizationID, ident.Payload.Principal.Service.ProjectID, nil
	}

	return "", "", nil
}
