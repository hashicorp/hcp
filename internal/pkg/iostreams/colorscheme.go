package iostreams

import (
	"github.com/muesli/termenv"
)

const (
	// Each constant defines the color code for the given color. A # prefix
	// indicates RGB colors, and a number indicates ANSI colors.
	black  = "0"
	white  = "7"
	red    = "#E52228"
	green  = "#008A22"
	orange = "#BB5A00"
	yellow = "#FFD814"
	gray   = "#C2C5CB"
)

// ColorScheme is used to style and color text according to the capabilities of
// the current terminal. It will automatically degrade the requested styling to
// what the system is capable for outputting.
type ColorScheme struct {
	profile termenv.Profile
}

// String is a stylized string.
type String struct {
	style termenv.Style
}

// String wraps the given string. The wrapped String can then have style or
// color applied to it.
func (cs *ColorScheme) String(s string) String {
	return String{
		style: cs.profile.String(s),
	}
}

// String implements the fmt.Stringer interface. The wrapped string will be
// printed with the appropriate control sequences applied to have the string
// stylized as requested.
func (s String) String() string {
	return s.style.String()
}

// SuccessIcon returns a success icon.
func (cs *ColorScheme) SuccessIcon() String {
	return cs.String("✓").Color(cs.Green())
}

// FailureIcon returns a failure icon.
func (cs *ColorScheme) FailureIcon() String {
	return cs.String("X").Color(cs.Red())
}

// WarningLabel returns a colored warning label.
func (cs *ColorScheme) WarningLabel() String {
	return cs.String("WARNING:").Color(cs.Orange())
}

// ErrorLabel returns a colored error label.
func (cs *ColorScheme) ErrorLabel() String {
	return cs.String("ERROR:").Color(cs.Red())
}

// Color applies the given color to the texts foreground.
func (s String) Color(c Color) String {
	return String{
		style: s.style.Foreground(c.color),
	}
}

// Background applies the given color to the texts background.
func (s String) Background(c Color) String {
	return String{
		style: s.style.Background(c.color),
	}
}

// Bold makes the string bold.
func (s String) Bold() String {
	return String{
		style: s.style.Bold(),
	}
}

// Faint makes the text faint.
func (s String) Faint() String {
	return String{
		style: s.style.Faint(),
	}
}

// Italic makes the text italic.
func (s String) Italic() String {
	return String{
		style: s.style.Italic(),
	}
}

// Underline makes the text underlined.
func (s String) Underline() String {
	return String{
		style: s.style.Underline(),
	}
}

// CrossOut makes the text have a cross through it middle height wise.
func (s String) CrossOut() String {
	return String{
		style: s.style.CrossOut(),
	}
}

// Blink makes the text blink.
func (s String) Blink() String {
	return String{
		style: s.style.Blink(),
	}
}

// Color is represents a color.
type Color struct {
	color termenv.Color
}

// White is the color white.
func (cs *ColorScheme) White() Color {
	return Color{
		color: cs.profile.Color(white),
	}
}

// Black is the color black.
func (cs *ColorScheme) Black() Color {
	return Color{
		color: cs.profile.Color(black),
	}
}

// Red is the color red.
func (cs *ColorScheme) Red() Color {
	return Color{
		color: cs.profile.Color(red),
	}
}

// Green is the color green.
func (cs *ColorScheme) Green() Color {
	return Color{
		color: cs.profile.Color(green),
	}
}

// Orange is the color orange.
func (cs *ColorScheme) Orange() Color {
	return Color{
		color: cs.profile.Color(orange),
	}
}

// Yellow is the color yellow.
func (cs *ColorScheme) Yellow() Color {
	return Color{
		color: cs.profile.Color(yellow),
	}
}

// Gray is the color gray.
func (cs *ColorScheme) Gray() Color {
	return Color{
		color: cs.profile.Color(gray),
	}
}

// RGB allows setting an RGB color in the format "#<hex>". If the terminal does
// not support TrueColor, the nearest approximate and supported color will be used.
func (cs *ColorScheme) RGB(hex string) Color {
	return Color{
		color: cs.profile.Color(hex),
	}
}
