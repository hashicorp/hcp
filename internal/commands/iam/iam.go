package iam

import (
	"github.com/hashicorp/hcp/internal/commands/iam/groups"
	"github.com/hashicorp/hcp/internal/commands/iam/roles"
	serviceprincipals "github.com/hashicorp/hcp/internal/commands/iam/serviceprincipals"
	"github.com/hashicorp/hcp/internal/commands/iam/users"
	"github.com/hashicorp/hcp/internal/commands/iam/workloadidentityproviders"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
)

func NewCmdIam(ctx *cmd.Context) *cmd.Command {
	cmd := &cmd.Command{
		Name:      "iam",
		ShortHelp: "Identity and access management.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ Bold "hcp iam" }} command group allows you to manage HCP identities including
		users, groups, and service principals.

		Service principal keys or workload identity providers may also be managed. When accessing
		HCP services from workloads that have an external identity provider, prefer using workload
		identity federation for more secure access to HCP.
		`),
	}

	cmd.AddChild(roles.NewCmdRoles(ctx))
	cmd.AddChild(users.NewCmdUsers(ctx))
	cmd.AddChild(groups.NewCmdGroups(ctx))
	cmd.AddChild(serviceprincipals.NewCmdServicePrincipals(ctx))
	cmd.AddChild(workloadidentityproviders.NewCmdWIPs(ctx))

	return cmd
}
