// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: MPL-2.0

package agent

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/stretchr/testify/require"
	"github.com/zclconf/go-cty/cty"
)

func TestConfig(t *testing.T) {
	t.Parallel()

	t.Run("can load the operations", func(t *testing.T) {
		t.Parallel()

		r := require.New(t)

		hcl := `
		group "test" {
			action "launch" {
				run {
					command = "./launch.sh"
				}
			}
		}
`

		cfg, err := ParseConfig(hcl)
		r.NoError(err)
		cfg.opWrapper = NoopOperations

		r.Equal([]string{"test"}, cfg.Groups())

		op, err := cfg.Action("test", "launch", nil)
		r.NoError(err)

		shell, ok := op.(*NoopWrapper).Operation.(*ShellOperation)
		r.True(ok)

		r.Equal([]string{"./launch.sh"}, shell.Arguments)
	})

	t.Run("splits shell operations using shell space rules", func(t *testing.T) {
		t.Parallel()

		r := require.New(t)

		hcl := `
		group "test" {
			action "launch" {
				run {
					command = "./launch.sh -delay=5ms   -restart"
				}
			}
		}
`

		cfg, err := ParseConfig(hcl)
		r.NoError(err)

		cfg.forceShell = "sh"
		cfg.opWrapper = NoopOperations

		op, err := cfg.Action("test", "launch", nil)
		r.NoError(err)

		shell, ok := op.(*NoopWrapper).Operation.(*ShellOperation)
		r.True(ok)

		r.Equal([]string{"sh", "-c", "./launch.sh -delay=5ms   -restart"}, shell.Arguments)
	})

	t.Run("can provide a list to run", func(t *testing.T) {
		t.Parallel()

		r := require.New(t)

		hcl := `
		group "test" {
			action "launch" {
				run {
					command = ["./launch.sh", "-delay=5ms", "-restart"]
				}
			}
		}
`

		cfg, err := ParseConfig(hcl)
		r.NoError(err)
		cfg.opWrapper = NoopOperations

		r.Equal([]string{"test"}, cfg.Groups())

		op, err := cfg.Action("test", "launch", nil)
		r.NoError(err)

		shell, ok := op.(*NoopWrapper).Operation.(*ShellOperation)
		r.True(ok)

		r.Equal([]string{"./launch.sh", "-delay=5ms", "-restart"}, shell.Arguments)
	})

	t.Run("can handle http operations", func(t *testing.T) {
		t.Parallel()

		r := require.New(t)

		hcl := `
		group "test" {
			action "launch" {
				http {
					url = "https://google.com"
				}
			}
		}
`

		cfg, err := ParseConfig(hcl)
		r.NoError(err)
		cfg.opWrapper = NoopOperations

		r.Equal([]string{"test"}, cfg.Groups())

		op, err := cfg.Action("test", "launch", nil)
		r.NoError(err)

		ho, ok := op.(*NoopWrapper).Operation.(*HTTPOperation)
		r.True(ok)

		r.Equal("https://google.com", ho.URL)
	})

	t.Run("can handle compound actions", func(t *testing.T) {
		t.Parallel()

		r := require.New(t)

		hcl := `
		group "test" {
			action "launch" {
				operation {
					http {
						url = "https://google.com"
					}
				}

				operation {
					http {
						url = "https://yahoo.com"
					}
				}
			}
		}
`

		cfg, err := ParseConfig(hcl)
		r.NoError(err)
		cfg.opWrapper = NoopOperations

		r.Equal([]string{"test"}, cfg.Groups())

		op, err := cfg.Action("test", "launch", nil)
		r.NoError(err)

		co, ok := op.(*NoopWrapper).Operation.(*CompoundOperation)
		r.True(ok)

		ho, ok := co.Operations[0].(*NoopWrapper).Operation.(*HTTPOperation)
		r.True(ok)

		r.Equal("https://google.com", ho.URL)

		ho, ok = co.Operations[1].(*NoopWrapper).Operation.(*HTTPOperation)
		r.True(ok)

		r.Equal("https://yahoo.com", ho.URL)
	})

	t.Run("can handle status operations", func(t *testing.T) {
		t.Parallel()

		r := require.New(t)

		hcl := `
		group "test" {
			action "launch" {
				status {
					message = "performing launch"
					values = { "attempt" = 3 }
					status = "running"
				}
			}
		}
`

		cfg, err := ParseConfig(hcl)
		r.NoError(err)
		cfg.opWrapper = NoopOperations

		r.Equal([]string{"test"}, cfg.Groups())

		op, err := cfg.Action("test", "launch", nil)
		r.NoError(err)

		ho, ok := op.(*NoopWrapper).Operation.(*StatusOperation)
		r.True(ok)

		r.Equal("performing launch", ho.Message)

		r.Equal("3", ho.Values["attempt"])

		r.Equal("running", ho.Status)
	})

	t.Run("can access input variables", func(t *testing.T) {
		t.Parallel()

		r := require.New(t)

		str := `
		group "test" {
			action "launch" {
				run {
					command = ["./launch.sh", "-delay", var.delay]
				}
			}
		}
`

		cfg, err := ParseConfig(str)
		r.NoError(err)
		cfg.opWrapper = NoopOperations

		r.Equal([]string{"test"}, cfg.Groups())

		var hctx hcl.EvalContext

		hctx.Variables = map[string]cty.Value{
			"var": cty.ObjectVal(map[string]cty.Value{
				"delay": cty.StringVal("5ms"),
			}),
		}

		op, err := cfg.Action("test", "launch", &hctx)
		r.NoError(err)

		shell, ok := op.(*NoopWrapper).Operation.(*ShellOperation)
		r.True(ok)

		r.Equal([]string{"./launch.sh", "-delay", "5ms"}, shell.Arguments)
	})

	t.Run("can specify docker options", func(t *testing.T) {
		t.Parallel()

		r := require.New(t)

		str := `
		group "test" {
			action "launch" {
				run {
					command = ["./launch.sh", "-delay", "5ms"]

					docker {
						image = "ubuntu"
					}
				}
			}
		}
`

		cfg, err := ParseConfig(str)
		r.NoError(err)
		cfg.opWrapper = NoopOperations

		r.Equal([]string{"test"}, cfg.Groups())

		op, err := cfg.Action("test", "launch", nil)
		r.NoError(err)

		shell, ok := op.(*NoopWrapper).Operation.(*ShellOperation)
		r.True(ok)

		r.Equal([]string{"./launch.sh", "-delay", "5ms"}, shell.Arguments)
		r.NotNil(shell.DockerOptions)
		r.Equal("ubuntu", shell.DockerOptions.Image)
	})
}
