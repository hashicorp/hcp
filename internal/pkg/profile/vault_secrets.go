// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package profile

import (
	"errors"

	"github.com/posener/complete"
)

// VaultSecrets is a named set of configuration for the HCP CLI. It captures
// vault secrets related configuration values such as app name.
type VaultSecretsConf struct {
	// AppName stores the app name against which the requests will be made.
	AppName string `hcl:"app_name"`
}

// Predict predicts the HCL key names and basic settable values
func (vsc *VaultSecretsConf) Predict(args complete.Args) []string {
	return nil
}

// Validate validates that the set values are valid. It validates parameters
// that do not require any communication with HCP.
func (vsc *VaultSecretsConf) Validate() error {
	if vsc == nil {
		return nil
	}
	if vsc.AppName == "" {
		return errors.New("app_name must be set")
	}
	return nil
}

func (vsc *VaultSecretsConf) isEmpty() bool {
	if vsc == nil {
		return true
	}

	if vsc.AppName != "" {
		return false
	}

	return true
}
