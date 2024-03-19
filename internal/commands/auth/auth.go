package auth

import (
	"github.com/hashicorp/hcp/internal/pkg/cmd"
)

func NewCmdAuth(ctx *cmd.Context) *cmd.Command {
	cmd := &cmd.Command{
		Name:      "auth",
		ShortHelp: "Authenticate to HCP.",
		LongHelp:  "The `hcp auth` command group lets you manage authentication to HCP.",
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
				Preamble: "Logout the CLI:",
				Command:  "$ hcp auth logout",
			},
		},
	}

	cmd.AddChild(NewCmdLogin(ctx))
	cmd.AddChild(NewCmdLogout(ctx))
	cmd.AddChild(NewCmdPrintAccessToken(ctx))
	return cmd
}
