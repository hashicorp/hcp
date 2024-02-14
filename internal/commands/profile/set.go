package profile

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/mitchellh/mapstructure"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/organization_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/client/project_service"
	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
)

func NewCmdSet(ctx *cmd.Context) *cmd.Command {
	opts := &SetOpts{
		Ctx:                 ctx.ShutdownCtx,
		IO:                  ctx.IO,
		Profile:             ctx.Profile,
		ProjectService:      project_service.New(ctx.HCP, nil),
		OrganizationService: organization_service.New(ctx.HCP, nil),
	}

	cmd := &cmd.Command{
		Name:      "set",
		ShortHelp: "Set a HCP CLI Property.",
		LongHelp: heredoc.New(ctx.IO).Mustf(`
		{{ Bold "hcp profile set" }} sets the specified property in your active profile.
		A property governs the behavior of a specific aspect of the HCP CLI. This could be
		setting the organization and project to target, or configuring the default output
		format across commands.

		To view all currently set properties, run {{ Bold "hcp profile display" }} or run
		{{ Bold "hcp profile read" }} to read the value of an individual property.

		To unset properties, use {{ Bold "hcp profile unset" }}.

		HCP CLI comes with a default profile but supports multiple. To create multiple
		configurations, use {{ Bold "hcp profile profiles create" }}, and {{ Bold "hcp profile profiles activate" }}
		to switch between them.
		`),
		Args: cmd.PositionalArguments{
			Autocomplete: opts.Profile,
			Args: []cmd.PositionalArgument{
				{
					Name: "COMPONENT/PROPERTY",
					Documentation: heredoc.New(ctx.IO).Must(`
					Property to be set. Note that COMPONENT/ is optional when referring to
					top-level profile fields, i.e., such as organization_id and project_id.
					Using component names is required for setting other properties like {{ Bold "core/output_format" }}.
					Consult the Available Properties section below for a comprehensive list of properties.
					`),
				},
				{
					Name:          "VALUE",
					Documentation: "Value to be set.",
				},
			},
		},
		AdditionalDocs: []cmd.DocSection{
			availablePropertiesDoc(ctx.IO),
		},
		RunF: func(c *cmd.Command, args []string) error {
			opts.Property = args[0]
			opts.Value = args[1]
			return setRun(opts)
		},
	}

	return cmd
}

type SetOpts struct {
	Ctx     context.Context
	IO      iostreams.IOStreams
	Profile *profile.Profile

	// Resource Manager client
	ProjectService      project_service.ClientService
	OrganizationService organization_service.ClientService

	// Arguments
	Property string
	Value    string
}

func setRun(opts *SetOpts) error {
	// Validate we are not changing the name
	if opts.Property == "name" {
		return fmt.Errorf("to update a profile name use %s",
			opts.IO.ColorScheme().String("hcp profile profiles rename").Bold())
	}

	// Validate we are setting a valid property
	if err := IsValidProperty(opts.Property); err != nil {
		return err
	}

	p := opts.Profile
	d, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		WeaklyTypedInput:     true,
		ErrorUnused:          true,
		Result:               p,
		TagName:              "hcl",
		IgnoreUntaggedFields: true,
	})
	if err != nil {
		return err
	}

	// Build the input
	input := map[string]any{}
	cur := input
	parts := strings.Split(opts.Property, "/")
	for i, p := range parts {
		if p == "" {
			return fmt.Errorf("property name following a \"/\" is required; empty property name is not allowed")
		}

		if i == len(parts)-1 {
			cur[p] = opts.Value
			continue
		}

		newLevel := map[string]any{}
		cur[p] = newLevel
		cur = newLevel
	}

	if err := d.Decode(input); err != nil {
		return convertDecodeError(err)
	}

	if err := p.Validate(); err != nil {
		return fmt.Errorf("invalid profile: %w", err)
	}

	// Check to see if the property being set is valid
	write := true
	if opts.Property == "project_id" {
		write, err = opts.validateProject()
	} else if opts.Property == "organization_id" {
		write, err = opts.validateOrg()
	}
	if err != nil {
		return err
	} else if !write {
		return nil
	}

	if err := p.Write(); err != nil {
		return err
	}

	fmt.Fprintf(opts.IO.Err(), "%s Property %q updated\n",
		opts.IO.ColorScheme().SuccessIcon(), opts.Property)

	return nil
}

func (o *SetOpts) validateProject() (bool, error) {
	if o.Property != "project_id" || !o.IO.CanPrompt() {
		return true, nil
	}

	// See if we can get the project.
	hasAccess := true
	params := &project_service.ProjectServiceGetParams{
		ID:      o.Value,
		Context: o.Ctx,
	}
	_, err := o.ProjectService.ProjectServiceGet(params, nil)
	if err != nil {
		var listErr *project_service.ProjectServiceGetDefault
		if errors.As(err, &listErr) {
			if listErr.IsCode(http.StatusForbidden) || listErr.IsCode(http.StatusNotFound) {
				hasAccess = false
			} else if listErr.IsCode(http.StatusBadRequest) {
				return false, fmt.Errorf("invalid project_id: %v", listErr.Payload.Message)
			} else {
				return false, fmt.Errorf("failed to check if principal has access to project %q: %w", o.Value, listErr)
			}
		} else {
			return false, fmt.Errorf("failed to check if principal has access to project %q: %w", o.Value, err)
		}
	}

	if !hasAccess {
		fmt.Fprintf(o.IO.Err(), "%s You do not appear to have access to project %q or it does not exist.\n",
			o.IO.ColorScheme().WarningLabel(), o.Value)

		prompt := fmt.Sprintf("\nAre you sure you wish to set the %q property", "project_id")
		res, err := o.IO.PromptConfirm(prompt)
		if err != nil {
			return false, fmt.Errorf("failed prompting for confirmation: %w", err)
		}

		return res, nil
	}

	return true, nil
}

func (o *SetOpts) validateOrg() (bool, error) {
	if o.Property != "organization_id" || !o.IO.CanPrompt() {
		return true, nil
	}

	// See if we can get the project.
	hasAccess := false
	params := &organization_service.OrganizationServiceListParams{
		Context: o.Ctx,
	}
	resp, err := o.OrganizationService.OrganizationServiceList(params, nil)
	if err != nil {
		return false, fmt.Errorf("failed to list organizations for principal: %w", err)
	}

	for _, org := range resp.Payload.Organizations {
		if org.ID == o.Value {
			hasAccess = true
			break
		}
	}

	if !hasAccess {
		fmt.Fprintf(o.IO.Err(), "%s You do not appear to be a member of organization %q.\n",
			o.IO.ColorScheme().WarningLabel(), o.Value)

		prompt := fmt.Sprintf("\nAre you sure you wish to set the %q property", "organization_id")
		res, err := o.IO.PromptConfirm(prompt)
		if err != nil {
			return false, fmt.Errorf("failed prompting for confirmation: %w", err)
		}

		return res, nil
	}

	return true, nil
}

// convertDecodeError converts the mapstructure decode error into a more
// contextual error.
func convertDecodeError(err error) error {
	mapErr := &mapstructure.Error{}
	if !errors.As(err, &mapErr) {
		return err
	}

	// We only expect a single error to ever occur
	if len(mapErr.Errors) > 1 {
		return err
	}

	// Parse an invalid key at the top-level
	errStr := mapErr.Errors[0]
	if strings.HasPrefix(errStr, "'' has invalid keys:") {
		parts := strings.Split(errStr, ": ")
		return fmt.Errorf("no top-level property with name %q", parts[1])
	}

	// Try to parse invalid keys within a component. This could occur if a user
	// runs "set core/bad-key value"
	var component, property string
	_, scanErr := fmt.Sscanf(strings.ReplaceAll(errStr, "'", ""), "%s has invalid keys: %s", &component, &property)
	if scanErr == nil {
		return fmt.Errorf("invalid property %q for component %q", property, component)
	}

	return errors.New(mapErr.Errors[0])
}
