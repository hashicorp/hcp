package profiles

import (
	"fmt"
	"slices"
	"strings"

	"github.com/hashicorp/hcp/internal/pkg/cmd"
	"github.com/hashicorp/hcp/internal/pkg/format"
	"github.com/hashicorp/hcp/internal/pkg/heredoc"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/hashicorp/hcp/internal/pkg/profile"
)

func NewCmdList(ctx *cmd.Context) *cmd.Command {
	opts := &ListOpts{
		IO:     ctx.IO,
		Output: ctx.Output,
	}
	cmd := &cmd.Command{
		Name:      "list",
		ShortHelp: "List existing HCP profiles.",
		LongHelp: heredoc.New(ctx.IO).Must(`
		The {{ template "mdCodeOrBold" "hcp profile profiles list" }} command lists existing HCP profiles.
		`),
		Examples: []cmd.Example{
			{
				Preamble: "To list existing profiles, run:",
				Command:  "$ hcp profile profiles list",
			},
		},
		NoAuthRequired: true,
		RunF: func(c *cmd.Command, args []string) error {
			l, err := profile.NewLoader()
			if err != nil {
				return err
			}
			opts.Profiles = l
			return listRun(opts)
		},
	}

	return cmd
}

type ListOpts struct {
	IO       iostreams.IOStreams
	Output   *format.Outputter
	Profiles *profile.Loader
}

func listRun(opts *ListOpts) error {
	profileNames, err := opts.Profiles.ListProfiles()
	if err != nil {
		return fmt.Errorf("failed to list profiles: %w", err)
	}

	profiles := make([]*profile.Profile, len(profileNames))
	for i, n := range profileNames {
		p, err := opts.Profiles.LoadProfile(n)
		if err != nil {
			return fmt.Errorf("failed to load profile %q: %w", n, err)
		}

		profiles[i] = p
	}

	// Sort the profiles based on name
	slices.SortFunc(profiles, func(p1, p2 *profile.Profile) int {
		return strings.Compare(p1.Name, p2.Name)
	})

	// Get the active profile
	active, err := opts.Profiles.GetActiveProfile()
	if err != nil {
		return fmt.Errorf("failed to get active profile: %w", err)
	}

	d := &profileDisplayer{
		profiles:      profiles,
		activeProfile: active.Name,
	}

	return opts.Output.Display(d)
}

type profileDisplayer struct {
	profiles      []*profile.Profile
	activeProfile string
}

func (p *profileDisplayer) DefaultFormat() format.Format { return format.Table }
func (p *profileDisplayer) Payload() any                 { return p.profiles }

func (p *profileDisplayer) FieldTemplates() []format.Field {
	return []format.Field{
		{
			Name:        "Name",
			ValueFormat: "{{ .Name }}",
		},
		{
			Name:        "Active",
			ValueFormat: fmt.Sprintf("{{ eq ( .Name ) %q }}", p.activeProfile),
		},
		{
			Name:        "Organization ID",
			ValueFormat: "{{ .OrganizationID }}",
		},
		{
			Name:        "Project ID",
			ValueFormat: "{{ .ProjectID }}",
		},
	}
}
