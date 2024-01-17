package roles

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/organization_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/models"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
)

func NewCmdList(ctx *cmd.Context, runF func(*ListOpts) error) *cmd.Command {
	opts := &ListOpts{
		Ctx:     ctx.ShutdownCtx,
		Profile: ctx.Profile,
		IO:      ctx.IO,
		Output:  ctx.Output,
		Client:  organization_service.New(ctx.HCP, nil),
	}
	cmd := &cmd.Command{
		Name:      "list",
		ShortHelp: "List an organization's roles.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ Bold "hcp iam roles list" }} command lists the roles that exist for an HCP organization.

		When referring to a role in an IAM binding, use the role's ID (e.g. "roles/admin").
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
	IO      iostreams.IOStreams
	Output  *format.Outputter

	Client organization_service.ClientService
}

func listRun(opts *ListOpts) error {
	req := organization_service.NewOrganizationServiceListRolesParamsWithContext(opts.Ctx)
	req.ID = opts.Profile.OrganizationID

	var roles []*models.HashicorpCloudResourcemanagerRole
	for {

		resp, err := opts.Client.OrganizationServiceListRoles(req, nil)
		if err != nil {
			return fmt.Errorf("failed to list organization roles: %w", err)
		}
		roles = append(roles, resp.Payload.Roles...)
		if resp.Payload.Pagination == nil || resp.Payload.Pagination.NextPageToken == "" {
			break
		}

		next := resp.Payload.Pagination.NextPageToken
		req.PaginationNextPageToken = &next
	}

	return opts.Output.Display(rolesDisplayer(roles))
}

type rolesDisplayer []*models.HashicorpCloudResourcemanagerRole

func (d rolesDisplayer) DefaultFormat() format.Format { return format.Table }
func (d rolesDisplayer) Payload() any                 { return d }

func (d rolesDisplayer) FieldTemplates() []format.Field {
	return []format.Field{
		{
			Name:        "ID",
			ValueFormat: "{{ .ID }}",
		},
		{
			Name:        "Title",
			ValueFormat: "{{ .Title }}",
		},
		{
			Name:        "Description",
			ValueFormat: "{{ .Description }}",
		},
	}
}
