package cmd

import (
	"testing"

	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/require"
)

func TestCommand_PersistentPrerun(t *testing.T) {
	t.Parallel()
	r := require.New(t)

	// Create the command tree
	root := &Command{Name: "root"}
	child := &Command{
		Name: "child",
		RunF: func(c *Command, args []string) error {
			return nil
		},
	}
	childContainer := &Command{Name: "child-group"}
	grandchild := &Command{
		Name: "grandchild",
		RunF: func(c *Command, args []string) error {
			return nil
		},
	}
	root.AddChild(child)
	root.AddChild(childContainer)
	childContainer.AddChild(grandchild)

	// Add the persistent preruns
	rootPreRunCount := 0
	containerPreRunCount := 0
	root.PersistentPreRun = func(c *Command, args []string) error {
		rootPreRunCount++
		return nil
	}
	childContainer.PersistentPreRun = func(c *Command, args []string) error {
		containerPreRunCount++
		return nil
	}

	// Run the grandchild and the child
	r.Zero(grandchild.Run(nil))
	r.Zero(child.Run(nil))

	// Expect the prerun commmands were called
	r.Equal(2, rootPreRunCount)
	r.Equal(1, containerPreRunCount)
}

func TestCommand_Flags(t *testing.T) {
	t.Parallel()
	r := require.New(t)

	// Create the command tree
	root := &Command{Name: "root"}
	rootFlag := root.persistentFlags().String("root-flag", "", "testing")

	seenFlags := 0
	child := &Command{
		Name: "child",
		RunF: func(c *Command, args []string) error {
			c.allFlags().VisitAll(func(_ *pflag.Flag) {
				seenFlags++
			})
			return nil
		},
	}
	root.AddChild(child)
	childFlag := child.allFlags().String("child-flag", "", "testing")

	r.Zero(child.Run([]string{"--root-flag=root-set", "--child-flag=child-set"}))
	r.Equal(2, seenFlags)
	r.Equal("root-set", *rootFlag)
	r.Equal("child-set", *childFlag)
}

func TestCommand_Logger(t *testing.T) {
	t.Parallel()
	r := require.New(t)

	// Create the command tree
	io := iostreams.Test()
	root := &Command{
		Name: "root",
		io:   io,
	}
	child := &Command{
		Name: "child",
		RunF: func(c *Command, args []string) error {
			c.Logger().Error("hello, world!")
			return nil
		},
	}
	root.AddChild(child)
	r.Zero(child.Run([]string{}))
	r.Contains(io.Error.String(), "hcp.child: hello, world!")
}
