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

func NewCmdSet(ctx *cmd.Context, runF func(opts *TFCConfigOpts) error) *cmd.Command {
	opts := &TFCConfigOpts{
		Ctx:            ctx.ShutdownCtx,
		Output:         ctx.Output,
		Profile:        ctx.Profile,
		IO:             ctx.IO,
		WaypointClient: waypoint_service.New(ctx.HCP, nil),
	}

	cmd := &cmd.Command{
		Name:      "set",
		ShortHelp: "Set TFC Config properties TFC Organization Name and TFC Team Token.",
		LongHelp: heredoc.New(ctx.IO).Mustf(`
        {{Bold "tfc-config set"}} Use: <<command>>{{ Bold "tfc-config set <<tfcOrg>> <<token>>" }}
		Set Terraform Cloud configuration, requires TFC Organization Name and TFC Team Token. There is only one 
		TFCConfig allowed per HCP Project ID.`),
		Args: cmd.PositionalArguments{
			Args: []cmd.PositionalArgument{
				{
					Name:          "TFC_ORG",
					Documentation: heredoc.New(ctx.IO).Must(`Name of the Terraform Cloud Organization`),
				},
				{
					Name: "TOKEN",
					Documentation: heredoc.New(ctx.IO).Must(`Terraform Cloud Team token for the TFC organization. 
						Team token must be set in order to perform HCP Waypoint commands. You can learn more about API tokens 
						at https://developer.hashicorp.com/terraform/cloud-docs/users-teams-organizations/api-tokens
						HCP Waypoint requires Team level access tokens in order to run correctly. Please ensure that your
						TFCConfig token has the correct permissions.`),
				},
			},
		},
		RunF: func(c *cmd.Command, args []string) error {
			opts.TfcOrg = args[0]
			opts.Token = args[1]
			if runF != nil {
				return runF(opts)
			}
			return setRun(opts)
		},
		PersistentPreRun: func(c *cmd.Command, args []string) error {
			return cmd.RequireOrgAndProject(ctx)
		},
	}
	return cmd
}

type TFCConfigOpts struct {
	Ctx     context.Context
	Profile *profile.Profile
	Output  *format.Outputter
	IO      iostreams.IOStreams

	TfcOrg         string
	Token          string
	WaypointClient waypoint_service.ClientService
}

func setRun(opts *TFCConfigOpts) error {
	nsID, err := GetNamespace(opts.Ctx, opts.WaypointClient, opts.Profile.OrganizationID, opts.Profile.ProjectID)
	if err != nil {
		return fmt.Errorf("error getting namespace: %s", err)
	}

	ns := &models.HashicorpCloudWaypointRefNamespace{ID: nsID}
	request := &models.HashicorpCloudWaypointCreateTFCConfigRequest{
		Namespace: ns,
		TfcConfig: &models.HashicorpCloudWaypointTFCConfig{
			OrganizationName: opts.TfcOrg,
			Token:            opts.Token,
		},
	}
	resp, err := opts.WaypointClient.WaypointServiceCreateTFCConfig(
		&waypoint_service.WaypointServiceCreateTFCConfigParams{
			Body:        request,
			NamespaceID: nsID,
			Context:     opts.Ctx,
		}, nil,
	)
	if err != nil {
		return fmt.Errorf("error setting TFC config: %s", err)

	}

	fmt.Fprintf(opts.IO.Err(), "%s TFC Config  %q updated\n", opts.IO.ColorScheme().SuccessIcon(), resp.Payload.TfcConfig.OrganizationName)

	return nil

}
