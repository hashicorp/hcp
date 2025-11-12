// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: MPL-2.0

package version

import (
	"fmt"

	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/hashicorp/hcp/version"
)

func NewCmdVersion(ctx *cmd.Context) *cmd.Command {
	cmd := &cmd.Command{
		Name:      "version",
		ShortHelp: "Display the HCP CLI version.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp version" }} command displays the HCP CLI version.
		`),
		RunF: func(c *cmd.Command, args []string) error {
			fmt.Fprintln(ctx.IO.Out(), version.GetHumanVersion())
			return nil
		},
		NoAuthRequired: true,
	}

	return cmd
}
