package auth

import (
	"github.com/hashicorp/hcp/internal/pkg/cmd"
)

func NewCmdAuth(ctx *cmd.Context) *cmd.Command {
	cmd := &cmd.Command{
		Name:      "auth",
		ShortHelp: "Authenticate to HCP",
		LongHelp:  "Manage authentication to HCP.",
		Examples: []cmd.Example{
			{
				Title:   "Browser Login",
				Command: "$ hcp auth login",
			},
			{
				Title:   "Service Principal Login",
				Command: "$ hcp auth login --client-id=spID --client-secret=spSecret",
			},
			{
				Title:   "Logout",
				Command: "$ hcp auth logout",
			},
		},
	}

	cmd.AddChild(NewCmdLogin(ctx))
	cmd.AddChild(NewCmdLogout(ctx))
	cmd.AddChild(NewCmdPrintAccessToken(ctx))
	return cmd
}
