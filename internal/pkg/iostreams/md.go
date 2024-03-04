package iostreams

import (
	"fmt"

	"github.com/muesli/termenv"
)

// IsMarkdownOutput is an interface that when met, indicates that the IOStreams
// instance is configured to output markdown.
type IsMarkdownOutput interface {
	MarkdownOutput() bool
}

type mdStream struct {
	*Testing
}

// MD returns a new IOStreams instance that is configured to output markdown.
func MD() IOStreams {
	return &mdStream{
		Testing: Test(),
	}
}

func (m *mdStream) MarkdownOutput() bool { return true }

func (m *mdStream) ColorScheme() *ColorScheme {
	return &ColorScheme{
		profile: termenv.Ascii,
		md:      true,
	}
}

// markdownString returns a string that is formatted for markdown.
func (s String) markdownString() string {
	mdText := s.rawString

	for _, e := range s.emphases {
		switch e {
		case EmphasisBold:
			mdText = "**" + mdText + "**"
		case EmphasisItalic:
			mdText = "_" + mdText + "_"
		case EmphasisUnderline:
			mdText = "<u>" + mdText + "</u>"
		case EmphasisCrossOut:
			mdText = "~~" + mdText + "~~"
		case EmphasisCode:
			mdText = "`" + mdText + "`"
		case EmphasisCodeBlock:
			mdText = fmt.Sprintf("```%s\n%s\n```", s.codeBlockExtension, mdText)
		}
	}

	return mdText
}
