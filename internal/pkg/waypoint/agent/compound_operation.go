// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package agent

import (
	"context"

	"github.com/hashicorp/go-hclog"
)

type CompoundOperation struct {
	Operations []Operation
}

func (c *CompoundOperation) Run(ctx context.Context, log hclog.Logger) (OperationStatus, error) {
	for _, op := range c.Operations {
		code, err := op.Run(ctx, log)
		if err != nil {
			return code, err
		}
	}

	return cleanStatus, nil
}
