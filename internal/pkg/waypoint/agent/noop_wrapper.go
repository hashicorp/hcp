// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package agent

import (
	"context"

	"github.com/hashicorp/go-hclog"
)

type NoopWrapper struct {
	Operation Operation
}

func (n *NoopWrapper) Run(ctx context.Context, log hclog.Logger) (OperationStatus, error) {
	return cleanStatus, nil
}

func NoopOperations(o Operation) Operation {
	return &NoopWrapper{Operation: o}
}
