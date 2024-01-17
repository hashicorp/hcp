package keys

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/hashicorp/hcp-sdk-go/auth"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/client/service_principals_service"
	"github.com/hashicorp/hcp/internal/commands/iam/serviceprincipals/helper"
	"github.com/hashicorp/hcp/internal/pkg/api/resourcename"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/flagvalue"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
)

func NewCmdCreate(ctx *cmd.Context, runF func(*CreateOpts) error) *cmd.Command {
	opts := &CreateOpts{
		Ctx:     ctx.ShutdownCtx,
		Profile: ctx.Profile,
		IO:      ctx.IO,
		Output:  ctx.Output,
		Client:  service_principals_service.New(ctx.HCP, nil),
	}

	cmd := &cmd.Command{
		Name:      "create",
		ShortHelp: "Create a new service principal key.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ Bold "hcp iam service-principals keys create" }} command creates a new service principal key.

		To output the generated keys to a credential file, pass the --output-cred-file flag. The credential file can be used
		to authenticate as the service principal. The benefit of using the credential file is that it avoids printing the
		Client ID and Client Secret to the terminal, and allows the credentials to be stored in a way that is less likely
		to leak into shell history. The HCP CLI allows authenticating via credential files using {{ Bold "hcp auth login --cred-file=PATH" }}.
		Prefer using credential files if your workflow allows it.

		To create a key for an organization service principal, pass the service principal's resource name or set the --project
		flag to "-" and pass its resource name suffix.
		`),
		Examples: []cmd.Example{
			{
				Preamble: `Create a new service principal key`,
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				$ hcp iam service-principals keys create my-service-principal
				`),
			},
			{
				Preamble: `Create a new service principal key specifying the resource name of the service principal`,
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				$ hcp iam service-principals keys create \
				  iam/project/123/service-principal/my-service-principal
				`),
			},
			{
				Preamble: `Output the new service principal key to a credential file`,
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				$ hcp iam service-principals keys create my-service-principal \
				  --output-cred-file=my-service-principal-creds.json
				`),
			},
		},
		Args: cmd.PositionalArguments{
			Args: []cmd.PositionalArgument{
				{
					Name:          "SP_NAME",
					Documentation: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Mustf(helper.SPNameArgDoc, "create a key for"),
				},
			},
		},
		Flags: cmd.Flags{
			Local: []*cmd.Flag{
				{
					Name:         "output-cred-file",
					DisplayValue: "PATH",
					Description:  "Output the created service principal key to a credential file. The file type must be json.",
					Value:        flagvalue.Simple(nil, &opts.CredentialFilePath),
				},
			},
		},
		RunF: func(c *cmd.Command, args []string) error {
			opts.Name = args[0]

			if opts.CredentialFilePath != nil && filepath.Ext(*opts.CredentialFilePath) != ".json" {
				return fmt.Errorf("credential file must be a json file")
			}

			if runF != nil {
				return runF(opts)
			}
			return createRun(opts)
		},
		PersistentPreRun: func(c *cmd.Command, args []string) error {
			return cmd.RequireOrganization(ctx)
		},
	}

	// Setup the autocomplete for the name argument
	cmd.Args.Autocomplete = helper.PredictSPResourceNameSuffix(ctx, cmd, opts.Client)

	return cmd
}

type CreateOpts struct {
	Ctx     context.Context
	IO      iostreams.IOStreams
	Profile *profile.Profile
	Output  *format.Outputter

	Name               string
	CredentialFilePath *string
	Client             service_principals_service.ClientService
}

func createRun(opts *CreateOpts) error {
	req := service_principals_service.NewServicePrincipalsServiceCreateServicePrincipalKeyParamsWithContext(opts.Ctx)
	req.ParentResourceName = helper.ResourceName(opts.Name, opts.Profile.OrganizationID, opts.Profile.ProjectID)

	resp, err := opts.Client.ServicePrincipalsServiceCreateServicePrincipalKey(req, nil)
	if err != nil {
		return fmt.Errorf("failed to create service principal key: %w", err)
	}

	// Create a credential file if requested
	if opts.CredentialFilePath != nil {
		cf := &auth.CredentialFile{
			Scheme: auth.CredentialFileSchemeServicePrincipal,
			Oauth: &auth.OauthConfig{
				ClientID:     resp.Payload.Key.ClientID,
				ClientSecret: resp.Payload.ClientSecret,
			},
		}

		if projectID, err := resourcename.ExtractProjectID(req.ParentResourceName); err == nil {
			cf.ProjectID = projectID
		}

		// Write the file to the requested path.
		if err := auth.WriteCredentialFile(*opts.CredentialFilePath, cf); err != nil {
			return fmt.Errorf("failed to write credential file: %w", err)
		}

		fmt.Fprintf(opts.IO.Err(), "%s Service principal credential file written to %q\n",
			opts.IO.ColorScheme().SuccessIcon(), *opts.CredentialFilePath)
		return nil
	}

	return opts.Output.Display(newSecretDisplayer(format.Pretty, resp.Payload.Key, resp.Payload.ClientSecret))
}
