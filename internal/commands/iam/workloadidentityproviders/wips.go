// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package workloadidentityproviders

import (
	"regexp"

	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
)

const (
	// WIPNameArgDoc is the documentation for accepting a workload identity
	// provider name as an argument.
	WIPNameArgDoc = `
	The resource name of the workload identity provider to %s. The format of the resource name is
	{{ Code "iam/project/PROJECT_ID/service-principal/SP_NAME/workload-identity-provider/WIP_NAME" }}.
	`
)

var (
	// WIPResourceName is a regex that matches a workload identity provider resource name
	WIPResourceName = regexp.MustCompile(`^iam/project/.+/service-principal/.+/workload-identity-provider/.+$`)
)

func NewCmdWIPs(ctx *cmd.Context) *cmd.Command {
	cmd := &cmd.Command{
		Name:      "workload-identity-providers",
		Aliases:   []string{"wips"},
		ShortHelp: "Manage Workload Identity Providers.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp iam workload-identity-providers" }} command group
		lets you create and manage Workload Identity Providers.

		Creating a workload identity provider creates a trust relationship
		between HCP and an external identity provider. Once created, a workload
		can exchange its external identity token for an HCP access token.

		HCP supports federating with AWS or any OIDC identity provider. This allows exchanging
		identity credentials for workloads running on AWS, GCP, Azure, GitHub Actions, Kubernetes,
		and more for an HCP Service Principal access token without having to store service principal
		credentials.

		To make exchanging external credentials as easy as possible, create a credential file using
		{{ template "mdCodeOrBold" "hcp iam workload-identity-providers create-cred-file" }}
		after creating your provider.

		The credential file contains details on how to source the external identity token and exchange
		it for an HCP access token. The {{ template "mdCodeOrBold" "hcp" }} CLI can be authenticated
		using a credential file by running {{ template "mdCodeOrBold" "hcp auth login --cred-file" }}.
		For programatic access, the HCP Go SDK can be used and authenticated using a credential file.
		`),
	}

	cmd.AddChild(NewCmdCreateAWS(ctx, nil))
	cmd.AddChild(NewCmdCreateOIDC(ctx, nil))
	cmd.AddChild(NewCmdCreateCredFile(ctx, nil))
	cmd.AddChild(NewCmdDelete(ctx, nil))
	cmd.AddChild(NewCmdList(ctx, nil))
	cmd.AddChild(NewCmdRead(ctx, nil))

	return cmd
}
