// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/client/groups_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/client/iam_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/resource_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/models"
	"github.com/hashicorp/hcp/internal/commands/iam/groups/helper"
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

		GroupsClient:   groups_service.New(ctx.HCP, nil),
		ResourceClient: resource_service.New(ctx.HCP, nil),
	}

	cmd := &cmd.Command{
		Name:      "set-policy",
		ShortHelp: "Set the IAM policy for a group.",
		LongHelp: heredoc.New(ctx.IO).Must(`
The {{ template "mdCodeOrBold" "hcp iam groups iam set-policy" }} command sets
the IAM policy for the group, given a group name and a file encoded in
JSON that contains the IAM policy. If adding or removing a single principal from
the policy, prefer using {{ template "mdCodeOrBold" "hcp iam groups iam add-binding" }}
and the related {{ template "mdCodeOrBold" "hcp iam groups iam delete-binding" }}.

The policy file is expected to be a file encoded in JSON that
contains the IAM policy.

The format for the policy JSON file is an object with the following format:

{{ define "bindings" -}} {
	"bindings": [
		{
			"role_id": "ROLE_ID",
			"members": [
				{
					"member_id": "PRINCIPAL_ID",
					"member_type": "USER"
				}
			]
		}
	],
	"etag": "ETAG"
} {{- end }}
{{- CodeBlock "bindings" "json" }}

If set, the etag of the policy must be equal to that of the existing policy. To view the
existing policy and its etag, run {{ template "mdCodeOrBold" "hcp iam groups iam read-policy --format=json" }}.
If unset, the existing policy's etag will be fetched and used.

Note that the only supported member_type is {{ template "mdCodeOrBold" "USER" }} and the only supported role_id is {{ template "mdCodeOrBold" "roles/iam.group-manager" }}".
		`),
		Examples: []cmd.Example{
			{
				Preamble: "Set the IAM Policy for a group:",
				Command: heredoc.New(ctx.IO, heredoc.WithPreserveNewlines()).Must(`
					$ cat >policy.json <<EOF
					{
						"bindings": [
							{
								"role_id": "roles/iam.group-manager",
								"members": [
									{
										"member_id": "97e2c752-4285-419e-a5cc-bf05ce811d7d",
										"member_type": "USER"
									}
								]
							},
						],
						"etag": "14124142"
					}
					EOF
					$ hcp iam groups iam set-policy \
						--group=Group-Name
					  --policy-file=policy.json \
				`),
			},
		},
		Flags: cmd.Flags{
			Local: []*cmd.Flag{
				{
					Name:         "group",
					Shorthand:    "g",
					DisplayValue: "NAME",
					Description:  "The name of the group to add the role binding to.",
					Value:        flagvalue.Simple("", &opts.GroupName),
					Autocomplete: helper.PredictGroupResourceNameSuffix(opts.Ctx, opts.Profile.OrganizationID, opts.GroupsClient),
					Required:     true,
				},
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
			resourceName := helper.ResourceName(opts.GroupName, ctx.Profile.OrganizationID)

			// Create our group IAM Updater
			u := &iamUpdater{
				resourceName: resourceName,
				client:       opts.ResourceClient,
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
			return cmd.RequireOrganization(ctx)
		},
	}

	return cmd
}

type SetPolicyOpts struct {
	Ctx     context.Context
	Profile *profile.Profile
	IO      iostreams.IOStreams

	Setter         iampolicy.Setter
	GroupName      string
	PolicyFile     string
	GroupsClient   groups_service.ClientService
	ResourceClient resource_service.ClientService
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
