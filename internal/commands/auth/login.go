// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/hashicorp/hcp-sdk-go/auth"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/client/iam_service"
	hcpconf "github.com/hashicorp/hcp-sdk-go/config"
	"github.com/hashicorp/hcp-sdk-go/httpclient"
	hcpAuth "github.com/hashicorp/hcp/internal/pkg/auth"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/flagvalue"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
	"github.com/hashicorp/hcp/version"
	"github.com/mitchellh/go-homedir"
	"github.com/posener/complete"
)

func NewCmdLogin(ctx *cmd.Context) *cmd.Command {
	opts := &LoginOpts{
		Ctx:     ctx.ShutdownCtx,
		IO:      ctx.IO,
		Profile: ctx.Profile,
		GetIAM: func(c hcpconf.HCPConfig) (iam_service.ClientService, error) {

			// Create a new HCP client since the one passed through the
			// cmd.Context may not be authenticated with the same principal that the
			// login command will be.
			hconfig := httpclient.Config{
				HCPConfig:     c,
				SourceChannel: version.GetSourceChannel(),
			}

			hcpClient, err := httpclient.New(hconfig)
			if err != nil {
				return nil, fmt.Errorf("failed to create HCP client: %w", err)
			}

			return iam_service.New(hcpClient, nil), nil
		},
		ConfigFn:      hcpconf.NewHCPConfig,
		CredentialDir: hcpAuth.CredentialsDir,
	}

	cmd := &cmd.Command{
		Name:      "login",
		ShortHelp: "Login to HCP.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp auth login" }} command lets you login to authenticate to HCP.

		If no arguments are provided, authentication occurs for your user principal by initiating a web browser login flow.

		To authenticate non-interactively, you may authenticate as a service principal. To do so, use the
		{{ template "mdCodeOrBold" "--client-id" }} and {{ template "mdCodeOrBold" "--client-secret" }}
		flags. A service principal may be created using {{ template "mdCodeOrBold" "hcp iam service-principals create" }}
		or via the {{ Link "HCP Portal" "https://portal.cloud.hashicorp.com" }}.

		If authenticating a workload using a Workload Identity Provider, a credential file may be used to authenticate
		by passing the Path to the credential file using {{ template "mdCodeOrBold" "--cred-file" }}. The command
		should be running in the environment that Workload Identity was previously configured to be able to retrieve
		and federate external credentials from.
		`),
		Examples: []cmd.Example{
			{
				Preamble: "Login interactively using a browser:",
				Command:  "$ hcp auth login",
			},
			{
				Preamble: "Login using service principal credentials:",
				Command:  "$ hcp auth login --client-id=spID --client-secret=spSecret",
			},
			{
				Preamble: "Login using Workload Identity credentials:",
				Command:  "$ hcp auth login --cred-file=workload_cred_file.json",
			},
		},
		Flags: cmd.Flags{
			Local: []*cmd.Flag{
				{
					Name:         "client-id",
					DisplayValue: "ID",
					Description:  "Service principal Client ID used to authenticate as the given service principal.",
					Value:        flagvalue.Simple("", &opts.ClientID),
				},
				{
					Name:         "client-secret",
					DisplayValue: "SECRET",
					Description:  "Service principal Client Secret used to authenticate as the given service principal.",
					Value:        flagvalue.Simple("", &opts.ClientSecret),
				},
				{
					Name:         "cred-file",
					DisplayValue: "PATH",
					Description: heredoc.New(ctx.IO, heredoc.WithNoWrap()).Must(`
				Path to the credential file used for workload identity federation (generated by
				{{ template "mdCodeOrBold" "hcp iam workload-identity-providers create-cred-file" }}) or service account
				credential key file.
				`),
					Value:        flagvalue.Simple("", &opts.CredentialFile),
					Autocomplete: complete.PredictFiles("*.json"),
				},
			},
		},
		RunF: func(c *cmd.Command, args []string) error {
			// Read our global flags
			opts.Quiet = ctx.GetGlobalFlags().Quiet
			return loginRun(opts)
		},
		NoAuthRequired: true,
	}

	return cmd
}

// NewConfigFunc is the function definition for retrieving a new HCPConfig
type NewConfigFunc func(opts ...hcpconf.HCPConfigOption) (hcpconf.HCPConfig, error)

// GetIAMClientFunc is the function definition for retrieving an IAM service client
// from a HCP Config.
type GetIAMClientFunc func(c hcpconf.HCPConfig) (iam_service.ClientService, error)

type LoginOpts struct {
	Ctx     context.Context
	IO      iostreams.IOStreams
	Profile *profile.Profile
	Quiet   bool

	// GetIAM retrieves an IAM service client using the passed HCP Config.
	GetIAM GetIAMClientFunc

	// ConfigFn is used to retrieve a new HCP Config
	ConfigFn NewConfigFunc

	// CredentialDir is the directory to store any necessary credential files.
	CredentialDir string

	// CredentialFile is the path to a credential file to use to login
	CredentialFile string

	// ClientID and ClientSecret are used to authenticate using service
	// principal credentials.
	ClientID     string
	ClientSecret string
}

func (o *LoginOpts) Validate() error {
	if o.IO == nil || o.Profile == nil || o.ConfigFn == nil || o.CredentialDir == "" {
		return fmt.Errorf("programmer error; missing required fields")
	}

	if o.CredentialFile != "" && (o.ClientID != "" || o.ClientSecret != "") {
		return fmt.Errorf("both credential file and client id/secret may not be set")
	}

	if (o.ClientID != "" || o.ClientSecret != "") && (o.ClientID == "" || o.ClientSecret == "") {
		return fmt.Errorf("both client id and client secret must be set")
	}

	return nil
}

func loginRun(opts *LoginOpts) error {
	if err := opts.Validate(); err != nil {
		return err
	}

	// Build our options
	var storeCredFile bool
	options := []hcpconf.HCPConfigOption{hcpconf.WithoutLogging()}
	if opts.CredentialFile != "" {
		options = append(options, hcpconf.WithCredentialFilePath(opts.CredentialFile))
		storeCredFile = true
	} else if opts.ClientID != "" {
		options = append(options, hcpconf.WithClientCredentials(opts.ClientID, opts.ClientSecret))
		storeCredFile = true
	} else {
		options = append(options, hcpconf.FromEnv())
	}

	// Get a HCP config
	hcpConfig, err := opts.ConfigFn(options...)
	if err != nil {
		return fmt.Errorf("error creating HCP config: %w", err)
	}

	// Try to get a token to ensure that we successfully authenticated.
	if _, err = hcpConfig.Token(); err != nil {
		return fmt.Errorf("unable to login to HCP: %w", err)
	}

	// Write any credential file necessary
	if storeCredFile {
		if err := writeCredFile(opts); err != nil {
			return err
		}
	}

	// If there is no organization or project set in the profile, attempt to
	// default it if the logging in principal is a service principal.
	if opts.Profile.OrganizationID == "" || opts.Profile.ProjectID == "" {
		// Get an IAM client using the new hcpConfig
		iam, err := opts.GetIAM(hcpConfig)
		if err != nil {
			return fmt.Errorf("failed to create IAM client: %w", err)
		}

		// Get the caller identity. If it is a service principal, we can set the
		// organization and potentially project automatically.
		callerIdentityParams := iam_service.NewIamServiceGetCallerIdentityParamsWithContext(opts.Ctx)
		ident, err := iam.IamServiceGetCallerIdentity(callerIdentityParams, nil)
		if err != nil {
			return fmt.Errorf("failed to get identity of principal logging in: %w", err)
		}

		didUpdate := false
		isSP := ident.Payload != nil && ident.Payload.Principal != nil && ident.Payload.Principal.Service != nil
		if opts.Profile.OrganizationID == "" && isSP {
			opts.Profile.OrganizationID = ident.Payload.Principal.Service.OrganizationID
			didUpdate = true
		}
		if opts.Profile.ProjectID == "" && isSP && ident.Payload.Principal.Service.ProjectID != "" {
			opts.Profile.ProjectID = ident.Payload.Principal.Service.ProjectID
			didUpdate = true
		}

		if didUpdate {
			if err := opts.Profile.Write(); err != nil {
				return fmt.Errorf("failed to update profile: %w", err)
			}
		}
	}

	cs := opts.IO.ColorScheme()
	if !opts.Quiet {
		fmt.Fprintln(opts.IO.Err(), cs.String("Successfully logged in!").Bold().Color(cs.Green()))

		// Check to see if we should ask the user to configure their profile
		if opts.Profile.OrganizationID == "" || opts.Profile.ProjectID == "" {
			fmt.Fprintln(opts.IO.Err())
			fmt.Fprintln(opts.IO.Err(), heredoc.New(opts.IO).Must(`
		No profile configuration detected. To configure {{ Bold "hcp" }} to execute commands against your
		desired organization and project, run:

		  {{ Bold "$ hcp profile init" }}
		`))
			fmt.Fprintln(opts.IO.Err())
		}
	}

	return nil
}

func writeCredFile(opts *LoginOpts) (err error) {
	// Open the destination file
	dir, err := homedir.Expand(opts.CredentialDir)
	if err != nil {
		return fmt.Errorf("failed to resolve hcp's credential directory path: %w", err)
	}

	// Ensure the path exists
	if err := os.MkdirAll(dir, 0766); err != nil {
		return fmt.Errorf("failed to create hcp's credential directory: %w", err)
	}

	// Create a credential file
	credFilePath := filepath.Join(dir, hcpAuth.CredFileName)
	cf, err := os.Create(credFilePath)
	if err != nil {
		return fmt.Errorf("failed to create hcp credential file: %w", err)
	}
	defer func() {
		err = cf.Close()
		if err != nil {
			err = fmt.Errorf("failed to close created credential file: %w", err)
		}
	}()

	// Copy the passed credential file contents
	if opts.CredentialFile != "" {
		source, err := os.Open(opts.CredentialFile)
		if err != nil {
			return fmt.Errorf("failed to open credential file: %w", err)
		}
		defer func() {
			err = source.Close()
			if err != nil {
				err = fmt.Errorf("failed to close passed credential file: %w", err)
			}
		}()

		if _, err := io.Copy(cf, source); err != nil {
			return fmt.Errorf("failed to store passed credential file: %w", err)
		}

		return nil
	}

	// Store the service principal credentials
	spCF := auth.CredentialFile{
		Scheme: auth.CredentialFileSchemeServicePrincipal,
		Oauth: &auth.OauthConfig{
			ClientID:     opts.ClientID,
			ClientSecret: opts.ClientSecret,
		},
	}

	e := json.NewEncoder(cf)
	e.SetIndent("", "  ")
	if err := e.Encode(spCF); err != nil {
		return fmt.Errorf("failed to store service principal credentials: %w", err)
	}

	return nil
}
