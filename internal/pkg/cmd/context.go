package cmd

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/hashicorp/go-hclog"
	hcpconf "github.com/hashicorp/hcp-sdk-go/config"
	"github.com/hashicorp/hcp/internal/pkg/auth"
	"github.com/hashicorp/hcp/internal/pkg/flagvalue"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
	"github.com/posener/complete"

	httptransport "github.com/go-openapi/runtime/client"
)

// Context passes global objects for constructing and invoking a command.
type Context struct {
	// IO is used to interact directly with IO or the terminal.
	IO iostreams.IOStreams

	// Profile is used to retrieve configuration.
	Profile *profile.Profile

	// Output is used to print structured output.
	Output *format.Outputter

	// HCP is the HTTP transport for connecting to HCP APIs.
	HCP *httptransport.Runtime

	// ShutdownCtx is a context that is cancelled if the user requests the
	// command to be shutdown. If a command can block for an extended amount of
	// time, the context should be used to exit early.
	ShutdownCtx context.Context

	// flags stores our global flags. Access must go through GetGlobalFlags()
	// which ensures flags are only accessed after the flags have been parsed
	// from the arguments.
	flags GlobalFlags
}

// GlobalFlags contains the global flags.
type GlobalFlags struct {
	// parsed stores if the flags have been parsed yet
	parsed bool

	// Unexported global flags. These should generally be access via other
	// helpers exported in the Context.
	project string
	profile string
	format  string
	debug   int

	// Quiet indicates the user has requested minimal output
	Quiet bool
}

func (ctx *Context) GetGlobalFlags() GlobalFlags {
	if !ctx.flags.parsed {
		panic("This is a programmer error. Only access global flags from within a run command. Otherwise flags haven't been parsed yet.")
	}

	return ctx.flags
}

// ConfigureRootCommand should be only called on the root command. It configures
// global flags and ensures that the context is configured based on any flags
// set during a command invocation.
func ConfigureRootCommand(ctx *Context, cmd *Command) {
	// Store the IO on the command, making it available to the entire tree.
	cmd.io = ctx.IO

	// Configure the global flags
	formats := []string{"pretty", "table", "json"}

	cmd.Flags.Persistent = append(cmd.Flags.Persistent, &Flag{
		Name:         "project",
		DisplayValue: "ID",
		Description:  "The HCP Project ID to use. If omitted the current project set in the configuration is used.",
		Value:        flagvalue.Simple[string]("", &ctx.flags.project),
		global:       true,
		Autocomplete: complete.PredictAnything,
	}, &Flag{
		Name:         "profile",
		DisplayValue: "NAME",
		Description:  "The profile to use. If omitted, the currently selected profile will be used.",
		Value:        flagvalue.Simple[string]("", &ctx.flags.profile),
		global:       true,
		Autocomplete: complete.PredictFunc(func(_ complete.Args) []string {
			l, err := profile.NewLoader()
			if err != nil {
				return nil
			}

			profiles, err := l.ListProfiles()
			if err != nil {
				return nil
			}

			return profiles
		}),
	}, &Flag{
		Name:         "format",
		DisplayValue: "FORMAT",
		Description:  "Sets the output format.",
		Value:        flagvalue.Enum[string](formats, "", &ctx.flags.format),
		global:       true,
		Autocomplete: complete.PredictSet(formats...),
	}, &Flag{
		Name:          "quiet",
		Description:   "Minimizes output and disables interactive prompting.",
		Value:         flagvalue.Simple[bool](false, &ctx.flags.Quiet),
		IsBooleanFlag: true,
		global:        true,
		Autocomplete:  complete.PredictNothing,
	}, &Flag{
		Name:          "debug",
		Description:   "Enable debug output.",
		Value:         flagvalue.Counter(0, &ctx.flags.debug),
		IsBooleanFlag: true,
		global:        true,
	})

	// Setup the pre-run command
	cmd.PersistentPreRun = func(c *Command, args []string) error {
		// Setup the HTTP logger. We retrieve the commands logger so the API
		// logger is named with the subcommand.
		ctx.HCP.SetLogger(newAPILogger(c.Logger()))
		ctx.HCP.Debug = true

		if err := ctx.applyGlobalFlags(c, args); err != nil {
			return err
		}

		return isAuthenticated(ctx, c, args)
	}
}

// applyGlobalFlags applies the global flags.
func (ctx *Context) applyGlobalFlags(c *Command, args []string) error {
	// Mark that we have parsed flags
	ctx.flags.parsed = true

	// Parse the profile first
	if p := ctx.flags.profile; p != "" {
		l, err := profile.NewLoader()
		if err != nil {
			return err
		}

		p, err := l.LoadProfile(ctx.flags.profile)
		if err != nil {
			return err
		}

		*ctx.Profile = *p
	}

	// Set the verbosity if the flag is set.
	verbosity := ctx.Profile.Core.GetVerbosity()
	switch ctx.flags.debug {
	case 0:
		// nothing
	case 1:
		verbosity = "debug"
	default:
		verbosity = "trace"
	}

	if verbosity != "" {
		l := hclog.LevelFromString(verbosity)
		if l == hclog.NoLevel {
			return fmt.Errorf("invalid log level: %q", verbosity)
		}

		c.Logger().SetLevel(l)
	}

	// Set the project if the flag is set.
	project := ctx.flags.project
	if project != "" {
		ctx.Profile.ProjectID = project
	}

	// Set the output format if the flag is set.
	f := ctx.flags.format
	if f == "" {
		f = ctx.Profile.Core.GetOutputFormat()
	}
	if f != "" {
		format, err := format.FromString(f)
		if err != nil {
			return err
		}

		ctx.Output.SetFormat(format)
	}

	// Disable color if set
	if ctx.Profile.Core != nil && ctx.Profile.Core.NoColor != nil && *ctx.Profile.Core.NoColor {
		ctx.IO.ForceNoColor()
	}

	// Set quiet on the IOStream if enabled.
	if ctx.flags.Quiet {
		ctx.IO.SetQuiet(true)
	}

	return nil
}

// ParseFlags can be used to parse the flags for a given command before it is
// run. This can be helpful in very specific cases such as accessing flags
// during autocompletion. The return args are the non-flag arguments.
func (ctx *Context) ParseFlags(c *Command, args []string) ([]string, error) {
	if err := c.parseFlags(args); err != nil {
		return nil, err
	}

	if err := ctx.applyGlobalFlags(c, args); err != nil {
		return nil, err
	}

	return c.allCommandFlags.Args(), nil
}

func isAuthenticated(ctx *Context, c *Command, args []string) error {
	if isTopLevelCmd(args) || c.NoAuthRequired {
		return nil
	}

	// Create the HCP Config
	hcpCfg, err := auth.GetHCPConfig(hcpconf.WithoutBrowserLogin())
	if err != nil {
		return fmt.Errorf("failed to instantiate HCP config: %w", err)
	}

	if tkn, err := hcpCfg.Token(); err != nil {
		return authHelp(ctx.IO)
	} else if !tkn.Expiry.After(time.Now()) {
		return authHelp(ctx.IO)
	}

	return nil
}

func authHelp(io iostreams.IOStreams) error {
	cs := io.ColorScheme()
	help := heredoc.Docf(`
No authentication detected. To get started with HCP CLI, please run:  %s`,
		cs.String("hcp auth login").Bold().String())

	return errors.New(help)
}

// Used to parse commands and skip loading hcp profile
func isTopLevelCmd(args []string) bool {
	if len(args) != 1 {
		return false
	}

	switch args[0] {
	case "version":
		return true
	case "-v":
		return true
	case "--version":
		return true
	case "-version":
		return true
	case "-h":
		return true
	case "--help":
		return true
	}
	return false
}

// logger implements the logging interface required by our openapi clients
type apiLogger struct {
	info  *log.Logger
	debug *log.Logger
}

func newAPILogger(l hclog.Logger) *apiLogger {
	l = l.Named("api_client")
	return &apiLogger{
		info:  l.StandardLogger(&hclog.StandardLoggerOptions{ForceLevel: hclog.Info}),
		debug: l.StandardLogger(&hclog.StandardLoggerOptions{ForceLevel: hclog.Debug}),
	}
}

// Printf prints an info log
func (l apiLogger) Printf(format string, args ...interface{}) {
	l.info.Printf(format, args...)
}

// Debugf prints a debug log
func (l apiLogger) Debugf(format string, args ...interface{}) {
	l.debug.Printf(format, args...)
}
