// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package profile

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProfile_IsValidProperty(t *testing.T) {
	t.Parallel()
	r := require.New(t)
	r.ErrorContains(IsValidProperty("project_dd"), "project_id")
	r.ErrorContains(IsValidProperty("core/no_colr"), "core/no_color")
}
