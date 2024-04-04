// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/go-openapi/runtime/client"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/models"
	"github.com/hashicorp/hcp/internal/pkg/api/iampolicy"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewCmdSetPolicy(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Name    string
		Args    []string
		Profile func(t *testing.T) *profile.Profile
		Error   string
		Expect  *SetPolicyOpts
	}{
		{
			Name:    "No Org",
			Profile: profile.TestProfile,
			Args:    []string{"--policy-file=test-policy.json"},
			Error:   "Organization ID must be configured",
		},
		{
			Name: "Too many args",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123")
			},
			Args:  []string{"--policy-file=test-policy.json", "foo", "bar"},
			Error: "no arguments allowed, but received 2",
		},
		{
			Name: "missing flag",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123")
			},
			Error: "missing required flag: --policy-file=PATH",
		},
		{
			Name: "Good",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123")
			},
			Args: []string{"--policy-file=test-policy.json"},
			Expect: &SetPolicyOpts{
				PolicyFile: "test-policy.json",
			},
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			r := require.New(t)

			// Create a context.
			io := iostreams.Test()
			ctx := &cmd.Context{
				IO:          io,
				Profile:     c.Profile(t),
				Output:      format.New(io),
				HCP:         &client.Runtime{},
				ShutdownCtx: context.Background(),
			}

			var gotOpts *SetPolicyOpts
			cmd := NewCmdSetPolicy(ctx, func(o *SetPolicyOpts) error {
				gotOpts = o
				return nil
			})
			cmd.SetIO(io)

			code := cmd.Run(c.Args)
			if c.Error != "" {
				r.NotZero(code)
				r.Contains(io.Error.String(), c.Error)
				return
			}

			r.Zero(code, io.Error.String())
			r.NotNil(gotOpts)
			r.Equal(c.Expect.PolicyFile, gotOpts.PolicyFile)
			r.NotNil(gotOpts.Setter)
		})
	}
}

func TestSetPolicyRun(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Name         string
		RespErr      error
		FileContent  string
		ExpectPolicy *models.HashicorpCloudResourcemanagerPolicy
		Error        string
		ParseError   bool
	}{
		{
			Name:        "Server error",
			FileContent: "{}",
			RespErr:     fmt.Errorf("failed to add policy"),
			Error:       "failed to add policy",
		},
		{
			Name: "bad file",
			FileContent: `{
  "bindingz": []
}`,
			ParseError: true,
			Error:      "failed to unmarshal policy file: json: unknown field \"bindingz\"",
		},
		{
			Name: "Good",
			FileContent: `{
  "bindings": [
    {
      "role_id": "roles/admin",
	  "members": [
	    {
		  "member_id": "8b70fa37-27d1-4618-b4ea-553fb776a6b6",
		  "member_type": "USER"
	    }
	  ]
    }
  ]
}`,
			ExpectPolicy: &models.HashicorpCloudResourcemanagerPolicy{
				Bindings: []*models.HashicorpCloudResourcemanagerPolicyBinding{
					{
						Members: []*models.HashicorpCloudResourcemanagerPolicyBindingMember{
							{
								MemberID:   "8b70fa37-27d1-4618-b4ea-553fb776a6b6",
								MemberType: models.HashicorpCloudResourcemanagerPolicyBindingMemberTypeUSER.Pointer(),
							},
						},
						RoleID: "roles/admin",
					},
				},
			},
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			r := require.New(t)

			io := iostreams.Test()
			setter := iampolicy.NewMockSetter(t)
			opts := &SetPolicyOpts{
				Ctx:    context.Background(),
				IO:     io,
				Setter: setter,
			}

			// Write the policy file
			f, err := os.CreateTemp("", "")
			r.NoError(err)
			defer func() {
				_ = os.Remove(f.Name())
			}()

			_, err = f.WriteString(c.FileContent)
			r.NoError(err)
			opts.PolicyFile = f.Name()

			if !c.ParseError {
				if c.RespErr != nil {
					setter.EXPECT().SetPolicy(mock.Anything, mock.Anything).Once().Return(nil, c.RespErr)
				} else {
					setter.EXPECT().SetPolicy(mock.Anything, mock.MatchedBy(func(policy *models.HashicorpCloudResourcemanagerPolicy) bool {
						return reflect.DeepEqual(policy, c.ExpectPolicy)
					})).Once().Return(&models.HashicorpCloudResourcemanagerPolicy{}, nil)
				}
			}

			// Run the command
			err = setPolicyRun(opts)
			if c.Error != "" {
				r.ErrorContains(err, c.Error)
				return
			}

			// Check we outputted
			r.NoError(err)
			r.Contains(io.Error.String(), `IAM Policy successfully set`)
		})
	}
}
