// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package workloadidentityproviders

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcp-sdk-go/auth"
	"github.com/hashicorp/hcp-sdk-go/auth/workload"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/flagvalue"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
)

const (
	// azureURL is the URL to obtain an Azure token from the Azure metadata service.
	azureURL = "http://169.254.169.254/metadata/identity/oauth2/token?api-version=2018-02-01&resource=%s"

	// azureClientIDQueryParam is the query parameter to specify the client ID when obtaining an Azure token.
	azureClientIDQueryParam = "&client_id=%s"

	// azureSubjectCredentialPointer is the JSON pointer to the access token in the Azure metadata service response.
	azureSubjectCredentialPointer = "/access_token"

	// gcpURL is the URL to obtain a GCP token from the GCP metadata service.
	gcpURL = "http://metadata/computeMetadata/v1/instance/service-accounts/default/identity?audience=%s&format=full"
)

var (
	// azureHeaders are the headers to send to the Azure metadata service.
	azureHeaders = map[string]string{
		"Metadata": "True",
	}

	// gcpHeaders are the headers to send to the GCP metadata service.
	gcpHeaders = map[string]string{
		"Metadata-Flavor": "Google",
	}
)

func NewCmdCreateCredFile(ctx *cmd.Context, runF func(*CreateCredFileOpts) error) *cmd.Command {
	opts := &CreateCredFileOpts{
		IO: ctx.IO,
	}

	cmd := &cmd.Command{
		Name:      "create-cred-file",
		ShortHelp: "Create a credential file.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp iam workload-identity-providers create-cred-file" }}
		command creates a credential file that allow access authenticating to HCP from a variety
		of external accounts.

		The generated credential file contains details on how to obtain the credential from the
		external identity provider and how to exchange them for an HCP access token.

		After creating the credential file, the HCP CLI can be authenticated by the workload by running
		{{ template "mdCodeOrBold" "hcp auth login --cred-file=PATH" }} where PATH is the path to
		the generated credential file.
		`),
		Examples: []cmd.Example{
			{
				Preamble: `Create a credential file for an AWS workload:`,
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				# Set the --imdsv1 flag if the AWS instance metadata service is using version 1.
				$ hcp iam workload-identity-providers create-cred-file \
				  iam/project/123/service-principal/my-sp/workload-identity-provider/aws \
				  --aws \
				  --output-file=credentials.json
				`),
			},
			{
				Preamble: `Create a credential file for a GCP workload:`,
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				$ hcp iam workload-identity-providers create-cred-file \
				  iam/project/123/service-principal/my-sp/workload-identity-provider/gcp \
				  --gcp \
				  --output-file=credentials.json
				`),
			},
			{
				Preamble: `Create a credential file for an Azure workload using a User Managed Identity:`,
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				$ hcp iam workload-identity-providers create-cred-file \
				  iam/project/123/service-principal/my-sp/workload-identity-provider/azure \
				  --azure \
				  --azure-resource=MANAGED_IDENTITY_CLIENT_ID \
				  --output-file=credentials.json
				`),
			},
			{
				Preamble: `Create a credential file for an Azure workload that has multiple User Managed Identities:`,
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				$ hcp iam workload-identity-providers create-cred-file \
				  iam/project/123/service-principal/my-sp/workload-identity-provider/azure \
				  --azure \
				  --azure-resource=MANAGED_IDENTITY_CLIENT_ID \
				  --azure-client-id=MANAGED_IDENTITY_CLIENT_ID \
				  --output-file=credentials.json
				`),
			},
			{
				Preamble: `Create a credential file for an Azure workload that is using a Managed Identity to authenticate as a Entra ID Application:`,
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				# ENTRA_ID_APP_ID_URL generally has the form "api://123-456-678-901"
				$ hcp iam workload-identity-providers create-cred-file \
				  iam/project/123/service-principal/my-sp/workload-identity-provider/azure \
				  --azure \
				  --azure-resource=ENTRA_ID_APP_ID_URI \
				  --azure-client-id=MANAGED_IDENTITY_CLIENT_ID \
				  --output-file=credentials.json
				`),
			},
			{
				Preamble: `Create a credential file that sources the token from a file:`,
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				# Assuming the file has the following JSON payload:
				# {
				#   "access_token": "eyJ0eXAiOiJKV1Qi...",
				#   ...
				# }
				$ hcp iam workload-identity-providers create-cred-file \
				  iam/project/123/service-principal/my-sp/workload-identity-provider/k8s \
				  --source-file=/var/run/secrets/tokens/hcp_token \
				  --source-json-pointer=/access_token \
				  --output-file=credentials.json
				`),
			},
			{
				Preamble: `Create a credential file that sources the token from a file:`,
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				# Assuming the file only contains the access token:
				$ hcp iam workload-identity-providers create-cred-file \
				  iam/project/123/service-principal/my-sp/workload-identity-provider/k8s \
				  --source-file \
				  --output-file=credentials.json
				`),
			},
			{
				Preamble: `Create a credential file that sources the token from an URL:`,
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				# Assuming the response has the following JSON payload:
				# {
				#   "access_token": "eyJ0eXAiOiJKV1Qi...",
				#   ...
				# }
				$ hcp iam workload-identity-providers create-cred-file \
				  iam/project/123/service-principal/my-sp/workload-identity-provider/example \
				  --source-url="https://example-oidc-provider.com/token" \
				  --source-json-pointer=/access_token \
				  --output-file=credentials.json
				`),
			},
			{
				Preamble: `Create a credential file that sources the token from an URL:`,
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				# Assuming the file only contains the access token:
				$ hcp iam workload-identity-providers create-cred-file \
				  iam/project/123/service-principal/my-sp/workload-identity-provider/example \
				  --source-url=https://example-oidc-provider.com/token \
				  --output-file=credentials.json
				`),
			},
			{
				Preamble: `Create a credential file that sources the token from an URL:`,
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				# To add headers to the request, use the --source-header flag:
				$ hcp iam workload-identity-providers create-cred-file \
				  iam/project/123/service-principal/my-sp/workload-identity-provider/example \
				  --source-url=https://example-oidc-provider.com/token \
				  --source-header=Metadata=True \
				  --source-header=Token=Identity \
				  --output-file=credentials.json
				`),
			},
			{
				Preamble: `Create a credential file that sources the token from an environment variable:`,
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				# Assuming the environment variable has the following JSON string value:
				# {
				#   "access_token": "eyJ0eXAiOiJKV1Qi...",
				#   ...
				# }
				$ hcp iam workload-identity-providers create-cred-file \
				  iam/project/123/service-principal/my-sp/workload-identity-provider/example \
				  --source-env=ACCESS_TOKEN \
				  --source-json-pointer=/access_token \
				  --output-file=credentials.json
				`),
			},
			{
				Preamble: `Create a credential file that sources the token from an environment variable:`,
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				# Assuming the environment variable only contains the access token:
				$ hcp iam workload-identity-providers create-cred-file \
				  iam/project/123/service-principal/my-sp/workload-identity-provider/example \
				  --source-env=ACCESS_TOKEN \
				  --output-file=credentials.json
				`),
			},
		},
		Args: cmd.PositionalArguments{
			Args: []cmd.PositionalArgument{
				{
					Name:          "WORKLOAD_IDENTITY_PROVIDER_NAME",
					Documentation: "The resource name of the provider for which the external identity will be exchanged against using the credential file.",
				},
			},
		},
		Flags: cmd.Flags{
			Local: []*cmd.Flag{
				{
					Name:         "output-file",
					DisplayValue: "PATH",
					Description:  "The path to output the credential file.",
					Value:        flagvalue.Simple("", &opts.OutputFile),
					Required:     true,
				},
				{
					Name:          "aws",
					Description:   "Set if exchanging an AWS workload identity.",
					Value:         flagvalue.Simple(false, &opts.AWS),
					IsBooleanFlag: true,
				},
				{
					Name:          "imdsv1",
					Description:   "Set if the AWS instance metadata service is using version 1.",
					Value:         flagvalue.Simple(false, &opts.IMDSv1),
					IsBooleanFlag: true,
				},
				{
					Name:          "azure",
					Description:   "Set if exchanging an Azure workload identity.",
					Value:         flagvalue.Simple(false, &opts.Azure),
					IsBooleanFlag: true,
				},
				{
					Name:         "azure-resource",
					DisplayValue: "URI",
					Description: heredoc.New(ctx.IO).Must(`
					The Azure Instance Metadata Service (IMDS) allows retrieving an access token for a specific resource.
					The audience (aud) claim in the returned token is set to the value of the resource parameter. As such,
					the azure-resource flag must be set to one of the allowed audiences for the Workload Identity Provider.

					The typical values for this flag are:

					{{ PreserveNewLines }}
					  * The Client ID of the User Assigned Managed Identity (UUID)
					  * The Application ID URI of the Microsoft Entra ID Application
					    ({{ template "mdCodeOrBold" "api://123-456-678-901" }}).
					{{ PreserveNewLines }}

					For more details on the resource parameter, see the
					{{ Link "Azure documentation" "https://learn.microsoft.com/en-us/entra/identity/managed-identities-azure-resources/how-to-use-vm-token#get-a-token-using-http" }}.
					`),
					Value: flagvalue.Simple("", &opts.AzureResource),
				},
				{
					Name:         "azure-client-id",
					DisplayValue: "ID",
					Description: heredoc.New(ctx.IO).Must(`
					In the case that the workload has multiple User Assigned Managed Identities,
					this flag specifies which Client ID should be used to retrieve the Azure identity token.

					If the workload only has one User Assigned Managed Identity, this flag is not required.
					`),
					Value: flagvalue.Simple("", &opts.AzureClientID),
				},
				{
					Name: "gcp",
					Description: heredoc.New(ctx.IO).Must(`
					Set if exchanging an GCP workload identity.

					It is assumed the workload identity provider was created
					with the issuer URI set to {{ template "mdCodeOrBold" "https://accounts.google.com" }}
					and the default allowed audiences.
					`),
					Value:         flagvalue.Simple(false, &opts.GCP),
					IsBooleanFlag: true,
				},
				{
					Name:         "source-url",
					DisplayValue: "URL",
					Description:  "URL to obtain the credential from.",
					Value:        flagvalue.Simple("", &opts.SourceURL),
				},
				{
					Name:         "source-file",
					DisplayValue: "PATH",
					Description:  "Path to file that contains the credential to exchange.",
					Value:        flagvalue.Simple("", &opts.SourceFile),
				},
				{
					Name:         "source-env",
					DisplayValue: "VAR",
					Description:  "The environment variable name that contains the credential to exchange.",
					Value:        flagvalue.Simple("", &opts.SourceEnvVar),
				},
				{
					Name:         "source-json-pointer",
					DisplayValue: "/PATH/TO/CREDENTIAL",
					Description: heredoc.New(ctx.IO).Must(`
A JSON pointer that indicates how to access the credential from a JSON.
If used with the {{ template "mdCodeOrBold" "source-url" }} flag, the pointer
is used to extract the credential from the JSON response from calling the URL.
If used with the {{ template "mdCodeOrBold" "source-file" }} flag, the pointer
is used to extract the credential read from the JSON file. Similarly, if used
with the {{ template "mdCodeOrBold" "source-env" }} flag, the pointer is used to
extract the credential from the environment variable whose value is a JSON
object.

As an example, if the JSON payload containing the credential file is:

{{ define "credentials" -}} {
  "access_token": "credentials",
  "nested": {
    "access_token": "nested-credentials"
  }
} {{- end }}
{{- CodeBlock "credentials" "json hideClipboard" }}

You can access the top level access token using the pointer
{{ template "mdCodeOrBold" "/access_token" }} and the nested access token can be
accessed using the pointer {{ template "mdCodeOrBold" "/nested/access_token" }}.
					`),
					Value: flagvalue.Simple("", &opts.CredentialJSONPointer),
				},
				{
					Name:         "source-header",
					DisplayValue: "KEY=VALUE",
					Description:  "Headers to send to the URL when obtaining the credential.",
					Value:        flagvalue.SimpleMap(nil, &opts.SourceURLHeaders),
					Repeatable:   true,
				},
			},
		},
		RunF: func(c *cmd.Command, args []string) error {
			opts.WIP = args[0]

			if runF != nil {
				return runF(opts)
			}
			return createCredFileRun(opts)
		},
		NoAuthRequired: true,
	}

	return cmd
}

type CreateCredFileOpts struct {
	IO iostreams.IOStreams

	WIP        string
	OutputFile string

	// Only one of these can be set
	AWS          bool
	Azure        bool
	GCP          bool
	SourceEnvVar string
	SourceURL    string
	SourceFile   string

	// AWS options
	IMDSv1 bool

	// Azure options
	AzureResource string
	AzureClientID string

	// JSON Options
	CredentialJSONPointer string

	// Headers to sent to SourceURL
	SourceURLHeaders map[string]string
}

func (c *CreateCredFileOpts) Validate() error {
	// Check that the resource name is valid
	if !WIPResourceName.MatchString(c.WIP) {
		return fmt.Errorf("invalid workload identity provider resource name: %s", c.WIP)
	}

	// Ensure we only received on of the source options
	sources := 0
	if c.AWS {
		sources++
	}
	if c.Azure {
		sources++
	}
	if c.GCP {
		sources++
	}
	if c.SourceEnvVar != "" {
		sources++
	}
	if c.SourceURL != "" {
		sources++
	}
	if c.SourceFile != "" {
		sources++
	}
	if sources != 1 {
		return fmt.Errorf("only one of --aws, --azure, --gcp, --source-env, --source-url, or --source-file can be set")
	}

	// Enusre that IMDSv1 is only set if AWS is set
	if c.IMDSv1 && !c.AWS {
		return fmt.Errorf("--imdsv1 can only be set if --aws is set")
	}

	// Ensure Azure resource is set if Azure is set
	if c.Azure && c.AzureResource == "" {
		return fmt.Errorf("--azure-resource must be set if --azure is set")
	}

	// Ensure Azure Resource is only set if Azure is set
	if c.AzureResource != "" && !c.Azure {
		return fmt.Errorf("--azure-resource can only be set if --azure is set")
	}

	// Ensure the Azure client ID is set if Azure is set and the resource is set
	if c.AzureClientID != "" && !c.Azure {
		return fmt.Errorf("--azure-client-id can only be set if --azure is set")
	}

	// Ensure no credential JSON Pointer if AWS/GCP/Azure set
	if (c.AWS || c.GCP || c.Azure) && c.CredentialJSONPointer != "" {
		return fmt.Errorf("--source-json-pointer can only be set if --source-url, --source-file, or --source-env is set")
	}

	// Ensure SourceURLHeaders is only set if SourceURL is set
	if len(c.SourceURLHeaders) > 0 && c.SourceURL == "" {
		return fmt.Errorf("--source-header can only be set if --source-url is set")
	}

	// Ensure the output file has a .json extension
	if !strings.HasSuffix(c.OutputFile, ".json") {
		return fmt.Errorf("output file must have a .json extension")
	}

	return nil
}

func createCredFileRun(opts *CreateCredFileOpts) error {
	if err := opts.Validate(); err != nil {
		return err
	}

	// Create the credential file
	cf := &auth.CredentialFile{
		ProjectID: "",
		Scheme:    auth.CredentialFileSchemeWorkload,
		Workload: &workload.IdentityProviderConfig{
			ProviderResourceName: opts.WIP,
		},
	}

	// Capture the credential JSON pointer if set.
	var format workload.CredentialFormat
	if opts.CredentialJSONPointer != "" {
		format.Type = workload.FormatTypeJSON
		format.SubjectCredentialPointer = opts.CredentialJSONPointer
	}

	// Configure based on the passed options.
	if opts.AWS {
		cf.Workload.AWS = &workload.AWSCredentialSource{
			IMDSv2: !opts.IMDSv1,
		}
	} else if opts.Azure {
		// Determine the IMDS URL based on the presence of the client ID.
		url := fmt.Sprintf(azureURL, opts.AzureResource)
		if opts.AzureClientID != "" {
			url += fmt.Sprintf(azureClientIDQueryParam, opts.AzureClientID)
		}

		cf.Workload.URL = &workload.URLCredentialSource{
			URL:     url,
			Headers: azureHeaders,
			CredentialFormat: workload.CredentialFormat{
				Type:                     workload.FormatTypeJSON,
				SubjectCredentialPointer: azureSubjectCredentialPointer,
			},
		}
	} else if opts.GCP {
		cf.Workload.URL = &workload.URLCredentialSource{
			URL:     fmt.Sprintf(gcpURL, opts.WIP),
			Headers: gcpHeaders,
		}
	} else if opts.SourceEnvVar != "" {
		cf.Workload.EnvironmentVariable = &workload.EnvironmentVariableCredentialSource{
			Var:              opts.SourceEnvVar,
			CredentialFormat: format,
		}
	} else if opts.SourceURL != "" {
		cf.Workload.URL = &workload.URLCredentialSource{
			URL:              opts.SourceURL,
			Headers:          opts.SourceURLHeaders,
			CredentialFormat: format,
		}
	} else if opts.SourceFile != "" {
		cf.Workload.File = &workload.FileCredentialSource{
			Path:             opts.SourceFile,
			CredentialFormat: format,
		}
	}

	if err := auth.WriteCredentialFile(opts.OutputFile, cf); err != nil {
		return fmt.Errorf("failed to write credential file: %w", err)
	}

	fmt.Fprintf(opts.IO.Err(), "Credential file written to %q\n", opts.OutputFile)
	return nil
}
