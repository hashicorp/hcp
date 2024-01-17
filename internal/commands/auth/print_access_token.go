package auth

import (
	"fmt"

	hcpconf "github.com/hashicorp/hcp-sdk-go/config"
	"github.com/hashicorp/hcp/internal/pkg/auth"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
)

func NewCmdPrintAccessToken(ctx *cmd.Context) *cmd.Command {
	cmd := &cmd.Command{
		Name:      "print-access-token",
		ShortHelp: "Print the access token for the authenticated account.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		Print an access token for the currently authenticated account.

		The output of this command can be used to set the {{ Italic
		"Authorization: Bearer <token>" }} HTTP header when manually making API requests.
		`),
		Examples: []cmd.Example{
			{
				Preamble: "To print the access token:",
				Command:  "$ hcp auth print-access-token",
			},
			{
				Preamble: "To use the access token when curling an API:",
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				$ curl https://api.cloud.hashicorp.com/iam/2019-12-10/caller-identity \
				  --header "Authorization: Bearer $(hcp auth print-access-token)"
				  `),
			},
		},
		RunF: func(c *cmd.Command, args []string) error {
			hcpCfg, err := auth.GetHCPConfig(hcpconf.WithoutBrowserLogin())
			if err != nil {
				return fmt.Errorf("failed to instantiate HCP config: %w", err)
			}

			tkn, err := hcpCfg.Token()
			if err != nil {
				return fmt.Errorf("failed to retrieve authenticated principal's access token: %w", err)
			}

			fmt.Fprintln(ctx.IO.Out(), tkn.AccessToken)
			return nil
		},
		NoAuthRequired: false,
	}

	return cmd
}
