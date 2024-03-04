package iam

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/client/iam_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/organization_service"
	"github.com/hashicorp/hcp/internal/pkg/api/iampolicy"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/flagvalue"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
)

func NewCmdDeleteBinding(ctx *cmd.Context, runF func(*DeleteBindingOpts) error) *cmd.Command {
	opts := &DeleteBindingOpts{
		Ctx: ctx.ShutdownCtx,
		IO:  ctx.IO,
	}

	cmd := &cmd.Command{
		Name:      "delete-binding",
		ShortHelp: "Delete an IAM policy binding for the organization.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		Deletes an IAM policy binding for the organization. A binding consists of a
		principal and a role.

		To view the existing role bindings, run {{ template "mdCodeOrBold" "hcp organizations iam read-policy" }}.
		`),
		Examples: []cmd.Example{
			{
				Preamble: `Delete a role binding for a principal previously granted role "roles/viewer":`,
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				$ hcp organizations iam delete-binding \
				  --member=ef938a22-09cf-4be9-b4d0-1f4587f80f53 \
				  --role=roles/viewer
				`),
			},
		},
		Flags: cmd.Flags{
			Local: []*cmd.Flag{
				{
					Name:         "member",
					DisplayValue: "PRINCIPAL_ID",
					Description:  "The ID of the principal to remove the role binding from.",
					Value:        flagvalue.Simple("", &opts.PrincipalID),
					Required:     true,
				},
				{
					Name:         "role",
					DisplayValue: "ROLE_ID",
					Description:  `The role ID (e.g. "roles/admin", "roles/contributor", "roles/viewer") to remove the member from.`,
					Value:        flagvalue.Simple("", &opts.Role),
					Required:     true,
					Autocomplete: iampolicy.AutocompleteRoles(opts.Ctx, ctx.Profile.OrganizationID, organization_service.New(ctx.HCP, nil)),
				},
			},
		},
		RunF: func(c *cmd.Command, args []string) error {
			// Create our IAM Updater
			u := &iamUpdater{
				orgID:  ctx.Profile.OrganizationID,
				client: organization_service.New(ctx.HCP, nil),
			}

			// Create the policy setter
			opts.Setter = iampolicy.NewSetter(
				ctx.Profile.OrganizationID,
				u,
				iam_service.New(ctx.HCP, nil),
				c.Logger())

			if runF != nil {
				return runF(opts)
			}

			return deleteBindingRun(opts)
		},
		PersistentPreRun: func(c *cmd.Command, args []string) error {
			return cmd.RequireOrganization(ctx)
		},
	}

	return cmd
}

type DeleteBindingOpts struct {
	Ctx context.Context
	IO  iostreams.IOStreams

	Setter      iampolicy.Setter
	PrincipalID string
	Role        string
}

func deleteBindingRun(opts *DeleteBindingOpts) error {
	_, err := opts.Setter.DeleteBinding(opts.Ctx, opts.PrincipalID, opts.Role)
	if err != nil {
		return err
	}

	fmt.Fprintf(opts.IO.Err(), "%s Principal %q binding to role %q deleted.\n",
		opts.IO.ColorScheme().SuccessIcon(), opts.PrincipalID, opts.Role)
	return nil
}
