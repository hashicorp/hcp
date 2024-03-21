package iam

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/client/iam_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/organization_service"
	"github.com/hashicorp/hcp/internal/pkg/api/iampolicy"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/flagvalue"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
)

func NewCmdAddBinding(ctx *cmd.Context, runF func(*AddBindingOpts) error) *cmd.Command {
	opts := &AddBindingOpts{
		Ctx:     ctx.ShutdownCtx,
		Profile: ctx.Profile,
		IO:      ctx.IO,
	}

	cmd := &cmd.Command{
		Name:      "add-binding",
		ShortHelp: "Add an IAM policy binding for the organization.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp organizations iam add-binding" }}
		command adds an IAM policy binding for the organization. A binding
		grants the specified principal the given role on the organization.

		To view the available roles to bind, run
		{{ template "mdCodeOrBold" "hcp iam roles list" }}.
		`),
		Examples: []cmd.Example{
			{
				Preamble: heredoc.New(ctx.IO).Must(`Bind a principal to role {{ template "mdCodeOrBold" "roles/viewer" }}:`),
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				$ hcp organizations iam add-binding \
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
					Description:  "The ID of the principal to add the role binding to.",
					Value:        flagvalue.Simple("", &opts.PrincipalID),
					Required:     true,
				},
				{
					Name:         "role",
					DisplayValue: "ROLE_ID",
					Description:  `The role ID (e.g. "roles/admin", "roles/contributor", "roles/viewer") to bind the member to.`,
					Value:        flagvalue.Simple("", &opts.Role),
					Required:     true,
					Autocomplete: iampolicy.AutocompleteRoles(opts.Ctx, opts.Profile.OrganizationID, organization_service.New(ctx.HCP, nil)),
				},
			},
		},
		RunF: func(c *cmd.Command, args []string) error {
			opts.Logger = c.Logger()

			// Create our IAM Updater
			u := &iamUpdater{
				orgID:  opts.Profile.OrganizationID,
				client: organization_service.New(ctx.HCP, nil),
			}

			// Create the policy setter
			opts.Setter = iampolicy.NewSetter(
				opts.Profile.OrganizationID,
				u,
				iam_service.New(ctx.HCP, nil),
				c.Logger(),
			)

			if runF != nil {
				return runF(opts)
			}

			return addIAMBindingRun(opts)
		},
		PersistentPreRun: func(c *cmd.Command, args []string) error {
			return cmd.RequireOrganization(ctx)
		},
	}

	return cmd
}

type AddBindingOpts struct {
	Ctx     context.Context
	Logger  hclog.Logger
	Profile *profile.Profile
	IO      iostreams.IOStreams

	Setter      iampolicy.Setter
	PrincipalID string
	Role        string
}

func addIAMBindingRun(opts *AddBindingOpts) error {
	_, err := opts.Setter.AddBinding(opts.Ctx, opts.PrincipalID, opts.Role)
	if err != nil {
		return err
	}

	fmt.Fprintf(opts.IO.Err(), "%s Principal %q bound to role %q.\n",
		opts.IO.ColorScheme().SuccessIcon(), opts.PrincipalID, opts.Role)
	return nil
}
