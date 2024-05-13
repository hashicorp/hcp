// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cmd

import (
	"errors"

	"github.com/MakeNowJust/heredoc/v2"
)

// RequireOrganization requires that the profile has a set organization ID.
func RequireOrganization(ctx *Context) error {
	if ctx.Profile.OrganizationID != "" {
		return nil
	}

	cs := ctx.IO.ColorScheme()
	help := heredoc.Docf(`%v

	Please run %v to interactively set the Organization ID, or run:

	%v`,
		cs.String("Organization ID must be configured before running the command.").Color(cs.Orange()),
		cs.String("hcp profile init").Bold(),
		cs.String("$ hcp profile set organization_id <organization_id>").Bold(),
	)

	return errors.New(help)
}
