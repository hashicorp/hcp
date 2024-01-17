package profile

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
	"github.com/mitchellh/mapstructure"
)

func NewCmdGet(ctx *cmd.Context) *cmd.Command {
	opts := &GetOpts{
		Ctx:     ctx.ShutdownCtx,
		IO:      ctx.IO,
		Profile: ctx.Profile,
	}

	cmd := &cmd.Command{
		Name:      "get",
		ShortHelp: "Get a HCP CLI Property",
		LongHelp: heredoc.New(ctx.IO).Mustf(`
		{{ Bold "hcp profile get" }} gets the specified property in your active profile.

		To view all currently set properties, run {{ Bold "hcp profile display" }}.
		`),
		Args: cmd.PositionalArguments{
			Autocomplete: opts.Profile,
			Args: []cmd.PositionalArgument{
				{
					Name: "component/property",
					Documentation: heredoc.New(ctx.IO).Must(`
					Property to be unset. Note that COMPONENT/ is optional when referring to
					top-level profile fields, i.e., such as organization_id and project_id.
					Using component names is required for setting other properties like {{ Bold "core/output_format" }}.
					Consult the Available Properties section below for a comprehensive list of properties.
					`),
				},
			},
		},
		AdditionalDocs: []cmd.DocSection{
			availablePropertiesDoc(ctx.IO),
		},
		NoAuthRequired: true,
		RunF: func(c *cmd.Command, args []string) error {
			opts.Property = args[0]

			return getRun(opts)
		},
	}

	return cmd
}

type GetOpts struct {
	Ctx     context.Context
	IO      iostreams.IOStreams
	Profile *profile.Profile

	Property string
}

func getRun(opts *GetOpts) error {
	if err := IsValidProperty(opts.Property); err != nil {
		return err
	}

	// Decode the existing profile into a map
	var data map[string]any
	dec, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		ErrorUnused:          true,
		Result:               &data,
		TagName:              "hcl",
		IgnoreUntaggedFields: true,
	})
	if err != nil {
		return err
	}

	if err := dec.Decode(opts.Profile); err != nil {
		return err
	}

	// Delete the key from the map
	parts := strings.Split(opts.Property, "/")
	level := data
	var value any
	for i, p := range parts {
		// This is the final property
		if i == len(parts)-1 {
			if _, ok := level[p]; !ok {
				return fmt.Errorf("property %q is not set", opts.Property)
			}

			value = level[p]
			break
		}

		// Retrieve the component
		nested, ok := level[p]
		if !ok {
			return fmt.Errorf("property %q is not set", opts.Property)
		}

		// Check if the retrieved element is a nested object
		sub, ok := nested.(map[string]any)
		if !ok {
			return fmt.Errorf("property %q is not set", opts.Property)
		}

		level = sub
	}

	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Pointer {
		value = v.Elem()
		if v.IsNil() {
			return fmt.Errorf("property %q is not set", opts.Property)
		}
	}

	fmt.Fprintf(opts.IO.Out(), "%v\n", value)
	return nil
}
