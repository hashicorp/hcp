package iam

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/client/iam_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/project_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/models"
	"github.com/hashicorp/hcp/internal/pkg/api/iampolicy"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/flagvalue"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
	"github.com/posener/complete"
)

func NewCmdSetPolicy(ctx *cmd.Context, runF func(*SetPolicyOpts) error) *cmd.Command {
	opts := &SetPolicyOpts{
		Ctx:     ctx.ShutdownCtx,
		Profile: ctx.Profile,
		IO:      ctx.IO,
	}

	cmd := &cmd.Command{
		Name:      "set-policy",
		ShortHelp: "Set the IAM policy for a project.",
		LongHelp: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
		    Sets the IAM policy for a project, given a project ID and a file encoded in
			JSON that contains the IAM policy.

			The format for the policy JSON file is an object with the following format:

			{
			  "bindings": [
			    {
			      "role_id": "ROLE_ID",
			      "members": [
				    {
			          "member_id": "PRINCIPAL_ID",
			          "member_type": "USER" | "GROUP" | "SERVICE_PRINCIPAL",
				    }
			      ]
			    }
			  ],
			  "etag": "ETAG",
			}

			If set, the etag of the policy must be equal to that of the existing policy. To view the
			existing policy and its etag, run {{ Bold "hcp projects iam read-policy --format=json" }}.
			If unset, the existing policy's etag will be fetched and used.
		`),
		Examples: []cmd.Example{
			{
				Preamble: "Set the IAM Policy for a project",
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
					$ cat >policy.json <<EOF
					{
					  "bindings": [
						{
						  "role_id": "roles/viewer",
						  "members": [
						    {
							  "member_id": "97e2c752-4285-419e-a5cc-bf05ce811d7d",
							  "member_type": "USER"
						    },
						    {
							  "member_id": "334514c1-4650-4699-891a-a7261fba9607",
							  "member_type": "GROUP"
						    }
						  ]
						},
						{
						  "role_id": "roles/admin",
						  "members": [
						    {
							  "member_id": "efa07942-17e8-4ef4-ae2d-ec51d32a0767",
							  "member_type": "SERVICE_PRINCIPAL"
						    }
						  ]
						}
					  ],
					  "etag": "14124142"
					}
					EOF
					$ hcp projects iam set-policy \
					  --policy-file=policy.json \
					  --project=8647ae06-ca65-467a-b72d-edba1f908fc8
				`),
			},
		},
		Flags: cmd.Flags{
			Local: []*cmd.Flag{
				{
					Name:         "policy-file",
					DisplayValue: "PATH",
					Description:  "The path to a file containing an IAM policy object.",
					Value:        flagvalue.Simple("", &opts.PolicyFile),
					Required:     true,
					Autocomplete: complete.PredictFiles("*.json"),
				},
			},
		},
		RunF: func(c *cmd.Command, args []string) error {
			// Create our project IAM Updater
			u := &iamUpdater{
				projectID: opts.Profile.ProjectID,
				client:    project_service.New(ctx.HCP, nil),
			}

			// Create the policy setter
			opts.Setter = iampolicy.NewSetter(
				opts.Profile.OrganizationID,
				u,
				iam_service.New(ctx.HCP, nil),
				c.Logger())

			if runF != nil {
				return runF(opts)
			}

			return setPolicyRun(opts)
		},
		PersistentPreRun: func(c *cmd.Command, args []string) error {
			return cmd.RequireOrgAndProject(ctx)
		},
	}

	return cmd
}

type SetPolicyOpts struct {
	Ctx     context.Context
	Profile *profile.Profile
	IO      iostreams.IOStreams

	Setter     iampolicy.Setter
	PolicyFile string
}

func setPolicyRun(opts *SetPolicyOpts) error {
	// Open the file
	f, err := os.Open(opts.PolicyFile)
	if err != nil {
		return fmt.Errorf("failed to open policy file: %w", err)
	}

	var p models.HashicorpCloudResourcemanagerPolicy
	d := json.NewDecoder(f)
	d.DisallowUnknownFields()
	if err := d.Decode(&p); err != nil {
		return fmt.Errorf("failed to unmarshal policy file: %w", err)
	}

	// Get the existing policy
	_, err = opts.Setter.SetPolicy(opts.Ctx, &p)
	if err != nil {
		return err
	}

	fmt.Fprintf(opts.IO.Err(), "%s IAM Policy successfully set.\n", opts.IO.ColorScheme().SuccessIcon())
	return nil
}
