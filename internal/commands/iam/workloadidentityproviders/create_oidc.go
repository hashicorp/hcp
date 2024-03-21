package workloadidentityproviders

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/client/service_principals_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/models"
	"github.com/hashicorp/hcp/internal/commands/iam/serviceprincipals/helper"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/flagvalue"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
)

func NewCmdCreateOIDC(ctx *cmd.Context, runF func(*CreateOIDCOpts) error) *cmd.Command {
	opts := &CreateOIDCOpts{
		Ctx:     ctx.ShutdownCtx,
		Profile: ctx.Profile,
		Output:  ctx.Output,
		IO:      ctx.IO,
		Client:  service_principals_service.New(ctx.HCP, nil),
	}

	cmd := &cmd.Command{
		Name:      "create-oidc",
		ShortHelp: "Create an OIDC Workload Identity Provider.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp iam workload-identity-providers create-oidc" }} command creates
		a new OIDC based workload identity provider.

		Common OIDC providers include Azure, GCP, Kubernetes Clusters, HashiCorp Vault, GitHub, GitLab, and more.

		When creating an OIDC provider, you must specify the issuer URL, the conditional access statement, and
		optionally the allowed audiences.

		The issuer URL is the URL of the OIDC provider that is allowed to exchange workload identities. The URL
		must be a valid URL that is reachable from the HCP control plane, and must match the issuer set in the
		response to the OIDC discovery endpoint (${issuer_url}/.well-known/openid-configuration).

		The conditional access statement must be set and is used to restrict which tokens issued by the OIDC
		provider are allowed to exchange their identity for an HCP access token. The condtional access statement
		is a hashicorp/go-bexpr string that is evaluated when exchanging tokens. It has access to all the claims
		in the external identity token and they can be accessed via the "jwt_claims.<claim_name>" syntax. An example
		conditional access statement that restricts access to a specific subject claim is 'jwt_claims.sub == "example"'.

		If unset, the allowed audiences will default to the resource name of the provider. The format will be:
		{{ template "mdCodeOrBold" "iam/project/PROJECT_ID/service-principal/SP_NAME/workload-identity-provider/WIP_NAME" }}.
		If set, the presented access token must have an audience that is contained in the set of allowed audiences.
		`),
		Examples: []cmd.Example{
			{
				Preamble: `Azure - Allow exchanging a User Managed Identity:`,
				Command: heredoc.New(ctx.IO, heredoc.WithNoWrap(), heredoc.WithPreserveNewlines()).Must(`
				$ hcp iam workload-identity-providers create-oidc azure-example-user-managed \
				  --service-principal=iam/project/PROJECT/service-principal/example-sp \
				  --issuer=https://sts.windows.net/AZURE_AD_TENANT_ID/ \
				  --allowed-audience=MANAGED_IDENTITY_CLIENT_ID \
				  --conditional-access='jwt_claims.sub == "MANAGED_IDENTITY_OBJECT_PRINCIPAL_ID"' \
				  --description="Azure User Managed Identity Example"
				`),
			},
			{
				Preamble: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
					GCP - Allow exchanging a Service Account Identity

					{{ Link "Full List of claims" "https://cloud.google.com/compute/docs/instances/verifying-instance-identity#payload" }}:
				`),
				Command: heredoc.New(ctx.IO, heredoc.WithNoWrap(), heredoc.WithPreserveNewlines()).Must(`
				$ hcp iam workload-identity-providers create-oidc gcp-example-service-account \
				  --service-principal=iam/project/PROJECT/service-principal/example-sp \
				  --issuer=https://accounts.google.com \
				  --conditional-access='jwt_claims.sub == "SERVICE_ACCOUNT_UNIQUE_ID"' \
				  --description="GCP Service Account Example"
				`),
			},
			{
				Preamble: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
					GitLab - Allow exchanging a GitLab

					{{ Link "Full list of claims" "https://docs.gitlab.com/ee/ci/secrets/id_token_authentication.html#token-payload" }}:
				`),
				Command: heredoc.New(ctx.IO, heredoc.WithNoWrap(), heredoc.WithPreserveNewlines()).Must(`
				$ hcp iam workload-identity-providers create-oidc gcp-example-service-account \
				  --service-principal=iam/project/PROJECT/service-principal/example-sp \
				  --issuer=https://gitlab.com \
				  --conditional-access='jwt_claims.project_path == "example-org/example-repo" and jwt_cliams.job_id == 302' \
				  --description="GitLab example-repo access for job 302"
				`),
			},
		},
		Args: cmd.PositionalArguments{
			Args: []cmd.PositionalArgument{
				{
					Name:          "PROVIDER_NAME",
					Documentation: "The name of the provider to create.",
				},
			},
		},
		Flags: cmd.Flags{
			Local: []*cmd.Flag{
				{
					Name:         "service-principal",
					DisplayValue: "RESOURCE_NAME",
					Description:  "The resource name of the service principal to create the provider for.",
					Value:        flagvalue.Simple("", &opts.SP),
					Required:     true,
				},
				{
					Name:         "issuer",
					DisplayValue: "URI",
					Description:  "The URL of the OIDC Issuer that is allowed to exchange workload identities.",
					Value:        flagvalue.Simple("", &opts.IssuerURI),
					Required:     true,
				},
				{
					Name:         "allowed-audience",
					DisplayValue: "AUD",
					Description: heredoc.New(ctx.IO).Must(`
					The set of audiences set on the access token that are allowed to exchange identities.
					The access token must have an audience that is contained in this set.

					If no audience is set, the default allowed audience will be the resource name of the provider. The format will be:
					{{ template "mdCodeOrBold" "iam/project/PROJECT_ID/service-principal/SP_NAME/workload-identity-provider/WIP_NAME" }}.
					`),
					Value:      flagvalue.SimpleSlice(nil, &opts.AllowedAudiences),
					Repeatable: true,
				},
				{
					Name:         "conditional-access",
					DisplayValue: "STATEMENT",
					Description: heredoc.New(ctx.IO).Must(`
					The conditional access statement is a hashicorp/go-bexpr string that is evaluated
					when exchanging tokens. It restricts which upstream identities are allowed to access
					the service principal.

					The conditional_access statement can access any claim from the external identity token using
					the {{ template "mdCodeOrBold" "jwt_claims.<claim_name>" }} syntax.
					As an example, access the subject claim with
					{{ template "mdCodeOrBold" "jwt_claims.sub" }}.
					`),
					Value:    flagvalue.Simple("", &opts.ConditionalAccess),
					Required: true,
				},
				{
					Name:         "description",
					DisplayValue: "TEXT",
					Description:  "A description of the provider.",
					Value:        flagvalue.Simple("", &opts.Description),
				},
			},
		},
		RunF: func(c *cmd.Command, args []string) error {
			opts.Name = args[0]

			if runF != nil {
				return runF(opts)
			}
			return createOIDCRun(opts)
		},
		PersistentPreRun: func(c *cmd.Command, args []string) error {
			return cmd.RequireOrgAndProject(ctx)
		},
	}

	cmd.Flags.Local[0].Autocomplete = helper.PredictSPResourceName(ctx, cmd, opts.Client)

	return cmd
}

type CreateOIDCOpts struct {
	Ctx     context.Context
	Profile *profile.Profile
	Output  *format.Outputter
	IO      iostreams.IOStreams

	Name              string
	SP                string
	IssuerURI         string
	AllowedAudiences  []string
	ConditionalAccess string
	Description       string
	Client            service_principals_service.ClientService
}

func createOIDCRun(opts *CreateOIDCOpts) error {
	if !helper.SPResourceName.MatchString(opts.SP) {
		return fmt.Errorf("invalid service principal resource name: %s", opts.SP)
	}

	req := service_principals_service.NewServicePrincipalsServiceCreateWorkloadIdentityProviderParamsWithContext(opts.Ctx)
	req.ParentResourceName = opts.SP

	req.Body = service_principals_service.ServicePrincipalsServiceCreateWorkloadIdentityProviderBody{
		Name: opts.Name,
		Provider: &models.HashicorpCloudIamWorkloadIdentityProvider{
			ConditionalAccess: opts.ConditionalAccess,
			Description:       opts.Description,
			OidcConfig: &models.HashicorpCloudIamOIDCWorkloadIdentityProviderConfig{
				AllowedAudiences: opts.AllowedAudiences,
				IssuerURI:        opts.IssuerURI,
			},
		},
	}

	resp, err := opts.Client.ServicePrincipalsServiceCreateWorkloadIdentityProvider(req, nil)
	if err != nil {
		return fmt.Errorf("failed to create workload identity provider: %w", err)
	}

	if err := opts.Output.Display(newDisplayer(format.Pretty, true, resp.Payload.Provider)); err != nil {
		return err
	}

	command := fmt.Sprintf(`$ hcp iam workload-identity-providers create-cred-file \
    %s \
    --output-file=creds.json \
    [REQUIRED FLAGS]`, resp.Payload.Provider.ResourceName)

	fmt.Fprintln(opts.IO.Err())
	fmt.Fprintf(opts.IO.Err(), `To create a credential file for the OIDC workload identity provider,
run the following command, passing the appropriate flags:

  %s`, opts.IO.ColorScheme().String(command).Bold())
	fmt.Fprintln(opts.IO.Err())

	return nil
}
