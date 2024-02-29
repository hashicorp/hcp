package workloadidentityproviders

import (
	"context"
	"os"
	"path"
	"testing"

	"github.com/go-openapi/runtime/client"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
	"github.com/stretchr/testify/require"
)

func TestNewCmdCreateCredFile(t *testing.T) {
	t.Parallel()

	wip := "iam/project/123/service-principal/example/workload-identity-provider/example"
	cases := []struct {
		Name    string
		Args    []string
		Profile func(t *testing.T) *profile.Profile
		Error   string
		Expect  *CreateCredFileOpts
	}{
		{
			Name: "Too many args",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123").SetProjectID("456")
			},
			Args: []string{
				"test", "extra",
				"--output-file", "config.json",
			},
			Error: "accepts 1 arg(s), received 2",
		},
		{
			Name: "Good AWS",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123").SetProjectID("456")
			},
			Args: []string{
				wip,
				"--output-file", "config.json",
				"--aws",
			},
			Expect: &CreateCredFileOpts{
				WIP:        wip,
				OutputFile: "config.json",
				AWS:        true,
			},
		},
		{
			Name: "Good all",
			Profile: func(t *testing.T) *profile.Profile {
				return profile.TestProfile(t).SetOrgID("123").SetProjectID("456")
			},
			Args: []string{
				wip,
				"--output-file", "config.json",
				"--aws", "--imdsv1",
				"--azure", "--azure-resource", "resource", "--azure-client-id", "client-id",
				"--gcp",
				"--source-env", "ENV_VAR",
				"--source-url", "https://example.com",
				"--source-file", "file.json",
				"--source-json-pointer", "/json-pointer",
				"--source-header", "header1=value1", "--source-header", "header2=value2",
			},
			Expect: &CreateCredFileOpts{
				WIP:                   wip,
				OutputFile:            "config.json",
				AWS:                   true,
				IMDSv1:                true,
				Azure:                 true,
				AzureResource:         "resource",
				AzureClientID:         "client-id",
				GCP:                   true,
				SourceEnvVar:          "ENV_VAR",
				SourceURL:             "https://example.com",
				SourceFile:            "file.json",
				CredentialJSONPointer: "/json-pointer",
				SourceURLHeaders:      map[string]string{"header1": "value1", "header2": "value2"},
			},
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			r := require.New(t)

			// Create a context.
			io := iostreams.Test()
			ctx := &cmd.Context{
				IO:          io,
				Profile:     c.Profile(t),
				Output:      format.New(io),
				HCP:         &client.Runtime{},
				ShutdownCtx: context.Background(),
			}

			var gotOpts *CreateCredFileOpts
			createCmd := NewCmdCreateCredFile(ctx, func(o *CreateCredFileOpts) error {
				gotOpts = o
				return nil
			})
			createCmd.SetIO(io)

			code := createCmd.Run(c.Args)
			if c.Error != "" {
				r.NotZero(code)
				r.Contains(io.Error.String(), c.Error)
				return
			}

			r.Zero(code, io.Error.String())
			r.NotNil(gotOpts)
			r.Equal(c.Expect.WIP, gotOpts.WIP)
			r.Equal(c.Expect.OutputFile, gotOpts.OutputFile)
			r.Equal(c.Expect.AWS, gotOpts.AWS)
			r.Equal(c.Expect.Azure, gotOpts.Azure)
			r.Equal(c.Expect.GCP, gotOpts.GCP)
			r.Equal(c.Expect.SourceEnvVar, gotOpts.SourceEnvVar)
			r.Equal(c.Expect.SourceURL, gotOpts.SourceURL)
			r.Equal(c.Expect.SourceFile, gotOpts.SourceFile)
			r.Equal(c.Expect.IMDSv1, gotOpts.IMDSv1)
			r.Equal(c.Expect.AzureResource, gotOpts.AzureResource)
			r.Equal(c.Expect.AzureClientID, gotOpts.AzureClientID)
			r.Equal(c.Expect.CredentialJSONPointer, gotOpts.CredentialJSONPointer)
			r.Equal(c.Expect.SourceURLHeaders, gotOpts.SourceURLHeaders)
		})
	}
}

func TestCreateCredFileOpts_Validate(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Name  string
		Opts  *CreateCredFileOpts
		Error string
	}{
		{
			Name: "Bad WIP",
			Opts: &CreateCredFileOpts{
				WIP: "bad",
			},
			Error: "invalid workload identity provider resource name",
		},
		{
			Name: "No sources",
			Opts: &CreateCredFileOpts{
				WIP:        "iam/project/123/service-principal/example/workload-identity-provider/example",
				OutputFile: "config.json",
			},
			Error: "only one of --aws, --azure, --gcp, --source-env, --source-url, or --source-file can be set",
		},
		{
			Name: "Too many sources",
			Opts: &CreateCredFileOpts{
				WIP:        "iam/project/123/service-principal/example/workload-identity-provider/example",
				OutputFile: "config.json",
				AWS:        true,
				Azure:      true,
			},
			Error: "only one of --aws, --azure, --gcp, --source-env, --source-url, or --source-file can be set",
		},
		{
			Name: "IMDSv1 without AWS",
			Opts: &CreateCredFileOpts{
				WIP:        "iam/project/123/service-principal/example/workload-identity-provider/example",
				OutputFile: "config.json",
				IMDSv1:     true,
				GCP:        true,
			},
			Error: "--imdsv1 can only be set if --aws is set",
		},
		{
			Name: "Azure without Azure Resource",
			Opts: &CreateCredFileOpts{
				WIP:        "iam/project/123/service-principal/example/workload-identity-provider/example",
				OutputFile: "config.json",
				Azure:      true,
			},
			Error: "--azure-resource must be set if --azure is set",
		},
		{
			Name: "Azure Resource without Azure",
			Opts: &CreateCredFileOpts{
				WIP:           "iam/project/123/service-principal/example/workload-identity-provider/example",
				OutputFile:    "config.json",
				AzureResource: "resource",
				GCP:           true,
			},
			Error: "--azure-resource can only be set if --azure is set",
		},
		{
			Name: "Azure Client ID without Azure",
			Opts: &CreateCredFileOpts{
				WIP:           "iam/project/123/service-principal/example/workload-identity-provider/example",
				OutputFile:    "config.json",
				AzureClientID: "client-id",
				GCP:           true,
			},
			Error: "--azure-client-id can only be set if --azure is set",
		},
		{
			Name: "Credential Pointer for AWS",
			Opts: &CreateCredFileOpts{
				WIP:                   "iam/project/123/service-principal/example/workload-identity-provider/example",
				OutputFile:            "config.json",
				AWS:                   true,
				CredentialJSONPointer: "/json-pointer",
			},
			Error: "--source-json-pointer can only be set if --source-url, --source-file, or --source-env is set",
		},
		{
			Name: "Credential Pointer for GCP",
			Opts: &CreateCredFileOpts{
				WIP:                   "iam/project/123/service-principal/example/workload-identity-provider/example",
				OutputFile:            "config.json",
				GCP:                   true,
				CredentialJSONPointer: "/json-pointer",
			},
			Error: "--source-json-pointer can only be set if --source-url, --source-file, or --source-env is set",
		},
		{
			Name: "Credential Pointer for Azure",
			Opts: &CreateCredFileOpts{
				WIP:                   "iam/project/123/service-principal/example/workload-identity-provider/example",
				OutputFile:            "config.json",
				Azure:                 true,
				AzureResource:         "resource",
				CredentialJSONPointer: "/json-pointer",
			},
			Error: "--source-json-pointer can only be set if --source-url, --source-file, or --source-env is set",
		},
		{
			Name: "Source URL Headers without Source URL",
			Opts: &CreateCredFileOpts{
				WIP:              "iam/project/123/service-principal/example/workload-identity-provider/example",
				OutputFile:       "config.json",
				SourceURLHeaders: map[string]string{"header1": "value1", "header2": "value2"},
				AWS:              true,
			},
			Error: "--source-header can only be set if --source-url is set",
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			r := require.New(t)

			err := c.Opts.Validate()
			if c.Error != "" {
				r.Error(err)
				r.Contains(err.Error(), c.Error)
				return
			}

			r.NoError(err)
		})
	}
}

func TestCreateCredFileRun(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Name     string
		Opts     *CreateCredFileOpts
		Expected string
	}{
		{
			Name: "AWS",
			Opts: &CreateCredFileOpts{
				WIP:        "iam/project/123/service-principal/example/workload-identity-provider/example",
				OutputFile: "config.json",
				AWS:        true,
			},
			Expected: `{
			  "scheme": "workload",
			  "workload": {
				"provider_resource_name": "iam/project/123/service-principal/example/workload-identity-provider/example",
				"aws": {
				  "imds_v2": true
				}
			  }
			}`,
		},
		{
			Name: "AWS with IMDSv1",
			Opts: &CreateCredFileOpts{
				WIP:        "iam/project/123/service-principal/example/workload-identity-provider/example",
				OutputFile: "config.json",
				AWS:        true,
				IMDSv1:     true,
			},
			Expected: `{
			  "scheme": "workload",
			  "workload": {
				"provider_resource_name": "iam/project/123/service-principal/example/workload-identity-provider/example",
				"aws": {}
			  }
			}`,
		},
		{
			Name: "Azure",
			Opts: &CreateCredFileOpts{
				WIP:           "iam/project/123/service-principal/example/workload-identity-provider/example",
				OutputFile:    "config.json",
				Azure:         true,
				AzureResource: "resource",
			},
			Expected: `{
			  "scheme": "workload",
			  "workload": {
				"provider_resource_name": "iam/project/123/service-principal/example/workload-identity-provider/example",
				"url": {
				  "url": "http://169.254.169.254/metadata/identity/oauth2/token?api-version=2018-02-01\u0026resource=resource",
				  "headers": {
					"Metadata": "True"
				  },
				  "format_type": "json",
				  "subject_cred_pointer": "/access_token"
				}
			  }
			}`,
		},
		{
			Name: "Azure Client ID",
			Opts: &CreateCredFileOpts{
				WIP:           "iam/project/123/service-principal/example/workload-identity-provider/example",
				OutputFile:    "config.json",
				Azure:         true,
				AzureResource: "resource",
				AzureClientID: "client-id",
			},
			Expected: `{
			  "scheme": "workload",
			  "workload": {
				"provider_resource_name": "iam/project/123/service-principal/example/workload-identity-provider/example",
				"url": {
				  "url": "http://169.254.169.254/metadata/identity/oauth2/token?api-version=2018-02-01\u0026resource=resource\u0026client_id=client-id",
				  "headers": {
					"Metadata": "True"
				  },
				  "format_type": "json",
				  "subject_cred_pointer": "/access_token"
				}
			  }
			}`,
		},
		{
			Name: "GCP",
			Opts: &CreateCredFileOpts{
				WIP:        "iam/project/123/service-principal/example/workload-identity-provider/example",
				OutputFile: "config.json",
				GCP:        true,
			},
			Expected: `{
			  "scheme": "workload",
			  "workload": {
				"provider_resource_name": "iam/project/123/service-principal/example/workload-identity-provider/example",
				"url": {
				  "url": "http://metadata/computeMetadata/v1/instance/service-accounts/default/identity?audience=iam/project/123/service-principal/example/workload-identity-provider/example\u0026format=full",
				  "headers": {
					"Metadata-Flavor": "Google"
				  }
				}
			  }
			}`,
		},
		{
			Name: "Source URL",
			Opts: &CreateCredFileOpts{
				WIP:        "iam/project/123/service-principal/example/workload-identity-provider/example",
				OutputFile: "config.json",
				SourceURL:  "https://example.com",
			},
			Expected: `{
			  "scheme": "workload",
			  "workload": {
				"provider_resource_name": "iam/project/123/service-principal/example/workload-identity-provider/example",
				"url": {
				  "url": "https://example.com"
				}
			  }
			}`,
		},
		{
			Name: "Source URL with JSON Pointer",
			Opts: &CreateCredFileOpts{
				WIP:                   "iam/project/123/service-principal/example/workload-identity-provider/example",
				OutputFile:            "config.json",
				SourceURL:             "https://example.com",
				CredentialJSONPointer: "/json-pointer",
			},
			Expected: `{
			  "scheme": "workload",
			  "workload": {
				"provider_resource_name": "iam/project/123/service-principal/example/workload-identity-provider/example",
				"url": {
				  "url": "https://example.com",
				  "format_type": "json",
				  "subject_cred_pointer": "/json-pointer"
				}
			  }
			}`,
		},
		{
			Name: "Source URL with Headers",
			Opts: &CreateCredFileOpts{
				WIP:              "iam/project/123/service-principal/example/workload-identity-provider/example",
				OutputFile:       "config.json",
				SourceURL:        "https://example.com",
				SourceURLHeaders: map[string]string{"header1": "value1", "header2": "value2"},
			},
			Expected: `{
			  "scheme": "workload",
			  "workload": {
				"provider_resource_name": "iam/project/123/service-principal/example/workload-identity-provider/example",
				"url": {
				  "url": "https://example.com",
				  "headers": {
					"header1": "value1",
					"header2": "value2"
				  }
				}
			  }
			}`,
		},
		{
			Name: "Source File",
			Opts: &CreateCredFileOpts{
				WIP:        "iam/project/123/service-principal/example/workload-identity-provider/example",
				OutputFile: "config.json",
				SourceFile: "file.json",
			},
			Expected: `{
			  "scheme": "workload",
			  "workload": {
				"provider_resource_name": "iam/project/123/service-principal/example/workload-identity-provider/example",
				"file": {
				  "path": "file.json"
				}
			  }
			}`,
		},
		{
			Name: "Source File JSON Pointer",
			Opts: &CreateCredFileOpts{
				WIP:                   "iam/project/123/service-principal/example/workload-identity-provider/example",
				OutputFile:            "config.json",
				SourceFile:            "file.json",
				CredentialJSONPointer: "/json-pointer",
			},
			Expected: `{
			  "scheme": "workload",
			  "workload": {
				"provider_resource_name": "iam/project/123/service-principal/example/workload-identity-provider/example",
				"file": {
				  "path": "file.json",
				  "format_type": "json",
				  "subject_cred_pointer": "/json-pointer"
				}
			  }
			}`,
		},
		{
			Name: "Source Env",
			Opts: &CreateCredFileOpts{
				WIP:          "iam/project/123/service-principal/example/workload-identity-provider/example",
				OutputFile:   "config.json",
				SourceEnvVar: "ENV_VAR",
			},
			Expected: `{
			  "scheme": "workload",
			  "workload": {
				"provider_resource_name": "iam/project/123/service-principal/example/workload-identity-provider/example",
				"env": {
				  "var": "ENV_VAR"
				}
			  }
			}`,
		},
		{
			Name: "Source Env JSON Pointer",
			Opts: &CreateCredFileOpts{
				WIP:                   "iam/project/123/service-principal/example/workload-identity-provider/example",
				OutputFile:            "config.json",
				SourceEnvVar:          "ENV_VAR",
				CredentialJSONPointer: "/json-pointer",
			},
			Expected: `{
			  "scheme": "workload",
			  "workload": {
				"provider_resource_name": "iam/project/123/service-principal/example/workload-identity-provider/example",
				"env": {
				  "var": "ENV_VAR",
				  "format_type": "json",
				  "subject_cred_pointer": "/json-pointer"
				}
			  }
			}`,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			r := require.New(t)

			// Add the IO to the options
			c.Opts.IO = iostreams.Test()

			// Determine the output file
			outputFile := path.Join(t.TempDir(), "config.json")
			c.Opts.OutputFile = outputFile

			// Run the command
			err := createCredFileRun(c.Opts)
			r.NoError(err)

			// Read the file
			out, err := os.ReadFile(outputFile)
			r.NoError(err)

			// Compare the output
			r.JSONEq(c.Expected, string(out), string(out))
		})
	}

}
