package versioncheck

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/hashicorp/hcp/internal/pkg/api/releasesapi/client/operations"
	mock_operations "github.com/hashicorp/hcp/internal/pkg/api/releasesapi/mocks"
	"github.com/hashicorp/hcp/internal/pkg/api/releasesapi/models"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type tester struct {
	c         *Checker
	io        *iostreams.Testing
	client    *mock_operations.MockClientService
	statePath string
}

func testChecker(t *testing.T, currentVersion string) *tester {
	return testCheckerWithStatePath(t, currentVersion, filepath.Join(t.TempDir(), "state.json"))
}

func testCheckerWithStatePath(t *testing.T, currentVersion, statePath string) *tester {
	t.Helper()
	tc := &tester{
		io:        iostreams.Test(),
		client:    mock_operations.NewMockClientService(t),
		statePath: statePath,
	}
	var err error
	tc.c, err = newChecker(tc.io, tc.statePath, currentVersion, tc.client, true)
	if err != nil {
		t.Fatalf("failed to create checker: %s", err)
	}

	return tc
}

func TestChecker_New(t *testing.T) {
	t.Parallel()
	r := require.New(t)
	path := filepath.Join(t.TempDir(), "newdir", "state.json")
	testCheckerWithStatePath(t, "1.2.3", path)

	// Check the dir was created.
	r.DirExists(filepath.Dir(path))
}

func TestChecker_Check_CancelCtx(t *testing.T) {
	t.Parallel()
	r := require.New(t)
	tc := testChecker(t, "1.2.3")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	tc.client.EXPECT().ListReleasesV1(mock.Anything, mock.Anything).Once().Return(nil, context.Canceled)
	err := tc.c.Check(ctx)
	r.NoError(err)
}

func TestChecker_Check_InCI(t *testing.T) {
	cases := []string{"CI", "BUILD_NUMBER", "RUN_ID"}
	for _, c := range cases {
		t.Run(c, func(t *testing.T) {
			r := require.New(t)
			tc := testChecker(t, "1.2.3")
			tc.c.skipCICheck = false
			t.Setenv(c, "1")
			r.NoError(tc.c.Check(context.Background()))
		})
	}
}

func TestChecker_Check_RecentUpdate(t *testing.T) {
	t.Parallel()
	r := require.New(t)
	tc := testChecker(t, "1.2.3")

	// Create a state file
	checkState := &versionCheckState{
		CheckedAt: time.Now(),
		path:      tc.statePath,
	}
	r.NoError(checkState.write())

	// Check if there is an upgrade available
	r.NoError(tc.c.Check(context.Background()))
}

func stringToPtr(s string) *string {
	return &s
}

func TestChecker_Check(t *testing.T) {
	t.Parallel()
	r := require.New(t)
	tc := testChecker(t, "1.2.3")

	// Create a state file
	checkState := &versionCheckState{
		CheckedAt: time.Now().Add(-time.Hour * 24 * 7), // 1 week ago
		path:      tc.statePath,
	}
	r.NoError(checkState.write())

	// Mock the API call
	expectedVersion := "1.2.5"
	changelog := "https://github.com/hashicorp/hcp/REAMDE.md"
	results := []*models.ProductReleaseResponseV1{
		{
			Version:      stringToPtr("1.3.0-beta1"),
			IsPrerelease: true,
			URLChangelog: changelog,
		},
		{
			Version:      stringToPtr(expectedVersion),
			IsPrerelease: false,
			URLChangelog: changelog,
		},
		{
			Version:      stringToPtr("1.2.4"),
			IsPrerelease: false,
			URLChangelog: changelog,
		},
	}
	tc.client.EXPECT().ListReleasesV1(mock.Anything, mock.Anything).Once().Return(&operations.ListReleasesV1OK{
		Payload: results,
	}, nil)

	// Check if there is an upgrade available
	r.NoError(tc.c.Check(context.Background()))

	cs := tc.c.getCheckState()
	r.Equal(expectedVersion, *cs.latestRelease.Version)

	// Display the results
	tc.c.Display()
	r.Contains(tc.io.Error.String(), expectedVersion)
	r.Contains(tc.io.Error.String(), "A new version of the HCP CLI is available")
	r.Contains(tc.io.Error.String(), changelog)

	// Check that the state file was written
	r.FileExists(tc.statePath)
	writtenState, err := readVersionCheckState(tc.statePath)
	r.NoError(err)
	r.WithinDuration(time.Now(), writtenState.CheckedAt, 100*time.Millisecond)
}

func TestChecker_Check_InvalidState_Read(t *testing.T) {
	t.Parallel()
	r := require.New(t)
	tc := testChecker(t, "1.2.3")
	r.NoError(os.WriteFile(tc.statePath, []byte("invalid"), 0777))

	expectedErr := fmt.Errorf("fine, just want to test that the api is still called")
	tc.client.EXPECT().ListReleasesV1(mock.Anything, mock.Anything).Once().Return(nil, expectedErr)
	r.ErrorIs(tc.c.Check(context.Background()), expectedErr)
}

func TestChecker_Check_Nil(t *testing.T) {
	t.Parallel()
	r := require.New(t)
	var c *Checker
	r.NoError(c.Check(context.Background()))
	c.Display()
}
