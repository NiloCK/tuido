package utils

// shamelessly stolen from
// https://github.com/noahgorstein/jqp/blob/92ce6eed480c70a5a8a1d67aab5e3e87052cb61d/tui/utils/utils.go

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// returns a string used for chroma syntax highlighting
func GetTerminalColorSupport() string {
	switch lipgloss.ColorProfile() {
	case termenv.Ascii:
		return "terminal"
	case termenv.ANSI:
		return "terminal16"
	case termenv.ANSI256:
		return "terminal256"
	case termenv.TrueColor:
		return "terminal16m"
	default:
		return "terminal"
	}
}
