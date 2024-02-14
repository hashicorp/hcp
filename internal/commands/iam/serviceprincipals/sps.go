package serviceprincipals

import (
	"github.com/hashicorp/hcp/internal/commands/iam/serviceprincipals/keys"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
)

func NewCmdServicePrincipals(ctx *cmd.Context) *cmd.Command {
	cmd := &cmd.Command{
		Name:      "service-principals",
		Aliases:   []string{"sp"},
		ShortHelp: "Create and manage service principals.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ Bold "hcp iam service-principals" }} command group is used to create and manage service principals.

		A service principals is a principal that is typically used by an application or workload that
		interacts with HCP. Your application uses the service principal to authenticate to HCP so that
		users aren't directly involved.

		Because service principals are principals, you can grant it permissions by granting a role. See the examples for guidance.
		`),
		Examples: []cmd.Example{
			{
				Preamble: `Create a new service principal and grant it "admin" on the project set in the profile:`,
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				$ hcp iam service-principals create my-app --format=json
				$ hcp projects add-iam-binding --member=my-app-sp-id --role=roles/admin
				`),
			},
		},
	}

	cmd.AddChild(NewCmdCreate(ctx, nil))
	cmd.AddChild(NewCmdDelete(ctx, nil))
	cmd.AddChild(NewCmdList(ctx, nil))
	cmd.AddChild(NewCmdRead(ctx, nil))

	cmd.AddChild(keys.NewCmdKeys(ctx))
	return cmd
}
