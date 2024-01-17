package resourcename

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExtractOrganizationID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		resourceName string
		want         string
		expectedErr  string
	}{
		{
			name:         "happy path",
			resourceName: "ns/organization/123/type-one/name1",
			want:         "123",
			expectedErr:  "",
		},
		{
			name:         "happy path nested",
			resourceName: "ns/organization/123/type-one/name1/type-two/name2",
			want:         "123",
			expectedErr:  "",
		},
		{
			name:         "ignore value validation",
			resourceName: "ns/organization/123/type-one/MIXEDcase",
			want:         "123",
			expectedErr:  "",
		},
		{
			name:         "just organization given",
			resourceName: "organization/123",
			want:         "123",
			expectedErr:  "",
		},
		{
			name:         "just project given",
			resourceName: "project/123",
			want:         "",
			expectedErr:  "invalid resource name: resource name must consist of a namespace, and a list of type and name parts",
		},
		{
			name:         "project given",
			resourceName: "ns/project/123/type-one/name1",
			want:         "",
			expectedErr:  "resource name doesn't specify an organization ID",
		},
		{
			name:         "nothing given",
			resourceName: "ns/type-other/123/type-one/name1",
			want:         "",
			expectedErr:  "resource name doesn't specify an organization ID",
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			r := require.New(t)
			got, err := ExtractOrganizationID(test.resourceName)
			if test.expectedErr != "" {
				r.ErrorContains(err, test.expectedErr)
			} else {
				r.NoError(err)
				r.Equal(test.want, got)
			}
		})

		t.Run(test.name+" - Must", func(t *testing.T) {
			t.Parallel()

			r := require.New(t)
			if test.expectedErr != "" {
				r.PanicsWithError(test.expectedErr, func() { MustExtractOrganizationID(test.resourceName) })
			} else {
				got := MustExtractOrganizationID(test.resourceName)
				r.Equal(test.want, got)
			}

		})
	}
}

func TestExtractProjectID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		resourceName string
		want         string
		expectedErr  string
	}{
		{
			name:         "happy path",
			resourceName: "ns/project/123/type-one/name1",
			want:         "123",
			expectedErr:  "",
		},
		{
			name:         "happy path nested",
			resourceName: "ns/project/123/type-one/name1/type-two/name2",
			want:         "123",
			expectedErr:  "",
		},
		{
			name:         "just project given",
			resourceName: "project/123",
			want:         "123",
			expectedErr:  "",
		},
		{
			name:         "just organization given",
			resourceName: "organization/123",
			want:         "",
			expectedErr:  "invalid resource name: resource name must consist of a namespace, and a list of type and name parts",
		},
		{
			name:         "ignore value validation",
			resourceName: "ns/project/123/type-one/MIXEDcase",
			want:         "123",
			expectedErr:  "",
		},
		{
			name:         "organization given",
			resourceName: "ns/organization/123/type-one/name1",
			want:         "",
			expectedErr:  "resource name doesn't specify a project ID",
		},
		{
			name:         "nothing given",
			resourceName: "ns/type-other/123/type-one/name1",
			want:         "",
			expectedErr:  "resource name doesn't specify a project ID",
		},
	}
	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			r := require.New(t)
			got, err := ExtractProjectID(test.resourceName)
			if test.expectedErr != "" {
				r.ErrorContains(err, test.expectedErr)
			} else {
				r.NoError(err)
				r.Equal(test.want, got)
			}
		})

		t.Run(test.name+" - Must", func(t *testing.T) {
			t.Parallel()

			r := require.New(t)
			if test.expectedErr != "" {
				r.PanicsWithError(test.expectedErr, func() { MustExtractProjectID(test.resourceName) })
			} else {
				got := MustExtractProjectID(test.resourceName)
				r.Equal(test.want, got)
			}

		})
	}
}

func TestExtractOrganizationOrProjectID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		resourceName  string
		wantOrgID     string
		wantProjectID string
		expectedErr   string
	}{
		{
			name:          "happy path project",
			resourceName:  "ns/project/123/type-one/name1",
			wantOrgID:     "",
			wantProjectID: "123",
			expectedErr:   "",
		},
		{
			name:          "happy path organization",
			resourceName:  "ns/organization/123/type-one/name1/type-two/name2",
			wantOrgID:     "123",
			wantProjectID: "",
			expectedErr:   "",
		},
		{
			name:          "just project given",
			resourceName:  "project/123",
			wantProjectID: "123",
			expectedErr:   "",
		},
		{
			name:         "just organization given",
			resourceName: "organization/123",
			wantOrgID:    "123",
			expectedErr:  "",
		},
		{
			name:          "ignore value validation",
			resourceName:  "ns/project/123/type-one/MIXEDcase",
			wantOrgID:     "",
			wantProjectID: "123",
			expectedErr:   "",
		},
		{
			name:          "organization and project given",
			resourceName:  "ns/organization/123/project/456/type-one/name1",
			wantOrgID:     "123",
			wantProjectID: "",
			expectedErr:   "",
		},
		{
			name:          "neither given",
			resourceName:  "ns/type-other/123/type-one/name1",
			wantOrgID:     "",
			wantProjectID: "",
			expectedErr:   "resource name doesn't specify an organization or project ID",
		},
	}
	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			r := require.New(t)
			gotOrg, gotProj, err := ExtractOrganizationOrProjectID(test.resourceName)
			if test.expectedErr != "" {
				r.ErrorContains(err, test.expectedErr)
			} else {
				r.NoError(err)
				r.Equal(test.wantOrgID, gotOrg)
				r.Equal(test.wantProjectID, gotProj)
			}
		})

		t.Run(test.name+" - Must", func(t *testing.T) {
			t.Parallel()

			r := require.New(t)
			if test.expectedErr != "" {
				r.PanicsWithError(test.expectedErr, func() { MustExtractOrganizationOrProjectID(test.resourceName) })
			} else {
				gotOrg, gotProj := MustExtractOrganizationOrProjectID(test.resourceName)
				r.Equal(test.wantOrgID, gotOrg)
				r.Equal(test.wantProjectID, gotProj)
			}

		})
	}
}

func TestExtractGeo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		resourceName string
		want         string
		expectedErr  string
	}{
		{
			name:         "happy path",
			resourceName: "ns/project/123/geo/us/type-one/name1",
			want:         "us",
			expectedErr:  "",
		},
		{
			name:         "happy path nested",
			resourceName: "ns/project/123/geo/us/type-one/name1/type-two/name2",
			want:         "us",
			expectedErr:  "",
		},
		{
			name:         "resource name without namespace",
			resourceName: "project/123",
			want:         "us",
			expectedErr:  "invalid resource name: resource name must consist of a namespace, and a list of type and name parts",
		},
		{
			name:         "ignore value validation",
			resourceName: "ns/project/123/geo/us/type-one/MIXEDcase",
			want:         "us",
			expectedErr:  "",
		},
		{
			name:         "no geo given",
			resourceName: "ns/project/123/type-one/name1",
			want:         "",
			expectedErr:  "resource name doesn't specify a geography",
		},
	}
	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			r := require.New(t)
			got, err := ExtractGeo(test.resourceName)
			if test.expectedErr != "" {
				r.ErrorContains(err, test.expectedErr)
			} else {
				r.NoError(err)
				r.Equal(test.want, got)
			}
		})

		t.Run(test.name+" - Must", func(t *testing.T) {
			t.Parallel()

			r := require.New(t)
			if test.expectedErr != "" {
				r.PanicsWithError(test.expectedErr, func() { MustExtractGeo(test.resourceName) })
			} else {
				got := MustExtractGeo(test.resourceName)
				r.Equal(test.want, got)
			}
		})
	}
}
