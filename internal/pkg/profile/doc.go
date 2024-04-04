// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

// Package profile stores the CLI configuration for the named profile.
//
// The profile stores common configuration values such as the organization and
// project ID to use when running commands. It also stores user settable values
// that customize CLI output; examples are setting the default output format or
// disabling color output.
//
// Individual services may store their configuration in the profile as well,
// within a detected stanza.
package profile
