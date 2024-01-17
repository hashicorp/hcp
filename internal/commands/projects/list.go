package projects

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/project_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/models"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
)

func NewCmdList(ctx *cmd.Context, runF func(*ListOpts) error) *cmd.Command {
	opts := &ListOpts{
		Ctx:     ctx.ShutdownCtx,
		Profile: ctx.Profile,
		IO:      ctx.IO,
		Output:  ctx.Output,
		Client:  project_service.New(ctx.HCP, nil),
	}

	cmd := &cmd.Command{
		Name:      "list",
		ShortHelp: "List HCP projects.",
		LongHelp:  "List HCP projects.",
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

	Client project_service.ClientService
}

func listRun(opts *ListOpts) error {
	req := project_service.NewProjectServiceListParamsWithContext(opts.Ctx)
	req.ScopeID = &opts.Profile.OrganizationID
	req.ScopeType = (*string)(models.HashicorpCloudResourcemanagerResourceIDResourceTypeORGANIZATION.Pointer())

	// TODO use access aware list.
	var projects []*models.HashicorpCloudResourcemanagerProject
	for {
		resp, err := opts.Client.ProjectServiceList(req, nil)
		if err != nil {
			return fmt.Errorf("failed to list projects: %w", err)
		}

		projects = append(projects, resp.Payload.Projects...)
		if resp.Payload.Pagination == nil || resp.Payload.Pagination.NextPageToken == "" {
			break
		}

		next := resp.Payload.Pagination.NextPageToken
		req.PaginationNextPageToken = &next
	}

	d := newDisplayer(format.Table, false, projects...)
	return opts.Output.Display(d)
}
