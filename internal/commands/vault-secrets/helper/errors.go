// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package helper

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-vault-secrets/stable/2023-06-13/models"
	"google.golang.org/grpc/codes"
)

// FmtErr formats an error into a human-readable string.
func FmtErr(err error) string {
	type payloader interface {
		GetPayload() *models.RPCStatus
	}

	switch e := err.(type) {
	case payloader:
		return fmt.Sprintf("Error: %s", FormatRPCStatus(e.GetPayload()))
	default:
		if strings.Contains(err.Error(), "(*models.RPCStatus) is not supported by the TextConsumer") {
			return "Possible network reachability issues with HCP platform. Please try again!"
		}
		return "Error: " + err.Error()
	}
}

// FormatRPCStatus formats an RPC status into a human-readable string.
func FormatRPCStatus(status *models.RPCStatus) string {
	codeToHelp := map[codes.Code]string{
		// codes.OK: "",
		// codes.Canceled: "",
		// codes.Unknown: "",
		// codes.InvalidArgument: "",
		// codes.DeadlineExceeded: "",
		// codes.NotFound: "",
		// codes.AlreadyExists: "",
		// codes.ResourceExhausted: "",
		// codes.FailedPrecondition: "",
		// codes.Aborted: "",
		// codes.OutOfRange: "",
		// codes.Unimplemented: "",
		// codes.Unavailable: "",
		// codes.DataLoss: "",
		codes.Internal:         "An internal error has occurred and our engineers have been notified.",
		codes.Unauthenticated:  "Authentication Failed. Please try using the hcp auth login command to get started.",
		codes.PermissionDenied: "the project/organization ID values configured do not match try running `hcp profile init` or `hcp profile init --vault-secrets`",
	}

	if status == nil {
		return codeToHelp[codes.Internal]
	}

	help, ok := codeToHelp[codes.Code(uint32(status.Code))]
	if !ok {
		if status.Message == "" {
			return codes.Code(uint32(status.Code)).String()
		}
		return fmt.Sprintf("%s - %s", codes.Code(uint32(status.Code)).String(), status.Message)
	}
	return help
}
