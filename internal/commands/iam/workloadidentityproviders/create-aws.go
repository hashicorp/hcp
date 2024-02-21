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

func NewCmdCreateAWS(ctx *cmd.Context, runF func(*CreateAWSOpts) error) *cmd.Command {
	opts := &CreateAWSOpts{
		Ctx:     ctx.ShutdownCtx,
		Profile: ctx.Profile,
		Output:  ctx.Output,
		IO:      ctx.IO,
		Client:  service_principals_service.New(ctx.HCP, nil),
	}

	cmd := &cmd.Command{
		Name:      "create-aws",
		ShortHelp: "Create an AWS Workload Identity Provider.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ Bold "hcp iam workload-identity-providers create-aws" }} command creates a new
		AWS workload identity provider.

		Once created, workloads running in the specified AWS account can exchange their AWS
		identity for an HCP access token which maps to the identity of the specified service
		principal.

		The conditional access statement can restrict which AWS roles are allowed to exchange
		their identity for an HCP access token. The condtional access statement is a hashicorp/go-bexpr
		string that is evaluated when exchanging tokens. It has access to the following variables:

		{{ PreserveNewLines }}
		  * "aws.arn": The AWS ARN associated with the calling entity.
		  * "aws.account_id": The AWS account ID number of the account that owns
			or contains the calling entity.
		  * "aws.user_id": The unique identifier of the calling entity.
		{{ PreserveNewLines }}

		An example conditional access statement that restricts access to a specific role is,
		'aws.arn matches "arn:aws:iam::123456789012:role/example-role/*"'.

		To aide in creating the conditional access statement, run {{ Bold "aws sts get-caller-identity" }}
		on the AWS workload to determine the values that will be available to the conditional access statement.
		`),
		Examples: []cmd.Example{
			{
				Preamble: `Create a provider that allows exchanging identities for AWS workloads with role "example-role":`,
				Command: heredoc.New(ctx.IO, heredoc.WithNoWrap(), heredoc.WithPreserveNewlines()).Must(`
				$ hcp iam workload-identity-providers create-aws aws-my-role \
				  --service-principal iam/project/PROJECT/service-principal/example-sp \
				  --account-id 123456789012 \
				  --conditional-access 'aws.arn matches "arn:aws:iam::123456789012:role/example-role/*"' \
				  --description "Allow exchanging AWS workloads that have role example-role"
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
					Name:         "account-id",
					DisplayValue: "AWS_ACCOUNT_ID",
					Description:  "The ID of the AWS account for which identity exchange will be allowed.",
					Value:        flagvalue.Simple("", &opts.AccountID),
					Required:     true,
				},
				{
					Name:         "conditional-access",
					DisplayValue: "STATEMENT",
					Description: heredoc.New(ctx.IO).Must(`
					conditional_access is a hashicorp/go-bexpr string that is evaluated when exchanging tokens.
					It restricts which upstream identities are allowed to access the service principal.

					The conditional_access statement can access the following variables:

					{{ PreserveNewLines }}
					  * "aws.arn": The AWS ARN associated with the calling entity.
					  * "aws.account_id": The AWS account ID number of the account that owns
					    or contains the calling entity.
					  * "aws.user_id": The unique identifier of the calling entity.
					{{ PreserveNewLines }}

					For details on the values of each variable, see the AWS documentation (https://docs.aws.amazon.com/STS/latest/APIReference/API_GetCallerIdentity.html#API_GetCallerIdentity_ResponseElements).
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
			return createAWSRun(opts)
		},
		PersistentPreRun: func(c *cmd.Command, args []string) error {
			return cmd.RequireOrgAndProject(ctx)
		},
	}

	cmd.Flags.Local[0].Autocomplete = helper.PredictSPResourceName(ctx, cmd, opts.Client)

	return cmd
}

type CreateAWSOpts struct {
	Ctx     context.Context
	Profile *profile.Profile
	Output  *format.Outputter
	IO      iostreams.IOStreams

	Name              string
	SP                string
	AccountID         string
	ConditionalAccess string
	Description       string
	Client            service_principals_service.ClientService
}

func createAWSRun(opts *CreateAWSOpts) error {
	req := service_principals_service.NewServicePrincipalsServiceCreateWorkloadIdentityProviderParamsWithContext(opts.Ctx)
	req.ParentResourceName = opts.SP

	req.Body = service_principals_service.ServicePrincipalsServiceCreateWorkloadIdentityProviderBody{
		Name: opts.Name,
		Provider: &models.HashicorpCloudIamWorkloadIdentityProvider{
			AwsConfig: &models.HashicorpCloudIamAWSWorkloadIdentityProviderConfig{
				AccountID: opts.AccountID,
			},
			ConditionalAccess: opts.ConditionalAccess,
			Description:       opts.Description,
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
    --aws --output-file=creds.json`, resp.Payload.Provider.ResourceName)

	fmt.Fprintln(opts.IO.Err())
	fmt.Fprintf(opts.IO.Err(), `To create a credential file for the AWS workload identity provider, run:

  %s`, opts.IO.ColorScheme().String(command).Bold())
	fmt.Fprintln(opts.IO.Err())

	return nil
}
