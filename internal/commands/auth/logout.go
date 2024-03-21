package auth

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

	hcpconf "github.com/hashicorp/hcp-sdk-go/config"
	"github.com/hashicorp/hcp/internal/pkg/auth"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
)

func NewCmdLogout(ctx *cmd.Context) *cmd.Command {
	cmd := &cmd.Command{
		Name:      "logout",
		ShortHelp: "Logout from HCP.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp auth logout" }} command logs out to remove access to HCP.
		`),
		Examples: []cmd.Example{
			{
				Preamble: "Logout the HCP CLI:",
				Command:  "$ hcp auth logout",
			},
		},
		RunF: func(c *cmd.Command, args []string) error {
			opts := &LogoutOpts{
				IO:            ctx.IO,
				CredentialDir: auth.CredentialsDir,
				logoutFn: func(conf hcpconf.HCPConfig) error {
					return conf.Logout()
				},
			}
			return logoutRun(opts)
		},
		NoAuthRequired: true,
	}

	return cmd
}

type LogoutOpts struct {
	IO iostreams.IOStreams

	// CredentialDir is the directory to store any necessary credential files.
	CredentialDir string

	// logoutFn is the logout function to invoke.
	logoutFn func(conf hcpconf.HCPConfig) error
}

func logoutRun(opts *LogoutOpts) error {
	hcpConfig, err := auth.GetHCPConfigFromDir(opts.CredentialDir)
	if err != nil {
		return fmt.Errorf("failed to instantiate HCP configuration: %w", err)
	}

	if err := opts.logoutFn(hcpConfig); err != nil {
		return fmt.Errorf("failed to logout from HCP: %w", err)
	}

	// Get the path the credential file
	credFilePath, err := auth.GetHCPCredFilePath(opts.CredentialDir)
	if err != nil {
		return fmt.Errorf("failed to resolve hcp's credential path: %w", err)
	}

	if err := os.Remove(credFilePath); err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("failed to remove cached credential file: %w", err)
		}
	}

	cs := opts.IO.ColorScheme()
	fmt.Fprintln(opts.IO.Err(), cs.String("Successfully logged out").Color(cs.Green()))

	return nil
}
