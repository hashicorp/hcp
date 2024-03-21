package tfcconfig

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2023-08-18/client/waypoint_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2023-08-18/models"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
)

func NewCmdCreate(ctx *cmd.Context, runF func(opts *TFCConfigCreateOpts) error) *cmd.Command {
	opts := &TFCConfigCreateOpts{
		Ctx:            ctx.ShutdownCtx,
		Output:         ctx.Output,
		Profile:        ctx.Profile,
		IO:             ctx.IO,
		WaypointClient: waypoint_service.New(ctx.HCP, nil),
	}

	cmd := &cmd.Command{
		Name:      "create",
		ShortHelp: "Set TFC Configuration.",
		LongHelp: heredoc.New(ctx.IO).Mustf(`
		The {{template "mdCodeOrBold" "hcp waypoint tfc-config create"}} command
		sets the TFC Organization Name and TFC Team token that will be used in Waypoint.

		There can only be one TFC Config set for each HCP Project.

		TFC Configs can be reviewed using the {{template "mdCodeOrBold" "hcp waypoint tfc-config read" }} command
		and removed with the {{template "mdCodeOrBold" "hcp waypoint tfc-config delete"}} command.
		`),
		Examples: []cmd.Example{
			{
				Preamble: `Create a new TFC Config in HCP Waypoint:`,
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
				$ hcp waypoint tfc-config create example-org <token>`),
			},
		},
		Args: cmd.PositionalArguments{
			Args: []cmd.PositionalArgument{
				{
					Name:          "TFC_ORG",
					Documentation: heredoc.New(ctx.IO).Must(`Name of the Terraform Cloud Organization.`),
				},
				{
					Name: "TOKEN",
					Documentation: heredoc.New(ctx.IO).Must(`
                    Terraform Cloud Team token for the TFC organization.

                    Team token must be set in order to perform HCP Waypoint commands.

                    Refer to the {{ Link "API tokens documentation" "https://developer.hashicorp.com/terraform/cloud-docs/users-teams-organizations/api-tokens" }} to learn more.

                    HCP Waypoint requires Team level access tokens in order to run correctly.

                    Please ensure that your TFCConfig token has the correct permissions.
					`),
				},
			},
		},
		RunF: func(c *cmd.Command, args []string) error {
			opts.TfcOrg = args[0]
			opts.Token = args[1]
			if runF != nil {
				return runF(opts)
			}
			return createRun(opts)
		},
		PersistentPreRun: func(c *cmd.Command, args []string) error {
			return cmd.RequireOrgAndProject(ctx)
		},
	}
	return cmd
}

type TFCConfigCreateOpts struct {
	Ctx     context.Context
	Profile *profile.Profile
	Output  *format.Outputter
	IO      iostreams.IOStreams

	TfcOrg         string
	Token          string
	WaypointClient waypoint_service.ClientService
}

func createRun(opts *TFCConfigCreateOpts) error {
	nsID, err := GetNamespace(opts.Ctx, opts.WaypointClient, opts.Profile.OrganizationID, opts.Profile.ProjectID)
	if err != nil {
		return fmt.Errorf("error getting namespace: %w", err)
	}

	ns := &models.HashicorpCloudWaypointRefNamespace{ID: nsID}
	resp, err := opts.WaypointClient.WaypointServiceCreateTFCConfig(
		&waypoint_service.WaypointServiceCreateTFCConfigParams{
			Body: &models.HashicorpCloudWaypointWaypointServiceCreateTFCConfigBody{
				Namespace: ns,
				TfcConfig: &models.HashicorpCloudWaypointTFCConfig{
					OrganizationName: opts.TfcOrg,
					Token:            opts.Token,
				},
			},
			NamespaceID: nsID,
			Context:     opts.Ctx,
		}, nil,
	)
	if err != nil {
		return fmt.Errorf("error creating TFC config: %w", err)

	}

	fmt.Fprintf(opts.IO.Err(), "%s TFC Config  %q created!\n", opts.IO.ColorScheme().SuccessIcon(), resp.Payload.TfcConfig.OrganizationName)

	return nil

}
