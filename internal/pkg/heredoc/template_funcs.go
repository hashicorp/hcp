package heredoc

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"golang.org/x/exp/maps"
)

const preserveNewLinesToken = "__preserveNewLines__"

// templateFuncs returns template helpers based on the IOStreams.
func (f *Formatter) templateFuncs() template.FuncMap {
	cs := f.io.ColorScheme()

	return template.FuncMap{
		"Color": func(values ...interface{}) (string, error) {
			s := cs.String(values[len(values)-1].(string))
			switch len(values) {
			case 2:
				c, err := getColor(cs, values[0].(string))
				if err != nil {
					return "", err
				}

				s = s.Color(c)
			case 3:
				foregroundColor, err := getColor(cs, values[0].(string))
				if err != nil {
					return "", fmt.Errorf("invalid foreground color: %w", err)
				}

				backgroundColor, err := getColor(cs, values[1].(string))
				if err != nil {
					return "", fmt.Errorf("invalid background color: %w", err)
				}

				s = s.
					Color(foregroundColor).
					Background(backgroundColor)
			}

			return s.String(), nil
		},
		"Bold":             styleFunc(cs, iostreams.String.Bold),
		"Faint":            styleFunc(cs, iostreams.String.Faint),
		"Italic":           styleFunc(cs, iostreams.String.Italic),
		"Underline":        styleFunc(cs, iostreams.String.Underline),
		"Blink":            styleFunc(cs, iostreams.String.Blink),
		"CrossOut":         styleFunc(cs, iostreams.String.CrossOut),
		"PreserveNewLines": func() string { return preserveNewLinesToken },
	}
}

func getColor(cs *iostreams.ColorScheme, c string) (iostreams.Color, error) {
	c = strings.ToLower(c)
	valid := map[string]func() iostreams.Color{
		"white":  cs.White,
		"black":  cs.Black,
		"red":    cs.Red,
		"green":  cs.Green,
		"orange": cs.Orange,
		"yellow": cs.Yellow,
		"gray":   cs.Gray,
	}

	if strings.HasPrefix(c, "#") {
		return cs.RGB(c), nil
	}

	color, ok := valid[c]
	if ok {
		return color(), nil
	}

	return cs.Black(), fmt.Errorf("unknown color. Must either be an RGB value (#<hex>) or one of %v", maps.Keys(valid))
}

func styleFunc(cs *iostreams.ColorScheme, f func(iostreams.String) iostreams.String) func(...interface{}) string {
	return func(values ...interface{}) string {
		s := cs.String(values[0].(string))
		return f(s).String()
	}
}
