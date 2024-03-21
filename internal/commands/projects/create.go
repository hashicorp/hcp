package projects

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-billing/preview/2020-11-05/client/billing_account_service"
	billingModels "github.com/hashicorp/hcp-sdk-go/clients/cloud-billing/preview/2020-11-05/models"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/project_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/models"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/flagvalue"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
)

const (
	// defaultBillingAccountID is the ID of the default/only billing account.
	defaultBillingAccountID = "default-account"
)

func NewCmdCreate(ctx *cmd.Context, runF func(*CreateOpts) error) *cmd.Command {
	opts := &CreateOpts{
		Ctx:           ctx.ShutdownCtx,
		Profile:       ctx.Profile,
		IO:            ctx.IO,
		Output:        ctx.Output,
		ProjectClient: project_service.New(ctx.HCP, nil),
		BillingClient: billing_account_service.New(ctx.HCP, nil),
	}

	cmd := &cmd.Command{
		Name:      "create",
		ShortHelp: "Create a new project.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp projects create" }} command creates
		a new project with the given name. The currently authenticated principal
		will be given role "admin" on the newly created project.
		`),
		Examples: []cmd.Example{
			{
				Preamble: "Creating a project with a description:",
				Command:  "$ hcp projects create example-project --description=\"my test project\"",
			},
		},
		Args: cmd.PositionalArguments{
			Args: []cmd.PositionalArgument{
				{
					Name:          "NAME",
					Documentation: "Name of the project to create.",
				},
			},
		},
		Flags: cmd.Flags{
			Local: []*cmd.Flag{
				{
					Name:         "description",
					DisplayValue: "DESCRIPTION",
					Description:  "An optional description for the project.",
					Value:        flagvalue.Simple("", &opts.Description),
				},
				{
					Name:          "set-as-default",
					Description:   "Set the newly created project as the default project in the active profile.",
					Value:         flagvalue.Simple(false, &opts.Default),
					IsBooleanFlag: true,
				},
			},
		},
		RunF: func(c *cmd.Command, args []string) error {
			opts.Name = args[0]
			if runF != nil {
				return runF(opts)
			}

			return createRun(opts)
		},
		PersistentPreRun: func(c *cmd.Command, args []string) error {
			return cmd.RequireOrganization(ctx)
		},
	}

	return cmd
}

type CreateOpts struct {
	Ctx         context.Context
	Profile     *profile.Profile
	IO          iostreams.IOStreams
	Output      *format.Outputter
	Name        string
	Description string
	Default     bool

	ProjectClient project_service.ClientService
	BillingClient billing_account_service.ClientService
}

func createRun(opts *CreateOpts) error {
	req := project_service.NewProjectServiceCreateParamsWithContext(opts.Ctx)
	req.Body = &models.HashicorpCloudResourcemanagerProjectCreateRequest{
		Description: opts.Description,
		Name:        opts.Name,
		Parent: &models.HashicorpCloudResourcemanagerResourceID{
			ID:   opts.Profile.OrganizationID,
			Type: models.HashicorpCloudResourcemanagerResourceIDResourceTypeORGANIZATION.Pointer(),
		},
	}

	resp, err := opts.ProjectClient.ProjectServiceCreate(req, nil)
	if err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}

	// Add to billing account
	if err := addToBillingAccount(opts, resp.Payload.Project.ID); err != nil {
		return err
	}

	// Display the created project
	d := format.NewDisplayer(resp.Payload.Project, format.Pretty, projectFields)
	if err := opts.Output.Display(d); err != nil {
		return err
	}

	if !opts.Default {
		return nil
	}

	opts.Profile.ProjectID = resp.Payload.Project.ID
	if err := opts.Profile.Write(); err != nil {
		return fmt.Errorf("failed to mark newly created project as the profile's default project: %w", err)
	}

	fmt.Fprintf(opts.IO.Err(), "\n%s Project %q set as default project in active profile.\n",
		opts.IO.ColorScheme().SuccessIcon(), resp.Payload.Project.ID)
	return nil
}

func addToBillingAccount(opts *CreateOpts, projectID string) error {
	req := billing_account_service.NewBillingAccountServiceGetParams()
	req.OrganizationID = opts.Profile.OrganizationID
	req.ID = defaultBillingAccountID
	resp, err := opts.BillingClient.BillingAccountServiceGet(req, nil)
	if err != nil {
		return fmt.Errorf("failed listing billing accounts: %w", err)
	}

	// Update the BA to include the new project ID
	ba := resp.Payload.BillingAccount
	updateReq := billing_account_service.NewBillingAccountServiceUpdateParams()
	updateReq.OrganizationID = ba.OrganizationID
	updateReq.ID = ba.ID
	updateReq.Body = &billingModels.Billing20201105UpdateBillingAccountRequest{
		OrganizationID: opts.Profile.OrganizationID,
		ID:             ba.ID,
		ProjectIds:     ba.ProjectIds,
		Name:           ba.Name,
		Country:        ba.Country,
	}

	updateReq.Body.ProjectIds = append(updateReq.Body.ProjectIds, projectID)

	_, err = opts.BillingClient.BillingAccountServiceUpdate(updateReq, nil)
	if err != nil {
		return fmt.Errorf("failed updating billing account: %w", err)
	}

	return nil
}
