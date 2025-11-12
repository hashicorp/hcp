// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: MPL-2.0

package appname

import (
	"errors"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/flagvalue"
)

var (
	// appName is the name of the Vault Secrets application. If not specified,
	// then the value from the active profile will be used.
	appName string
)

func set(val string) {
	appName = val
}

// Get returns the Vault Secrets application name.
func Get() string {
	return appName
}

// Flag returns a flag value for the Vault Secrets application name.
func Flag() flagvalue.Value {
	return flagvalue.Simple("", &appName)
}

// Require requires that the profile has a set organization and project ID along with
// the Vault Secrets application name.
func Require(ctx *cmd.Context) error {
	err := cmd.RequireOrgAndProject(ctx)
	if err != nil {
		return err
	}

	if appName == "" && ctx.Profile.VaultSecrets != nil {
		set(ctx.Profile.VaultSecrets.AppName)
	}

	if appName != "" || ctx.Profile.VaultSecrets != nil && ctx.Profile.VaultSecrets.AppName != "" {
		return nil
	}

	cs := ctx.IO.ColorScheme()
	help := heredoc.Docf(`%v

	Set the app name using the --app flag or set the app name on your active profile one of the following commands:

	%v
	%v`,
		cs.String("Vault Secrets application name must set.").Color(cs.Orange()),
		cs.String("$ hcp profile set vault-secrets/app <app_name>").Bold(),
		cs.String("$ hcp profile init --vault-secrets").Bold(),
	)

	return errors.New(help)
}
