package organizations

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
		ShortHelp: "List organizations.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		"The {{ template "mdCodeOrBold" "hcp organizations list" }} command
		lists the organizations the authenticated principal is a member of."
		`),
		RunF: func(c *cmd.Command, args []string) error {
			if runF != nil {
				return runF(opts)
			}
			return listRun(opts)
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
	req := organization_service.NewOrganizationServiceListParamsWithContext(opts.Ctx)

	var orgs []*models.HashicorpCloudResourcemanagerOrganization
	for {
		resp, err := opts.Client.OrganizationServiceList(req, nil)
		if err != nil {
			return fmt.Errorf("failed to list organizations: %w", err)
		}

		orgs = append(orgs, resp.Payload.Organizations...)
		if resp.Payload.Pagination == nil || resp.Payload.Pagination.NextPageToken == "" {
			break
		}

		next := resp.Payload.Pagination.NextPageToken
		req.PaginationNextPageToken = &next
	}

	d := newDisplayer(format.Table, false, orgs...)
	return opts.Output.Display(d)
}
