# HCP CLI Codebase Documentation

This directory contains some documentation about the `hcp` CLI codebase, aimed at
readers who are interested in making code contributions.

If you're looking for information on using `hcp`, please instead refer to the
[HCP CLI Documentation](https://developer.hashicorp.com/hcp/docs/cli).

## Building and testing the CLI

To build the CLI and output the binary to the `bin/` directory, run:

```sh
make go/build
```

To install the CLI to your GOPATH, run:

```sh
make go/install
```

To run the tests, run:

```sh
make go/test
make go/lint
```

### Mocks

To add a new mock, edit `.mockery.yaml` and add the package you want mocked. To
regenerate the mocks, either after adding a new mock package or after updating
the dependencies, run:

```sh
# mockery version >= v2.38.0
make go/mocks
```

## Authoring a command

### Command structure

Since the HCP CLI is shared amongst multiple products, care must be taken
to not pollute the top level command with potentially conflicting commands. To
avoid this, each product is placed in its own subcommand:

```sh
hcp packer
hcp waypoint
hcp vault-secrets
hcp vault-radar
```

Subcommands can have further nested subcommands themselves. When a service has
an object that has CRUD operations on it, the object should likely be a
subcommand with further nested commands:

```sh
$ hcp vault-secrets apps [create, list, delete, read, update]
```

This structure naturally has a layout of chaining nouns together with the last
command being a verb such as create, delete, list, etc. Other than the top-level
noun, nouns should be **pluralized**. As an example prefer `hcp iam groups` to `hcp iam
group`.

To help users become familiar with our CLI, we should be consistent with common
verbs across commands. Prefer using:

* **create**: creating a new resource
* **list**: listing resources
* **delete**: deleting a resource
* **read**: reading an existing resource in more detail
* **update**: updating an existing resource

### Command PersistentRun Functions

If a command requires that a project or organization ID is set, the following
`PersistentPreRun` function should be used:

```go
// Require only the organization to be set
cmd := &cmd.Command{
  ...
  PersistentPreRun: func(c *cmd.Command, args []string) error {
    return cmd.RequireOrganization(ctx)
  },
}

// Require the organization and project to be set
cmd := &cmd.Command{
  ...
  PersistentPreRun: func(c *cmd.Command, args []string) error {
    return cmd.RequireOrgAndProject(ctx)
  },
}
```

### Command Documentation.

#### Short Help

Short help must be a single sentence that describes the command. It must start
with a capital letter and end with a period. It should be a concise description
of what the command does.

#### Long Help

Long help should be a more detailed description of the command. If certain
commands are expected to be used together, this is a good place to mention that.
Long help should be written in full sentences and should be grammatically
correct, and it should end with a period.

All long help should use the `heredoc` package to format the help text. The
package allows formatting the help text (making text bold, colored, italic, etc)
and allows for outputting text appropriate for the command line and generated
markdown documentation. For more detail see the [heredoc
documentation](contributing/heredoc).

```go

For consistency, commands share a common first sentence. For command groups it
is:

```sh
The {{ template "mdCodeOrBold" "hcp iam groups" }} command group ...
```

And for commands it is:

```sh
The {{ template "mdCodeOrBold" "hcp iam groups create" }} command ...
```

Replace `hcp iam groups` and `hcp iam groups create` with the appropriate command.

#### Examples Help

Commands can include examples to help users understand how to use the command.
Commands should err towards having examples, even if they appear simple.

Example preambles should be full sentences that start with a capital letter, and
end in a colon.

The example themselves should start with either a `#` for a comment or a `$` for
the command. A comment should be reserved for rare cases where additional detail
needs to be provided that doesn't fit into the preamble.

```sh
$ hcp iam groups create example
```

When the example contains flags, display the flag as `--flag-name=value` instead
of `--flag-name value`. This ensures that examples are consistent across
commands and leaves no room for ambiguity as to whether the flag takes a value
or not (the value can not be mistaken for a positional argument).

To handle long commands that don't fit on a single line, use a backslash `\` to
split the command into multiple lines. The subsequent lines should be indented
by two spaces.

```sh
$ hcp iam groups create example \
  --description="This is an example group."
```

### Flags vs Arguments

Positional arguments should generally be used to specify the resource being
acted upon or created. Flags should be used for all other options.

It should be very rare to have a command that requires more than one positional
argument.

#### Flags

When defining a flag, ensure you mark whether the flag is required or not. The
CLI framework will validate that all required flags are set before running the
command.

Setting the flag's DisplayValue is highly recommend as well. The DisplayValue is
used when displaying flag usage. As an example, a flag defined as follows:

```go
{
    Name:         "member",
    DisplayValue: "ID",
    Description:  "The ID of the principal to add to the group.",
    Value:        flagvalue.Simple("", &opts.Member),
},
```

Will have help output that looks like:

```sh
FLAGS
  --member=ID
    The ID of the user principal to add to the group.
```

The DisplayValue provides additional context to help the user understand what
the expected value should be.

### Command Input / Output

Commands **should not** directly use `os.Stdout`, `os.Stderr`, or `os.Stdin`.
Instead, use the following which are passed to your command either via the
`cmd.Command` or the `cmd.Context`.

* `ctx.IO`: Provides `In()`, `Out()`, and `Err()`, as well as helpers for
  prompting and reading secrets.
* `ctx.Output`: Use this for structured output. The outputter will automatically
  be configured by the `--format` flag and can output JSON, table format, or
  pretty printed key-value pairs.
* `cmd.Logger()`: Use this for accessing a logger for your command.

#### IO.Out() vs IO.Err()

Intuitively one may assume to output errors to `ctx.IO.Err()` and everything
else to `ctx.IO.Out()`. However, this is not the case. Out and Err map the CLIs
stdout and stderr file descriptors. To allow chaining of commands, it is
important the stdout is reserved for the primary output of the command and all
other help output should be sent to stderr.

* Out: The primary output for your command should go to stdout. Anything that is machine readable should also go to stdout—this is where piping sends things by default.
* Err: Log messages, errors, and so on should all be sent to stderr. This means that when commands are piped together, these messages are displayed to the user and not fed into the next command.

When the CLI is invoked with `--quiet`, all Err output is suppressed.

As an example, if your command deletes a resource and you want to display a
success message to the user, you should output that to Err as it is not machine
readable and is intended only for humans. By writing to Err, we also make it
easier for users to write scripts that suppress these non critical messages.

```go
fmt.Fprintf(opts.IO.Err(), "Deleted application %q\n", appID)
```

#### Logs

To emit logs from your command, capture the logger from the command passed into
the RunF function. This logger is automatically configured to be prefixed with
the command name and is set to the appropriate log level based on the
invocation.

```go
RunF: func(c *cmd.Command, args []string) error {
    opts.Logger = c.Logger()
    ...
},
```

To display logs, run the command with `--debug` for debug logging or `--debug
--debug` for trace logging.

#### Structured Output

To output the results of a command in a structured format use the passed
`ctx.Output` object. This object supports outputting objects in multiple formats
(JSON, Table, Pretty) and is automatically configured by the `--format` flag or
a profile configuration.

There are multiple ways to emit output, each providing increasing flexibility.

```go
// Assuming we are outputting the following resource
type Project struct {
    ID   string
    Name string
    Metadata *ProjectMetadata
    CreatedAt time.Time
}

type ProjectMetadata struct {
    Owner string
    Description string
}
```

1. The simplest way is to use the `Show` method:

The `Show` method uses reflection to determine the fields in an object to display.
It supports being passed a single object or a slice of objects. The fields will
be displayed alpahbetically and any CamelCase fields will become space
deliminated (e.g. `MyField` becomes `My Field`) when outputted..

The second parameter to Show is the default format to display the output with.

```go
func getProject(opts *GetOptions) error {
    resp, err := ... // Get the resource
    respPayload := resp.GetPayload().Project
    return opts.Output.Show(respPayload, format.Pretty)
}
```

This will output the following:

```sh
ID: 1234
Name: My Project
Metadata Owner: 21328901902
Metadata Description: My Project Description
Created At: 2021-01-01T00:00:00Z
```

2. Using `NewDisplayer`:

The outputter has a method called `Display` that takes a `Displayer` interface.
`NewDisplayer` is a helper function that creates a `Displayer` from a struct.
NewDisplayer gives you control over the names of fields, the order they are
displayed in, and the formatting of the output.

```go
func getProject(opts *GetOptions) error {
    resp, err := ... // Get the resource
    respPayload := resp.GetPayload().Project

    // Create the fields. The fields allow setting the outputting name directly
    // and the value is a text/template which allows additional formatting.
    projectFields = []format.Field{
        format.NewField("Name", "{{ .Name }}"),
        format.NewField("Project ID", "{{ .ID }}"), // Display Project ID instead of ID
        format.NewField("Owner", "{{ .Metadata.Owner }}"), // Access sub-structs
        format.NewField("Description", "{{ .Metadata.Description }}"),
        format.NewField("Created At", "{{ .CreatedAt }}"),
    }

    // Display the created project
    d := format.NewDisplayer(resp.Payload.Project, format.Pretty, projectFields)
    return opts.Output.Display(d)
}
```

This will output the following:

```sh
Project ID: 1234
Name: My Project
Owner: 21328901902
Description: My Project Description
Created At: 2021-01-01T00:00:00Z
```

3. Directly implementing the `Displayer` interface:

For ultimate flexibility, you can implement the `Displayer` interface directly.
There are two main use cases for this. The first is if you want to customize how
a field is displayed you can wrap the object that is being displayed in a custom
type and implement a function on the wrapped type. Then you can use the
text/template to invoke that function.

```
type wrappedProject struct {
    *Project
}

func (p *wrappedProject) Owner() (string, error) {
    resp, err := iam.GetUser(p.Metadata.Owner)
    if err != nil "
        return "", err
    }

    return resp.User.Name, nil
}

func (d *ProjectDisplayer) FieldTemplates() []Field {
    return []Field{
        NewField("Name", "{{ .Name }}"),
        NewField("Project ID", "{{ .ID }}"),
        NewField("Owner", "{{ .Owner }}"),
        NewField("Description", "{{ .Metadata.Description }}"),
        NewField("Created At", "{{ .CreatedAt }}"),
    }
}
...
```

The second use case is if you want to output a different object for JSON than
the table/pretty format. This can be accomplished by implementing the
`TemplatePayload` interface. To see an example of this, see how [IAM policies are
displayed](https://github.com/hashicorp/hcp/blob/v0.1.0/internal/pkg/api/iampolicy/displayer.go#L73-L86).

#### Errors

If a command results in an error, return the error directly. The CLI framework
will handle displaying the error message to the user.

```go

func good() error {
    if err := doSomething(); err != nil {
        return fmt.Error("failed to do the thing: %w", err)
    }
    ...
}

// This will cause the error to be displayed twice.
func bad() error {
    if err := doSomething(); err != nil {
        fmt.Fprintf(opts.IO.Err(), "failed to do the thing: %v\n", err)
        return fmt.Error("failed to do the thing: %w", err)
    }
    ...
}
```

#### Prompting

When the command is a destructive action, prompt the user for confirmation. This
can be done as follows:

```go
func deleteGroup(opts *DeleteOptions) error {
    if opts.IO.CanPrompt() {
        ok, err := opts.IO.PromptConfirm("The group will be deleted.\n\nDo you want to continue")
        if err != nil {
            return fmt.Errorf("failed to retrieve confirmation: %w", err)
        }

        if !ok {
            return nil
        }
    }

    // Do delete
    ...
}
```

If the user invokes the command with `--quiet`, the prompt will be suppressed.

#### Reading secrets

Commands **should not** accept secrets from environment variables, flags, or
positional arguments. This will leak the secret to `ps` output and potentially
shell history.

Instead prefer reading secrets from a file or from stdin, interactively or
non-interactively. As an example have a flag called `--secret-file`:

* `--secret-file=/path/to/secret/file`: Read the secret from the passed file.
* `--secret-file=-`: Read the secret from stdin.

```sh
cat /path/to/secret/file | hcp secrets create --secret-file=-
```

* If `--secret-file` is not passed, prompt the user for the secret.

```go
func getSecret(io iostreams.IOStreams) (string, error) {
    fmt.Fprintln(io.Err(), "Enter the plaintext secret to upload:")
    data, err := io.ReadSecret()
    ...
}
```

### Validating/generating the developer.hashicorp.com documentation

The `hcp` CLI generates the command documentation found on
[developer.hashicorp.com](https://developer.hashicorp.com/hcp/docs/cli/commands).
To check that the generated documentation for your command is correct before
committing the code, generate the documentation and run the
developer.hashicorp.com documentation locally.

First checkout the [`hcp-docs`
repository](https://github.com/hashicorp/hcp-docs). The following commands
assume the `hcp-docs` repository is checked out in the same parent directory as
`hcp`.

```sh
$ make docs/gen
$ make docs/move
$ cd ../hcp-docs
$ make website
```

## Releasing

> Releasing is currently a manual process, please reach out to [#team-cloud-core-platform](https://hashicorp.enterprise.slack.com/archives/C073FTXFLTA) for assistance.

If it is your first time releasing, following the onboarding [steps
here](https://hashicorp.atlassian.net/wiki/spaces/RELENG/pages/2301263888/Part+1+Onboarding+Pre-Requisites#Steps-for-each-member-of-your-team-to-complete).

To get the main branch ready for a release ensure the following:

- [ ] The version in `version/VERSION` is updated to the desired version.
- [ ] The `CHANGELOG.md` is updated with the new version and the changes. The
changelog can be generated using `LAST_RELEASE_GIT_TAG=v0.x.y make changelog/build`
- [ ] Changes since the last release are manually tested and working.

Then follow the [release steps outlined
here](https://hashicorp.atlassian.net/wiki/spaces/RELENG/pages/2303492328/Trigger+a+Staging+Release).

After a successful release:

- [ ] Tag the release commit with the version that was released.
- [ ] Update the `cmd/VERSION` file to the next version with `-dev` appended.
- [ ] Update the developer.hashicorp.com documentation by following the steps
  outlined in the "Validating/generating the developer.hashicorp.com documentation" section.
  PR the changes to the `hcp-docs` repository, have a team member review them,
  and merge.
