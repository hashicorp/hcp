package profile

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
	"github.com/mitchellh/mapstructure"
)

func NewCmdUnset(ctx *cmd.Context) *cmd.Command {
	opts := &UnsetOpts{
		Ctx:     ctx.ShutdownCtx,
		IO:      ctx.IO,
		Profile: ctx.Profile,
	}

	cmd := &cmd.Command{
		Name:      "unset",
		ShortHelp: "Unset a HCP CLI Property.",
		LongHelp: heredoc.New(ctx.IO).Mustf(`
		{{ Bold "hcp profile unset" }} unsets the specified property in your active profile.

		To view all currently set properties, run {{ Bold "hcp profile display" }}.
		`),
		Args: cmd.PositionalArguments{
			Autocomplete: opts.Profile,
			Args: []cmd.PositionalArgument{
				{
					Name: "COMPONENT/PROPERTY",
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
			l, err := profile.NewLoader()
			if err != nil {
				return err
			}
			opts.Profiles = l

			return unsetRun(opts)
		},
	}

	return cmd
}

type UnsetOpts struct {
	Ctx     context.Context
	IO      iostreams.IOStreams
	Profile *profile.Profile

	Property string
	Profiles *profile.Loader
}

func unsetRun(opts *UnsetOpts) error {
	// Validate we are not changing the name
	if opts.Property == "name" {
		return fmt.Errorf("to update a profile name use %s",
			opts.IO.ColorScheme().String("hcp profile profiles rename").Bold())
	}

	if err := IsValidProperty(opts.Property); err != nil {
		return err
	}

	// Decode the existing profile into a map
	var data map[string]any
	dec, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		WeaklyTypedInput:     true,
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
	didDelete := false
	for i, p := range parts {
		// This is the final property
		if i == len(parts)-1 {
			if _, ok := level[p]; !ok {
				break
			}

			delete(level, p)
			didDelete = true
			break
		}

		// Retrieve the component
		nested, ok := level[p]
		if !ok {
			break
		}

		// Check if the retrieved element is a nested object
		sub, ok := nested.(map[string]any)
		if !ok {
			break
		}

		level = sub
	}

	if didDelete {
		p, err := opts.Profiles.NewProfile(opts.Profile.Name)
		if err != nil {
			return err
		}

		dec2, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
			WeaklyTypedInput:     true,
			ErrorUnused:          true,
			Result:               p,
			TagName:              "hcl",
			IgnoreUntaggedFields: true,
		})
		if err != nil {
			return err
		}

		if err := dec2.Decode(data); err != nil {
			return convertDecodeError(err)
		}

		if err := p.Validate(); err != nil {
			return fmt.Errorf("invalid profile: %w", err)
		}

		if err := p.Write(); err != nil {
			return err
		}
	}

	cs := opts.IO.ColorScheme()
	fmt.Fprintf(opts.IO.Err(), "%s Property %q unset\n", cs.SuccessIcon(), opts.Property)
	return nil
}
