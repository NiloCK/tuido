package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var ( // header styles
	activeTabBorder = lipgloss.Border{
		Top:         "─",
		Bottom:      " ",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "┘",
		BottomRight: "└",
	}

	tabBorder = lipgloss.Border{
		Top:         "─",
		Bottom:      "─",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "┴",
		BottomRight: "┴",
	}

	tabStyle       lipgloss.Style = lipgloss.NewStyle().Border(tabBorder).Padding(0, 2)
	activeTabStyle lipgloss.Style = tabStyle.Copy().Bold(true).Border(activeTabBorder, true)

	tabGapStyle lipgloss.Style = tabStyle.Copy().BorderTop(false).BorderLeft(false).BorderRight(false)
)

func (t tui) header() string {
	var todoTab, doneTab string

	if t.view == todo {
		todoTab = activeTabStyle.Render(string(todo))
	} else {
		todoTab = tabStyle.Render(string(todo))
	}

	if t.view == done {
		doneTab = activeTabStyle.Render(string(done))
	} else {
		doneTab = tabStyle.Render(string(done))
	}

	tabs := lipgloss.JoinHorizontal(lipgloss.Bottom, todoTab, doneTab, tabGapStyle.Render(t.filter.View()))

	gap := tabGapStyle.Render(strings.Repeat(" ", max(0, t.w-lipgloss.Width(tabs)-2)))

	return lipgloss.JoinHorizontal(lipgloss.Bottom, tabs, gap) + "\n\n"

}

func (t tui) View() string {
	selected := lipgloss.NewStyle().Bold(true)
	if len(t.renderSelection) == 0 { // init population
		t.populateRenderSelection()
	}

	ret := t.header() // todo: stringbuilder
	for i, item := range t.renderSelection {

		if i == int(t.selection) {
			ret += "> "
			ret += selected.Render(item.String())
		} else {
			ret += "  "
			ret += item.String()
		}
		ret += "\n"
	}
	return ret
}

func max(a, b int) int {
	if a >= b {
		return a
	}
	return b
}
