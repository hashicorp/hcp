// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package auth

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/hcp-sdk-go/auth"
	"github.com/hashicorp/hcp-sdk-go/config"
	hcpAuth "github.com/hashicorp/hcp/internal/pkg/auth"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/stretchr/testify/require"
)

func TestLogout(t *testing.T) {
	t.Parallel()
	r := require.New(t)

	loggedOut := false
	io := iostreams.Test()
	opts := &LogoutOpts{
		IO:            io,
		CredentialDir: t.TempDir(),
		logoutFn: func(conf config.HCPConfig) error {
			loggedOut = true
			r.NotNil(conf)
			return nil
		},
	}

	// Run logout without any credentials
	r.NoError(logoutRun(opts))
	r.Equal(io.Error.String(), "Successfully logged out\n")
	r.True(loggedOut)

	// Write a credential file to disk
	credFilePath := filepath.Join(opts.CredentialDir, hcpAuth.CredFileName)
	f, err := os.Create(credFilePath)
	r.NoError(err)
	defer func() { r.NoError(f.Close()) }()

	// Write a credential file
	cf := auth.CredentialFile{
		Scheme: auth.CredentialFileSchemeServicePrincipal,
		Oauth: &auth.OauthConfig{
			ClientID:     "123",
			ClientSecret: "456",
		},
	}
	r.NoError(json.NewEncoder(f).Encode(cf))

	// Run logout again
	r.NoError(logoutRun(opts))

	// Ensure the credfile was removed
	_, err = os.Stat(credFilePath)
	r.ErrorIs(err, fs.ErrNotExist)
}
