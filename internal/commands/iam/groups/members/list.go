package members

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/client/groups_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/models"
	"github.com/hashicorp/hcp/internal/commands/iam/groups/helper"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/flagvalue"
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
		Client:  groups_service.New(ctx.HCP, nil),
	}

	cmd := &cmd.Command{
		Name:      "list",
		ShortHelp: "List the the members of a group.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp iam groups members list" }} command lists the members of a group.
		`),
		Examples: []cmd.Example{
			{
				Preamble: `List the members of "team-platform":`,
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				$ hcp iam groups members list --group=team-platform \
				  --description "Team Platform engineering group"
				`),
			},
		},
		Flags: cmd.Flags{
			Local: []*cmd.Flag{
				{
					Name:         "group",
					Shorthand:    "g",
					DisplayValue: "NAME",
					Description:  heredoc.New(ctx.IO).Mustf(helper.GroupNameArgDoc, "list membership from"),
					Value:        flagvalue.Simple("", &opts.GroupName),
					Autocomplete: helper.PredictGroupResourceNameSuffix(opts.Ctx, opts.Profile.OrganizationID, opts.Client),
				},
			},
		},
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

	GroupName string
	Client    groups_service.ClientService
}

func listRun(opts *ListOpts) error {
	req := groups_service.NewGroupsServiceListGroupMembersParamsWithContext(opts.Ctx)
	req.ResourceName = helper.ResourceName(opts.GroupName, opts.Profile.OrganizationID)

	var groups membersDisplayer
	for {

		resp, err := opts.Client.GroupsServiceListGroupMembers(req, nil)
		if err != nil {
			return fmt.Errorf("failed to list group members: %w", err)
		}
		groups = append(groups, resp.Payload.Members...)
		if resp.Payload.Pagination == nil || resp.Payload.Pagination.NextPageToken == "" {
			break
		}

		next := resp.Payload.Pagination.NextPageToken
		req.PaginationNextPageToken = &next
	}

	return opts.Output.Display(groups)
}

type membersDisplayer []*models.HashicorpCloudIamGroupMember

func (d membersDisplayer) DefaultFormat() format.Format {
	return format.Table
}

func (d membersDisplayer) Payload() any {
	return d
}

func (d membersDisplayer) FieldTemplates() []format.Field {
	return []format.Field{
		{
			Name:        "Name",
			ValueFormat: "{{ .Name }}",
		},
		{
			Name:        "ID",
			ValueFormat: "{{ .ID }}",
		},
		{
			Name:        "Email",
			ValueFormat: "{{ .Email }}",
		},
	}
}
