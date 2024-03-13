package users

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/client/iam_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/models"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/hashicorp/hcp/internal/pkg/profile"
)

func NewCmdList(ctx *cmd.Context, runF func(*ListOpts) error) *cmd.Command {
	opts := &ListOpts{
		Ctx:     ctx.ShutdownCtx,
		Profile: ctx.Profile,
		Output:  ctx.Output,
		Client:  iam_service.New(ctx.HCP, nil),
	}

	cmd := &cmd.Command{
		Name:      "list",
		ShortHelp: "List the organization's users.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp iam users list" }} command lists the users for an HCP organization.
		`),
		RunF: func(c *cmd.Command, args []string) error {
			if runF != nil {
				return runF(opts)
			}

			return listRun(opts)
		},
		PersistentPreRun: func(c *cmd.Command, args []string) error {
			return cmd.RequireOrganization(ctx)
		},
	}

	return cmd
}

type ListOpts struct {
	Ctx     context.Context
	Profile *profile.Profile
	Output  *format.Outputter

	Client iam_service.ClientService
}

func listRun(opts *ListOpts) error {
	req := iam_service.NewIamServiceListUserPrincipalsByOrganizationParamsWithContext(opts.Ctx)
	req.OrganizationID = opts.Profile.OrganizationID

	var users []*models.HashicorpCloudIamUserPrincipal
	for {

		resp, err := opts.Client.IamServiceListUserPrincipalsByOrganization(req, nil)
		if err != nil {
			return fmt.Errorf("failed to list users: %w", err)
		}
		users = append(users, resp.Payload.UserPrincipals...)
		if resp.Payload.Pagination == nil || resp.Payload.Pagination.NextPageToken == "" {
			break
		}

		next := resp.Payload.Pagination.NextPageToken
		req.PaginationNextPageToken = &next
	}

	return opts.Output.Display(newDisplayer(format.Table, false, users...))
}
